package tui

import (
	"fmt"
	"strings"

	gclassroom "google.golang.org/api/classroom/v1"
)

type namedItem struct {
	title string
	desc  string
}

func (i namedItem) Title() string       { return i.title }
func (i namedItem) Description() string { return i.desc }
func (i namedItem) FilterValue() string { return i.title + " " + i.desc }

type courseItem struct {
	course *gclassroom.Course
}

func (i courseItem) Title() string {
	if i.course == nil || strings.TrimSpace(i.course.Name) == "" {
		return "(untitled class)"
	}
	return i.course.Name
}

func (i courseItem) Description() string {
	if i.course == nil {
		return ""
	}
	section := strings.TrimSpace(i.course.Section)
	if section == "" {
		section = "No section"
	}
	return fmt.Sprintf("%s | %s | %s", section, i.course.CourseState, i.course.Id)
}

func (i courseItem) FilterValue() string {
	if i.course == nil {
		return ""
	}
	return i.course.Name + " " + i.course.Section + " " + i.course.Id
}
