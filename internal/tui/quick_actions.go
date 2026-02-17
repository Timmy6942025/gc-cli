package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/classroom"
)

type quickActionDef struct {
	title string
	desc  string
	steps []quickActionStep
	build func(m *model, values map[string]string) (string, error)
}

type quickActionStep struct {
	key         string
	label       string
	placeholder string
	optional    bool
	defaultFn   func(m *model, values map[string]string) string
	resolve     func(m *model, values map[string]string, raw string) (string, error)
}

func (m *model) openActionMenu() {
	defs := m.contextActions()
	if len(defs) == 0 {
		m.status = "No actions available in this view."
		return
	}

	items := make([]list.Item, 0, len(defs))
	for _, def := range defs {
		items = append(items, namedItem{title: def.title, desc: def.desc})
	}

	m.commandMode = false
	m.commandInput.Blur()
	m.commandInput.SetValue("")

	m.actionMode = true
	m.actionSelecting = true
	m.actionDefs = defs
	m.activeAction = nil
	m.actionStep = 0
	m.actionValues = map[string]string{}
	m.actionList.SetItems(items)
	m.actionList.Select(0)
	m.status = "Select an action and press enter."
}

func (m *model) handleActionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.actionMode {
		return m, nil
	}

	if m.actionSelecting {
		switch msg.Type {
		case tea.KeyEsc:
			m.closeActionMode("Action cancelled.")
			return m, nil
		case tea.KeyEnter:
			idx := m.actionList.Index()
			if idx < 0 || idx >= len(m.actionDefs) {
				return m, nil
			}
			def := m.actionDefs[idx]
			m.activeAction = &def
			m.actionSelecting = false
			m.actionStep = 0
			m.actionValues = map[string]string{}
			m.prepareActionStep()
			return m, nil
		default:
			var cmd tea.Cmd
			m.actionList, cmd = m.actionList.Update(msg)
			return m, cmd
		}
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.closeActionMode("Action cancelled.")
		return m, nil
	case tea.KeyEnter:
		if m.activeAction == nil {
			m.closeActionMode("Action cancelled.")
			return m, nil
		}
		step := m.activeAction.steps[m.actionStep]
		raw := strings.TrimSpace(m.actionInput.Value())
		if raw == "" && step.defaultFn != nil {
			raw = strings.TrimSpace(step.defaultFn(m, m.actionValues))
		}
		if raw == "" && !step.optional {
			m.status = step.label + " is required."
			return m, nil
		}

		resolved := raw
		if step.resolve != nil && raw != "" {
			v, err := step.resolve(m, m.actionValues, raw)
			if err != nil {
				m.status = err.Error()
				return m, nil
			}
			resolved = v
		}
		m.actionValues[step.key] = strings.TrimSpace(resolved)
		m.actionStep++

		if m.actionStep >= len(m.activeAction.steps) {
			cmdRaw, err := m.activeAction.build(m, m.actionValues)
			if err != nil {
				m.status = err.Error()
				return m, nil
			}
			m.closeActionMode("Running action...")
			return m, m.runCLICommand(cmdRaw)
		}

		m.prepareActionStep()
		return m, nil
	default:
		var cmd tea.Cmd
		m.actionInput, cmd = m.actionInput.Update(msg)
		return m, cmd
	}
}

func (m *model) prepareActionStep() {
	if m.activeAction == nil || m.actionStep >= len(m.activeAction.steps) {
		return
	}
	step := m.activeAction.steps[m.actionStep]
	defaultValue := ""
	if step.defaultFn != nil {
		defaultValue = step.defaultFn(m, m.actionValues)
	}
	m.actionInput.Prompt = step.label + ": "
	m.actionInput.Placeholder = step.placeholder
	m.actionInput.SetValue(defaultValue)
	m.actionInput.CursorEnd()
	m.actionInput.Focus()
	m.status = fmt.Sprintf("%s (%d/%d)", m.activeAction.title, m.actionStep+1, len(m.activeAction.steps))
}

