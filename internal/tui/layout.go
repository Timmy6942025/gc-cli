package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
)

func (m *model) resize(width, height int) {
	m.width = width
	m.height = height

	left := max(20, min(28, width/5))
	middle := max(32, min(48, width/3))
	right := width - left - middle - 8
	if right < 30 {
		right = 30
	}

	m.leftWidth = left
	m.middleWidth = middle
	m.rightWidth = right

	listHeight := max(10, height-10)
	m.globalList.SetSize(max(16, left-4), listHeight)
	m.courseList.SetSize(max(24, middle-4), listHeight)
	m.classTabs.SetSize(max(24, right-4), 6)

	m.courseInfo.SetColumns([]table.Column{
		{Title: "Field", Width: 14},
		{Title: "Value", Width: max(18, right-20)},
	})
	m.courseInfo.SetHeight(5)

	m.content.Width = max(24, right-6)
	m.content.Height = max(6, listHeight-18)
}

func cycleList(l *list.Model, step int) {
	total := len(l.Items())
	if total == 0 {
		return
	}
	idx := l.Index() + step
	for idx < 0 {
		idx += total
	}
	idx = idx % total
	l.Select(idx)
}

func fallback(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
