package cacp

import (
	"fmt"
	"time"
)

// Broker error codes (5001-7008 range)
const (
	// Authentication errors (5001-5010)
	ErrorCodeInvalidCredentials    = 5001
	ErrorCodeTokenExpired         = 5002
	ErrorCodeForbidden             = 5003
	ErrorCodeUnauthorized          = 5004

	// Task errors (6001-6010)
	ErrorCodeTaskNotFound         = 6001
	ErrorCodeTaskValidationError  = 6002
	ErrorCodeTaskStateError       = 6004
	ErrorCodeTaskRetryError       = 6005

	// Group errors (7001-7010)
	ErrorCodeGroupNotFound        = 7001
	ErrorCodeGroupValidationError = 7002
	ErrorCodeMemberNotFound       = 7003
	ErrorCodeMemberError          = 7004

	// Validation errors (1001-1999)
	ErrorCodeValidationError      = 1001
	ErrorCodeMissingRequiredField = 1002

	// Agent errors (2001-2999)
	ErrorCodeAgentNotFound        = 2001
	ErrorCodeAgentValidationError = 2002

	// Message errors (3001-3999)
	ErrorCodeMessageNotFound      = 3001
	ErrorCodeMessageDeliveryError = 3002
)

// ErrorRequestID represents the request ID associated with an error.
type ErrorRequestID string

// Error types for the SDK.

// APIError represents a general API error.
type APIError struct {
	Message    string
	Code       string
	StatusCode int
	RequestID  ErrorRequestID
}

func (e *APIError) Error() string {
	if e.Code != "" {
		if e.RequestID != "" {
			return "[" + e.Code + "] " + e.Message + " (request_id: " + string(e.RequestID) + ")"
		}
		return "[" + e.Code + "] " + e.Message
	}
	if e.RequestID != "" {
		return e.Message + " (request_id: " + string(e.RequestID) + ")"
	}
	return e.Message
}

// AuthenticationError represents an authentication failure.
type AuthenticationError struct {
	Message   string
	ErrorCode int
	RequestID ErrorRequestID
}

func (e *AuthenticationError) Error() string {
	if e.ErrorCode > 0 {
		if e.RequestID != "" {
			return fmt.Sprintf("[%d] %s (request_id: %s)", e.ErrorCode, e.Message, e.RequestID)
		}
		return fmt.Sprintf("[%d] %s", e.ErrorCode, e.Message)
	}
	if e.RequestID != "" {
		return e.Message + " (request_id: " + string(e.RequestID) + ")"
	}
	return e.Message
}

// ValidationError represents a validation failure.
type ValidationError struct {
	Message   string
	Field     string
	ErrorCode int
	RequestID ErrorRequestID
}

func (e *ValidationError) Error() string {
	msg := ""
	if e.Field != "" {
		msg = "validation error on field '" + e.Field + "': " + e.Message
	} else {
		msg = "validation error: " + e.Message
	}
	if e.ErrorCode > 0 {
		if e.RequestID != "" {
			return fmt.Sprintf("[%d] %s (request_id: %s)", e.ErrorCode, msg, e.RequestID)
		}
		return fmt.Sprintf("[%d] %s", e.ErrorCode, msg)
	}
	if e.RequestID != "" {
		return msg + " (request_id: " + string(e.RequestID) + ")"
	}
	return msg
}

// NotFoundError represents a resource not found error.
type NotFoundError struct {
	Message   string
	Code      string
	ErrorCode int
	RequestID ErrorRequestID
}

func (e *NotFoundError) Error() string {
	if e.ErrorCode > 0 {
		if e.RequestID != "" {
			return fmt.Sprintf("[%d] %s (request_id: %s)", e.ErrorCode, e.Message, e.RequestID)
		}
		return fmt.Sprintf("[%d] %s", e.ErrorCode, e.Message)
	}
	if e.RequestID != "" {
		return e.Message + " (request_id: " + string(e.RequestID) + ")"
	}
	return e.Message
}

// AgentNotFoundError represents an agent not found error.
type AgentNotFoundError struct {
	AgentID   string
	RequestID ErrorRequestID
}

func (e *AgentNotFoundError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("agent not found: %s (request_id: %s)", e.AgentID, e.RequestID)
	}
	return "agent not found: " + e.AgentID
}

// ConnectionError represents a connection failure.
type ConnectionError struct {
	Message   string
	RequestID ErrorRequestID
}

func (e *ConnectionError) Error() string {
	if e.RequestID != "" {
		return "connection error: " + e.Message + " (request_id: " + string(e.RequestID) + ")"
	}
	return "connection error: " + e.Message
}

// ForbiddenError represents a forbidden access error.
type ForbiddenError struct {
	Message   string
	ErrorCode int
	RequestID ErrorRequestID
}

func (e *ForbiddenError) Error() string {
	if e.ErrorCode > 0 {
		if e.RequestID != "" {
			return fmt.Sprintf("[%d] %s (request_id: %s)", e.ErrorCode, e.Message, e.RequestID)
		}
		return fmt.Sprintf("[%d] %s", e.ErrorCode, e.Message)
	}
	if e.RequestID != "" {
		return e.Message + " (request_id: " + string(e.RequestID) + ")"
	}
	return e.Message
}

