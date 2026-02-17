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
		return m, nil
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
			"Loaded %d classes (Teaching: %d, Enrolled: %d) at %s",
			len(m.allCourses),
			len(m.teachingCourses),
			len(m.enrolledCourses),
			time.Now().Format(time.Kitchen),
		)
		m.refreshPanels()
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.NextTab):
			if m.classMode {
				cycleList(&m.classTabs, 1)
			} else {
				previous := m.selectedCourseID()
				cycleList(&m.globalList, 1)
				m.applyCourseView(previous)
			}
			m.refreshPanels()
			return m, nil
		case key.Matches(msg, m.keys.PrevTab):
			if m.classMode {
				cycleList(&m.classTabs, -1)
			} else {
				previous := m.selectedCourseID()
				cycleList(&m.globalList, -1)
				m.applyCourseView(previous)
			}
			m.refreshPanels()
			return m, nil
		case key.Matches(msg, m.keys.OpenClass):
			if m.selectedCourse() != nil {
				m.classMode = true
				m.refreshPanels()
			}
			return m, nil
		case key.Matches(msg, m.keys.Back):
			m.classMode = false
			m.refreshPanels()
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			m.status = "Refreshing classes from Google Classroom..."
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
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Keep class list navigation consistent in both global and class modes.
	// Class tabs are intentionally switched via tab/shift-tab (or [/]).
	m.courseList, cmd = m.courseList.Update(msg)
	cmds = append(cmds, cmd)

	if _, isKey := msg.(tea.KeyMsg); !isKey {
		m.content, cmd = m.content.Update(msg)
		cmds = append(cmds, cmd)
	}

	if prevCourseID != m.selectedCourseID() {
		m.refreshPanels()
	}

	return m, tea.Batch(cmds...)
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
			m.content.SetContent("No class selected.")
			return
		}
		tab := m.currentClassTab()
		switch tab {
		case "Stream":
			m.content.SetContent(fmt.Sprintf("Stream\n\nUse `gc-cli stream list --course %s` to view posts\nand `gc-cli stream post --course %s --text ...` to publish.", course.Id, course.Id))
		case "Classwork":
			m.content.SetContent(fmt.Sprintf("Classwork\n\nUse `gc-cli classwork list --course %s`\nto browse assignments/materials.\nUse create/edit/publish/schedule for full lifecycle.", course.Id))
		case "People":
			m.content.SetContent(fmt.Sprintf("People\n\nUse `gc-cli people list --course %s`\nto view teachers/students.\nUse invite/remove to manage roster.", course.Id))
		case "Grades":
			m.content.SetContent(fmt.Sprintf("Grades\n\nUse `gc-cli grades list --course %s`\nfor gradebook, or `gc-cli submissions grade ...`\nfor direct draft/assigned grade updates.", course.Id))
		default:
			m.content.SetContent("")
		}
		return
	}

	switch m.currentGlobalView() {
	case "Home":
		if course == nil {
			m.content.SetContent("Home\n\nSelect a class to enter Stream/Classwork/People/Grades.")
			return
		}
		m.content.SetContent(fmt.Sprintf("Home\n\nSelected class: %s\nCourse ID: %s\n\nPress enter to open class tabs.\nPress o to open this class in Classroom web.", course.Name, course.Id))
	case "To-do":
		m.content.SetContent("To-do\n\nUse `gc-cli to-do list` to view pending student work.")
	case "Calendar":
		m.content.SetContent("Calendar\n\nPress o to open Google Calendar in your browser.")
	case "Teaching":
		m.content.SetContent(fmt.Sprintf("Teaching\n\nShowing %d teaching classes.\nUse tab/shift+tab to switch views.", len(m.teachingCourses)))
	case "Enrolled":
		m.content.SetContent(fmt.Sprintf("Enrolled\n\nShowing %d enrolled classes.\nUse tab/shift+tab to switch views.", len(m.enrolledCourses)))
	default:
		m.content.SetContent("")
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
