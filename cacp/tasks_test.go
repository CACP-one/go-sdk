package cacp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTasksAPI_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tasks" {
			t.Errorf("Expected path /v1/tasks, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"id": "task_abc123",
			"task_type": "data-processing",
			"status": "pending",
			"priority": "normal",
			"payload": {"data_source": "s3://bucket/data.csv"},
			"retry_count": 0,
			"max_retries": 3,
			"created_at": "2024-01-01T00:00:00Z",
			"metadata": {}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	create := &TaskCreate{
		TaskType: "data-processing",
		Payload:  map[string]interface{}{"data_source": "s3://bucket/data.csv"},
		Priority: "normal",
	}

	task, err := client.Tasks().Create(context.Background(), create)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if task.ID != "task_abc123" {
		t.Errorf("Expected task ID 'task_abc123', got '%s'", task.ID)
	}
	if task.TaskType != "data-processing" {
		t.Errorf("Expected task type 'data-processing', got '%s'", task.TaskType)
	}
	if task.Status != TaskStatusPending {
		t.Errorf("Expected status 'pending', got '%s'", task.Status)
	}
}

func TestTasksAPI_CreateWithSchedule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"id": "task_def456",
			"task_type": "scheduled-job",
			"status": "pending",
			"priority": "high",
			"retry_count": 0,
			"max_retries": 5,
			"created_at": "2024-01-01T00:00:00Z",
			"scheduled_at": "2024-01-02T00:00:00Z",
			"metadata": {}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	scheduledAt := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	create := &TaskCreate{
		TaskType:    "scheduled-job",
		Priority:    "high",
		MaxRetries:  5,
		ScheduledAt: &scheduledAt,
	}

	task, err := client.Tasks().Create(context.Background(), create)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if task.MaxRetries != 5 {
		t.Errorf("Expected max_retries 5, got %d", task.MaxRetries)
	}
	if task.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", task.Priority)
	}
	if task.ScheduledAt == nil {
		t.Error("Expected scheduled_at to be set")
	} else if !task.ScheduledAt.Equal(scheduledAt) {
		t.Errorf("Expected scheduled_at %v, got %v", scheduledAt, *task.ScheduledAt)
	}
}

func TestTasksAPI_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tasks/task_xyz789" {
			t.Errorf("Expected path /v1/tasks/task_xyz789, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "task_xyz789",
			"task_type": "data-processing",
			"status": "running",
			"priority": "normal",
			"payload": {"data_source": "s3://bucket/data.csv"},
			"result": null,
			"error_message": null,
			"retry_count": 1,
			"max_retries": 3,
			"created_at": "2024-01-01T00:00:00Z",
			"started_at": "2024-01-01T00:01:00Z",
			"metadata": {}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	task, err := client.Tasks().Get(context.Background(), "task_xyz789")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if task.ID != "task_xyz789" {
		t.Errorf("Expected task ID 'task_xyz789', got '%s'", task.ID)
	}
	if task.Status != TaskStatusRunning {
		t.Errorf("Expected status 'running', got '%s'", task.Status)
	}
	if task.RetryCount != 1 {
		t.Errorf("Expected retry_count 1, got %d", task.RetryCount)
	}
}

func TestTasksAPI_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tasks" {
			t.Errorf("Expected path /v1/tasks, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"tasks": [
				{
					"id": "task_1",
					"task_type": "data-processing",
					"status": "running",
					"priority": "normal",
					"retry_count": 0,
					"max_retries": 3,
					"created_at": "2024-01-01T00:00:00Z",
					"metadata": {}
				},
				{
					"id": "task_2",
					"task_type": "analysis",
					"status": "running",
					"priority": "normal",
					"retry_count": 1,
					"max_retries": 3,
					"created_at": "2024-01-01T00:00:00Z",
					"metadata": {}
				}
			],
			"count": 2
		}`))
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	options := &TaskListOptions{
		Status: TaskStatusRunning,
		Limit:  10,
	}

	taskList, err := client.Tasks().List(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if taskList.Total != 2 {
		t.Errorf("Expected total 2, got %d", taskList.Total)
	}
	if len(taskList.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(taskList.Tasks))
	}
	for _, task := range taskList.Tasks {
		if task.Status != TaskStatusRunning {
			t.Errorf("Expected all tasks to have status 'running', got '%s'", task.Status)
		}
	}
}

func TestTasksAPI_Cancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tasks/task_xyz789/cancel" {
			t.Errorf("Expected path /v1/tasks/task_xyz789/cancel, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "task_xyz789",
			"task_type": "data-processing",
			"status": "cancelled",
			"priority": "normal",
			"retry_count": 1,
			"max_retries": 3,
			"created_at": "2024-01-01T00:00:00Z",
			"completed_at": "2024-01-01T00:05:00Z",
			"metadata": {}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	task, err := client.Tasks().Cancel(context.Background(), "task_xyz789")
	if err != nil {
		t.Fatalf("Failed to cancel task: %v", err)
	}

	if task.Status != TaskStatusCancelled {
		t.Errorf("Expected status 'cancelled', got '%s'", task.Status)
	}
}

func TestTasksAPI_Retry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tasks/task_xyz789/retry" {
			t.Errorf("Expected path /v1/tasks/task_xyz789/retry, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "task_xyz789",
			"task_type": "data-processing",
			"status": "queued",
			"priority": "normal",
			"retry_count": 2,
			"max_retries": 3,
			"created_at": "2024-01-01T00:00:00Z",
			"metadata": {}
		}`))
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	task, err := client.Tasks().Retry(context.Background(), "task_xyz789")
	if err != nil {
		t.Fatalf("Failed to retry task: %v", err)
	}

	if task.Status != TaskStatusQueued {
		t.Errorf("Expected status 'queued', got '%s'", task.Status)
	}
	if task.RetryCount != 2 {
		t.Errorf("Expected retry_count 2, got %d", task.RetryCount)
	}
}