func (m *model) closeActionMode(status string) {
	m.actionMode = false
	m.actionSelecting = false
	m.actionDefs = nil
	m.activeAction = nil
	m.actionStep = 0
	m.actionValues = map[string]string{}
	m.actionInput.Blur()
	m.actionInput.SetValue("")
	if strings.TrimSpace(status) != "" {
		m.status = status
	}
}

func (m *model) contextActions() []quickActionDef {
	if !m.classMode {
		if m.currentGlobalView() == "To-do" {
			return []quickActionDef{
				{
					title: "Turn in / Mark as done",
					desc:  "Submit a To-do assignment",
					steps: []quickActionStep{
						{key: "todo_ref", label: "To-do item (#)", placeholder: "#1", resolve: resolveTodoRef},
					},
					build: func(_ *model, values map[string]string) (string, error) {
						courseID, courseWorkID, submissionID, err := parseTodoRef(values["todo_ref"])
						if err != nil {
							return "", err
						}
						return "submissions turn-in --course " + quoteArg(courseID) + " --course-work " + quoteArg(courseWorkID) + " --submission " + quoteArg(submissionID), nil
					},
				},
				{
					title: "Unsubmit / Reclaim",
					desc:  "Undo turn in for a To-do item",
					steps: []quickActionStep{
						{key: "todo_ref", label: "To-do item (#)", placeholder: "#1", resolve: resolveTodoRef},
					},
					build: func(_ *model, values map[string]string) (string, error) {
						courseID, courseWorkID, submissionID, err := parseTodoRef(values["todo_ref"])
						if err != nil {
							return "", err
						}
						return "submissions unsubmit --course " + quoteArg(courseID) + " --course-work " + quoteArg(courseWorkID) + " --submission " + quoteArg(submissionID), nil
					},
				},
			}
		}
		return nil
	}

	course := m.selectedCourse()
	if course == nil {
		return nil
	}
	courseID := course.Id

	switch m.currentClassTab() {
	case "Stream":
		return []quickActionDef{
			{
				title: "Post announcement",
				desc:  "Create a Stream announcement",
				steps: []quickActionStep{
					{key: "text", label: "Text", placeholder: "Announcement text"},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					return "stream post --course " + quoteArg(courseID) + " --text " + quoteArg(values["text"]), nil
				},
			},
			{
				title: "Edit announcement",
				desc:  "Edit an existing announcement",
				steps: []quickActionStep{
					{key: "announcement", label: "Announcement (# or ID)", placeholder: "#1 or announcement ID", resolve: resolveAnnouncementRef},
					{key: "text", label: "New text", placeholder: "Updated announcement text"},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					return "stream edit --course " + quoteArg(courseID) + " --announcement " + quoteArg(values["announcement"]) + " --text " + quoteArg(values["text"]), nil
				},
			},
			{
				title: "Delete announcement",
				desc:  "Delete a Stream announcement",
				steps: []quickActionStep{
					{key: "announcement", label: "Announcement (# or ID)", placeholder: "#1 or announcement ID", resolve: resolveAnnouncementRef},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					return "stream delete --course " + quoteArg(courseID) + " --announcement " + quoteArg(values["announcement"]), nil
				},
			},
		}
	case "Classwork":
		return []quickActionDef{
			{
				title: "Create classwork",
				desc:  "Create assignment/question/material",
				steps: []quickActionStep{
					{key: "title", label: "Title", placeholder: "Classwork title"},
					{key: "description", label: "Description", placeholder: "Optional description", optional: true},
					{key: "type", label: "Type", placeholder: "ASSIGNMENT|MATERIAL|SHORT_ANSWER_QUESTION|MULTIPLE_CHOICE_QUESTION", optional: true, defaultFn: func(_ *model, _ map[string]string) string { return "ASSIGNMENT" }},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					cmd := "classwork create --course " + quoteArg(courseID) + " --title " + quoteArg(values["title"])
					if v := strings.TrimSpace(values["description"]); v != "" {
						cmd += " --description " + quoteArg(v)
					}
					if v := strings.TrimSpace(values["type"]); v != "" {
						cmd += " --type " + quoteArg(strings.ToUpper(v))
					}
					return cmd, nil
				},
			},
			{
				title: "Edit classwork",
				desc:  "Edit classwork title/description",
				steps: []quickActionStep{
					{key: "course_work", label: "Classwork (# or ID)", placeholder: "#1 or course-work ID", resolve: resolveCourseWorkRef},
					{key: "title", label: "New title", placeholder: "Optional new title", optional: true},
					{key: "description", label: "New description", placeholder: "Optional new description", optional: true},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					title := strings.TrimSpace(values["title"])
					desc := strings.TrimSpace(values["description"])
					if title == "" && desc == "" {
						return "", fmt.Errorf("set at least title or description")
					}
					cmd := "classwork edit --course " + quoteArg(courseID) + " --course-work " + quoteArg(values["course_work"])
					if title != "" {
						cmd += " --title " + quoteArg(title)
					}
					if desc != "" {
						cmd += " --description " + quoteArg(desc)
					}
					return cmd, nil
				},
			},
			{
				title: "Publish classwork",
				desc:  "Publish a classwork draft",
				steps: []quickActionStep{
					{key: "course_work", label: "Classwork (# or ID)", placeholder: "#1 or course-work ID", resolve: resolveCourseWorkRef},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					return "classwork publish --course " + quoteArg(courseID) + " --course-work " + quoteArg(values["course_work"]), nil
				},
			},
			{
				title: "Schedule classwork",
				desc:  "Schedule classwork publication",
				steps: []quickActionStep{
					{key: "course_work", label: "Classwork (# or ID)", placeholder: "#1 or course-work ID", resolve: resolveCourseWorkRef},
					{key: "when", label: "When (RFC3339)", placeholder: "2026-03-01T09:00:00-05:00"},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					return "classwork schedule --course " + quoteArg(courseID) + " --course-work " + quoteArg(values["course_work"]) + " --when " + quoteArg(values["when"]), nil
				},
			},
			{
				title: "Delete classwork",
				desc:  "Delete a classwork item",
				steps: []quickActionStep{
					{key: "course_work", label: "Classwork (# or ID)", placeholder: "#1 or course-work ID", resolve: resolveCourseWorkRef},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					return "classwork delete --course " + quoteArg(courseID) + " --course-work " + quoteArg(values["course_work"]), nil
				},
			},
			{
				title: "Turn in / Mark as done",
				desc:  "Submit your classwork",
				steps: []quickActionStep{
					{key: "course_work", label: "Classwork (# or ID)", placeholder: "#1 or course-work ID", resolve: resolveCourseWorkRef},
				},
				build: func(m *model, values map[string]string) (string, error) {
					courseWorkID := strings.TrimSpace(values["course_work"])
					submissionID, err := m.resolveMySubmissionID(courseID, courseWorkID)
					if err != nil {
						return "", err
					}
					return "submissions turn-in --course " + quoteArg(courseID) + " --course-work " + quoteArg(courseWorkID) + " --submission " + quoteArg(submissionID), nil
				},
			},
			{
				title: "Unsubmit / Reclaim",
				desc:  "Take back your turned-in work",
				steps: []quickActionStep{
					{key: "course_work", label: "Classwork (# or ID)", placeholder: "#1 or course-work ID", resolve: resolveCourseWorkRef},
				},
				build: func(m *model, values map[string]string) (string, error) {
					courseWorkID := strings.TrimSpace(values["course_work"])
					submissionID, err := m.resolveMySubmissionID(courseID, courseWorkID)
					if err != nil {
						return "", err
					}
					return "submissions unsubmit --course " + quoteArg(courseID) + " --course-work " + quoteArg(courseWorkID) + " --submission " + quoteArg(submissionID), nil
				},
			},
		}
	case "People":
		return []quickActionDef{
			{
				title: "Invite person",
				desc:  "Invite student or teacher",
				steps: []quickActionStep{
					{key: "role", label: "Role", placeholder: "student or teacher", defaultFn: func(_ *model, _ map[string]string) string { return "student" }, resolve: resolveRole},
					{key: "user", label: "User", placeholder: "email or user ID"},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					return "people invite --course " + quoteArg(courseID) + " --role " + quoteArg(values["role"]) + " --user " + quoteArg(values["user"]), nil
				},
			},
			{
				title: "Remove person",
				desc:  "Remove student or teacher",
				steps: []quickActionStep{
					{key: "role", label: "Role", placeholder: "student or teacher", defaultFn: func(_ *model, _ map[string]string) string { return "student" }, resolve: resolveRole},
					{key: "user", label: "User (# or ID/email)", placeholder: "#1 or user ID/email", resolve: resolvePersonRef},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					return "people remove --course " + quoteArg(courseID) + " --role " + quoteArg(values["role"]) + " --user " + quoteArg(values["user"]), nil
				},
			},
		}
	case "Grades":
		return []quickActionDef{
			{
				title: "Set draft grade",
				desc:  "Update Draft grade",
				steps: []quickActionStep{
					{key: "course_work", label: "Classwork (# or ID)", placeholder: "#1 or course-work ID", resolve: resolveCourseWorkRef},
					{key: "submission", label: "Submission ID", placeholder: "student submission ID"},
					{key: "value", label: "Grade value", placeholder: "e.g. 95"},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					if _, err := strconv.ParseFloat(strings.TrimSpace(values["value"]), 64); err != nil {
						return "", fmt.Errorf("invalid grade value: %v", err)
					}
					return "grades set-draft --course " + quoteArg(courseID) + " --course-work " + quoteArg(values["course_work"]) + " --submission " + quoteArg(values["submission"]) + " --value " + quoteArg(values["value"]), nil
				},
			},
			{
				title: "Set assigned grade",
				desc:  "Update Assigned grade",
				steps: []quickActionStep{
					{key: "course_work", label: "Classwork (# or ID)", placeholder: "#1 or course-work ID", resolve: resolveCourseWorkRef},
					{key: "submission", label: "Submission ID", placeholder: "student submission ID"},
					{key: "value", label: "Grade value", placeholder: "e.g. 95"},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					if _, err := strconv.ParseFloat(strings.TrimSpace(values["value"]), 64); err != nil {
						return "", fmt.Errorf("invalid grade value: %v", err)
					}
					return "grades set-assigned --course " + quoteArg(courseID) + " --course-work " + quoteArg(values["course_work"]) + " --submission " + quoteArg(values["submission"]) + " --value " + quoteArg(values["value"]), nil
				},
			},
			{
				title: "Return submission",
				desc:  "Return graded work",
				steps: []quickActionStep{
					{key: "course_work", label: "Classwork (# or ID)", placeholder: "#1 or course-work ID", resolve: resolveCourseWorkRef},
					{key: "submission", label: "Submission ID", placeholder: "student submission ID"},
				},
				build: func(_ *model, values map[string]string) (string, error) {
					return "grades return --course " + quoteArg(courseID) + " --course-work " + quoteArg(values["course_work"]) + " --submission " + quoteArg(values["submission"]), nil
				},
			},
		}
	default:
		return nil
	}
}

