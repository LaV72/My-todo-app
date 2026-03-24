package models

// TaskCreateRequest represents the payload for creating a new task
type TaskCreateRequest struct {
	Title       string              `json:"title" validate:"required,min=1,max=200"`
	Description string              `json:"description" validate:"max=2000"`
	Priority    int                 `json:"priority" validate:"required,min=1,max=5"`
	Deadline    *Deadline           `json:"deadline,omitempty"`
	Category    string              `json:"category,omitempty"`
	Objectives  []ObjectiveRequest  `json:"objectives,omitempty"`
	Notes       string              `json:"notes,omitempty" validate:"omitempty,max=5000"`
	Reward      int                 `json:"reward,omitempty" validate:"omitempty,min=0,max=99999"`
	Tags        []string            `json:"tags,omitempty" validate:"omitempty,max=20,dive,min=1,max=50"`
}

// TaskUpdateRequest represents the payload for updating a task
type TaskUpdateRequest struct {
	Title       *string             `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string             `json:"description,omitempty" validate:"omitempty,max=2000"`
	Priority    *int                `json:"priority,omitempty" validate:"omitempty,min=1,max=5"`
	Deadline    *Deadline           `json:"deadline,omitempty"`
	Category    *string             `json:"category,omitempty"`
	Notes       *string             `json:"notes,omitempty" validate:"omitempty,max=5000"`
	Reward      *int                `json:"reward,omitempty" validate:"omitempty,min=0,max=99999"`
	Tags        []string            `json:"tags,omitempty" validate:"omitempty,max=20,dive,min=1,max=50"`
}

// TaskStatusUpdateRequest represents a request to update only the task status
type TaskStatusUpdateRequest struct {
	Status TaskStatus `json:"status" validate:"required"`
}

// ObjectiveRequest represents an objective in a create/update request
type ObjectiveRequest struct {
	Text  string `json:"text" validate:"required,min=1,max=500"`
	Order int    `json:"order,omitempty" validate:"omitempty,min=0,max=10000"`
}

// ObjectiveUpdateRequest represents an objective update
type ObjectiveUpdateRequest struct {
	Text      *string `json:"text,omitempty" validate:"omitempty,min=1,max=500"`
	Completed *bool   `json:"completed,omitempty"`
}

// CategoryCreateRequest represents the payload for creating a category
type CategoryCreateRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Color string `json:"color" validate:"required,hexcolor"`
	Icon  string `json:"icon,omitempty" validate:"omitempty,max=50"`
	Type  string `json:"type" validate:"required,oneof=main side"`
}

// CategoryUpdateRequest represents the payload for updating a category
type CategoryUpdateRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Color *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
	Icon  *string `json:"icon,omitempty" validate:"omitempty,max=50"`
}

// BulkTaskCreateRequest represents a request to create multiple tasks
type BulkTaskCreateRequest struct {
	Tasks []TaskCreateRequest `json:"tasks" validate:"required,min=1,max=50,dive"`
}

// BulkTaskDeleteRequest represents a request to delete multiple tasks
type BulkTaskDeleteRequest struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// TaskReorderRequest represents a request to reorder tasks
type TaskReorderRequest struct {
	IDs []string `json:"ids" validate:"required,min=1,max=1000"`
}
