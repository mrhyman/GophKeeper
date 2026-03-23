package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gophkeeper/pkg/models"
)

var (
	listTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	listBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(60)

	listSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	listNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	listMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			MarginLeft(6)

	listEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			Italic(true).
			MarginTop(1).
			MarginBottom(1)

	listSearchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	listCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C"))

	listHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			MarginTop(1)

	listConfirmDeleteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF4672")).
				Bold(true).
				MarginTop(1)
)

// SecretListModel — модель экрана списка секретов.
type SecretListModel struct {
	allSecrets    []*models.Secret
	filtered      []*models.Secret
	filter        models.SecretType
	cursor        int
	search        textinput.Model
	searchActive  bool
	confirmDelete bool
	pageSize      int
	pageOffset    int
}

// NewSecretListModel создаёт новую модель списка секретов.
func NewSecretListModel(secrets []*models.Secret, filter models.SecretType) SecretListModel {
	search := textinput.New()
	search.Placeholder = "поиск по имени..."
	search.CharLimit = 64
	search.Width = 40

	m := SecretListModel{
		allSecrets: secrets,
		filter:     filter,
		search:     search,
		pageSize:   15,
	}

	m.applyFilters()

	return m
}

// Init инициализирует экран.
func (m SecretListModel) Init() tea.Cmd {
	return nil
}

// Update обрабатывает события.
func (m SecretListModel) Update(msg tea.Msg) (SecretListModel, tea.Cmd) {
	// Если активен поиск — обрабатываем ввод
	if m.searchActive {
		return m.updateSearch(msg)
	}

	// Если ожидаем подтверждения удаления
	if m.confirmDelete {
		return m.updateConfirmDelete(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {

		// Навигация
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.moveCursorUp()

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.moveCursorDown()

		// Пагинация
		case key.Matches(msg, key.NewBinding(key.WithKeys("pgup", "ctrl+u"))):
			m.pageUp()

		case key.Matches(msg, key.NewBinding(key.WithKeys("pgdown", "ctrl+d"))):
			m.pageDown()

		// Home / End
		case key.Matches(msg, key.NewBinding(key.WithKeys("home", "g"))):
			m.cursor = 0
			m.pageOffset = 0

		case key.Matches(msg, key.NewBinding(key.WithKeys("end", "G"))):
			m.cursor = len(m.filtered) - 1
			m.adjustPageOffset()

		// Открыть секрет
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m.openSelected()

		// Создать новый
		case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
			return m.createNew()

		// Удалить
		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			if len(m.filtered) > 0 {
				m.confirmDelete = true
			}

		// Поиск
		case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
			m.searchActive = true
			m.search.SetValue("")
			return m, m.search.Focus()

		// Назад
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg {
				return NavigateMsg{Screen: 1} // ScreenMainMenu
			}

		// Выход
		case key.Matches(msg, key.NewBinding(key.WithKeys("q"))):
			return m, func() tea.Msg {
				return NavigateMsg{Screen: 1}
			}
		}
	}

	return m, nil
}

// View отрисовывает экран.
func (m SecretListModel) View() string {
	var b strings.Builder

	// Заголовок
	title := m.titleText()
	b.WriteString(listTitleStyle.Render(title))
	b.WriteString("  ")
	b.WriteString(listCountStyle.Render(fmt.Sprintf("(%d)", len(m.filtered))))
	b.WriteString("\n\n")

	// Строка поиска
	if m.searchActive {
		b.WriteString(listSearchStyle.Render("🔍 "))
		b.WriteString(m.search.View())
		b.WriteString("\n\n")
	} else if m.search.Value() != "" {
		b.WriteString(listSearchStyle.Render(fmt.Sprintf("🔍 \"%s\"", m.search.Value())))
		b.WriteString("  ")
		b.WriteString(listMetaStyle.Render("(/ — изменить, esc — сбросить)"))
		b.WriteString("\n\n")
	}

	// Список
	if len(m.filtered) == 0 {
		b.WriteString(listEmptyStyle.Render("  Нет записей. Нажмите n для создания."))
		b.WriteString("\n")
	} else {
		end := m.pageOffset + m.pageSize
		if end > len(m.filtered) {
			end = len(m.filtered)
		}

		visible := m.filtered[m.pageOffset:end]

		for i, secret := range visible {
			globalIdx := m.pageOffset + i
			b.WriteString(m.renderItem(secret, globalIdx))
			b.WriteString("\n")

			// Метаинформация для выбранного
			if globalIdx == m.cursor {
				meta := m.renderMeta(secret)
				if meta != "" {
					b.WriteString(listMetaStyle.Render(meta))
					b.WriteString("\n")
				}
			}
		}

		// Индикатор страниц
		if len(m.filtered) > m.pageSize {
			page := m.pageOffset/m.pageSize + 1
			totalPages := (len(m.filtered)-1)/m.pageSize + 1
			b.WriteString("\n")
			b.WriteString(listCountStyle.Render(
				fmt.Sprintf("  страница %d/%d", page, totalPages),
			))
			b.WriteString("\n")
		}
	}

	// Подтверждение удаления
	if m.confirmDelete && len(m.filtered) > 0 {
		secret := m.filtered[m.cursor]
		b.WriteString("\n")
		b.WriteString(listConfirmDeleteStyle.Render(
			fmt.Sprintf("  Удалить \"%s\"? (y — да, n — нет)", secret.Name),
		))
	}

	// Подсказки
	b.WriteString("\n")
	b.WriteString(listHintStyle.Render(
		"↑↓/jk: навигация • enter: открыть • n: создать • d: удалить • /: поиск • esc: назад",
	))

	return listBoxStyle.Render(b.String())
}

func (m SecretListModel) GetFilter() models.SecretType {
	return m.filter
}

// --- Навигация ---

func (m *SecretListModel) moveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.pageOffset {
			m.pageOffset = m.cursor
		}
	}
}