func TestTaskModels(t *testing.T) {
	t.Run("TaskStatus constants", func(t *testing.T) {
		if TaskStatusPending != "pending" {
			t.Errorf("Expected TaskStatusPending to be 'pending', got '%s'", TaskStatusPending)
		}
		if TaskStatusQueued != "queued" {
			t.Errorf("Expected TaskStatusQueued to be 'queued', got '%s'", TaskStatusQueued)
		}
		if TaskStatusRunning != "running" {
			t.Errorf("Expected TaskStatusRunning to be 'running', got '%s'", TaskStatusRunning)
		}
		if TaskStatusCompleted != "completed" {
			t.Errorf("Expected TaskStatusCompleted to be 'completed', got '%s'", TaskStatusCompleted)
		}
		if TaskStatusFailed != "failed" {
			t.Errorf("Expected TaskStatusFailed to be 'failed', got '%s'", TaskStatusFailed)
		}
		if TaskStatusCancelled != "cancelled" {
			t.Errorf("Expected TaskStatusCancelled to be 'cancelled', got '%s'", TaskStatusCancelled)
		}
	})

	t.Run("TaskCreate validation", func(t *testing.T) {
		ctx := context.Background()

		client, err := NewClient(&Config{BaseURL: "http://localhost:4001", APIKey: "test"})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		_, err = client.Tasks().Create(ctx, nil)
		if err == nil {
			t.Error("Expected error for nil TaskCreate")
		}

		create := &TaskCreate{
			Payload: map[string]interface{}{"data": "value"},
		}
		_, err = client.Tasks().Create(ctx, create)
		if err == nil {
			t.Error("Expected error for missing task_type")
		}
	})

	t.Run("Task validation", func(t *testing.T) {
		ctx := context.Background()

		client, err := NewClient(&Config{BaseURL: "http://localhost:4001", APIKey: "test"})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		_, err = client.Tasks().Get(ctx, "")
		if err == nil {
			t.Error("Expected error for empty task ID")
		}
	})
}

func TestTasksAPI_ErrorHandling(t *testing.T) {
	t.Run("Create with missing task_type", func(t *testing.T) {
		client, err := NewClient(&Config{
			BaseURL: "http://localhost:4001",
			APIKey:  "test-key",
		})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		create := &TaskCreate{
			Payload: map[string]interface{}{"data": "value"},
		}

		_, err = client.Tasks().Create(context.Background(), create)
		if err == nil {
			t.Error("Expected error for missing task_type")
		}
	})

	t.Run("Get with empty task ID", func(t *testing.T) {
		client, err := NewClient(&Config{
			BaseURL: "http://localhost:4001",
			APIKey:  "test-key",
		})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		_, err = client.Tasks().Get(context.Background(), "")
		if err == nil {
			t.Error("Expected error for empty task ID")
		}
	})

	t.Run("Cancel with empty task ID", func(t *testing.T) {
		client, err := NewClient(&Config{
			BaseURL: "http://localhost:4001",
			APIKey:  "test-key",
		})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		_, err = client.Tasks().Cancel(context.Background(), "")
		if err == nil {
			t.Error("Expected error for empty task ID")
		}
	})

	t.Run("Retry with empty task ID", func(t *testing.T) {
		client, err := NewClient(&Config{
			BaseURL: "http://localhost:4001",
			APIKey:  "test-key",
		})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		_, err = client.Tasks().Retry(context.Background(), "")
		if err == nil {
			t.Error("Expected error for empty task ID")
		}
	})
}