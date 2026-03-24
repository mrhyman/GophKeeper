package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	authDomain "gophkeeper/internal/client/domain/auth"
)

// Индексы полей и кнопок.
const (
	loginFieldLogin = iota
	loginFieldPassword
	loginFieldMasterPassword
	loginButtonSubmit
	loginButtonToggleMode
)

const loginTotalFields = 5

var (
	loginTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	loginBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 3).
			Width(50)

	loginFocusedButton = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 2).
				MarginTop(1)

	loginBlurredButton = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C6C6C")).
				Padding(0, 2).
				MarginTop(1)

	loginErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4672")).
			MarginTop(1)

	loginHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			MarginTop(1)
)

// LoginModel — модель экрана входа/регистрации.
type LoginModel struct {
	inputs     []textinput.Model
	focusIndex int
	isRegister bool
	loading    bool
	err        error
}

// NewLoginModel создаёт новую модель экрана входа.
func NewLoginModel() LoginModel {
	inputs := make([]textinput.Model, 3)

	// Login
	inputs[loginFieldLogin] = textinput.New()
	inputs[loginFieldLogin].Placeholder = "логин"
	inputs[loginFieldLogin].CharLimit = 64
	inputs[loginFieldLogin].Width = 30
	inputs[loginFieldLogin].Focus()

	// Password
	inputs[loginFieldPassword] = textinput.New()
	inputs[loginFieldPassword].Placeholder = "пароль"
	inputs[loginFieldPassword].CharLimit = 128
	inputs[loginFieldPassword].Width = 30
	inputs[loginFieldPassword].EchoMode = textinput.EchoPassword
	inputs[loginFieldPassword].EchoCharacter = '•'

	// Master password
	inputs[loginFieldMasterPassword] = textinput.New()
	inputs[loginFieldMasterPassword].Placeholder = "мастер-пароль (для шифрования)"
	inputs[loginFieldMasterPassword].CharLimit = 128
	inputs[loginFieldMasterPassword].Width = 30
	inputs[loginFieldMasterPassword].EchoMode = textinput.EchoPassword
	inputs[loginFieldMasterPassword].EchoCharacter = '•'

	return LoginModel{
		inputs:     inputs,
		focusIndex: 0,
	}
}

// Init инициализирует экран.
func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

// loginResultMsg — результат асинхронной операции логина/регистрации.
type loginResultMsg struct {
	masterPassword string
	err            error
}

// Update обрабатывает события.
func (m LoginModel) Update(msg tea.Msg, authService *authDomain.Service) (LoginModel, tea.Cmd) {
	switch msg := msg.(type) {

	case loginResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		// Успешный вход — переходим в главное меню
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen:     1, // ScreenMainMenu
				MasterPass: msg.masterPassword,
			}
		}

	case tea.KeyMsg:
		// Сброс ошибки при любом вводе
		m.err = nil

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "shift+tab", "up", "down"))):
			return m.moveFocus(msg)

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m.handleEnter(authService)
		}
	}

	// Обновляем активное поле ввода
	return m.updateInputs(msg)
}

// View отрисовывает экран.
func (m LoginModel) View() string {
	var b strings.Builder

	// Заголовок
	mode := "Вход"
	if m.isRegister {
		mode = "Регистрация"
	}
	b.WriteString(loginTitleStyle.Render(fmt.Sprintf("🔐 GophKeeper — %s", mode)))
	b.WriteString("\n\n")

	// Поля ввода
	for i, input := range m.inputs {
		label := m.fieldLabel(i)
		fmt.Fprintf(&b, "%s\n", label)
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	// Кнопка отправки
	submitLabel := "[ Войти ]"
	if m.isRegister {
		submitLabel = "[ Зарегистрироваться ]"
	}
	if m.loading {
		submitLabel = "[ Загрузка... ]"
	}

	if m.focusIndex == loginButtonSubmit {
		b.WriteString(loginFocusedButton.Render(submitLabel))
	} else {
		b.WriteString(loginBlurredButton.Render(submitLabel))
	}
	b.WriteString("\n")

	// Кнопка переключения режима
	toggleLabel := "[ Нет аккаунта? Регистрация ]"
	if m.isRegister {
		toggleLabel = "[ Есть аккаунт? Войти ]"
	}

	if m.focusIndex == loginButtonToggleMode {
		b.WriteString(loginFocusedButton.Render(toggleLabel))
	} else {
		b.WriteString(loginBlurredButton.Render(toggleLabel))
	}

	// Ошибка
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(loginErrorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error())))
	}

	// Подсказки
	b.WriteString("\n")
	b.WriteString(loginHintStyle.Render("tab: переключить • enter: подтвердить • ctrl+c: выход"))

	return loginBoxStyle.Render(b.String())
}

