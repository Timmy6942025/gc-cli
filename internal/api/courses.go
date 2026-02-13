package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type Course struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Section           string          `json:"section"`
	Description       string          `json:"descriptionHeading"`
	Room              string          `json:"room"`
	OwnerID           string          `json:"ownerId"`
	CourseState       string          `json:"courseState"`
	EnrollmentCode    string          `json:"enrollmentCode"`
	CourseTheme       string          `json:"courseTheme"`
	AlternateLink     string          `json:"alternateLink"`
	TeacherGroupEmail string          `json:"teacherGroupEmail"`
	CourseGroupEmail  string          `json:"courseGroupEmail"`
	TeacherFolder     json.RawMessage `json:"teacherFolder,omitempty"`
	CloningOptions    json.RawMessage `json:"cloningOptions,omitempty"`
}

type CourseList struct {
	Courses       []Course `json:"courses"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}

func (c *Client) ListCourses(ctx context.Context, pageSize int) ([]Course, string, error) {
	var allCourses []Course
	var pageToken string

	for {
		params := buildListParams(pageSize, pageToken)
		resp, err := c.get(ctx, "/courses", params)
		if err != nil {
			return nil, "", fmt.Errorf("failed to list courses: %w", err)
		}

		var result CourseList
		if err := json.Unmarshal(resp, &result); err != nil {
			return nil, "", fmt.Errorf("failed to parse course list: %w", err)
		}

		allCourses = append(allCourses, result.Courses...)

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	return allCourses, pageToken, nil
}

func (c *Client) GetCourse(ctx context.Context, courseID string) (*Course, error) {
	endpoint := fmt.Sprintf("/courses/%s", url.PathEscape(courseID))
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get course %s: %w", courseID, err)
	}

	var course Course
	if err := json.Unmarshal(resp, &course); err != nil {
		return nil, fmt.Errorf("failed to parse course: %w", err)
	}

	return &course, nil
}
