package tui

import "github.com/charmbracelet/lipgloss"

type uiStyles struct {
	title        lipgloss.Style
	panel        lipgloss.Style
	panelFocused lipgloss.Style
	muted        lipgloss.Style
	status       lipgloss.Style
	command      lipgloss.Style
	modal        lipgloss.Style
}

func defaultStyles() uiStyles {
	basePanel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	return uiStyles{
		title:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		panel:        basePanel.BorderForeground(lipgloss.Color("238")),
		panelFocused: basePanel.BorderForeground(lipgloss.Color("45")),
		muted:        lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		command: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("24")).
			Padding(0, 1),
		modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("81")).
			Background(lipgloss.Color("235")).
			Padding(0, 1),
	}
}
