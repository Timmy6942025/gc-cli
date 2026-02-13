package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type StudentSubmission struct {
	ID                    string          `json:"id"`
	CourseID              string          `json:"courseId"`
	CourseWorkID          string          `json:"courseWorkId"`
	UserID                string          `json:"userId"`
	State                 string          `json:"state"`
	AssignedGrade         float64         `json:"assignedGrade,omitempty"`
	DraftGrade            float64         `json:"draftGrade,omitempty"`
	SubmittedTimestamp    time.Time       `json:"submittedTimestamp,omitempty"`
	ReturnTimestamp       time.Time       `json:"returnTimestamp,omitempty"`
	CourseWorkMaterial    json.RawMessage `json:"courseWorkMaterial,omitempty"`
	AssignmentSubmission  json.RawMessage `json:"assignmentSubmission,omitempty"`
	MultiChoiceSubmission json.RawMessage `json:"multipleChoiceSubmission,omitempty"`
	ShortAnswerSubmission json.RawMessage `json:"shortAnswerSubmission,omitempty"`
	Attachment            json.RawMessage `json:"attachment,omitempty"`
	AlternateLink         string          `json:"alternateLink,omitempty"`
	CourseWorkType        string          `json:"courseWorkType,omitempty"`
	SubmissionHistory     json.RawMessage `json:"submissionHistory,omitempty"`
}

type StudentSubmissionList struct {
	StudentSubmissions []StudentSubmission `json:"studentSubmissions"`
	NextPageToken      string              `json:"nextPageToken,omitempty"`
}

func (c *Client) ListStudentSubmissions(ctx context.Context, courseID, courseWorkID string, pageSize int) ([]StudentSubmission, string, error) {
	var allSubmissions []StudentSubmission
	var pageToken string

	for {
		params := buildListParams(pageSize, pageToken)
		endpoint := fmt.Sprintf("/courses/%s/courseWork/%s/studentSubmissions", url.PathEscape(courseID), url.PathEscape(courseWorkID))
		resp, err := c.get(ctx, endpoint, params)
		if err != nil {
			return nil, "", fmt.Errorf("failed to list submissions for coursework %s in course %s: %w", courseWorkID, courseID, err)
		}

		var result StudentSubmissionList
		if err := json.Unmarshal(resp, &result); err != nil {
			return nil, "", fmt.Errorf("failed to parse submission list: %w", err)
		}

		allSubmissions = append(allSubmissions, result.StudentSubmissions...)

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	return allSubmissions, pageToken, nil
}

func (c *Client) GetStudentSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*StudentSubmission, error) {
	endpoint := fmt.Sprintf("/courses/%s/courseWork/%s/studentSubmissions/%s",
		url.PathEscape(courseID), url.PathEscape(courseWorkID), url.PathEscape(submissionID))
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission %s for coursework %s in course %s: %w", submissionID, courseWorkID, courseID, err)
	}

	var sub StudentSubmission
	if err := json.Unmarshal(resp, &sub); err != nil {
		return nil, fmt.Errorf("failed to parse submission: %w", err)
	}

	return &sub, nil
}

type SubmissionUpdate struct {
	AssignedGrade         float64         `json:"assignedGrade,omitempty"`
	DraftGrade            float64         `json:"draftGrade,omitempty"`
	AssignmentSubmission  json.RawMessage `json:"assignmentSubmission,omitempty"`
	MultiChoiceSubmission json.RawMessage `json:"multipleChoiceSubmission,omitempty"`
	ShortAnswerSubmission json.RawMessage `json:"shortAnswerSubmission,omitempty"`
}

func (c *Client) PatchStudentSubmission(ctx context.Context, courseID, courseWorkID, submissionID string, update *SubmissionUpdate) (*StudentSubmission, error) {
	endpoint := fmt.Sprintf("/courses/%s/courseWork/%s/studentSubmissions/%s",
		url.PathEscape(courseID), url.PathEscape(courseWorkID), url.PathEscape(submissionID))

	body, err := json.Marshal(update)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal submission update: %w", err)
	}

	resp, err := c.patch(ctx, endpoint, nil, body)
	if err != nil {
		return nil, fmt.Errorf("failed to patch submission %s for coursework %s in course %s: %w", submissionID, courseWorkID, courseID, err)
	}

	var sub StudentSubmission
	if err := json.Unmarshal(resp, &sub); err != nil {
		return nil, fmt.Errorf("failed to parse submission response: %w", err)
	}

	return &sub, nil
}

func (c *Client) GetMySubmission(ctx context.Context, courseID, courseWorkID string) (*StudentSubmission, error) {
	endpoint := fmt.Sprintf("/courses/%s/courseWork/%s/studentSubmissions/me",
		url.PathEscape(courseID), url.PathEscape(courseWorkID))
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get my submission for coursework %s in course %s: %w", courseWorkID, courseID, err)
	}

	var sub StudentSubmission
	if err := json.Unmarshal(resp, &sub); err != nil {
		return nil, fmt.Errorf("failed to parse submission: %w", err)
	}

	return &sub, nil
}

type Attachment struct {
	DriveFile    *DriveFile    `json:"driveFile,omitempty"`
	YouTubeVideo *YouTubeVideo `json:"youtubeVideo,omitempty"`
	Link         *Link         `json:"link,omitempty"`
	Form         *Form         `json:"form,omitempty"`
}

type DriveFile struct {
	ID            string              `json:"id,omitempty"`
	Title         string              `json:"title,omitempty"`
	AlternateLink string              `json:"alternateLink,omitempty"`
	FileRef       *DriveFileReference `json:"driveFile,omitempty"`
}

type DriveFileReference struct {
	ID string `json:"id"`
}

type YouTubeVideo struct {
	ID            string `json:"id"`
	AlternateLink string `json:"alternateLink,omitempty"`
}

type Link struct {
	URL          string `json:"url,omitempty"`
	Title        string `json:"title,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
}

type Form struct {
	FormURL      string `json:"formUrl,omitempty"`
	Title        string `json:"title,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
}

type AssignmentSubmission struct {
	Attachments []Attachment `json:"attachments,omitempty"`
}