func resolveAnnouncementRef(m *model, _ map[string]string, raw string) (string, error) {
	course := m.selectedCourse()
	if course == nil {
		return "", fmt.Errorf("no class selected")
	}
	return resolveByIndexOrRaw(raw, m.streamByCourse[course.Id], func(item *gclassroom.Announcement) string { return strings.TrimSpace(item.Id) }, "announcement")
}

func resolveCourseWorkRef(m *model, _ map[string]string, raw string) (string, error) {
	course := m.selectedCourse()
	if course == nil {
		return "", fmt.Errorf("no class selected")
	}
	return resolveByIndexOrRaw(raw, m.classworkByCourse[course.Id], func(item *gclassroom.CourseWork) string { return strings.TrimSpace(item.Id) }, "classwork")
}

func resolveRole(_ *model, _ map[string]string, raw string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "student", "students":
		return "student", nil
	case "teacher", "teachers":
		return "teacher", nil
	default:
		return "", fmt.Errorf("role must be student or teacher")
	}
}

func resolvePersonRef(m *model, values map[string]string, raw string) (string, error) {
	course := m.selectedCourse()
	if course == nil {
		return "", fmt.Errorf("no class selected")
	}
	role := strings.ToLower(strings.TrimSpace(values["role"]))
	switch role {
	case "teacher", "teachers":
		return resolveByIndexOrRaw(raw, m.teachersByCourse[course.Id], func(item *gclassroom.Teacher) string { return strings.TrimSpace(item.UserId) }, "teacher")
	default:
		return resolveByIndexOrRaw(raw, m.studentsByCourse[course.Id], func(item *gclassroom.Student) string { return strings.TrimSpace(item.UserId) }, "student")
	}
}