func (m *SecretListModel) moveCursorDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
		m.adjustPageOffset()
	}
}

func (m *SecretListModel) pageUp() {
	m.cursor -= m.pageSize
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.pageOffset -= m.pageSize
	if m.pageOffset < 0 {
		m.pageOffset = 0
	}
}

func (m *SecretListModel) pageDown() {
	m.cursor += m.pageSize
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	m.adjustPageOffset()
}

func (m *SecretListModel) adjustPageOffset() {
	if m.cursor >= m.pageOffset+m.pageSize {
		m.pageOffset = m.cursor - m.pageSize + 1
	}
}

// --- Фильтрация и поиск ---

func (m *SecretListModel) applyFilters() {
	m.filtered = make([]*models.Secret, 0)

	query := strings.ToLower(strings.TrimSpace(m.search.Value()))

	for _, s := range m.allSecrets {
		// Фильтр по типу
		if m.filter != 0 && s.Type != m.filter {
			continue
		}

		// Фильтр по поисковому запросу
		if query != "" && !strings.Contains(strings.ToLower(s.Name), query) {
			continue
		}

		m.filtered = append(m.filtered, s)
	}

	// Сброс курсора если вышел за границы
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.pageOffset = 0
}

// --- Обработка поиска ---

func (m SecretListModel) updateSearch(msg tea.Msg) (SecretListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.searchActive = false
			m.search.Blur()
			m.applyFilters()
			return m, nil

		case "esc":
			m.searchActive = false
			m.search.Blur()
			m.search.SetValue("")
			m.applyFilters()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)

	// Живой поиск при вводе
	m.applyFilters()

	return m, cmd
}

// --- Подтверждение удаления ---

func (m SecretListModel) updateConfirmDelete(msg tea.Msg) (SecretListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "д", "Д":
			m.confirmDelete = false
			if m.cursor < len(m.filtered) {
				secret := m.filtered[m.cursor]
				return m, func() tea.Msg {
					return deleteSecretMsg{SecretID: secret.ID}
				}
			}

		case "n", "N", "н", "Н", "esc":
			m.confirmDelete = false
		}
	}

	return m, nil
}

// deleteSecretMsg — внутреннее сообщение для удаления секрета.
type deleteSecretMsg struct {
	SecretID string
}

// --- Действия ---

func (m SecretListModel) openSelected() (SecretListModel, tea.Cmd) {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return m, nil
	}

	secret := m.filtered[m.cursor]
	return m, func() tea.Msg {
		return NavigateMsg{
			Screen: 3, // ScreenSecretView
			Secret: secret,
		}
	}
}

func (m SecretListModel) createNew() (SecretListModel, tea.Cmd) {
	secretType := m.filter
	if secretType == 0 {
		secretType = models.SecretTypeLoginPassword // по умолчанию
	}

	return m, func() tea.Msg {
		return NavigateMsg{
			Screen:     4, // ScreenSecretEdit
			SecretType: secretType,
		}
	}
}

// --- Рендеринг ---

func (m SecretListModel) renderItem(secret *models.Secret, index int) string {
	cursor := "  "
	style := listNormalStyle

	if index == m.cursor {
		cursor = "▸ "
		style = listSelectedStyle
	}

	icon := secretTypeIcon(secret.Type)
	updated := formatTimestamp(secret.UpdatedAt)

	line := fmt.Sprintf("%s%s %s", cursor, icon, secret.Name)
	timestamp := listCountStyle.Render(fmt.Sprintf("  %s  v%d", updated, secret.Version))

	return style.Render(line) + timestamp
}

func (m SecretListModel) renderMeta(secret *models.Secret) string {
	if len(secret.Metadata) == 0 {
		return ""
	}

	var parts []string
	for k, v := range secret.Metadata {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v))
	}

	return strings.Join(parts, " • ")
}

func (m SecretListModel) titleText() string {
	switch m.filter {
	case models.SecretTypeLoginPassword:
		return "🔑 Пароли"
	case models.SecretTypeTextData:
		return "📝 Заметки"
	case models.SecretTypeBinaryData:
		return "📦 Файлы"
	case models.SecretTypeBankCard:
		return "💳 Банковские карты"
	default:
		return "📋 Все записи"
	}
}

// --- Утилиты ---

func secretTypeIcon(t models.SecretType) string {
	switch t {
	case models.SecretTypeLoginPassword:
		return "🔑"
	case models.SecretTypeTextData:
		return "📝"
	case models.SecretTypeBinaryData:
		return "📦"
	case models.SecretTypeBankCard:
		return "💳"
	default:
		return "❓"
	}
}

func formatTimestamp(ts int64) string {
	if ts == 0 {
		return "—"
	}

	t := time.Unix(ts, 0)
	now := time.Now()

	// Сегодня — показываем время
	if t.YearDay() == now.YearDay() && t.Year() == now.Year() {
		return t.Format("15:04")
	}

	// В этом году — день и месяц
	if t.Year() == now.Year() {
		return t.Format("02 Jan")
	}

	// Другой год
	return t.Format("02.01.2006")
}
