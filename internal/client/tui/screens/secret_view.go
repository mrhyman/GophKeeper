package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	secretsDomain "gophkeeper/internal/client/domain/secrets"
	"gophkeeper/pkg/models"
)

var (
	viewTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	viewBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 3).
			Width(60)

	viewLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			Width(18)

	viewValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	viewHiddenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			Italic(true)

	viewMetaHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				MarginTop(1)

	viewMetaKeyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C6C6C"))

	viewMetaValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	viewActionSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 2)

	viewActionNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C6C6C")).
				Padding(0, 2)

	viewHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			MarginTop(1)

	viewErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4672")).
			MarginTop(1)

	viewConfirmStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF4672")).
				Bold(true).
				MarginTop(1)
)

// viewAction описывает действие на экране просмотра.
type viewAction struct {
	icon  string
	label string
}

// SecretViewModel — модель экрана просмотра секрета.
type SecretViewModel struct {
	secret         *models.Secret
	secretsService *secretsDomain.Service

	// Расшифрованные данные
	loginPassword *models.LoginPassword
	textData      *models.TextData
	binaryData    *models.BinaryData
	bankCard      *models.BankCard

	// Состояние
	revealed      bool // показаны ли скрытые поля (пароль, CVV)
	decryptErr    error
	confirmDelete bool
	actionCursor  int
	actions       []viewAction
}

// NewSecretViewModel создаёт новую модель просмотра секрета.
func NewSecretViewModel(secret *models.Secret, secretsService *secretsDomain.Service) SecretViewModel {
	m := SecretViewModel{
		secret:         secret,
		secretsService: secretsService,
		actions: []viewAction{
			{icon: "👁", label: "Показать/скрыть"},
			{icon: "✏️", label: "Редактировать"},
			{icon: "🗑", label: "Удалить"},
			{icon: "←", label: "Назад"},
		},
	}

	// Расшифровываем данные
	m.decrypt()

	return m
}

// Init инициализирует экран.
func (m SecretViewModel) Init() tea.Cmd {
	return nil
}

// Update обрабатывает события.
func (m SecretViewModel) Update(msg tea.Msg) (SecretViewModel, tea.Cmd) {
	if m.confirmDelete {
		return m.updateConfirmDelete(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// Навигация по действиям
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			m.actionCursor--
			if m.actionCursor < 0 {
				m.actionCursor = len(m.actions) - 1
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			m.actionCursor++
			if m.actionCursor >= len(m.actions) {
				m.actionCursor = 0
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			m.actionCursor++
			if m.actionCursor >= len(m.actions) {
				m.actionCursor = 0
			}

		// Выполнить действие
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m.handleAction()

		// Быстрые клавиши
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			m.revealed = !m.revealed

		case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
			return m.editSecret()

		case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
			m.confirmDelete = true

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			return m.goBack()
		}
	}

	return m, nil
}

// View отрисовывает экран.
func (m SecretViewModel) View() string {
	var b strings.Builder

	// Заголовок
	icon := secretTypeIcon(m.secret.Type)
	b.WriteString(viewTitleStyle.Render(fmt.Sprintf("%s %s", icon, m.secret.Name)))
	b.WriteString("\n\n")

	// Ошибка дешифрования
	if m.decryptErr != nil {
		b.WriteString(viewErrorStyle.Render(
			fmt.Sprintf("✗ Ошибка расшифровки: %s", m.decryptErr.Error()),
		))
		b.WriteString("\n\n")
	} else {
		// Типизированные поля
		b.WriteString(m.renderFields())
		b.WriteString("\n")
	}

	// Системная информация
	b.WriteString(m.renderSystemInfo())
	b.WriteString("\n")

	// Метаданные
	if len(m.secret.Metadata) > 0 {
		b.WriteString(m.renderMetadata())
		b.WriteString("\n")
	}

	// Действия
	b.WriteString("\n")
	b.WriteString(m.renderActions())

	// Подтверждение удаления
	if m.confirmDelete {
		b.WriteString("\n")
		b.WriteString(viewConfirmStyle.Render(
			fmt.Sprintf("Удалить \"%s\"? (y — да, n — нет)", m.secret.Name),
		))
	}

	// Подсказки
	b.WriteString("\n")
	b.WriteString(viewHintStyle.Render(
		"←→/hl: действия • enter: выполнить • r: показать/скрыть • e: редактировать • d: удалить • esc: назад",
	))

	return viewBoxStyle.Render(b.String())
}

// --- Расшифровка ---

func (m *SecretViewModel) decrypt() {
	switch m.secret.Type {
	case models.SecretTypeLoginPassword:
		var lp models.LoginPassword
		if err := m.secretsService.Decrypt(m.secret, &lp); err != nil {
			m.decryptErr = err
			return
		}
		m.loginPassword = &lp

	case models.SecretTypeTextData:
		var td models.TextData
		if err := m.secretsService.Decrypt(m.secret, &td); err != nil {
			m.decryptErr = err
			return
		}
		m.textData = &td

	case models.SecretTypeBinaryData:
		var bd models.BinaryData
		if err := m.secretsService.Decrypt(m.secret, &bd); err != nil {
			m.decryptErr = err
			return
		}
		m.binaryData = &bd

	case models.SecretTypeBankCard:
		var bc models.BankCard
		if err := m.secretsService.Decrypt(m.secret, &bc); err != nil {
			m.decryptErr = err
			return
		}
		m.bankCard = &bc
	}
}

// --- Рендеринг полей ---

func (m SecretViewModel) renderFields() string {
	var b strings.Builder

	switch m.secret.Type {
	case models.SecretTypeLoginPassword:
		if m.loginPassword == nil {
			break
		}
		b.WriteString(m.renderField("Логин:", m.loginPassword.Login, false))
		b.WriteString(m.renderField("Пароль:", m.loginPassword.Password, true))
		if m.loginPassword.URI != "" {
			b.WriteString(m.renderField("URL:", m.loginPassword.URI, false))
		}

	case models.SecretTypeTextData:
		if m.textData == nil {
			break
		}
		b.WriteString(m.renderField("Текст:", m.textData.Text, false))

	case models.SecretTypeBinaryData:
		if m.binaryData == nil {
			break
		}
		b.WriteString(m.renderField("Файл:", m.binaryData.FileName, false))
		size := formatBytes(len(m.binaryData.Data))
		b.WriteString(m.renderField("Размер:", size, false))

	case models.SecretTypeBankCard:
		if m.bankCard == nil {
			break
		}
		b.WriteString(m.renderField("Номер:", m.formatCardNumber(), true))
		b.WriteString(m.renderField("Владелец:", m.bankCard.Holder, false))
		exp := fmt.Sprintf("%02d/%d", m.bankCard.ExpMonth, m.bankCard.ExpYear)
		b.WriteString(m.renderField("Срок:", exp, false))
		b.WriteString(m.renderField("CVV:", m.bankCard.CVV, true))
	}

	return b.String()
}

func (m SecretViewModel) renderField(label, value string, sensitive bool) string {
	l := viewLabelStyle.Render(label)

	var v string
	if sensitive && !m.revealed {
		v = viewHiddenStyle.Render("••••••••")
	} else {
		v = viewValueStyle.Render(value)
	}

	return fmt.Sprintf("%s %s\n", l, v)
}

func (m SecretViewModel) renderSystemInfo() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(viewMetaHeaderStyle.Render("Информация"))
	b.WriteString("\n")

	typeName := secretTypeName(m.secret.Type)
	b.WriteString(m.renderField("Тип:", typeName, false))
	b.WriteString(m.renderField("Версия:", fmt.Sprintf("%d", m.secret.Version), false))
	b.WriteString(m.renderField("Обновлён:", formatTimestamp(m.secret.UpdatedAt), false))
	b.WriteString(m.renderField("ID:", m.secret.ID, false))

	return b.String()
}

func (m SecretViewModel) renderMetadata() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(viewMetaHeaderStyle.Render("Метаданные"))
	b.WriteString("\n")

	for k, v := range m.secret.Metadata {
		key := viewMetaKeyStyle.Render(fmt.Sprintf("  %s:", k))
		val := viewMetaValueStyle.Render(fmt.Sprintf(" %s", v))
		b.WriteString(key + val + "\n")
	}

	return b.String()
}

