package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Цвета
	colorPrimary   = lipgloss.Color("#7D56F4")
	colorSecondary = lipgloss.Color("#6C6C6C")
	colorSuccess   = lipgloss.Color("#04B575")
	colorError     = lipgloss.Color("#FF4672")
	colorWarning   = lipgloss.Color("#FFB627")

	// Стили элементов
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	// Контейнеры
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	// Элементы списка
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Строка статуса
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(lipgloss.Color("#AAAAAA")).
			Padding(0, 1)

	// Поля ввода
	focusedInputStyle = lipgloss.NewStyle().
				Foreground(colorPrimary)

	blurredInputStyle = lipgloss.NewStyle().
				Foreground(colorSecondary)

	// Кнопки
	focusedButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(colorPrimary).
				Padding(0, 2)

	blurredButtonStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Padding(0, 2)
)
