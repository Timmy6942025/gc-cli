package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type CourseWork struct {
	ID                         string          `json:"id"`
	CourseID                   string          `json:"courseId"`
	Title                      string          `json:"title"`
	Description                string          `json:"description"`
	State                      string          `json:"state"`
	WorkType                   string          `json:"workType"`
	MaxPoints                  int64           `json:"maxPoints,omitempty"`
	DueDate                    *Date           `json:"dueDate,omitempty"`
	DueTime                    *TimeOfDay      `json:"dueTime,omitempty"`
	ScheduledDate              *Date           `json:"scheduledDate,omitempty"`
	ScheduledTime              *TimeOfDay      `json:"scheduledTime,omitempty"`
	AllowLateSubmission        bool            `json:"allowLateSubmission"`
	SubmissionModificationTime time.Time       `json:"submissionModificationTime,omitempty"`
	CreateTime                 time.Time       `json:"createTime,omitempty"`
	UpdateTime                 time.Time       `json:"updateTime,omitempty"`
	DraftGrade                 json.RawMessage `json:"draftGrade,omitempty"`
	AssignedGrade              json.RawMessage `json:"assignedGrade,omitempty"`
	CourseWorkMaterial         json.RawMessage `json:"courseWorkMaterial,omitempty"`
	Assignment                 json.RawMessage `json:"assignment,omitempty"`
	MultipleChoiceQuestion     json.RawMessage `json:"multipleChoiceQuestion,omitempty"`
	AlternateLink              string          `json:"alternateLink,omitempty"`
	TeacherFolder              json.RawMessage `json:"teacherFolder,omitempty"`
	TopicID                    string          `json:"topicId,omitempty"`
	GradeCategory              json.RawMessage `json:"gradeCategory,omitempty"`
}

type Date struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type TimeOfDay struct {
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type CourseWorkList struct {
	CourseWork    []CourseWork `json:"courseWork"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

func (c *Client) ListCourseWork(ctx context.Context, courseID string, pageSize int) ([]CourseWork, string, error) {
	var allCourseWork []CourseWork
	var pageToken string

	for {
		params := buildListParams(pageSize, pageToken)
		endpoint := fmt.Sprintf("/courses/%s/courseWork", url.PathEscape(courseID))
		resp, err := c.get(ctx, endpoint, params)
		if err != nil {
			return nil, "", fmt.Errorf("failed to list coursework for course %s: %w", courseID, err)
		}

		var result CourseWorkList
		if err := json.Unmarshal(resp, &result); err != nil {
			return nil, "", fmt.Errorf("failed to parse coursework list: %w", err)
		}

		allCourseWork = append(allCourseWork, result.CourseWork...)

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	return allCourseWork, pageToken, nil
}

func (c *Client) GetCourseWork(ctx context.Context, courseID, courseWorkID string) (*CourseWork, error) {
	endpoint := fmt.Sprintf("/courses/%s/courseWork/%s", url.PathEscape(courseID), url.PathEscape(courseWorkID))
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get coursework %s in course %s: %w", courseWorkID, courseID, err)
	}

	var cw CourseWork
	if err := json.Unmarshal(resp, &cw); err != nil {
		return nil, fmt.Errorf("failed to parse coursework: %w", err)
	}

	return &cw, nil
}
