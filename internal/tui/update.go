package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/classroom"
	"github.com/timothy/gc-cli/internal/platform"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		m.refreshPanels()
		return m, m.maybeLoadCurrentContent(false)
	case coursesMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Refresh failed: %v", msg.err)
			m.refreshPanels()
			return m, nil
		}
		previous := m.selectedCourseID()
		m.allCourses = msg.courses
		m.teachingCourses = msg.teachingCourses
		m.enrolledCourses = msg.enrolledCourses
		m.applyCourseView(previous)
		m.status = fmt.Sprintf(
			"Loaded %d classes (Teaching: %d, Enrolled: %d) at %s (Focus: %s)",
			len(m.allCourses),
			len(m.teachingCourses),
			len(m.enrolledCourses),
			time.Now().Format(time.Kitchen),
			m.focusLabel(),
		)
		m.refreshPanels()
		return m, m.maybeLoadCurrentContent(true)
	case contentLoadedMsg:
		delete(m.contentLoading, msg.key)
		if courseID, tab, ok := parseClassContentKey(msg.key); ok {
			switch tab {
			case "Stream":
				m.streamByCourse[courseID] = append([]*gclassroom.Announcement(nil), msg.streamItems...)
			case "Classwork":
				m.classworkByCourse[courseID] = append([]*gclassroom.CourseWork(nil), msg.courseWorkItems...)
			case "People":
				m.teachersByCourse[courseID] = append([]*gclassroom.Teacher(nil), msg.teacherItems...)
				m.studentsByCourse[courseID] = append([]*gclassroom.Student(nil), msg.studentItems...)
			}
		}
		if msg.key == "global:todo" {
			m.todoItems = append([]classroom.TodoItem(nil), msg.todoItems...)
		}
		if msg.err != nil && strings.TrimSpace(msg.content) == "" {
			m.contentCache[msg.key] = fmt.Sprintf("Load failed:\n\n%v", msg.err)
			m.status = fmt.Sprintf("Failed to load %s: %v", friendlyContentKey(msg.key), msg.err)
		} else {
			m.contentCache[msg.key] = msg.content
			if msg.status != "" {
				m.status = msg.status
			}
			if msg.err != nil {
				m.status = msg.status + " (partial)"
			}
		}
		if m.currentContentKey() == msg.key {
			m.refreshPanels()
		}
		return m, nil
	case commandResultMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Command failed: %v", msg.err)
		} else {
			m.status = "Command finished. Press r to refresh Classroom data."
			m.resetContentCache()
		}
		output := strings.TrimSpace(msg.output)
		if output == "" {
			output = "(no output)"
		}
		m.setViewportContent(
			"command-output:"+time.Now().Format(time.RFC3339Nano),
			fmt.Sprintf("Command Output\n\n$ gc-cli %s\n\n%s", strings.TrimSpace(msg.command), output),
		)
		m.focus = focusContent
		return m, nil
	case tea.KeyMsg:
		if m.actionMode {
			return m.handleActionKey(msg)
		}
		if m.commandMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.commandMode = false
				m.commandInput.Blur()
				m.commandInput.SetValue("")
				m.status = "Command cancelled"
				return m, nil
			case tea.KeyEnter:
				raw := strings.TrimSpace(m.commandInput.Value())
				m.commandMode = false
				m.commandInput.Blur()
				m.commandInput.SetValue("")
				if raw == "" {
					m.status = "Command cancelled"
					return m, nil
				}
				m.status = "Running command..."
				return m, m.runCLICommand(raw)
			default:
				var cmd tea.Cmd
				m.commandInput, cmd = m.commandInput.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.ActionMenu):
			m.openActionMenu()
			return m, nil
		case key.Matches(msg, m.keys.Command):
			m.commandMode = true
			m.commandInput.SetValue(m.defaultCommandTemplate())
			m.commandInput.CursorEnd()
			m.commandInput.Focus()
			m.status = "Command mode active. Enter to run, esc to cancel."
			return m, nil
		case key.Matches(msg, m.keys.NextPane):
			m.moveFocus(1)
			m.refreshPanels()
			return m, m.maybeLoadCurrentContent(false)
		case key.Matches(msg, m.keys.PrevPane):
			m.moveFocus(-1)
			m.refreshPanels()
			return m, m.maybeLoadCurrentContent(false)
		case key.Matches(msg, m.keys.NextViewTab):
			previous := m.selectedCourseID()
			if m.classMode {
				cycleList(&m.classTabs, 1)
			} else {
				cycleList(&m.globalList, 1)
				m.applyCourseView(previous)
			}
			m.refreshPanels()
			return m, m.maybeLoadCurrentContent(false)
		case key.Matches(msg, m.keys.PrevViewTab):
			previous := m.selectedCourseID()
			if m.classMode {
				cycleList(&m.classTabs, -1)
			} else {
				cycleList(&m.globalList, -1)
				m.applyCourseView(previous)
			}
			m.refreshPanels()
			return m, m.maybeLoadCurrentContent(false)
		case key.Matches(msg, m.keys.OpenClass):
			if m.selectedCourse() != nil {
				m.classMode = true
				m.focus = focusTabs
				m.refreshPanels()
			}
			return m, m.maybeLoadCurrentContent(false)
		case key.Matches(msg, m.keys.Back):
			m.classMode = false
			if m.focus == focusTabs {
				m.focus = focusCourses
			}
			m.refreshPanels()
			return m, m.maybeLoadCurrentContent(false)
		case key.Matches(msg, m.keys.Refresh):
			m.status = fmt.Sprintf("Refreshing classes from Google Classroom... (Focus: %s)", m.focusLabel())
			m.resetContentCache()
			m.refreshPanels()
			return m, m.refreshCourses()
		case key.Matches(msg, m.keys.OpenWeb):
			m.openCurrentInBrowser()
			m.refreshPanels()
			return m, nil
		case key.Matches(msg, m.keys.ToggleHelp):
			m.help.ShowAll = !m.help.ShowAll
			m.refreshPanels()
			return m, nil
		}
	}

	prevCourseID := m.selectedCourseID()
	prevGlobalView := m.currentGlobalView()
	prevClassTab := m.currentClassTab()

	var cmd tea.Cmd
	switch m.focus {
	case focusGlobal:
		m.globalList, cmd = m.globalList.Update(msg)
	case focusTabs:
		if m.classMode {
			m.classTabs, cmd = m.classTabs.Update(msg)
		} else {
			m.focus = focusCourses
			m.courseList, cmd = m.courseList.Update(msg)
		}
	case focusContent:
		m.content, cmd = m.content.Update(msg)
	default:
		m.courseList, cmd = m.courseList.Update(msg)
	}

	changed := false
	if m.currentGlobalView() != prevGlobalView {
		m.applyCourseView(prevCourseID)
		changed = true
	}
	if prevCourseID != m.selectedCourseID() || prevGlobalView != m.currentGlobalView() || prevClassTab != m.currentClassTab() {
		m.refreshPanels()
		changed = true
	}

	loadCmd := m.maybeLoadCurrentContent(false)
	if changed {
		return m, tea.Batch(cmd, loadCmd)
	}
	if loadCmd != nil {
		return m, tea.Batch(cmd, loadCmd)
	}
	return m, cmd
}