func (e *ConnectionError) Error() string {
	return "connection error: " + e.Message
}

// RateLimitError represents a rate limit exceeded error.
type RateLimitError struct {
	Message    string
	RetryAfter time.Duration
	RequestID  ErrorRequestID
}

func (e *RateLimitError) Error() string {
	msg := ""
	if e.RetryAfter > 0 {
		msg = fmt.Sprintf("%s. Retry after %v", e.Message, e.RetryAfter)
	} else {
		msg = e.Message
	}
	if e.RequestID != "" {
		return msg + " (request_id: " + string(e.RequestID) + ")"
	}
	return msg
}

// MessageError represents a message-related error.
type MessageError struct {
	Message   string
	MessageID string
	RequestID ErrorRequestID
}

func (e *MessageError) Error() string {
	if e.MessageID != "" {
		if e.RequestID != "" {
			return fmt.Sprintf("message error (%s): %s (request_id: %s)", e.MessageID, e.Message, e.RequestID)
		}
		return fmt.Sprintf("message error (%s): %s", e.MessageID, e.Message)
	}
	if e.RequestID != "" {
		return "message error: " + e.Message + " (request_id: " + string(e.RequestID) + ")"
	}
	return "message error: " + e.Message
}

// MessageError represents a message-related error.
type MessageError struct {
	Message   string
	MessageID string
}

func (e *MessageError) Error() string {
	if e.MessageID != "" {
		return fmt.Sprintf("message error (%s): %s", e.MessageID, e.Message)
	}
	return "message error: " + e.Message
}

// RPCError represents an RPC call error.
type RPCError struct {
	Message   string
	Method    string
	Code      int
	RequestID ErrorRequestID
}

func (e *RPCError) Error() string {
	msg := ""
	if e.Method != "" {
		msg = fmt.Sprintf("RPC error calling '%s': %s", e.Method, e.Message)
	} else {
		msg = "RPC error: " + e.Message
	}
	if e.RequestID != "" {
		return msg + " (request_id: " + string(e.RequestID) + ")"
	}
	return msg
}

// WebSocketError represents a WebSocket error.
type WebSocketError struct {
	Message   string
	RequestID ErrorRequestID
}

func (e *WebSocketError) Error() string {
	if e.RequestID != "" {
		return "websocket error: " + e.Message + " (request_id: " + string(e.RequestID) + ")"
	}
	return "websocket error: " + e.Message
}

// TimeoutError represents a timeout error.
type TimeoutError struct {
	Operation string
	Timeout   time.Duration
	RequestID ErrorRequestID
}

func (e *TimeoutError) Error() string {
	msg := ""
	if e.Timeout > 0 {
		msg = fmt.Sprintf("operation '%s' timed out after %v", e.Operation, e.Timeout)
	} else {
		msg = fmt.Sprintf("operation '%s' timed out", e.Operation)
	}
	if e.RequestID != "" {
		return msg + " (request_id: " + string(e.RequestID) + ")"
	}
	return msg
}

// TaskNotFoundError represents a task not found error.
type TaskNotFoundError struct {
	TaskID    string
	RequestID ErrorRequestID
}

func (e *TaskNotFoundError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("task not found: %s (request_id: %s)", e.TaskID, e.RequestID)
	}
	return "task not found: " + e.TaskID
}

// TaskStateError represents an invalid task state error.
type TaskStateError struct {
	Message      string
	ErrorCode    int
	TaskID       string
	CurrentStatus TaskStatus
	RequestID    ErrorRequestID
}

func (e *TaskStateError) Error() string {
	msg := ""
	if e.TaskID != "" && e.CurrentStatus != "" {
		msg = fmt.Sprintf("task state error (%s, status: %s): %s", e.TaskID, e.CurrentStatus, e.Message)
	} else {
		msg = "task state error: " + e.Message
	}
	if e.ErrorCode > 0 {
		if e.RequestID != "" {
			return fmt.Sprintf("[%d] %s (request_id: %s)", e.ErrorCode, msg, e.RequestID)
		}
		return fmt.Sprintf("[%d] %s", e.ErrorCode, msg)
	}
	if e.RequestID != "" {
		return msg + " (request_id: " + string(e.RequestID) + ")"
	}
	return msg
}

// TaskError represents a task-related error.
type TaskError struct {
	Message   string
	ErrorCode int
	TaskID    string
	RequestID ErrorRequestID
}

func (e *TaskError) Error() string {
	msg := ""
	if e.TaskID != "" {
		msg = fmt.Sprintf("task error (%s): %s", e.TaskID, e.Message)
	} else {
		msg = "task error: " + e.Message
	}
	if e.ErrorCode > 0 {
		if e.RequestID != "" {
			return fmt.Sprintf("[%d] %s (request_id: %s)", e.ErrorCode, msg, e.RequestID)
		}
		return fmt.Sprintf("[%d] %s", e.ErrorCode, msg)
	}
	if e.RequestID != "" {
		return msg + " (request_id: " + string(e.RequestID) + ")"
	}
	return msg
}

