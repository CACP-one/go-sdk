package cacp

import (
	"context"
	"time"
)

// TaskStatus represents the status of a task.
type TaskStatus string

// Task status constants.
const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Task represents a task in the system.
type Task struct {
	ID              string                 `json:"id"`
	TaskType        string                 `json:"task_type"`
	Status          TaskStatus             `json:"status"`
	Priority        string                 `json:"priority"`
	SenderAgentID   string                 `json:"sender_agent_id,omitempty"`
	RecipientAgentID string                 `json:"recipient_agent_id,omitempty"`
	Payload         map[string]interface{} `json:"payload"`
	Result          map[string]interface{} `json:"result,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	RetryCount      int                    `json:"retry_count"`
	MaxRetries      int                    `json:"max_retries"`
	CreatedAt       time.Time              `json:"created_at"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TaskCreate represents a request to create a task.
type TaskCreate struct {
	TaskType        string                 `json:"task_type"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
	Priority        string                 `json:"priority,omitempty"`
	RecipientAgentID string                 `json:"recipient_agent_id,omitempty"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
	MaxRetries      int                    `json:"max_retries,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TaskListOptions represents options for listing tasks.
type TaskListOptions struct {
	Status          TaskStatus `json:"status,omitempty"`
	TaskType        string     `json:"task_type,omitempty"`
	SenderAgentID   string     `json:"sender_agent_id,omitempty"`
	RecipientAgentID string     `json:"recipient_agent_id,omitempty"`
	Priority        string     `json:"priority,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	Offset          int        `json:"offset,omitempty"`
}

// TaskList represents a paginated list of tasks.
type TaskList struct {
	Tasks []*Task `json:"tasks"`
	Total int     `json:"total"`
	Limit int     `json:"limit"`
	Offset int    `json:"offset"`
}

// TasksAPI provides access to task-related operations.
type TasksAPI struct {
	client *Client
}

func newTasksAPI(client *Client) *TasksAPI {
	return &TasksAPI{client: client}
}

// Create creates a new task.
func (t *TasksAPI) Create(ctx context.Context, create *TaskCreate) (*Task, error) {
	if create == nil {
		return nil, &ValidationError{Message: "create is required", Field: "create"}
	}
	if create.TaskType == "" {
		return nil, &ValidationError{Message: "task_type is required", Field: "task_type"}
	}

	var task Task
	err := t.client.post(ctx, "/v1/tasks", create, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Get retrieves a task by ID.
func (t *TasksAPI) Get(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, &ValidationError{Message: "task ID is required", Field: "task_id"}
	}

	var task Task
	err := t.client.get(ctx, "/v1/tasks/"+taskID, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// List retrieves a list of tasks with optional filters.
func (t *TasksAPI) List(ctx context.Context, options *TaskListOptions) (*TaskList, error) {
	var result struct {
		Tasks []Task `json:"tasks"`
		Count int    `json:"count"`
	}
	
	err := t.client.get(ctx, "/v1/tasks", &result)
	if err != nil {
		return nil, err
	}

	taskList := &TaskList{
		Tasks: make([]*Task, 0, len(result.Tasks)),
		Total: result.Count,
	}
	
	for i := range result.Tasks {
		taskList.Tasks = append(taskList.Tasks, &result.Tasks[i])
	}
	
	if options != nil {
		taskList.Limit = options.Limit
		taskList.Offset = options.Offset
	} else {
		taskList.Limit = 100
		taskList.Offset = 0
	}
	
	return taskList, nil
}

// Cancel cancels a task.
func (t *TasksAPI) Cancel(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, &ValidationError{Message: "task ID is required", Field: "task_id"}
	}

	var task Task
	err := t.client.post(ctx, "/v1/tasks/"+taskID+"/cancel", nil, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Retry retries a failed task.
func (t *TasksAPI) Retry(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, &ValidationError{Message: "task ID is required", Field: "task_id"}
	}

	var task Task
	err := t.client.post(ctx, "/v1/tasks/"+taskID+"/retry", nil, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}