func (m *model) moveFocus(step int) {
	areas := []focusArea{focusGlobal, focusCourses}
	if m.classMode {
		areas = append(areas, focusTabs)
	}
	areas = append(areas, focusContent)
	if len(areas) == 0 {
		m.focus = focusCourses
		return
	}
	idx := 0
	for i, area := range areas {
		if area == m.focus {
			idx = i
			break
		}
	}
	idx += step
	for idx < 0 {
		idx += len(areas)
	}
	m.focus = areas[idx%len(areas)]
}

func (m *model) focusLabel() string {
	switch m.focus {
	case focusGlobal:
		return "Global Views"
	case focusTabs:
		return "Class Tabs"
	case focusContent:
		return "Content"
	default:
		return "Classes"
	}
}

func (m *model) refreshPanels() {
	m.updateCourseInfoTable()
	m.updateContent()
}

func (m *model) updateCourseInfoTable() {
	course := m.selectedCourse()
	if course == nil {
		m.courseInfo.SetRows([]table.Row{{"Name", "-"}, {"Course ID", "-"}, {"State", "-"}, {"Section", "-"}, {"Enrollment code", "-"}})
		return
	}
	m.courseInfo.SetRows([]table.Row{
		{"Name", course.Name},
		{"Course ID", course.Id},
		{"State", fallback(course.CourseState, "-")},
		{"Section", fallback(course.Section, "-")},
		{"Enrollment code", fallback(course.EnrollmentCode, "-")},
	})
}

