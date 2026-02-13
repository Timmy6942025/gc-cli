package tui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/timboy697/gc-cli/internal/api"
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
	ViewCoursePicker
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

	SelectedCoursework   int
	SelectedGrade        int
	SelectedAnnouncement int

	SelectedCourseID   string
	SelectedCourseName string
	CoursePickerIndex  int

	APIClient *api.Client

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

func New(cfg *config.Config, client *api.Client) Model {
	menuItems := []MenuItem{
		{"Classes", "View your enrolled classes", ViewCourses},
		{"Classwork", "View assignments and deadlines", ViewCoursework},
		{"Grades", "Check your grades and scores", ViewGrades},
		{"Announcements", "Class announcements", ViewAnnouncements},
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
		APIClient:    client,
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

	case ViewCoursePicker:
		return m.handleCoursePickerKey(msg)

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

func (m Model) handleCoursePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Up) {
		if m.CoursePickerIndex > 0 {
			m.CoursePickerIndex--
		}
		m.Viewport.SetContent(m.renderCoursePicker())
		return m, nil
	}

	if key.Matches(msg, keys.Down) {
		if m.CoursePickerIndex < len(m.Courses)-1 {
			m.CoursePickerIndex++
		}
		m.Viewport.SetContent(m.renderCoursePicker())
		return m, nil
	}

	if key.Matches(msg, keys.Select) {
		m.handleCoursePickerSelect()
		return m, nil
	}

	return m, nil
}

