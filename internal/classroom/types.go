package classroom

import (
	"time"
)

// Domain snapshots persisted in local cache for stale-aware reads.
type CourseSnapshot struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Section        string    `json:"section,omitempty"`
	Description    string    `json:"description,omitempty"`
	OwnerID        string    `json:"owner_id,omitempty"`
	EnrollmentCode string    `json:"enrollment_code,omitempty"`
	State          string    `json:"state,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CourseWorkSnapshot struct {
	CourseID     string    `json:"course_id"`
	CourseWorkID string    `json:"course_work_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	WorkType     string    `json:"work_type"`
	State        string    `json:"state"`
	TopicID      string    `json:"topic_id,omitempty"`
	MaxPoints    float64   `json:"max_points,omitempty"`
	DueAt        time.Time `json:"due_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type StudentSubmissionSnapshot struct {
	CourseID      string    `json:"course_id"`
	CourseWorkID  string    `json:"course_work_id"`
	SubmissionID  string    `json:"submission_id"`
	UserID        string    `json:"user_id"`
	State         string    `json:"state"`
	AssignedGrade float64   `json:"assigned_grade,omitempty"`
	DraftGrade    float64   `json:"draft_grade,omitempty"`
	Late          bool      `json:"late"`
	UpdateTime    time.Time `json:"update_time"`
}

type RosterSnapshot struct {
	CourseID       string   `json:"course_id"`
	TeacherUserIDs []string `json:"teacher_user_ids"`
	StudentUserIDs []string `json:"student_user_ids"`
}

type GradebookSnapshot struct {
	CourseID     string                      `json:"course_id"`
	CourseWorkID string                      `json:"course_work_id"`
	Submissions  []StudentSubmissionSnapshot `json:"submissions"`
}

type ParityStatus struct {
	Feature string `json:"feature"`
	Status  string `json:"status"`
	Notes   string `json:"notes,omitempty"`
}

type ListParams struct {
	PageSize  int64
	PageToken string
}

type TodoItem struct {
	CourseID        string    `json:"course_id"`
	CourseName      string    `json:"course_name"`
	CourseWorkID    string    `json:"course_work_id"`
	CourseWorkTitle string    `json:"course_work_title"`
	DueDate         time.Time `json:"due_date,omitempty"`
	SubmissionID    string    `json:"submission_id"`
	SubmissionState string    `json:"submission_state"`
	Late            bool      `json:"late"`
}