func (m *model) updateContent() {
	course := m.selectedCourse()

	if m.classMode {
		if course == nil {
			m.setViewportContent("class:none", "No class selected.")
			return
		}
		tab := m.currentClassTab()
		key := m.currentContentKey()
		if m.contentLoading[key] {
			m.setViewportContent(key, fmt.Sprintf("%s\n\nLoading from Google Classroom API...", tab))
			return
		}
		if text, ok := m.contentCache[key]; ok {
			m.setViewportContent(key, text)
			return
		}
		m.setViewportContent(key, fmt.Sprintf("%s\n\nNo data loaded yet. Press r to refresh.", tab))
		return
	}

	switch m.currentGlobalView() {
	case "Home":
		if course == nil {
			m.setViewportContent("global:home", "Home\n\nSelect a class to enter Stream/Classwork/People/Grades.")
			return
		}
		m.setViewportContent("global:home", fmt.Sprintf("Home\n\nSelected class: %s\nCourse ID: %s\n\nUse tab/shift+tab to move focus between panes.\nPress enter to open class tabs, o to open this class in Classroom web.", course.Name, course.Id))
	case "To-do":
		key := m.currentContentKey()
		if m.contentLoading[key] {
			m.setViewportContent(key, "To-do\n\nLoading from Google Classroom API...")
			return
		}
		if text, ok := m.contentCache[key]; ok {
			m.setViewportContent(key, text)
			return
		}
		m.setViewportContent(key, "To-do\n\nNo data loaded yet. Press r to refresh.")
	case "Calendar":
		m.setViewportContent("global:calendar", "Calendar\n\nPress o to open Google Calendar in your browser.")
	case "Teaching":
		m.setViewportContent("global:teaching", fmt.Sprintf("Teaching\n\nShowing %d teaching classes.\nUse up/down on Global Views, or [ and ] for quick view switching.", len(m.teachingCourses)))
	case "Enrolled":
		m.setViewportContent("global:enrolled", fmt.Sprintf("Enrolled\n\nShowing %d enrolled classes.\nUse up/down on Global Views, or [ and ] for quick view switching.", len(m.enrolledCourses)))
	default:
		m.setViewportContent("global:unknown", "")
	}
}

func (m *model) setViewportContent(key, text string) {
	m.content.SetContent(text)
	if key != m.lastContentKey {
		m.content.GotoTop()
		m.lastContentKey = key
	}
}

func (m *model) openCurrentInBrowser() {
	if !m.classMode && m.currentGlobalView() == "Calendar" {
		h := m.resolver.Resolve("calendar", "", nil)
		if h.Blocked {
			m.status = h.Reason
			return
		}
		if err := platform.OpenURL(h.URL); err != nil {
			m.status = fmt.Sprintf("Open browser failed: %v", err)
			return
		}
		m.status = "Opened Google Calendar in browser"
		return
	}

	course := m.selectedCourse()
	if course == nil {
		m.status = "No class selected"
		return
	}
	h := m.resolver.Resolve("course", course.Id, nil)
	if h.Blocked {
		m.status = h.Reason
		return
	}
	if err := platform.OpenURL(h.URL); err != nil {
		m.status = fmt.Sprintf("Open browser failed: %v", err)
		return
	}
	m.status = "Opened class in Classroom web"
}

