package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gophkeeper/pkg/models"
)

var (
	menuTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	menuBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 3).
			Width(50)

	menuSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	menuNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	menuDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			MarginLeft(4)

	menuHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			MarginTop(1)
)

// menuItem описывает пункт главного меню.
type menuItem struct {
	icon        string
	title       string
	description string
	action      func() tea.Msg
}

// MainMenuModel — модель экрана главного меню.
type MainMenuModel struct {
	items  []menuItem
	cursor int
}

// NewMainMenuModel создаёт новую модель главного меню.
func NewMainMenuModel() MainMenuModel {
	items := []menuItem{
		{
			icon:        "🔑",
			title:       "Пароли",
			description: "Логины и пароли",
			action: func() tea.Msg {
				return NavigateMsg{
					Screen: 2, // ScreenSecretList
					Filter: models.SecretTypeLoginPassword,
				}
			},
		},
		{
			icon:        "📝",
			title:       "Заметки",
			description: "Текстовые данные",
			action: func() tea.Msg {
				return NavigateMsg{
					Screen: 2,
					Filter: models.SecretTypeTextData,
				}
			},
		},
		{
			icon:        "📦",
			title:       "Файлы",
			description: "Бинарные данные",
			action: func() tea.Msg {
				return NavigateMsg{
					Screen: 2,
					Filter: models.SecretTypeBinaryData,
				}
			},
		},
		{
			icon:        "💳",
			title:       "Банковские карты",
			description: "Данные карт",
			action: func() tea.Msg {
				return NavigateMsg{
					Screen: 2,
					Filter: models.SecretTypeBankCard,
				}
			},
		},
		{
			icon:        "📋",
			title:       "Все записи",
			description: "Показать все секреты",
			action: func() tea.Msg {
				return NavigateMsg{
					Screen: 2,
					Filter: 0, // без фильтра
				}
			},
		},
		{
			icon:        "🔄",
			title:       "Синхронизация",
			description: "Синхронизировать с сервером",
			action: func() tea.Msg {
				return NavigateMsg{
					Screen: 5, // ScreenSyncStatus
				}
			},
		},
		{
			icon:        "🚪",
			title:       "Выход",
			description: "Выйти из аккаунта",
			action: func() tea.Msg {
				return LogoutMsg{
					Reason:  LogoutReasonUser,
					Message: "Вы вышли из аккаунта",
				}
			},
		},
	}

	return MainMenuModel{
		items:  items,
		cursor: 0,
	}
}

// Init инициализирует экран.
func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

// Update обрабатывает события.
func (m MainMenuModel) Update(msg tea.Msg) (MainMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.items) - 1
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m.handleSelect()

		// Быстрые клавиши
		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			m.cursor = 0
			return m.handleSelect()
		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			m.cursor = 1
			return m.handleSelect()
		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			m.cursor = 2
			return m.handleSelect()
		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			m.cursor = 3
			return m.handleSelect()
		case key.Matches(msg, key.NewBinding(key.WithKeys("5"))):
			m.cursor = 4
			return m.handleSelect()
		case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
			m.cursor = 5
			return m.handleSelect()
		}
	}

	return m, nil
}

// View отрисовывает экран.
func (m MainMenuModel) View() string {
	var b strings.Builder

	b.WriteString(menuTitleStyle.Render("🔐 GophKeeper"))
	b.WriteString("\n\n")

	for i, item := range m.items {
		// Разделитель перед "Синхронизация"
		if i == 5 {
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333333")).
				Render("  ──────────────────────────────"))
			b.WriteString("\n\n")
		}

		cursor := "  "
		style := menuNormalStyle

		if i == m.cursor {
			cursor = "▸ "
			style = menuSelectedStyle
		}

		// Номер для быстрого доступа
		shortcut := ""
		if i < 5 {
			shortcut = fmt.Sprintf("[%d] ", i+1)
		} else if i == 5 {
			shortcut = "[s] "
		}

		line := fmt.Sprintf("%s%s%s %s", cursor, shortcut, item.icon, item.title)
		b.WriteString(style.Render(line))
		b.WriteString("\n")

		// Описание для выбранного пункта
		if i == m.cursor {
			b.WriteString(menuDescStyle.Render(item.description))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(menuHintStyle.Render("↑↓/jk: навигация • enter/1-5: выбрать • s: синхронизация • q: выход"))

	return menuBoxStyle.Render(b.String())
}

// handleSelect обрабатывает выбор пункта меню.
func (m MainMenuModel) handleSelect() (MainMenuModel, tea.Cmd) {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := m.items[m.cursor]
	return m, func() tea.Msg {
		return item.action()
	}
}