func (m *Model) handleCoursePickerSelect() {
	if m.CoursePickerIndex < 0 || m.CoursePickerIndex >= len(m.Courses) {
		return
	}

	m.SelectedCourseID = m.Courses[m.CoursePickerIndex].ID
	m.SelectedCourseName = m.Courses[m.CoursePickerIndex].Name

	switch m.PreviousView {
	case ViewCoursework:
		m.CurrentView = ViewCoursework
		m.loadCoursework()
	case ViewGrades:
		m.CurrentView = ViewGrades
		m.loadGrades()
	case ViewAnnouncements:
		m.CurrentView = ViewAnnouncements
		m.loadAnnouncements()
	default:
		m.CurrentView = ViewMainMenu
	}
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
	if m.CurrentView == ViewCoursework || m.CurrentView == ViewGrades || m.CurrentView == ViewAnnouncements {
		if key.Matches(msg, keys.Up) {
			m.scrollUp()
			return m, nil
		}
		if key.Matches(msg, keys.Down) {
			m.scrollDown()
			return m, nil
		}
	}

	if m.CurrentView == ViewCoursePicker {
		if key.Matches(msg, keys.Up) {
			if m.CoursePickerIndex > 0 {
				m.CoursePickerIndex--
			}
			m.updateViewport(m.renderCoursePicker())
			return m, nil
		}
		if key.Matches(msg, keys.Down) {
			if m.CoursePickerIndex < len(m.Courses)-1 {
				m.CoursePickerIndex++
			}
			m.updateViewport(m.renderCoursePicker())
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

func (m *Model) scrollUp() {
	switch m.CurrentView {
	case ViewCoursework:
		if m.SelectedCoursework > 0 {
			m.SelectedCoursework--
			m.Viewport.SetContent(m.renderCoursework())
		}
	case ViewGrades:
		if m.SelectedGrade > 0 {
			m.SelectedGrade--
			m.Viewport.SetContent(m.renderGrades())
		}
	case ViewAnnouncements:
		if m.SelectedAnnouncement > 0 {
			m.SelectedAnnouncement--
			m.Viewport.SetContent(m.renderAnnouncements())
		}
	}
}

func (m *Model) scrollDown() {
	switch m.CurrentView {
	case ViewCoursework:
		if m.SelectedCoursework < len(m.Coursework)-1 {
			m.SelectedCoursework++
			m.Viewport.SetContent(m.renderCoursework())
		}
	case ViewGrades:
		if m.SelectedGrade < len(m.Grades)-1 {
			m.SelectedGrade++
			m.Viewport.SetContent(m.renderGrades())
		}
	case ViewAnnouncements:
		if m.SelectedAnnouncement < len(m.Announcements)-1 {
			m.SelectedAnnouncement++
			m.Viewport.SetContent(m.renderAnnouncements())
		}
	}
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
	m.LoadingMsg = "Loading classes..."

	// Use real API
	courses, _, err := m.APIClient.ListCourses(context.Background(), 100)
	if err != nil {
		m.IsLoading = false
		m.ErrorMsg = fmt.Sprintf("Failed to load classes: %v", err)
		m.CurrentView = ViewError
		m.updateViewport(m.renderError())
		return
	}

	m.Courses = make([]CourseItem, 0, len(courses))
	for _, c := range courses {
		if c.CourseState == "ACTIVE" {
			m.Courses = append(m.Courses, CourseItem{
				ID:      c.ID,
				Name:    c.Name,
				Section: c.Section,
				Desc:    c.Description,
				Room:    c.Room,
			})
		}
	}

	m.IsLoading = false
	m.updateViewport(m.renderCourses())
}

func (m *Model) loadCoursesForPicker() {
	m.IsLoading = true
	m.LoadingMsg = "Loading classes..."

	courses, _, err := m.APIClient.ListCourses(context.Background(), 100)
	if err != nil {
		m.IsLoading = false
		m.ErrorMsg = fmt.Sprintf("Failed to load classes: %v", err)
		m.CurrentView = ViewError
		m.updateViewport(m.renderError())
		return
	}

	m.Courses = make([]CourseItem, 0, len(courses))
	for _, c := range courses {
		if c.CourseState == "ACTIVE" {
			m.Courses = append(m.Courses, CourseItem{
				ID:      c.ID,
				Name:    c.Name,
				Section: c.Section,
				Desc:    c.Description,
				Room:    c.Room,
			})
		}
	}

	m.IsLoading = false
	m.updateViewport(m.renderCoursePicker())
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

func (m *Model) loadCoursework() {
	if m.AuthState != AuthAuthenticated {
		m.CurrentView = ViewAuthRequired
		m.ErrorMsg = "Please authenticate first using 'gc-cli auth login'"
		return
	}

	if m.SelectedCourseID == "" {
		m.CurrentView = ViewCoursePicker
		m.loadCoursesForPicker()
		return
	}

	m.IsLoading = true
	m.LoadingMsg = "Loading classwork..."

	coursework, _, err := m.APIClient.ListCourseWork(context.Background(), m.SelectedCourseID, 100)
	if err != nil {
		m.IsLoading = false
		m.ErrorMsg = fmt.Sprintf("Failed to load classwork: %v", err)
		m.CurrentView = ViewError
		m.updateViewport(m.renderError())
		return
	}

	m.Coursework = make([]CourseworkItem, 0, len(coursework))
	for _, cw := range coursework {
		if cw.State == "PUBLISHED" {
			status := StatusPending
			if cw.DueDate != nil {
				dueDate := time.Date(cw.DueDate.Year, time.Month(cw.DueDate.Month), cw.DueDate.Day, 23, 59, 59, 0, time.UTC)
				if time.Now().After(dueDate) {
					status = StatusOverdue
				}
			}

			var dueDateStr, dueTimeStr string
			if cw.DueDate != nil {
				dueDateStr = fmt.Sprintf("%d/%02d/%02d", cw.DueDate.Year, cw.DueDate.Month, cw.DueDate.Day)
			}
			if cw.DueTime != nil {
				dueTimeStr = fmt.Sprintf("%02d:%02d", cw.DueTime.Hours, cw.DueTime.Minutes)
			}

			m.Coursework = append(m.Coursework, CourseworkItem{
				ID:          cw.ID,
				CourseID:    m.SelectedCourseID,
				CourseName:  m.SelectedCourseName,
				AssignTitle: cw.Title,
				Desc:        cw.Description,
				State:       cw.State,
				DueDate:     dueDateStr,
				DueTime:     dueTimeStr,
				Points:      cw.MaxPoints,
				Status:      status,
				WorkType:    cw.WorkType,
			})
		}
	}

	m.SelectedCoursework = 0
	m.sortCourseworkByDueDate()
	m.IsLoading = false
	m.updateViewport(m.renderCoursework())
}

func (m *Model) loadGrades() {
	if m.AuthState != AuthAuthenticated {
		m.CurrentView = ViewAuthRequired
		m.ErrorMsg = "Please authenticate first using 'gc-cli auth login'"
		return
	}

	if m.SelectedCourseID == "" {
		m.CurrentView = ViewCoursePicker
		m.loadCoursesForPicker()
		return
	}

	m.IsLoading = true
	m.LoadingMsg = "Loading grades..."

	coursework, _, err := m.APIClient.ListCourseWork(context.Background(), m.SelectedCourseID, 100)
	if err != nil {
		m.IsLoading = false
		m.ErrorMsg = fmt.Sprintf("Failed to load grades: %v", err)
		m.CurrentView = ViewError
		m.updateViewport(m.renderError())
		return
	}

	m.Grades = make([]GradeItem, 0)
	for _, cw := range coursework {
		if cw.State != "PUBLISHED" {
			continue
		}
		submission, err := m.APIClient.GetMySubmission(context.Background(), m.SelectedCourseID, cw.ID)
		if err != nil {
			continue
		}

		if submission.AssignedGrade > 0 || submission.DraftGrade > 0 {
			grade := submission.AssignedGrade
			if grade == 0 && submission.DraftGrade > 0 {
				grade = submission.DraftGrade
			}

			m.Grades = append(m.Grades, GradeItem{
				CourseName:  m.SelectedCourseName,
				Assignment:  cw.Title,
				Score:       fmt.Sprintf("%.1f", grade),
				MaxScore:    fmt.Sprintf("%d", cw.MaxPoints),
				SubmittedAt: submission.SubmittedTimestamp.Format("2006-01-02"),
			})
		}
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

	if m.SelectedCourseID == "" {
		m.CurrentView = ViewCoursePicker
		m.loadCoursesForPicker()
		return
	}

	m.IsLoading = true
	m.LoadingMsg = "Loading announcements..."

	announcements, _, err := m.APIClient.ListAnnouncements(context.Background(), m.SelectedCourseID, 100)
	if err != nil {
		m.IsLoading = false
		m.ErrorMsg = fmt.Sprintf("Failed to load announcements: %v", err)
		m.CurrentView = ViewError
		m.updateViewport(m.renderError())
		return
	}

	m.Announcements = make([]AnnouncementItem, 0, len(announcements))
	for _, a := range announcements {
		m.Announcements = append(m.Announcements, AnnouncementItem{
			CourseName:    m.SelectedCourseName,
			AnnounceTitle: a.Text,
			Text:          a.Text,
			PostedAt:      a.CreationTime.Format("2006-01-02"),
		})
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
				Render("No classes found"),
		)
	}

	var output string
	output += sectionTitleStyle.Width(m.Width-8).Render("Your Classes") + "\n\n"

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

func (m Model) renderCoursePicker() string {
	if len(m.Courses) == 0 {
		return contentStyle.Width(m.Width - 4).Height(m.Height - 6).Render(
			"\n\n\n" + lipgloss.NewStyle().
				Foreground(textMuted).
				Align(lipgloss.Center).
				Width(m.Width-8).
				Render("No classes found"),
		)
	}

	var output string
	output += sectionTitleStyle.Width(m.Width-8).Render("Select a Class") + "\n\n"

	for i, course := range m.Courses {
		isSelected := i == m.CoursePickerIndex

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

		content := fmt.Sprintf("%s %s (%s)", courseNum, courseName, section)

		output += itemStyle.Render(content) + "\n\n"
	}

	hint := lipgloss.NewStyle().
		Foreground(textMuted).
		Width(m.Width - 8).
		Render("↑↓: select  •  enter: confirm  •  esc: back")

	output += "\n" + hint

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
	output += sectionTitleStyle.Width(m.Width-8).Render("Your Classwork") + "\n\n"

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
	case ViewCoursePicker:
		status = "↑↓/jk: select  •  enter: confirm  •  esc: back"
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
	token, err := auth.TokenFromFile(cfg.Auth.TokenFile)
	if err != nil {
		return fmt.Errorf("not authenticated: run 'gc-cli auth login' first")
	}

	authCfg := auth.NewConfig(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.TokenFile)
	client, err := api.NewClientFromToken(context.Background(), authCfg.OAuth2Config(), token)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	p := tea.NewProgram(
		New(cfg, client),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return err
	}

	return nil
}
