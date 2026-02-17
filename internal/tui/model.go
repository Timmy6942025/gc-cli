package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	gclassroom "google.golang.org/api/classroom/v1"

	"github.com/timothy/gc-cli/internal/classroom"
	"github.com/timothy/gc-cli/internal/webhandoff"
)

type model struct {
	client   classroom.ClassroomClient
	resolver webhandoff.HandoffResolver

	keys appKeyMap
	help help.Model

	globalList list.Model
	courseList list.Model
	classTabs  list.Model
	courseInfo table.Model
	content    viewport.Model

	courses         []*gclassroom.Course
	allCourses      []*gclassroom.Course
	teachingCourses []*gclassroom.Course
	enrolledCourses []*gclassroom.Course

	status string

	classMode bool

	width       int
	height      int
	leftWidth   int
	middleWidth int
	rightWidth  int
}

type coursesMsg struct {
	courses         []*gclassroom.Course
	teachingCourses []*gclassroom.Course
	enrolledCourses []*gclassroom.Course
	err             error
}

func NewModel(client classroom.ClassroomClient, resolver webhandoff.HandoffResolver) tea.Model {
	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(0)

	globalItems := []list.Item{
		namedItem{title: "Home", desc: "Overview"},
		namedItem{title: "To-do", desc: "Student work queue"},
		namedItem{title: "Calendar", desc: "Class schedule"},
		namedItem{title: "Teaching", desc: "Classes you teach"},
		namedItem{title: "Enrolled", desc: "Classes you attend"},
	}
	globalList := list.New(globalItems, delegate, 24, 12)
	globalList.Title = "Global Views"
	globalList.SetShowHelp(false)
	globalList.SetShowFilter(false)
	globalList.SetShowStatusBar(false)
	globalList.SetShowPagination(false)
	globalList.DisableQuitKeybindings()

	courseList := list.New([]list.Item{}, delegate, 44, 12)
	courseList.Title = "Classes"
	courseList.SetShowHelp(false)
	courseList.SetShowFilter(false)
	courseList.SetShowStatusBar(false)
	courseList.SetShowPagination(false)
	courseList.DisableQuitKeybindings()
	courseList.SetFilteringEnabled(false)

	tabs := list.New([]list.Item{
		namedItem{title: "Stream", desc: "Announcements and posts"},
		namedItem{title: "Classwork", desc: "Assignments and materials"},
		namedItem{title: "People", desc: "Teachers and students"},
		namedItem{title: "Grades", desc: "Draft and assigned grades"},
	}, delegate, 28, 6)
	tabs.Title = "Class Tabs"
	tabs.SetShowHelp(false)
	tabs.SetShowFilter(false)
	tabs.SetShowStatusBar(false)
	tabs.SetShowPagination(false)
	tabs.DisableQuitKeybindings()
	tabs.SetFilteringEnabled(false)

	courseInfo := table.New(
		table.WithColumns([]table.Column{{Title: "Field", Width: 14}, {Title: "Value", Width: 36}}),
		table.WithRows([]table.Row{{"Name", "-"}, {"Course ID", "-"}, {"State", "-"}}),
		table.WithFocused(false),
		table.WithHeight(5),
	)

	content := viewport.New(36, 10)
	content.SetContent("Loading courses...")

	helper := help.New()
	helper.ShowAll = false

	m := &model{
		client:      client,
		resolver:    resolver,
		keys:        newKeyMap(),
		help:        helper,
		globalList:  globalList,
		courseList:  courseList,
		classTabs:   tabs,
		courseInfo:  courseInfo,
		content:     content,
		status:      "Press r to refresh, enter to open class tabs, tab/[ ] to switch view/tab, ? for help.",
		classMode:   false,
		leftWidth:   24,
		middleWidth: 44,
		rightWidth:  44,
	}
	m.refreshPanels()
	return m
}

func (m *model) Init() tea.Cmd {
	if m.client == nil {
		m.status = "Not authenticated. Run gc-cli auth login."
		m.refreshPanels()
		return nil
	}
	return m.refreshCourses()
}

func Run(client classroom.ClassroomClient, resolver webhandoff.HandoffResolver) error {
	p := tea.NewProgram(NewModel(client, resolver), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
