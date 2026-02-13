package tui

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/timboy697/gc-cli/internal/auth"
	"github.com/timboy697/gc-cli/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

type ViewType int

const (
	ViewMainMenu ViewType = iota
	ViewCourses
	ViewCoursework
	ViewGrades
	ViewAnnouncements
	ViewLoading
	ViewError
	ViewAuthRequired
)

type AuthState int

const (
	AuthUnknown AuthState = iota
	AuthNotAuthenticated
	AuthAuthenticated
)

type MenuItem struct {
	title       string
	description string
	view        ViewType
}

func (m MenuItem) Title() string       { return m.title }
func (m MenuItem) Description() string { return m.description }
func (m MenuItem) FilterValue() string { return m.title }

type Model struct {
	CurrentView  ViewType
	PreviousView ViewType
	AuthState    AuthState

	Menu         list.Model
	SelectedMenu int

	Courses       []CourseItem
	Coursework    []CourseworkItem
	Grades        []GradeItem
	Announcements []AnnouncementItem

	SelectedCoursework int

	Viewport viewport.Model

	IsLoading  bool
	LoadingMsg string

	ErrorMsg string

	Config *config.Config

	Width  int
	Height int
}

type CourseItem struct {
	ID      string
	Name    string
	Section string
	Desc    string
	Room    string
}

func (c CourseItem) Title() string       { return c.Name }
func (c CourseItem) Description() string { return c.Section }
func (c CourseItem) FilterValue() string { return c.Name }

type GradeItem struct {
	CourseName  string
	Assignment  string
	Score       string
	MaxScore    string
	SubmittedAt string
}

func (g GradeItem) Title() string { return g.Assignment }
func (g GradeItem) Description() string {
	return fmt.Sprintf("%s — %s/%s", g.CourseName, g.Score, g.MaxScore)
}
func (g GradeItem) FilterValue() string { return g.Assignment }

type AnnouncementItem struct {
	CourseName    string
	AnnounceTitle string
	Text          string
	PostedAt      string
}

func (a AnnouncementItem) Title() string { return a.AnnounceTitle }
func (a AnnouncementItem) Description() string {
	return fmt.Sprintf("%s — %s", a.CourseName, a.PostedAt)
}
func (a AnnouncementItem) FilterValue() string { return a.AnnounceTitle }

type CourseworkStatus int

const (
	StatusPending CourseworkStatus = iota
	StatusTurnedIn
	StatusReturned
	StatusOverdue
	StatusDraft
)

type CourseworkItem struct {
	ID          string
	CourseID    string
	CourseName  string
	AssignTitle string
	Desc        string
	State       string
	DueDate     string
	DueTime     string
	Points      int64
	Status      CourseworkStatus
	WorkType    string
}

func (c CourseworkItem) Title() string { return c.AssignTitle }
func (c CourseworkItem) Description() string {
	return fmt.Sprintf("%s — %s — %s", c.CourseName, c.DueDate, c.StatusString())
}
func (c CourseworkItem) FilterValue() string { return c.AssignTitle }

func (c CourseworkItem) StatusString() string {
	switch c.Status {
	case StatusTurnedIn:
		return "TURNED_IN"
	case StatusReturned:
		return "RETURNED"
	case StatusOverdue:
		return "OVERDUE"
	case StatusDraft:
		return "DRAFT"
	default:
		return "NEW"
	}
}

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Select   key.Binding
	Back     key.Binding
	Quit     key.Binding
	Refresh  key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "back"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "select"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "go back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "page down"),
	),
}