func (m SecretViewModel) renderActions() string {
	var parts []string

	for i, action := range m.actions {
		label := fmt.Sprintf("%s %s", action.icon, action.label)
		if i == m.actionCursor {
			parts = append(parts, viewActionSelectedStyle.Render(label))
		} else {
			parts = append(parts, viewActionNormalStyle.Render(label))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}

// --- Действия ---

func (m SecretViewModel) handleAction() (SecretViewModel, tea.Cmd) {
	switch m.actionCursor {
	case 0: // Показать/скрыть
		m.revealed = !m.revealed
		return m, nil
	case 1: // Редактировать
		return m.editSecret()
	case 2: // Удалить
		m.confirmDelete = true
		return m, nil
	case 3: // Назад
		return m.goBack()
	}
	return m, nil
}

func (m SecretViewModel) editSecret() (SecretViewModel, tea.Cmd) {
	return m, func() tea.Msg {
		return NavigateMsg{
			Screen: 4, // ScreenSecretEdit
			Secret: m.secret,
		}
	}
}

func (m SecretViewModel) goBack() (SecretViewModel, tea.Cmd) {
	return m, func() tea.Msg {
		return NavigateMsg{Screen: 2} // ScreenSecretList
	}
}

func (m SecretViewModel) updateConfirmDelete(msg tea.Msg) (SecretViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "д", "Д":
			m.confirmDelete = false
			secretID := m.secret.ID
			return m, func() tea.Msg {
				return DeleteSecretMsg{SecretID: secretID}
			}
		case "n", "N", "н", "Н", "esc":
			m.confirmDelete = false
		}
	}
	return m, nil
}

// --- Форматирование ---

func (m SecretViewModel) formatCardNumber() string {
	if m.bankCard == nil {
		return ""
	}

	num := m.bankCard.Number
	if !m.revealed && len(num) >= 4 {
		return "•••• •••• •••• " + num[len(num)-4:]
	}

	// Форматируем по 4 цифры
	var parts []string
	for i := 0; i < len(num); i += 4 {
		end := i + 4
		if end > len(num) {
			end = len(num)
		}
		parts = append(parts, num[i:end])
	}
	return strings.Join(parts, " ")
}

func secretTypeName(t models.SecretType) string {
	switch t {
	case models.SecretTypeLoginPassword:
		return "Логин/Пароль"
	case models.SecretTypeTextData:
		return "Текстовые данные"
	case models.SecretTypeBinaryData:
		return "Бинарные данные"
	case models.SecretTypeBankCard:
		return "Банковская карта"
	default:
		return "Неизвестный"
	}
}

func formatBytes(n int) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
