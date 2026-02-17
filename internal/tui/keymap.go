package tui

import "github.com/charmbracelet/bubbles/key"

type appKeyMap struct {
	UpDown      key.Binding
	NextPane    key.Binding
	PrevPane    key.Binding
	NextViewTab key.Binding
	PrevViewTab key.Binding
	OpenClass   key.Binding
	Back        key.Binding
	OpenWeb     key.Binding
	Refresh     key.Binding
	ToggleHelp  key.Binding
	Quit        key.Binding
}

func newKeyMap() appKeyMap {
	return appKeyMap{
		UpDown: key.NewBinding(
			key.WithKeys("up", "down", "j", "k"),
			key.WithHelp("up/down", "navigate focused pane"),
		),
		NextPane: key.NewBinding(
			key.WithKeys("tab", "right"),
			key.WithHelp("tab/right", "next pane"),
		),
		PrevPane: key.NewBinding(
			key.WithKeys("shift+tab", "left"),
			key.WithHelp("shift+tab/left", "previous pane"),
		),
		NextViewTab: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next view/tab"),
		),
		PrevViewTab: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "previous view/tab"),
		),
		OpenClass: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open class tabs"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back to global"),
		),
		OpenWeb: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh courses"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle full help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k appKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.UpDown, k.NextPane, k.NextViewTab, k.OpenClass, k.OpenWeb, k.Refresh, k.Quit}
}

func (k appKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.UpDown, k.NextPane, k.PrevPane, k.NextViewTab, k.PrevViewTab},
		{k.OpenClass, k.Back, k.OpenWeb, k.Refresh, k.ToggleHelp, k.Quit},
	}
}