var (
	bgPrimary       = lipgloss.Color("#0f0f14")
	bgSecondary     = lipgloss.Color("#18181f")
	bgTertiary      = lipgloss.Color("#22222a")
	bgHighlight     = lipgloss.Color("#2d2d3a")
	textPrimary     = lipgloss.Color("#e8e8ed")
	textSecondary   = lipgloss.Color("#9898a6")
	textMuted       = lipgloss.Color("#5c5c6e")
	accentPrimary   = lipgloss.Color("#7c6fff")
	accentSecondary = lipgloss.Color("#ff6b9d")
	accentTertiary  = lipgloss.Color("#4ecdc4")
	successColor    = lipgloss.Color("#5fd068")
	errorColor      = lipgloss.Color("#ff6b6b")
	warningColor    = lipgloss.Color("#ffd93d")
	borderColor     = lipgloss.Color("#3a3a4a")

	windowStyle = lipgloss.NewStyle().
			Background(bgPrimary).
			Foreground(textPrimary).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Background(bgSecondary).
			Foreground(accentPrimary).
			Bold(true).
			Padding(1, 2).
			Width(0).
			Align(lipgloss.Center)

	contentStyle = lipgloss.NewStyle().
			Background(bgSecondary).
			Foreground(textPrimary).
			Padding(1, 2)

	loadingStyle = lipgloss.NewStyle().
			Background(bgPrimary).
			Foreground(accentPrimary).
			Bold(true).
			Padding(2, 0)

	errorStyle = lipgloss.NewStyle().
			Background(bgPrimary).
			Foreground(errorColor).
			Padding(2, 0)

	statusBarStyle = lipgloss.NewStyle().
			Background(bgTertiary).
			Foreground(textSecondary).
			Padding(0, 2).
			Height(1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1)

	listStyle = lipgloss.NewStyle().
			Background(bgPrimary)

	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(accentPrimary).
				Bold(true).
				Padding(0, 0, 1, 0)

	infoLabelStyle = lipgloss.NewStyle().
			Foreground(textSecondary).
			Width(15).
			Align(lipgloss.Right)

	infoValueStyle = lipgloss.NewStyle().
			Foreground(textPrimary)
)

func New(cfg *config.Config) Model {
	menuItems := []MenuItem{
		{"Courses", "View your enrolled courses", ViewCourses},
		{"Coursework", "View assignments and deadlines", ViewCoursework},
		{"Grades", "Check your grades and scores", ViewGrades},
		{"Announcements", "View course announcements", ViewAnnouncements},
		{"Quit", "Exit the application", ViewMainMenu},
	}

	items := make([]list.Item, len(menuItems))
	for i := range menuItems {
		items[i] = menuItems[i]
	}

	delegate := list.NewDefaultDelegate()
	menuList := list.New(items, delegate, 0, 0)
	menuList.SetShowHelp(false)
	menuList.SetShowStatusBar(false)
	menuList.SetFilteringEnabled(false)
	menuList.SetShowPagination(false)

	authState := AuthNotAuthenticated
	if cfg != nil && auth.TokenExists(cfg.Auth.TokenFile) {
		authState = AuthAuthenticated
	}

	return Model{
		CurrentView:  ViewMainMenu,
		PreviousView: ViewMainMenu,
		AuthState:    authState,
		Menu:         menuList,
		SelectedMenu: 0,
		Config:       cfg,
		IsLoading:    false,
		LoadingMsg:   "Loading...",
		Width:        80,
		Height:       24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Viewport.Width = msg.Width - 4
		m.Viewport.Height = msg.Height - 6
		m.Menu.SetSize(msg.Width-4, msg.Height-6)
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.IsLoading {
		return m, nil
	}

	switch m.CurrentView {
	case ViewMainMenu:
		m.Menu, cmd = m.Menu.Update(msg)
		cmds = append(cmds, cmd)

	case ViewCourses, ViewCoursework, ViewGrades, ViewAnnouncements:
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Quit) {
		if m.CurrentView == ViewMainMenu {
			return m, tea.Quit
		}
		m.PreviousView = m.CurrentView
		m.CurrentView = ViewMainMenu
		return m, nil
	}

	if key.Matches(msg, keys.Back) {
		if m.CurrentView != ViewMainMenu {
			m.PreviousView = m.CurrentView
			m.CurrentView = ViewMainMenu
		}
		return m, nil
	}

	switch m.CurrentView {
	case ViewMainMenu:
		return m.handleMainMenuKey(msg)

	case ViewCourses, ViewCoursework, ViewGrades, ViewAnnouncements:
		return m.handleContentKey(msg)

	case ViewAuthRequired:
		if key.Matches(msg, keys.Select) {
			m.PreviousView = m.CurrentView
			m.CurrentView = ViewMainMenu
		}
	}

	return m, nil
}

func (m Model) handleMainMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Up) {
		if m.Menu.Index() > 0 {
			m.Menu.CursorUp()
		}
		return m, nil
	}

	if key.Matches(msg, keys.Down) {
		if m.Menu.Index() < len(m.Menu.Items())-1 {
			m.Menu.CursorDown()
		}
		return m, nil
	}

	if key.Matches(msg, keys.Select) {
		return m.selectMenuItem()
	}

	if key.Matches(msg, keys.Left) {
		return m, nil
	}

	if key.Matches(msg, keys.Right) {
		return m.selectMenuItem()
	}

	return m, nil
}