// GroupNotFoundError represents a group not found error.
type GroupNotFoundError struct {
	GroupID   string
	RequestID ErrorRequestID
}

func (e *GroupNotFoundError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("group not found: %s (request_id: %s)", e.GroupID, e.RequestID)
	}
	return "group not found: " + e.GroupID
}

// GroupError struct {
	Message   string
	ErrorCode int
	GroupID   string
	RequestID ErrorRequestID
}

func (e *GroupError) Error() string {
	msg := ""
	if e.GroupID != "" {
		msg = fmt.Sprintf("group error (%s): %s", e.GroupID, e.Message)
	} else {
		msg = "group error: " + e.Message
	}
	if e.ErrorCode > 0 {
		if e.RequestID != "" {
			return fmt.Sprintf("[%d] %s (request_id: %s)", e.ErrorCode, msg, e.RequestID)
		}
		return fmt.Sprintf("[%d] %s", e.ErrorCode, msg)
	}
	if e.RequestID != "" {
		return msg + " (request_id: " + string(e.RequestID) + ")"
	}
	return msg
}

// MemberError represents a group member-related error.
type MemberError struct {
	Message   string
	ErrorCode int
	GroupID   string
	AgentID   string
	RequestID ErrorRequestID
}

func (e *MemberError) Error() string {
	msg := ""
	if e.GroupID != "" && e.AgentID != "" {
		msg = fmt.Sprintf("member error (group: %s, agent: %s): %s", e.GroupID, e.AgentID, e.Message)
	} else {
		msg = "member error: " + e.Message
	}
	if e.ErrorCode > 0 {
		if e.RequestID != "" {
			return fmt.Sprintf("[%d] %s (request_id: %s)", e.ErrorCode, msg, e.RequestID)
		}
		return fmt.Sprintf("[%d] %s", e.ErrorCode, msg)
	}
	if e.RequestID != "" {
		return msg + " (request_id: " + string(e.RequestID) + ")"
	}
	return msg
}

// ErrorFromResponse creates an appropriate error from an API error response.
func ErrorFromResponse(apiErr *ErrorResponse, requestID ErrorRequestID) error {
	if apiErr == nil {
		return nil
	}

	message := apiErr.Error
	if message == "" {
		message = "an error occurred"
	}

	// Try to parse error code from the response
	var code int
	if apiErr.Code != "" {
		if c, err := fmt.Sscanf(apiErr.Code, "%d", &code); err == nil && c == 1 {
			// Successfully parsed numeric code
		}
	}

	// Map broker error codes to specific error types
	switch code {
	case ErrorCodeInvalidCredentials, ErrorCodeTokenExpired, ErrorCodeUnauthorized:
		return &AuthenticationError{
			Message:   message,
			ErrorCode: code,
			RequestID: requestID,
		}
	case ErrorCodeForbidden:
		return &ForbiddenError{
			Message:   message,
			ErrorCode: code,
			RequestID: requestID,
		}
	case ErrorCodeTaskNotFound:
		return &TaskNotFoundError{
			TaskID:    "",
			RequestID: requestID,
		}
	case ErrorCodeTaskValidationError:
		return &TaskError{
			Message:   message,
			ErrorCode: code,
			RequestID: requestID,
		}
	case ErrorCodeTaskStateError:
		return &TaskStateError{
			Message:   message,
			ErrorCode: code,
			RequestID: requestID,
		}
	case ErrorCodeTaskRetryError:
		return &TaskError{
			Message:   message,
			ErrorCode: code,
			RequestID: requestID,
		}
	case ErrorCodeGroupNotFound:
		return &GroupNotFoundError{
			GroupID:   "",
			RequestID: requestID,
		}
	case ErrorCodeGroupValidationError:
		return &GroupError{
			Message:   message,
			ErrorCode: code,
			RequestID: requestID,
		}
	case ErrorCodeMemberNotFound, ErrorCodeMemberError:
		return &MemberError{
			Message:   message,
			ErrorCode: code,
			RequestID: requestID,
		}
	case ErrorCodeValidationError, ErrorCodeMissingRequiredField:
		field := ""
		if apiErr.Details != nil {
			if f, ok := apiErr.Details["field"].(string); ok {
				field = f
			}
		}
		return &ValidationError{
			Message:   message,
			Field:     field,
			ErrorCode: code,
			RequestID: requestID,
		}
	case ErrorCodeAgentNotFound:
		return &AgentNotFoundError{
			AgentID:   "",
			RequestID: requestID,
		}
	case ErrorCodeAgentValidationError:
		return &AgentNotFoundError{
			AgentID:   "",
			RequestID: requestID,
		}
	case ErrorCodeMessageNotFound, ErrorCodeMessageDeliveryError:
		return &MessageError{
			Message:   message,
			RequestID: requestID,
		}
	default:
		// Fallback to generic errors
		return &APIError{
			Message:    message,
			Code:       apiErr.Code,
			StatusCode: 0,
			RequestID:  requestID,
		}
	}
}