func (m *model) refreshCourses() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		teaching, teachErr := fetchAllCourses(ctx, m.client, classroom.CourseListParams{
			PageSize:     100,
			TeacherID:    "me",
			CourseStates: []string{"ACTIVE"},
		})
		enrolled, enrollErr := fetchAllCourses(ctx, m.client, classroom.CourseListParams{
			PageSize:     100,
			StudentID:    "me",
			CourseStates: []string{"ACTIVE"},
		})

		all := mergeCourses(teaching, enrolled)
		if len(all) == 0 {
			fallback, fallbackErr := fetchAllCourses(ctx, m.client, classroom.CourseListParams{
				PageSize:     100,
				CourseStates: []string{"ACTIVE"},
			})
			if fallbackErr != nil {
				if teachErr != nil {
					return coursesMsg{err: teachErr}
				}
				if enrollErr != nil {
					return coursesMsg{err: enrollErr}
				}
				return coursesMsg{err: fallbackErr}
			}
			all = fallback
		}

		return coursesMsg{
			courses:         all,
			teachingCourses: normalizeCourseList(teaching),
			enrolledCourses: normalizeCourseList(enrolled),
			err:             nil,
		}
	}
}

func fetchAllCourses(ctx context.Context, client classroom.ClassroomClient, params classroom.CourseListParams) ([]*gclassroom.Course, error) {
	var out []*gclassroom.Course
	pageToken := strings.TrimSpace(params.PageToken)
	for {
		pageParams := params
		pageParams.PageToken = pageToken
		items, next, err := client.ListCoursesFiltered(ctx, pageParams)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
		if next == "" {
			break
		}
		pageToken = next
	}
	return normalizeCourseList(out), nil
}

func mergeCourses(groups ...[]*gclassroom.Course) []*gclassroom.Course {
	byID := map[string]*gclassroom.Course{}
	for _, list := range groups {
		for _, c := range list {
			if c == nil || strings.TrimSpace(c.Id) == "" {
				continue
			}
			byID[c.Id] = c
		}
	}
	out := make([]*gclassroom.Course, 0, len(byID))
	for _, c := range byID {
		out = append(out, c)
	}
	return normalizeCourseList(out)
}

func normalizeCourseList(courses []*gclassroom.Course) []*gclassroom.Course {
	out := make([]*gclassroom.Course, 0, len(courses))
	for _, c := range courses {
		if c == nil {
			continue
		}
		if c.CourseState != "" && c.CourseState != "ACTIVE" {
			continue
		}
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool {
		leftName := strings.ToLower(strings.TrimSpace(out[i].Name))
		rightName := strings.ToLower(strings.TrimSpace(out[j].Name))
		if leftName == rightName {
			return out[i].Id < out[j].Id
		}
		return leftName < rightName
	})
	return out
}

func (m *model) applyCourseView(previousCourseID string) {
	switch m.currentGlobalView() {
	case "Teaching":
		m.courses = append([]*gclassroom.Course(nil), m.teachingCourses...)
	case "Enrolled":
		m.courses = append([]*gclassroom.Course(nil), m.enrolledCourses...)
	default:
		m.courses = append([]*gclassroom.Course(nil), m.allCourses...)
	}
	m.setCourseItems(previousCourseID)
}

func (m *model) setCourseItems(previousCourseID string) {
	items := make([]list.Item, 0, len(m.courses))
	for _, c := range m.courses {
		items = append(items, courseItem{course: c})
	}
	m.courseList.SetItems(items)
	if len(items) == 0 {
		m.courseList.Select(0)
		return
	}
	if previousCourseID == "" {
		m.courseList.Select(0)
		return
	}
	for i, item := range items {
		ci, ok := item.(courseItem)
		if ok && ci.course != nil && ci.course.Id == previousCourseID {
			m.courseList.Select(i)
			return
		}
	}
	m.courseList.Select(0)
}

func (m *model) selectedCourse() *gclassroom.Course {
	selected := m.courseList.SelectedItem()
	ci, ok := selected.(courseItem)
	if !ok {
		return nil
	}
	return ci.course
}

func (m *model) selectedCourseID() string {
	course := m.selectedCourse()
	if course == nil {
		return ""
	}
	return course.Id
}

func (m *model) currentGlobalView() string {
	selected := m.globalList.SelectedItem()
	i, ok := selected.(namedItem)
	if !ok {
		return "Home"
	}
	return i.title
}

func (m *model) currentClassTab() string {
	selected := m.classTabs.SelectedItem()
	i, ok := selected.(namedItem)
	if !ok {
		return "Stream"
	}
	return i.title
}