func (m Model) selectMenuItem() (tea.Model, tea.Cmd) {
	selected := m.Menu.Index()
	if selected < 0 || selected >= len(m.Menu.Items()) {
		return m, nil
	}

	item := m.Menu.Items()[selected]
	menuItem, ok := item.(MenuItem)
	if !ok {
		return m, nil
	}

	switch menuItem.view {
	case ViewCourses:
		m.PreviousView = m.CurrentView
		m.CurrentView = ViewCourses
		m.loadCourses()
	case ViewCoursework:
		m.PreviousView = m.CurrentView
		m.CurrentView = ViewCoursework
		m.loadCoursework()
	case ViewGrades:
		m.PreviousView = m.CurrentView
		m.CurrentView = ViewGrades
		m.loadGrades()
	case ViewAnnouncements:
		m.PreviousView = m.CurrentView
		m.CurrentView = ViewAnnouncements
		m.loadAnnouncements()
	case ViewMainMenu:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleContentKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.CurrentView == ViewCoursework {
		if key.Matches(msg, keys.Up) {
			if m.SelectedCoursework > 0 {
				m.SelectedCoursework--
			}
			m.Viewport.SetContent(m.renderCoursework())
			return m, nil
		}
		if key.Matches(msg, keys.Down) {
			if m.SelectedCoursework < len(m.Coursework)-1 {
				m.SelectedCoursework++
			}
			m.Viewport.SetContent(m.renderCoursework())
			return m, nil
		}
	}

	if key.Matches(msg, keys.Refresh) {
		switch m.CurrentView {
		case ViewCourses:
			m.loadCourses()
		case ViewCoursework:
			m.loadCoursework()
		case ViewGrades:
			m.loadGrades()
		case ViewAnnouncements:
			m.loadAnnouncements()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.CurrentView == ViewMainMenu && msg.Type == tea.MouseLeft {
		menuHeight := m.Height - 6
		itemHeight := 3
		firstItemY := 2

		if msg.Y >= firstItemY && msg.Y < firstItemY+menuHeight {
			clickedIndex := (msg.Y - firstItemY) / itemHeight
			if clickedIndex >= 0 && clickedIndex < len(m.Menu.Items()) {
				m.Menu.Select(clickedIndex)
				return m.selectMenuItem()
			}
		}
	}

	return m, nil
}

func (m *Model) loadCourses() {
	if m.AuthState != AuthAuthenticated {
		m.CurrentView = ViewAuthRequired
		m.ErrorMsg = "Please authenticate first using 'gc-cli auth login'"
		return
	}

	m.IsLoading = true
	m.LoadingMsg = "Loading courses..."

	time.Sleep(500 * time.Millisecond)

	m.Courses = []CourseItem{
		{ID: "course-1", Name: "CS 101: Introduction to Computer Science", Section: "Fall 2024", Desc: "Fundamental concepts of programming", Room: "Building A, Room 101"},
		{ID: "course-2", Name: "MATH 201: Linear Algebra", Section: "Fall 2024", Desc: "Vector spaces, linear transformations", Room: "Building B, Room 205"},
		{ID: "course-3", Name: "PHYS 150: General Physics I", Section: "Fall 2024", Desc: "Mechanics, thermodynamics, waves", Room: "Science Building, Room 302"},
	}

	m.IsLoading = false
	m.updateViewport(m.renderCourses())
}

func (m *Model) loadCoursework() {
	if m.AuthState != AuthAuthenticated {
		m.CurrentView = ViewAuthRequired
		m.ErrorMsg = "Please authenticate first using 'gc-cli auth login'"
		return
	}

	m.IsLoading = true
	m.LoadingMsg = "Loading coursework..."

	time.Sleep(500 * time.Millisecond)

	m.Coursework = []CourseworkItem{
		{ID: "cw-1", CourseID: "course-1", CourseName: "CS 101", AssignTitle: "Programming Assignment 1", Desc: "Implement a basic calculator", State: "PUBLISHED", DueDate: "2024-09-15", DueTime: "23:59", Points: 100, Status: StatusReturned, WorkType: "ASSIGNMENT"},
		{ID: "cw-2", CourseID: "course-1", CourseName: "CS 101", AssignTitle: "Quiz 1: Variables and Data Types", Desc: "Online quiz on data types", State: "PUBLISHED", DueDate: "2024-09-20", DueTime: "23:59", Points: 20, Status: StatusReturned, WorkType: "QUIZ"},
		{ID: "cw-3", CourseID: "course-1", CourseName: "CS 101", AssignTitle: "Programming Assignment 2", Desc: "OOP concepts", State: "PUBLISHED", DueDate: "2024-10-15", DueTime: "23:59", Points: 100, Status: StatusTurnedIn, WorkType: "ASSIGNMENT"},
		{ID: "cw-4", CourseID: "course-2", CourseName: "MATH 201", AssignTitle: "Homework 1: Vectors", Desc: "Problems from Chapter 1", State: "PUBLISHED", DueDate: "2024-09-18", DueTime: "23:59", Points: 50, Status: StatusReturned, WorkType: "ASSIGNMENT"},
		{ID: "cw-5", CourseID: "course-2", CourseName: "MATH 201", AssignTitle: "Homework 2: Matrices", Desc: "Problems from Chapter 2", State: "PUBLISHED", DueDate: "2024-09-25", DueTime: "23:59", Points: 50, Status: StatusTurnedIn, WorkType: "ASSIGNMENT"},
		{ID: "cw-6", CourseID: "course-3", CourseName: "PHYS 150", AssignTitle: "Lab Report 1: Motion", Desc: "Motion experiment writeup", State: "PUBLISHED", DueDate: "2024-09-22", DueTime: "17:00", Points: 50, Status: StatusReturned, WorkType: "ASSIGNMENT"},
		{ID: "cw-7", CourseID: "course-2", CourseName: "MATH 201", AssignTitle: "Midterm Exam", Desc: "Covers chapters 1-3", State: "PUBLISHED", DueDate: "2024-10-01", DueTime: "14:00", Points: 100, Status: StatusOverdue, WorkType: "EXAM"},
		{ID: "cw-8", CourseID: "course-1", CourseName: "CS 101", AssignTitle: "Lab 3: Debugging", Desc: "Debugging practice", State: "DRAFT", DueDate: "", DueTime: "", Points: 25, Status: StatusDraft, WorkType: "ASSIGNMENT"},
	}

	m.SelectedCoursework = 0
	m.sortCourseworkByDueDate()
	m.IsLoading = false
	m.updateViewport(m.renderCoursework())
}

func (m *Model) sortCourseworkByDueDate() {
	sort.SliceStable(m.Coursework, func(i, j int) bool {
		if m.Coursework[i].DueDate == "" && m.Coursework[j].DueDate == "" {
			return false
		}
		if m.Coursework[i].DueDate == "" {
			return false
		}
		if m.Coursework[j].DueDate == "" {
			return true
		}
		return m.Coursework[i].DueDate < m.Coursework[j].DueDate
	})
}

func (m *Model) loadGrades() {
	if m.AuthState != AuthAuthenticated {
		m.CurrentView = ViewAuthRequired
		m.ErrorMsg = "Please authenticate first using 'gc-cli auth login'"
		return
	}

	m.IsLoading = true
	m.LoadingMsg = "Loading grades..."

	time.Sleep(500 * time.Millisecond)

	m.Grades = []GradeItem{
		{CourseName: "CS 101", Assignment: "Programming Assignment 1", Score: "95", MaxScore: "100", SubmittedAt: "2024-09-15"},
		{CourseName: "CS 101", Assignment: "Quiz 1", Score: "18", MaxScore: "20", SubmittedAt: "2024-09-20"},
		{CourseName: "MATH 201", Assignment: "Homework 1", Score: "90", MaxScore: "100", SubmittedAt: "2024-09-18"},
		{CourseName: "MATH 201", Assignment: "Midterm Exam", Score: "82", MaxScore: "100", SubmittedAt: "2024-10-10"},
		{CourseName: "PHYS 150", Assignment: "Lab Report 1", Score: "48", MaxScore: "50", SubmittedAt: "2024-09-22"},
	}

	m.IsLoading = false
	m.updateViewport(m.renderGrades())
}

func (m *Model) loadAnnouncements() {
	if m.AuthState != AuthAuthenticated {
		m.CurrentView = ViewAuthRequired
		m.ErrorMsg = "Please authenticate first using 'gc-cli auth login'"
		return
	}

	m.IsLoading = true
	m.LoadingMsg = "Loading announcements..."

	time.Sleep(500 * time.Millisecond)

	m.Announcements = []AnnouncementItem{
		{CourseName: "CS 101", AnnounceTitle: "Assignment 2 Posted", Text: "The second programming assignment has been posted. Due October 15th.", PostedAt: "2024-10-01"},
		{CourseName: "MATH 201", AnnounceTitle: "Office Hours Change", Text: "Office hours this week will be Thursday 2-4 PM.", PostedAt: "2024-10-02"},
		{CourseName: "PHYS 150", AnnounceTitle: "Lab Safety Reminder", Text: "Please review lab safety procedures before your session.", PostedAt: "2024-09-28"},
		{CourseName: "CS 101", AnnounceTitle: "Guest Lecture Next Week", Text: "Guest speaker from Google next Tuesday.", PostedAt: "2024-10-03"},
	}

	m.IsLoading = false
	m.updateViewport(m.renderAnnouncements())
}

func (m *Model) updateViewport(content string) {
	m.Viewport.SetContent(content)
}

func (m Model) View() string {
	var content string

	switch m.CurrentView {
	case ViewMainMenu:
		content = m.renderMainMenu()

	case ViewCourses:
		if m.IsLoading {
			content = m.renderLoading()
		} else {
			content = m.Viewport.View()
		}

	case ViewCoursework:
		if m.IsLoading {
			content = m.renderLoading()
		} else {
			content = m.Viewport.View()
		}

	case ViewGrades:
		if m.IsLoading {
			content = m.renderLoading()
		} else {
			content = m.Viewport.View()
		}

	case ViewAnnouncements:
		if m.IsLoading {
			content = m.renderLoading()
		} else {
			content = m.Viewport.View()
		}

	case ViewAuthRequired:
		content = m.renderAuthRequired()

	case ViewLoading:
		content = m.renderLoading()

	case ViewError:
		content = m.renderError()
	}

	header := m.renderHeader()
	statusBar := m.renderStatusBar()

	output := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		statusBar,
	)

	return windowStyle.Height(m.Height).Render(output)
}

func (m Model) renderHeader() string {
	var title string

	switch m.CurrentView {
	case ViewMainMenu:
		title = " Google Classroom CLI "
	case ViewCourses:
		title = " Courses "
	case ViewCoursework:
		title = " Assignments "
	case ViewGrades:
		title = " Grades "
	case ViewAnnouncements:
		title = " Announcements "
	case ViewAuthRequired:
		title = " Authentication Required "
	case ViewLoading:
		title = " Loading... "
	case ViewError:
		title = " Error "
	default:
		title = " gc-cli "
	}

	return headerStyle.Width(m.Width - 2).Render(title)
}

func (m Model) renderMainMenu() string {
	menuView := m.Menu.View()

	menuBorder := borderStyle.
		Width(m.Width - 4).
		Height(m.Height - 6).
		Render(menuView)

	return menuBorder
}

func (m Model) renderCourses() string {
	if len(m.Courses) == 0 {
		return contentStyle.Width(m.Width - 4).Height(m.Height - 6).Render(
			"\n\n\n" + lipgloss.NewStyle().
				Foreground(textMuted).
				Align(lipgloss.Center).
				Width(m.Width-8).
				Render("No courses found"),
		)
	}

	var output string
	output += sectionTitleStyle.Width(m.Width-8).Render("Your Courses") + "\n\n"

	for i, course := range m.Courses {
		courseNum := lipgloss.NewStyle().
			Foreground(accentPrimary).
			Bold(true).
			Render(fmt.Sprintf("%d.", i+1))

		courseName := lipgloss.NewStyle().
			Foreground(textPrimary).
			Bold(true).
			Render(course.Name)

		section := lipgloss.NewStyle().
			Foreground(accentTertiary).
			Render(course.Section)

		desc := lipgloss.NewStyle().
			Foreground(textSecondary).
			Render(course.Desc)

		room := lipgloss.NewStyle().
			Foreground(textMuted).
			Render("📍 " + course.Room)

		output += fmt.Sprintf("%s %s (%s)\n%s\n%s\n\n", courseNum, courseName, section, desc, room)
	}

	return contentStyle.Width(m.Width - 4).Render(output)
}

func (m Model) renderCoursework() string {
	if len(m.Coursework) == 0 {
		return contentStyle.Width(m.Width - 4).Height(m.Height - 6).Render(
			"\n\n\n" + lipgloss.NewStyle().
				Foreground(textMuted).
				Align(lipgloss.Center).
				Width(m.Width-8).
				Render("No assignments found"),
		)
	}

	var output string
	output += sectionTitleStyle.Width(m.Width-8).Render("Your Assignments") + "\n\n"

	output += lipgloss.NewStyle().
		Foreground(textMuted).
		Width(m.Width-8).
		Render("✓ RETURNED  ◐ TURNED_IN  ✗ OVERDUE  ○ NEW") + "\n\n"

	for i, cw := range m.Coursework {
		isSelected := i == m.SelectedCoursework

		var itemStyle lipgloss.Style
		if isSelected {
			itemStyle = lipgloss.NewStyle().
				Background(bgHighlight).
				Foreground(textPrimary).
				Padding(1, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accentPrimary).
				Width(m.Width - 8)
		} else {
			itemStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Padding(1, 1).
				Width(m.Width - 8)
		}

		entryNum := lipgloss.NewStyle().
			Foreground(accentPrimary).
			Bold(true).
			Render(fmt.Sprintf("%d.", i+1))

		title := lipgloss.NewStyle().
			Foreground(textPrimary).
			Bold(true).
			Render(cw.Title())

		course := lipgloss.NewStyle().
			Foreground(accentTertiary).
			Render(cw.CourseName)

		var statusColor lipgloss.Color
		var statusIcon string
		switch cw.Status {
		case StatusReturned:
			statusColor = successColor
			statusIcon = "✓"
		case StatusTurnedIn:
			statusColor = warningColor
			statusIcon = "◐"
		case StatusOverdue:
			statusColor = errorColor
			statusIcon = "✗"
		case StatusDraft:
			statusColor = textMuted
			statusIcon = "○"
		default:
			statusColor = textSecondary
			statusIcon = "○"
		}

		status := lipgloss.NewStyle().
			Foreground(statusColor).
			Bold(true).
			Render(fmt.Sprintf("%s %s", statusIcon, cw.StatusString()))

		dueDate := cw.DueDate
		if cw.DueTime != "" {
			dueDate += " " + cw.DueTime
		}
		if dueDate == "" {
			dueDate = "-"
		}

		due := lipgloss.NewStyle().
			Foreground(textSecondary).
			Render("Due: " + dueDate)

		points := lipgloss.NewStyle().
			Foreground(textMuted).
			Render(fmt.Sprintf("%d pts", cw.Points))

		workType := lipgloss.NewStyle().
			Foreground(textMuted).
			Render(cw.WorkType)

		content := fmt.Sprintf("%s %s\n  %s  •  %s  •  %s\n  %s  •  %s",
			entryNum, title, course, status, due, points, workType)

		output += itemStyle.Render(content) + "\n\n"
	}

	return contentStyle.Width(m.Width - 4).Render(output)
}

func (m Model) renderGrades() string {
	if len(m.Grades) == 0 {
		return contentStyle.Width(m.Width - 4).Height(m.Height - 6).Render(
			"\n\n\n" + lipgloss.NewStyle().
				Foreground(textMuted).
				Align(lipgloss.Center).
				Width(m.Width-8).
				Render("No grades found"),
		)
	}

	var output string
	output += sectionTitleStyle.Width(m.Width-8).Render("Your Grades") + "\n\n"

	for i, grade := range m.Grades {
		entryNum := lipgloss.NewStyle().
			Foreground(accentPrimary).
			Bold(true).
			Render(fmt.Sprintf("%d.", i+1))

		assignment := lipgloss.NewStyle().
			Foreground(textPrimary).
			Bold(true).
			Render(grade.Assignment)

		course := lipgloss.NewStyle().
			Foreground(accentTertiary).
			Render(grade.CourseName)

		scoreColor := textPrimary
		if grade.Score == grade.MaxScore {
			scoreColor = successColor
		} else if grade.Score == "0" || grade.Score == "" {
			scoreColor = errorColor
		}

		score := lipgloss.NewStyle().
			Foreground(scoreColor).
			Bold(true).
			Render(fmt.Sprintf("%s/%s", grade.Score, grade.MaxScore))

		submitted := lipgloss.NewStyle().
			Foreground(textMuted).
			Render("Submitted: " + grade.SubmittedAt)

		output += fmt.Sprintf("%s %s\n  %s — %s\n  %s\n\n", entryNum, assignment, course, score, submitted)
	}

	return contentStyle.Width(m.Width - 4).Render(output)
}

func (m Model) renderAnnouncements() string {
	if len(m.Announcements) == 0 {
		return contentStyle.Width(m.Width - 4).Height(m.Height - 6).Render(
			"\n\n\n" + lipgloss.NewStyle().
				Foreground(textMuted).
				Align(lipgloss.Center).
				Width(m.Width-8).
				Render("No announcements found"),
		)
	}

	var output string
	output += sectionTitleStyle.Width(m.Width-8).Render("Course Announcements") + "\n\n"

	for i, ann := range m.Announcements {
		annNum := lipgloss.NewStyle().
			Foreground(accentPrimary).
			Bold(true).
			Render(fmt.Sprintf("%d.", i+1))

		title := lipgloss.NewStyle().
			Foreground(textPrimary).
			Bold(true).
			Render(ann.Title())

		course := lipgloss.NewStyle().
			Foreground(accentTertiary).
			Render(ann.CourseName)

		date := lipgloss.NewStyle().
			Foreground(textMuted).
			Render(ann.PostedAt)

		text := lipgloss.NewStyle().
			Foreground(textSecondary).
			Width(m.Width - 12).
			Render(ann.Text)

		output += fmt.Sprintf("%s %s\n  📚 %s — %s\n\n%s\n\n", annNum, title, course, date, text)
	}

	return contentStyle.Width(m.Width - 4).Render(output)
}

func (m Model) renderLoading() string {
	loadingContent := lipgloss.NewStyle().
		Foreground(accentPrimary).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.Width - 8).
		Render("⟳ " + m.LoadingMsg)

	return lipgloss.Place(
		m.Width-4,
		m.Height-6,
		lipgloss.Center,
		lipgloss.Center,
		loadingStyle.Width(m.Width-4).Height(m.Height-6).Render(loadingContent),
	)
}

func (m Model) renderError() string {
	errorContent := lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.Width - 8).
		Render("⚠ " + m.ErrorMsg)

	return lipgloss.Place(
		m.Width-4,
		m.Height-6,
		lipgloss.Center,
		lipgloss.Center,
		errorStyle.Width(m.Width-4).Height(m.Height-6).Render(errorContent),
	)
}

