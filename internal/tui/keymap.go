package tui

import "github.com/charmbracelet/bubbles/key"

type appKeyMap struct {
	UpDown     key.Binding
	NextTab    key.Binding
	PrevTab    key.Binding
	OpenClass  key.Binding
	Back       key.Binding
	OpenWeb    key.Binding
	Refresh    key.Binding
	ToggleHelp key.Binding
	Quit       key.Binding
}

func newKeyMap() appKeyMap {
	return appKeyMap{
		UpDown: key.NewBinding(
			key.WithKeys("up", "down", "j", "k"),
			key.WithHelp("up/down", "navigate classes"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next view/tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous view/tab"),
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
	return []key.Binding{k.UpDown, k.NextTab, k.OpenClass, k.OpenWeb, k.Refresh, k.Quit}
}

func (k appKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.UpDown, k.NextTab, k.PrevTab, k.OpenClass, k.Back},
		{k.OpenWeb, k.Refresh, k.ToggleHelp, k.Quit},
	}
}
