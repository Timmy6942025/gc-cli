package tui

import (
	"context"
	"fmt"
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
		m.courses = msg.courses
		m.setCourseItems(previous)
		m.status = fmt.Sprintf("Loaded %d classes at %s", len(msg.courses), time.Now().Format(time.Kitchen))
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
				cycleList(&m.globalList, 1)
			}
			m.refreshPanels()
			return m, nil
		case key.Matches(msg, m.keys.PrevTab):
			if m.classMode {
				cycleList(&m.classTabs, -1)
			} else {
				cycleList(&m.globalList, -1)
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
			m.status = "Refreshing classes..."
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

	prev := m.selectedCourseID()
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.courseList, cmd = m.courseList.Update(msg)
	cmds = append(cmds, cmd)

	m.content, cmd = m.content.Update(msg)
	cmds = append(cmds, cmd)

	if prev != m.selectedCourseID() {
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
		m.content.SetContent("Teaching\n\nUse `gc-cli classes list` to browse classes you teach.")
	case "Enrolled":
		m.content.SetContent("Enrolled\n\nUse `gc-cli classes list` to browse classes you are enrolled in.")
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
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		courses, _, err := m.client.ListCourses(ctx, classroom.ListParams{PageSize: 200})
		return coursesMsg{courses: courses, err: err}
	}
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