func (m Model) renderAuthRequired() string {
	title := lipgloss.NewStyle().
		Foreground(accentSecondary).
		Bold(true).
		Width(m.Width - 8).
		Align(lipgloss.Center).
		Render("🔒 Authentication Required")

	message := lipgloss.NewStyle().
		Foreground(textSecondary).
		Width(m.Width - 8).
		Align(lipgloss.Center).
		Render("Please authenticate first using:\n\n  gc-cli auth login\n\nThen run 'gc-cli tui' again.")

	hint := lipgloss.NewStyle().
		Foreground(textMuted).
		Width(m.Width - 8).
		Align(lipgloss.Center).
		Render("Press ESC or ← to go back")

	content := lipgloss.NewStyle().
		Width(m.Width-4).
		Height(m.Height-6).
		Background(bgSecondary).
		Padding(2, 0).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				"\n\n\n",
				title,
				"\n",
				message,
				"\n\n\n",
				hint,
			),
		)

	return content
}

func (m Model) renderStatusBar() string {
	var status string

	switch m.CurrentView {
	case ViewMainMenu:
		status = "↑↓/jk: navigate  •  enter/l: select  •  q: quit"
	case ViewCourses, ViewCoursework, ViewGrades, ViewAnnouncements:
		status = "↑↓/jk: scroll  •  r: refresh  •  esc/q: back"
	case ViewAuthRequired:
		status = "esc: go back"
	default:
		status = "q: quit"
	}

	authStatus := "Not logged in"
	if m.AuthState == AuthAuthenticated {
		authStatus = "✓ Logged in"
	}

	authStyle := statusBarStyle
	if m.AuthState == AuthAuthenticated {
		authStyle = authStyle.Foreground(successColor)
	} else {
		authStyle = authStyle.Foreground(warningColor)
	}

	statusBar := lipgloss.NewStyle().
		Width(m.Width).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				statusBarStyle.Width(m.Width-len(authStatus)-3).Render(status),
				authStyle.Render(authStatus),
			),
		)

	return statusBar
}

func Run(cfg *config.Config) error {
	p := tea.NewProgram(
		New(cfg),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return err
	}

	return nil
}
