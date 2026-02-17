package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/classroom"
)

type contentLoadedMsg struct {
	key             string
	content         string
	status          string
	err             error
	streamItems     []*gclassroom.Announcement
	courseWorkItems []*gclassroom.CourseWork
	teacherItems    []*gclassroom.Teacher
	studentItems    []*gclassroom.Student
}

type gradeSummaryRow struct {
	workID          string
	workTitle       string
	workType        string
	maxPoints       float64
	total           int
	turnedIn        int
	returned        int
	late            int
	assignedAverage float64
	assignedCount   int
	err             error
}

func (m *model) currentContentKey() string {
	if m.classMode {
		course := m.selectedCourse()
		if course == nil {
			return ""
		}
		return "class:" + strings.TrimSpace(course.Id) + ":" + m.currentClassTab()
	}
	if m.currentGlobalView() == "To-do" {
		return "global:todo"
	}
	return ""
}

func (m *model) maybeLoadCurrentContent(force bool) tea.Cmd {
	if m.client == nil {
		return nil
	}
	key := m.currentContentKey()
	if key == "" {
		return nil
	}
	if force {
		delete(m.contentCache, key)
	}
	if m.contentLoading[key] {
		return nil
	}
	if !force {
		if _, ok := m.contentCache[key]; ok {
			return nil
		}
	}
	m.contentLoading[key] = true

	if key == "global:todo" {
		return m.loadTodoContent(key)
	}

	course := m.selectedCourse()
	if course == nil {
		delete(m.contentLoading, key)
		return nil
	}
	return m.loadClassTabContent(course, m.currentClassTab(), key)
}

func (m *model) resetContentCache() {
	m.contentCache = map[string]string{}
	m.contentLoading = map[string]bool{}
	m.streamByCourse = map[string][]*gclassroom.Announcement{}
	m.classworkByCourse = map[string][]*gclassroom.CourseWork{}
	m.teachersByCourse = map[string][]*gclassroom.Teacher{}
	m.studentsByCourse = map[string][]*gclassroom.Student{}
}

func (m *model) loadTodoContent(key string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		items, err := m.client.BuildTodo(ctx)
		if err != nil {
			return contentLoadedMsg{key: key, err: err}
		}

		var b strings.Builder
		b.WriteString("To-do\n\n")
		if len(items) == 0 {
			b.WriteString("No pending work.")
		} else {
			for i, item := range items {
				fmt.Fprintf(&b, "%d. %s\n", i+1, item.CourseWorkTitle)
				fmt.Fprintf(&b, "   Class: %s\n", item.CourseName)
				fmt.Fprintf(&b, "   Submission: %s  State: %s\n", item.SubmissionID, item.SubmissionState)
				if !item.DueDate.IsZero() {
					fmt.Fprintf(&b, "   Due: %s\n", item.DueDate.Local().Format("Jan 2, 2006 3:04 PM"))
				} else {
					b.WriteString("   Due: -\n")
				}
				if item.Late {
					b.WriteString("   Late: yes\n")
				}
				b.WriteString("\n")
			}
		}

		return contentLoadedMsg{
			key:     key,
			content: strings.TrimSpace(b.String()),
			status:  fmt.Sprintf("Loaded To-do (%d items)", len(items)),
		}
	}
}

