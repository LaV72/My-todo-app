package models

import "time"

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	StatusActive     TaskStatus = "active"
	StatusInProgress TaskStatus = "in_progress"
	StatusComplete   TaskStatus = "complete"
	StatusFailed     TaskStatus = "failed"
	StatusArchived   TaskStatus = "archived"
)

// Task represents a to-do item in the quest journal
type Task struct {
	ID          string       `json:"id" db:"id"`
	Title       string       `json:"title" db:"title"`
	Description string       `json:"description" db:"description"`
	Priority    int          `json:"priority" db:"priority"`               // 1-5 stars
	Deadline    Deadline     `json:"deadline"`                             // Embedded deadline info
	Category    string       `json:"category" db:"category"`               // Category ID
	Status      TaskStatus   `json:"status" db:"status"`
	Objectives  []Objective  `json:"objectives"`                           // Sub-tasks
	Notes       string       `json:"notes" db:"notes"`
	Reward      int          `json:"reward" db:"reward"`                   // Points/BP
	Tags        []string     `json:"tags"`                                 // Tags for filtering
	Order       int          `json:"order" db:"order_index"`               // Manual ordering
	CreatedAt   time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time    `json:"updatedAt" db:"updated_at"`
	CompletedAt *time.Time   `json:"completedAt,omitempty" db:"completed_at"`

	// Computed fields (not stored in DB, calculated on demand)
	Progress  float64 `json:"progress" db:"-"`    // 0.0 - 1.0
	IsOverdue bool    `json:"isOverdue" db:"-"`
	DaysLeft  *int    `json:"daysLeft,omitempty" db:"-"`
}

// Deadline represents the deadline information for a task
type Deadline struct {
	Type string     `json:"type" db:"deadline_type"` // "short", "medium", "long", "none"
	Date *time.Time `json:"date,omitempty" db:"deadline_date"`
}

// Objective represents a sub-task within a main task
type Objective struct {
	ID        string    `json:"id" db:"id"`
	TaskID    string    `json:"taskId" db:"task_id"`
	Text      string    `json:"text" db:"text"`
	Completed bool      `json:"completed" db:"completed"`
	Order     int       `json:"order" db:"order_index"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// Category represents a grouping for tasks (Main/Side quests, projects, contexts)
type Category struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Color     string    `json:"color" db:"color"`         // Hex color code
	Icon      string    `json:"icon" db:"icon"`           // Icon identifier
	Type      string    `json:"type" db:"type"`           // "main" or "side"
	Order     int       `json:"order" db:"order_index"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// Stats represents overall statistics
type Stats struct {
	TotalTasks             int                `json:"totalTasks"`
	ActiveTasks            int                `json:"activeTasks"`
	CompletedTasks         int                `json:"completedTasks"`
	FailedTasks            int                `json:"failedTasks"`
	TotalRewards           int                `json:"totalRewards"`
	CompletionRate         float64            `json:"completionRate"`          // 0.0 - 1.0
	AverageTimeToComplete  float64            `json:"averageTimeToComplete"`   // hours
	StreakDays             int                `json:"streakDays"`
	CategoryStats          map[string]int     `json:"categoryStats"`
	PriorityStats          map[int]int        `json:"priorityStats"`
}

// DailyStat represents statistics for a single day
type DailyStat struct {
	Date           time.Time `json:"date" db:"date"`
	TasksCompleted int       `json:"tasksCompleted" db:"tasks_completed"`
	TasksCreated   int       `json:"tasksCreated" db:"tasks_created"`
	RewardsEarned  int       `json:"rewardsEarned" db:"rewards_earned"`
}

// CategoryStat represents statistics for a category
type CategoryStat struct {
	CategoryID     string  `json:"categoryId"`
	TotalTasks     int     `json:"totalTasks"`
	CompletedTasks int     `json:"completedTasks"`
	CompletionRate float64 `json:"completionRate"`
}

// TaskFilter represents filtering options for querying tasks
type TaskFilter struct {
	Status           []TaskStatus
	Priority         []int
	Categories       []string
	Tags             []string
	DeadlineType     string
	DateFrom         *time.Time
	DateTo           *time.Time
	IncludeCompleted bool
	SortBy           string // "priority", "deadline", "created_at", etc.
	SortOrder        string // "asc", "desc"
	Limit            int
	Offset           int
}