func resolveTodoRef(m *model, _ map[string]string, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("to-do item is required")
	}
	if !strings.HasPrefix(raw, "#") {
		if _, _, _, err := parseTodoRef(raw); err != nil {
			return "", err
		}
		return raw, nil
	}
	idx, err := strconv.Atoi(strings.TrimPrefix(raw, "#"))
	if err != nil || idx < 1 {
		return "", fmt.Errorf("invalid to-do shortcut %q", raw)
	}
	if idx > len(m.todoItems) {
		return "", fmt.Errorf("to-do shortcut %q out of range (1-%d)", raw, len(m.todoItems))
	}
	item := m.todoItems[idx-1]
	if strings.TrimSpace(item.CourseID) == "" || strings.TrimSpace(item.CourseWorkID) == "" || strings.TrimSpace(item.SubmissionID) == "" {
		return "", fmt.Errorf("selected to-do item is missing course-work/submission IDs")
	}
	return joinTodoRef(item.CourseID, item.CourseWorkID, item.SubmissionID), nil
}

func (m *model) resolveMySubmissionID(courseID, courseWorkID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	submissions, _, err := m.client.ListStudentSubmissions(ctx, courseID, courseWorkID, "me", classroom.ListParams{PageSize: 5})
	if err != nil {
		return "", fmt.Errorf("load your submission for classwork %s: %w", courseWorkID, err)
	}
	if len(submissions) == 0 {
		return "", fmt.Errorf("no submission found for you in classwork %s", courseWorkID)
	}
	submissionID := strings.TrimSpace(submissions[0].Id)
	if submissionID == "" {
		return "", fmt.Errorf("submission ID is empty for classwork %s", courseWorkID)
	}
	return submissionID, nil
}