func (m *model) loadClassTabContent(course *gclassroom.Course, tab, key string) tea.Cmd {
	courseID := strings.TrimSpace(course.Id)
	courseName := strings.TrimSpace(course.Name)

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		switch tab {
		case "Stream":
			items, next, err := m.client.ListAnnouncements(ctx, courseID, classroomListParams(50))
			if err != nil {
				return contentLoadedMsg{key: key, err: err}
			}
			return contentLoadedMsg{
				key:         key,
				content:     formatStreamContent(courseName, items, next),
				status:      fmt.Sprintf("Loaded Stream (%d announcements)", len(items)),
				streamItems: items,
			}
		case "Classwork":
			items, next, err := m.client.ListCourseWork(ctx, courseID, classroomListParams(80))
			if err != nil {
				return contentLoadedMsg{key: key, err: err}
			}
			topics, _, topicErr := m.client.ListTopics(ctx, courseID, classroomListParams(200))
			if topicErr != nil {
				topics = nil
			}
			return contentLoadedMsg{
				key:             key,
				content:         formatClassworkContent(courseName, topics, items, next),
				status:          fmt.Sprintf("Loaded Classwork (%d items)", len(items)),
				courseWorkItems: items,
			}
		case "People":
			teachers, _, teacherErr := m.client.ListTeachers(ctx, courseID, classroomListParams(200))
			students, _, studentErr := m.client.ListStudents(ctx, courseID, classroomListParams(500))
			if teacherErr != nil && studentErr != nil {
				return contentLoadedMsg{key: key, err: fmt.Errorf("list people: teachers: %v, students: %v", teacherErr, studentErr)}
			}
			var partialErr error
			if teacherErr != nil || studentErr != nil {
				partialErr = fmt.Errorf("people data partially loaded")
			}
			return contentLoadedMsg{
				key:          key,
				content:      formatPeopleContent(courseName, teachers, students, teacherErr, studentErr),
				status:       fmt.Sprintf("Loaded People (%d teachers, %d students)", len(teachers), len(students)),
				err:          partialErr,
				teacherItems: teachers,
				studentItems: students,
			}
		case "Grades":
			workItems, _, err := m.client.ListCourseWork(ctx, courseID, classroomListParams(50))
			if err != nil {
				return contentLoadedMsg{key: key, err: err}
			}
			summaries := make([]gradeSummaryRow, 0, len(workItems))
			hasSubmissionErrors := false
			for _, work := range workItems {
				row := gradeSummaryRow{
					workID:    work.Id,
					workTitle: work.Title,
					workType:  work.WorkType,
					maxPoints: work.MaxPoints,
				}
				submissions, _, subErr := m.client.ListStudentSubmissions(ctx, courseID, work.Id, "", classroomListParams(300))
				if subErr != nil {
					submissions, _, subErr = m.client.ListStudentSubmissions(ctx, courseID, work.Id, "me", classroomListParams(300))
				}
				if subErr != nil {
					row.err = subErr
					hasSubmissionErrors = true
					summaries = append(summaries, row)
					continue
				}

				row.total = len(submissions)
				for _, sub := range submissions {
					if sub.Late {
						row.late++
					}
					switch sub.State {
					case "TURNED_IN":
						row.turnedIn++
					case "RETURNED":
						row.returned++
						row.assignedAverage += sub.AssignedGrade
						row.assignedCount++
					}
				}
				summaries = append(summaries, row)
			}

			sort.SliceStable(summaries, func(i, j int) bool {
				left := strings.ToLower(strings.TrimSpace(summaries[i].workTitle))
				right := strings.ToLower(strings.TrimSpace(summaries[j].workTitle))
				if left == right {
					return summaries[i].workID < summaries[j].workID
				}
				return left < right
			})

			return contentLoadedMsg{
				key:     key,
				content: formatGradesContent(courseName, summaries),
				status:  fmt.Sprintf("Loaded Grades (%d classwork items)", len(summaries)),
				err:     nilIfFalse(hasSubmissionErrors),
			}
		default:
			return contentLoadedMsg{
				key:     key,
				content: tab + "\n\nNo renderer registered for this tab.",
				status:  "Loaded " + tab,
			}
		}
	}
}

func classroomListParams(pageSize int64) classroom.ListParams {
	return classroom.ListParams{PageSize: pageSize}
}

func formatStreamContent(courseName string, items []*gclassroom.Announcement, next string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Stream\n\nClass: %s\n\n", fallback(courseName, "-"))
	if len(items) == 0 {
		b.WriteString("No announcements in Stream.")
		return b.String()
	}
	for i, ann := range items {
		text := strings.TrimSpace(ann.Text)
		if text == "" {
			text = "(no text)"
		}
		fmt.Fprintf(&b, "%d. %s\n", i+1, truncateLine(text, 140))
		fmt.Fprintf(&b, "   Updated: %s  ID: %s\n\n", formatAPITime(ann.UpdateTime), ann.Id)
	}
	if strings.TrimSpace(next) != "" {
		fmt.Fprintf(&b, "More announcements available. next_page_token=%s\n", next)
	}
	return strings.TrimSpace(b.String())
}

func formatClassworkContent(courseName string, topics []*gclassroom.Topic, items []*gclassroom.CourseWork, next string) string {
	topicByID := map[string]string{}
	for _, t := range topics {
		if t == nil || strings.TrimSpace(t.TopicId) == "" {
			continue
		}
		topicByID[t.TopicId] = strings.TrimSpace(t.Name)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Classwork\n\nClass: %s\n\n", fallback(courseName, "-"))
	if len(items) == 0 {
		b.WriteString("No classwork found.")
		return b.String()
	}
	for i, work := range items {
		workType := strings.TrimSpace(work.WorkType)
		if workType == "" {
			workType = "-"
		}
		fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, workType, fallback(strings.TrimSpace(work.Title), "(untitled classwork)"))
		fmt.Fprintf(&b, "   State: %s  Points: %.2f  Due: %s\n", fallback(strings.TrimSpace(work.State), "-"), work.MaxPoints, formatDue(work))
		topic := strings.TrimSpace(topicByID[work.TopicId])
		if topic == "" {
			topic = "-"
		}
		fmt.Fprintf(&b, "   Topic: %s  ID: %s\n\n", topic, work.Id)
	}
	if strings.TrimSpace(next) != "" {
		fmt.Fprintf(&b, "More classwork available. next_page_token=%s\n", next)
	}
	return strings.TrimSpace(b.String())
}