// moveFocus перемещает фокус между полями и кнопками.
func (m LoginModel) moveFocus(msg tea.KeyMsg) (LoginModel, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		m.focusIndex++
		if m.focusIndex >= loginTotalFields {
			m.focusIndex = 0
		}
	case "shift+tab", "up":
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = loginTotalFields - 1
		}
	}

	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}

	return m, tea.Batch(cmds...)
}

// handleEnter обрабатывает нажатие Enter.
func (m LoginModel) handleEnter(authService *authDomain.Service) (LoginModel, tea.Cmd) {
	switch m.focusIndex {

	// Фокус на полях ввода — перемещаем на следующее
	case loginFieldLogin, loginFieldPassword, loginFieldMasterPassword:
		m.focusIndex++
		cmds := make([]tea.Cmd, len(m.inputs))
		for i := range m.inputs {
			if i == m.focusIndex {
				cmds[i] = m.inputs[i].Focus()
			} else {
				m.inputs[i].Blur()
			}
		}
		return m, tea.Batch(cmds...)

	// Кнопка отправки
	case loginButtonSubmit:
		return m.submit(authService)

	// Кнопка переключения режима
	case loginButtonToggleMode:
		m.isRegister = !m.isRegister
		m.focusIndex = 0
		m.err = nil
		cmds := make([]tea.Cmd, len(m.inputs))
		for i := range m.inputs {
			if i == 0 {
				cmds[i] = m.inputs[i].Focus()
			} else {
				m.inputs[i].Blur()
			}
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// submit выполняет логин или регистрацию.
func (m LoginModel) submit(authService *authDomain.Service) (LoginModel, tea.Cmd) {
	login := strings.TrimSpace(m.inputs[loginFieldLogin].Value())
	password := m.inputs[loginFieldPassword].Value()
	masterPass := m.inputs[loginFieldMasterPassword].Value()

	// Валидация
	if login == "" {
		m.err = fmt.Errorf("введите логин")
		return m, nil
	}
	if password == "" {
		m.err = fmt.Errorf("введите пароль")
		return m, nil
	}
	if masterPass == "" {
		m.err = fmt.Errorf("введите мастер-пароль")
		return m, nil
	}

	m.loading = true
	isRegister := m.isRegister

	// Асинхронный запрос к серверу
	return m, func() tea.Msg {
		var err error
		if isRegister {
			err = authService.Register(login, password)
		} else {
			err = authService.Login(login, password)
		}
		return loginResultMsg{
			masterPassword: masterPass,
			err:            err,
		}
	}
}

// updateInputs обновляет поля ввода.
func (m LoginModel) updateInputs(msg tea.Msg) (LoginModel, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

// fieldLabel возвращает подпись для поля.
func (m LoginModel) fieldLabel(index int) string {
	focused := index == m.focusIndex
	var label string

	switch index {
	case loginFieldLogin:
		label = "Логин:"
	case loginFieldPassword:
		label = "Пароль:"
	case loginFieldMasterPassword:
		label = "Мастер-пароль:"
	}

	if focused {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true).
			Render(label)
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C6C6C")).
		Render(label)
}
