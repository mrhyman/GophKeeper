package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Цвета
	colorSuccess = lipgloss.Color("#04B575")
	colorError   = lipgloss.Color("#FF4672")

	// Стили элементов
	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Строка статуса
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(lipgloss.Color("#AAAAAA")).
			Padding(0, 1)
)