func formatPeopleContent(courseName string, teachers []*gclassroom.Teacher, students []*gclassroom.Student, teacherErr error, studentErr error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "People\n\nClass: %s\n\n", fallback(courseName, "-"))

	fmt.Fprintf(&b, "Teachers (%d)\n", len(teachers))
	if teacherErr != nil {
		fmt.Fprintf(&b, "- Could not load teachers: %v\n", teacherErr)
	} else {
		for i, t := range teachers {
			profile := t.Profile
			fmt.Fprintf(&b, "%d. %s <%s> (%s)\n", i+1, profileName(profile), profileEmail(profile), t.UserId)
		}
		if len(teachers) == 0 {
			b.WriteString("- None\n")
		}
	}

	b.WriteString("\n")
	fmt.Fprintf(&b, "Students (%d)\n", len(students))
	if studentErr != nil {
		fmt.Fprintf(&b, "- Could not load students: %v\n", studentErr)
	} else {
		for i, s := range students {
			profile := s.Profile
			fmt.Fprintf(&b, "%d. %s <%s> (%s)\n", i+1, profileName(profile), profileEmail(profile), s.UserId)
		}
		if len(students) == 0 {
			b.WriteString("- None\n")
		}
	}

	return strings.TrimSpace(b.String())
}

func formatGradesContent(courseName string, rows []gradeSummaryRow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Grades\n\nClass: %s\n\n", fallback(courseName, "-"))
	if len(rows) == 0 {
		b.WriteString("No classwork available for gradebook.")
		return b.String()
	}
	for i, row := range rows {
		fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, fallback(strings.TrimSpace(row.workType), "-"), fallback(strings.TrimSpace(row.workTitle), "(untitled classwork)"))
		fmt.Fprintf(&b, "   ID: %s  Max points: %.2f\n", row.workID, row.maxPoints)
		if row.err != nil {
			fmt.Fprintf(&b, "   Grade data unavailable: %s\n\n", row.err.Error())
			continue
		}
		fmt.Fprintf(&b, "   Submissions: %d  Turned in: %d  Returned: %d  Late: %d\n", row.total, row.turnedIn, row.returned, row.late)
		if row.assignedCount > 0 {
			fmt.Fprintf(&b, "   Assigned grade average (returned): %.2f\n\n", row.assignedAverage/float64(row.assignedCount))
		} else {
			b.WriteString("   Assigned grade average (returned): -\n\n")
		}
	}
	return strings.TrimSpace(b.String())
}

func formatDue(work *gclassroom.CourseWork) string {
	if work == nil || work.DueDate == nil {
		return "-"
	}
	hour, minute, second := 23, 59, 59
	if work.DueTime != nil {
		hour = int(work.DueTime.Hours)
		minute = int(work.DueTime.Minutes)
		second = int(work.DueTime.Seconds)
	}
	t := time.Date(
		int(work.DueDate.Year),
		time.Month(work.DueDate.Month),
		int(work.DueDate.Day),
		hour,
		minute,
		second,
		0,
		time.Local,
	)
	return t.Format("Jan 2, 2006 3:04 PM")
}

func formatAPITime(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}
	return t.Local().Format("Jan 2, 2006 3:04 PM")
}

func profileName(p *gclassroom.UserProfile) string {
	if p == nil {
		return "-"
	}
	if p.Name != nil && strings.TrimSpace(p.Name.FullName) != "" {
		return strings.TrimSpace(p.Name.FullName)
	}
	if strings.TrimSpace(p.EmailAddress) != "" {
		return strings.TrimSpace(p.EmailAddress)
	}
	if strings.TrimSpace(p.Id) != "" {
		return strings.TrimSpace(p.Id)
	}
	return "-"
}

func profileEmail(p *gclassroom.UserProfile) string {
	if p == nil || strings.TrimSpace(p.EmailAddress) == "" {
		return "-"
	}
	return strings.TrimSpace(p.EmailAddress)
}

func truncateLine(v string, max int) string {
	v = strings.TrimSpace(v)
	if max <= 0 || len(v) <= max {
		return v
	}
	if max <= 3 {
		return v[:max]
	}
	return v[:max-3] + "..."
}

func nilIfFalse(flag bool) error {
	if flag {
		return fmt.Errorf("one or more grade rows could not be loaded")
	}
	return nil
}

func friendlyContentKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "content"
	}
	if key == "global:todo" {
		return "To-do"
	}
	if strings.HasPrefix(key, "class:") {
		parts := strings.Split(key, ":")
		if len(parts) >= 3 {
			return parts[2]
		}
		return "class tab"
	}
	return key
}

func parseClassContentKey(key string) (courseID string, tab string, ok bool) {
	parts := strings.Split(strings.TrimSpace(key), ":")
	if len(parts) != 3 || parts[0] != "class" {
		return "", "", false
	}
	return parts[1], parts[2], true
}