func resolveByIndexOrRaw[T any](raw string, items []T, idFn func(item T) string, label string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("%s reference is required", label)
	}
	if !strings.HasPrefix(raw, "#") {
		return raw, nil
	}
	index, err := strconv.Atoi(strings.TrimPrefix(raw, "#"))
	if err != nil || index < 1 {
		return "", fmt.Errorf("invalid %s shortcut %q", label, raw)
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no loaded %s items; use full ID instead", label)
	}
	if index > len(items) {
		return "", fmt.Errorf("%s shortcut %q out of range (1-%d)", label, raw, len(items))
	}
	resolved := strings.TrimSpace(idFn(items[index-1]))
	if resolved == "" {
		return "", fmt.Errorf("selected %s has no ID", label)
	}
	return resolved, nil
}

func joinTodoRef(courseID, courseWorkID, submissionID string) string {
	return strings.TrimSpace(courseID) + "::" + strings.TrimSpace(courseWorkID) + "::" + strings.TrimSpace(submissionID)
}

func parseTodoRef(raw string) (courseID, courseWorkID, submissionID string, err error) {
	parts := strings.Split(strings.TrimSpace(raw), "::")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid to-do reference; expected #<n> or course::courseWork::submission")
	}
	courseID = strings.TrimSpace(parts[0])
	courseWorkID = strings.TrimSpace(parts[1])
	submissionID = strings.TrimSpace(parts[2])
	if courseID == "" || courseWorkID == "" || submissionID == "" {
		return "", "", "", fmt.Errorf("invalid to-do reference; missing IDs")
	}
	return courseID, courseWorkID, submissionID, nil
}

func quoteArg(v string) string {
	return strconv.Quote(strings.TrimSpace(v))
}
