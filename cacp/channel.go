package cacp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	phoenixHeartbeatInterval = 30 * time.Second
	phoenixVSN               = "2.0.0"
)

type MessageHandler func(payload map[string]interface{})

type PhoenixMessage struct {
	Ref     string                 `json:"ref"`
	Event   string                 `json:"event"`
	Topic   string                 `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
}

func (m *PhoenixMessage) toTuple() []interface{} {
	return []interface{}{m.Ref, m.Event, m.Topic, m.Payload}
}

func (m *PhoenixMessage) toJSON() ([]byte, error) {
	return json.Marshal(m.toTuple())
}

func messageFromTuple(tuple []interface{}) (*PhoenixMessage, error) {
	if len(tuple) < 4 {
		return nil, fmt.Errorf("invalid Phoenix message tuple: %v", tuple)
	}

	payload, ok := tuple[3].(map[string]interface{})
	if !ok {
		payload = make(map[string]interface{})
	}

	return &PhoenixMessage{
		Ref:     fmt.Sprintf("%v", tuple[0]),
		Event:   fmt.Sprintf("%v", tuple[1]),
		Topic:   fmt.Sprintf("%v", tuple[2]),
		Payload: payload,
	}, nil
}

func messageFromJSON(data []byte) (*PhoenixMessage, error) {
	var tuple []interface{}
	if err := json.Unmarshal(data, &tuple); err != nil {
		return nil, err
	}
	return messageFromTuple(tuple)
}

type PhoenixChannel struct {
	topic  string
	params map[string]interface{}

	joined     bool
	joinDone   chan struct{}
	mu         sync.RWMutex
	handlers   map[string][]MessageHandler
	client     *PhoenixChannelClient
}

func (ch *PhoenixChannel) isJoined() bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.joined
}

func (ch *PhoenixChannel) waitUntilJoined(ctx context.Context) error {
	select {
	case <-ch.joinDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (ch *PhoenixChannel) markJoined() {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.joined = true
	close(ch.joinDone)
}

func (ch *PhoenixChannel) resetJoin() {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.joined = false
	ch.joinDone = make(chan struct{})
}

func (ch *PhoenixChannel) emit(event string, payload map[string]interface{}) {
	ch.mu.RLock()
	handlers := ch.handlers[event]
	ch.mu.RUnlock()

	for _, handler := range handlers {
		go func(h MessageHandler) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("panic in event handler: %v\n", r)
				}
			}()
			h(payload)
		}(handler)
	}
}

type PhoenixChannelClient struct {
	client *Client

	conn           *websocket.Conn
	connected      bool
	mu             sync.RWMutex
	refCounter     int

	channels       map[string]*PhoenixChannel
	pendingReqs    map[string]chan map[string]interface{}

	globalHandlers []func(*PhoenixMessage)
	messageChan    chan map[string]interface{}

	heartbeatTicker *time.Ticker
	heartbeatStop   chan struct{}

	done chan struct{}
}

func newPhoenixChannelClient(client *Client) *PhoenixChannelClient {
	return &PhoenixChannelClient{
		client:        client,
		channels:      make(map[string]*PhoenixChannel),
		pendingReqs:   make(map[string]chan map[string]interface{}),
		messageChan:   make(chan map[string]interface{}, 100),
		heartbeatStop: make(chan struct{}),
		done:          make(chan struct{}),
	}
}

func (p *PhoenixChannelClient) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

func (p *PhoenixChannelClient) Connect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connected {
		return nil
	}

	wsURL := fmt.Sprintf("%s/websocket?vsn=%s", p.client.WebSocketURL(), phoenixVSN)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return &ConnectionError{Message: err.Error()}
	}

	p.conn = conn
	p.connected = true

	go p.heartbeatLoop()
	go p.receiveLoop()

	return nil
}

func (p *PhoenixChannelClient) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil
	}

	close(p.heartbeatStop)

	if p.heartbeatTicker != nil {
		p.heartbeatTicker.Stop()
	}

	for _, ch := range p.channels {
		p.leaveChannel(ch.topic)
	}

	if p.conn != nil {
		p.conn.Close()
	}

	close(p.done)
	p.connected = false

	return nil
}

func (p *PhoenixChannelClient) Channel(topic string, params map[string]interface{}) *PhoenixChannel {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ch, ok := p.channels[topic]; ok {
		return ch
	}

	ch := &PhoenixChannel{
		topic:     topic,
		params:    params,
		joinDone:  make(chan struct{}),
		handlers:  make(map[string][]MessageHandler),
		client:    p,
	}
	p.channels[topic] = ch
	return ch
}

func (p *PhoenixChannelClient) JoinChannel(ctx context.Context, topic string, params map[string]interface{}) error {
 ch := p.Channel(topic, params)

	if ch.isJoined() {
		return nil
	}

	p.mu.Lock()
	ref := fmt.Sprintf("%d", p.nextRef())
	p.mu.Unlock()

	msg := &PhoenixMessage{
		Ref:     ref,
		Event:   "phx_join",
		Topic:   topic,
		Payload: params,
	}

	replyChan := make(chan map[string]interface{}, 1)
	p.mu.Lock()
	p.pendingReqs[ref] = replyChan
	p.mu.Unlock()

	if err := p.sendMessage(msg); err != nil {
		p.mu.Lock()
		delete(p.pendingReqs, ref)
		p.mu.Unlock()
		return err
	}

	select {
	case reply := <-replyChan:
		if status, ok := reply["status"].(string); ok && status != "ok" {
			return &WebSocketError{Message: fmt.Sprintf("failed to join channel %s: %v", topic, reply)}
		}
		ch.markJoined()
		fmt.Printf("Joined channel: %s\n", topic)
		return nil
	case <-time.After(5 * time.Second):
		p.mu.Lock()
		delete(p.pendingReqs, ref)
		p.mu.Unlock()
		return &WebSocketError{Message: fmt.Sprintf("timeout joining channel %s", topic)}
	case <-ctx.Done():
		p.mu.Lock()
		delete(p.pendingReqs, ref)
		p.mu.Unlock()
		return ctx.Err()
	}
}

func (p *PhoenixChannelClient) LeaveChannel(topic string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch, ok := p.channels[topic]
	if !ok || !ch.isJoined() {
		return nil
	}

	ref := fmt.Sprintf("%d", p.nextRef())
	msg := &PhoenixMessage{
		Ref:     ref,
		Event:   "phx_leave",
		Topic:   topic,
		Payload: map[string]interface{}{},
	}

	if err := p.sendMessage(msg); err != nil {
		fmt.Printf("Error sending leave for channel %s: %v\n", topic, err)
	}

	ch.resetJoin()
	fmt.Printf("Left channel: %s\n", topic)
	return nil
}

func (p *PhoenixChannelClient) Push(topic, event string, payload map[string]interface{}) error {
	p.mu.Lock()
	ref := fmt.Sprintf("%d", p.nextRef())
	p.mu.Unlock()

	msg := &PhoenixMessage{
		Ref:     ref,
		Event:   event,
		Topic:   topic,
		Payload: payload,
	}

	return p.sendMessage(msg)
}

func (p *PhoenixChannelClient) Request(ctx context.Context, topic, event string, payload map[string]interface{}) (map[string]interface{}, error) {
	p.mu.Lock()
	ref := fmt.Sprintf("%d", p.nextRef())
	p.mu.Unlock()

	msg := &PhoenixMessage{
		Ref:     ref,
		Event:   event,
		Topic:   topic,
		Payload: payload,
	}

	replyChan := make(chan map[string]interface{}, 1)
	p.mu.Lock()
	p.pendingReqs[ref] = replyChan
	p.mu.Unlock()

	if err := p.sendMessage(msg); err != nil {
		p.mu.Lock()
		delete(p.pendingReqs, ref)
		p.mu.Unlock()
		return nil, err
	}

	select {
	case reply := <-replyChan:
		if status, ok := reply["status"].(string); ok && status != "ok" {
			return nil, &WebSocketError{Message: fmt.Sprintf("phx_reply error: %v", reply)}
		}
		return reply, nil
	case <-time.After(5 * time.Second):
		p.mu.Lock()
		delete(p.pendingReqs, ref)
		p.mu.Unlock()
		return nil, &WebSocketError{Message: fmt.Sprintf("timeout waiting for reply to %s on %s", event, topic)}
	case <-ctx.Done():
		p.mu.Lock()
		delete(p.pendingReqs, ref)
		p.mu.Unlock()
		return nil, ctx.Err()
	}
}

func (p *PhoenixChannelClient) Subscribe(agentID string) {
	topic := fmt.Sprintf("agent:%s", agentID)
	p.Channel(topic, nil)
}

func (p *PhoenixChannelClient) Unsubscribe(agentID string) error {
	topic := fmt.Sprintf("agent:%s", agentID)
	return p.LeaveChannel(topic)
}

func (p *PhoenixChannelClient) Send(ctx context.Context, toAgent string, content map[string]interface{}, messageType string, fromAgent string, metadata map[string]interface{}) error {
	topic := fmt.Sprintf("agent:%s", toAgent)

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"content":      content,
			"message_type": messageType,
		},
	}

	if fromAgent != "" {
		if msg, ok := payload["message"].(map[string]interface{}); ok {
			msg["sender"] = map[string]interface{}{"agent_id": fromAgent}
		}
	}

	if metadata != nil {
		if msg, ok := payload["message"].(map[string]interface{}); ok {
			msg["metadata"] = metadata
		}
	}

	return p.Push(topic, "send", payload)
}

func (p *PhoenixChannelClient) SendRPC(ctx context.Context, toAgent, method string, params map[string]interface{}, requestID, fromAgent string) error {
	topic := fmt.Sprintf("agent:%s", toAgent)

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"method": method,
			"params": params,
		},
	}

	if requestID != "" {
		if msg, ok := payload["message"].(map[string]interface{}); ok {
			msg["correlation_id"] = requestID
		}
	}

	if fromAgent != "" {
		if msg, ok := payload["message"].(map[string]interface{}); ok {
			msg["sender"] = map[string]interface{}{"agent_id": fromAgent}
		}
	}

	return p.Push(topic, "rpc_request", payload)
}

func (p *PhoenixChannelClient) Messages() <-chan map[string]interface{} {
	return p.messageChan
}

func (p *PhoenixChannelClient) OnGlobalMessage(handler func(*PhoenixMessage)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.globalHandlers = append(p.globalHandlers, handler)
}

func (p *PhoenixChannelClient) sendMessage(msg *PhoenixMessage) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected || p.conn == nil {
		return &WebSocketError{Message: "not connected"}
	}

	data, err := msg.toJSON()
	if err != nil {
		return err
	}

	return p.conn.WriteMessage(websocket.TextMessage, data)
}

func (p *PhoenixChannelClient) nextRef() int {
	p.refCounter++
	return p.refCounter
}

func (p *PhoenixChannelClient) heartbeatLoop() {
	p.heartbeatTicker = time.NewTicker(phoenixHeartbeatInterval)
	defer p.heartbeatTicker.Stop()

	for {
		select {
		case <-p.heartbeatTicker.C:
			p.mu.RLock()
			if !p.connected {
				p.mu.RUnlock()
				return
			}

			msg := &PhoenixMessage{
				Ref:     fmt.Sprintf("%d", p.nextRef()),
				Event:   "heartbeat",
				Topic:   "phoenix",
				Payload: map[string]interface{}{},
			}

			if err := p.sendMessage(msg); err != nil {
				fmt.Printf("Heartbeat error: %v\n", err)
			}

			p.mu.RUnlock()
		case <-p.heartbeatStop:
			return
		case <-p.done:
			return
		}
	}
}

func (p *PhoenixChannelClient) receiveLoop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic in receiveLoop: %v\n", r)
		}
	}()

	for {
		select {
		case <-p.done:
			return
		default:
			p.mu.RLock()
			conn := p.conn
			p.mu.RUnlock()

			if conn == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			_, data, err := conn.ReadMessage()
			if err != nil {
				p.mu.Lock()
				p.connected = false
				p.mu.Unlock()
				return
			}

			msg, err := messageFromJSON(data)
			if err != nil {
				fmt.Printf("Error parsing message: %v\n", err)
				continue
			}

			p.handleMessage(msg)
		}
	}
}

func (p *PhoenixChannelClient) handleMessage(msg *PhoenixMessage) {
	fmt.Printf("Received: %s on %s\n", msg.Event, msg.Topic)

	if msg.Event == "phx_reply" {
		p.mu.Lock()
		if ch, ok := p.pendingReqs[msg.Ref]; ok {
			select {
			case ch <- msg.Payload:
			default:
			}
			delete(p.pendingReqs, msg.Ref)
		}
		p.mu.Unlock()
	} else if msg.Event == "phx_error" {
		p.mu.Lock()
		if ch, ok := p.pendingReqs[msg.Ref]; ok {
			close(ch)
			delete(p.pendingReqs, msg.Ref)
		}
		p.mu.Unlock()
	} else if msg.Event == "phx_close" {
		p.mu.RLock()
		ch, ok := p.channels[msg.Topic]
		p.mu.RUnlock()

		if ok {
			ch.resetJoin()
		}
	} else {
		p.mu.RLock()
		ch, ok := p.channels[msg.Topic]
		p.mu.RUnlock()

		if ok {
			ch.emit(msg.Event, msg.Payload)
		}

		if msg.Event == "message" || msg.Event == "rpc_response" || msg.Event == "rpc_error" {
			if message, ok := msg.Payload["message"].(map[string]interface{}); ok {
				select {
				case p.messageChan <- message:
				default:
				}
			}
		}

		p.mu.RLock()
		handlers := make([]func(*PhoenixMessage), len(p.globalHandlers))
		copy(handlers, p.globalHandlers)
		p.mu.RUnlock()

		for _, handler := range handlers {
			go func(h func(*PhoenixMessage)) {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("panic in global handler: %v\n", r)
					}
				}()
				h(msg)
			}(handler)
		}
	}
}