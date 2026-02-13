package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Announcement struct {
	ID                 string          `json:"id"`
	CourseID           string          `json:"courseId"`
	Text               string          `json:"text"`
	State              string          `json:"state"`
	AlternateLink      string          `json:"alternateLink"`
	CreationTime       time.Time       `json:"creationTime"`
	UpdateTime         time.Time       `json:"updateTime"`
	ScheduledTime      time.Time       `json:"scheduledTime,omitempty"`
	AssigneeMode       string          `json:"assigneeMode,omitempty"`
	CourseWorkMaterial json.RawMessage `json:"courseWorkMaterial,omitempty"`
	TopicID            string          `json:"topicId,omitempty"`
	CreatorUserID      string          `json:"creatorUserId,omitempty"`
}

type AnnouncementList struct {
	Announcements []Announcement `json:"announcements"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
}

func (c *Client) ListAnnouncements(ctx context.Context, courseID string, pageSize int) ([]Announcement, string, error) {
	var allAnnouncements []Announcement
	var pageToken string

	for {
		params := buildListParams(pageSize, pageToken)
		endpoint := fmt.Sprintf("/courses/%s/announcements", url.PathEscape(courseID))
		resp, err := c.get(ctx, endpoint, params)
		if err != nil {
			return nil, "", fmt.Errorf("failed to list announcements for course %s: %w", courseID, err)
		}

		var result AnnouncementList
		if err := json.Unmarshal(resp, &result); err != nil {
			return nil, "", fmt.Errorf("failed to parse announcement list: %w", err)
		}

		allAnnouncements = append(allAnnouncements, result.Announcements...)

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	return allAnnouncements, pageToken, nil
}

func (c *Client) GetAnnouncement(ctx context.Context, courseID, announcementID string) (*Announcement, error) {
	endpoint := fmt.Sprintf("/courses/%s/announcements/%s", url.PathEscape(courseID), url.PathEscape(announcementID))
	resp, err := c.get(ctx, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get announcement %s in course %s: %w", announcementID, courseID, err)
	}

	var announcement Announcement
	if err := json.Unmarshal(resp, &announcement); err != nil {
		return nil, fmt.Errorf("failed to parse announcement: %w", err)
	}

	return &announcement, nil
}
