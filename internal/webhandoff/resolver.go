package webhandoff

import (
	"fmt"
	"net/url"
	"strings"
)

const baseURL = "https://classroom.google.com"

type Handoff struct {
	Feature  string `json:"feature"`
	CourseID string `json:"course_id,omitempty"`
	URL      string `json:"url,omitempty"`
	Reason   string `json:"reason"`
	Blocked  bool   `json:"blocked"`
}

type HandoffResolver interface {
	Resolve(feature, courseID string, args map[string]string) Handoff
}

type Resolver struct{}

func (Resolver) Resolve(feature, courseID string, args map[string]string) Handoff {
	f := normalizeFeature(feature)
	switch f {
	case "class-settings", "invite-code", "stream-controls", "meet-link":
		if courseID == "" {
			return Handoff{Feature: feature, Reason: "course id is required", Blocked: true}
		}
		return Handoff{Feature: feature, CourseID: courseID, URL: fmt.Sprintf("%s/c/%s/settings", baseURL, url.PathEscape(courseID)), Reason: "This setting is currently managed in Classroom web.", Blocked: false}
	case "analytics":
		if courseID == "" {
			return Handoff{Feature: feature, URL: baseURL, Reason: "Open Classroom analytics in web.", Blocked: false}
		}
		return Handoff{Feature: feature, CourseID: courseID, URL: fmt.Sprintf("%s/c/%s", baseURL, url.PathEscape(courseID)), Reason: "Classroom analytics is available in web.", Blocked: false}
	case "calendar":
		return Handoff{Feature: feature, URL: "https://calendar.google.com", Reason: "Calendar is opened in web.", Blocked: false}
	case "course":
		if courseID == "" {
			return Handoff{Feature: feature, Reason: "course id is required", Blocked: true}
		}
		return Handoff{Feature: feature, CourseID: courseID, URL: fmt.Sprintf("%s/c/%s", baseURL, url.PathEscape(courseID)), Reason: "Open this course in Classroom web.", Blocked: false}
	default:
		if rawURL := strings.TrimSpace(args["url"]); rawURL != "" {
			return Handoff{Feature: feature, CourseID: courseID, URL: rawURL, Reason: "External handoff URL.", Blocked: false}
		}
		return Handoff{Feature: feature, CourseID: courseID, Reason: "unsupported handoff feature", Blocked: true}
	}
}

func normalizeFeature(feature string) string {
	f := strings.TrimSpace(strings.ToLower(feature))
	f = strings.ReplaceAll(f, "_", "-")
	f = strings.ReplaceAll(f, " ", "-")
	return f
}
