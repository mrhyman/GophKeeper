package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	authDomain "gophkeeper/internal/client/domain/auth"
	secretsDomain "gophkeeper/internal/client/domain/secrets"
	syncDomain "gophkeeper/internal/client/domain/sync"
	"gophkeeper/internal/client/tui/screens"
	"gophkeeper/pkg/buildinfo"
)

// Services объединяет все доменные сервисы для передачи в TUI.
type Services struct {
	Auth    *authDomain.Service
	Secrets *secretsDomain.Service
	Sync    *syncDomain.Service
}

// Model — корневая модель bubbletea-приложения.
type Model struct {
	services     Services
	keys         KeyMap
	activeScreen Screen
	width        int
	height       int

	// Экраны
	loginScreen      screens.LoginModel
	mainMenuScreen   screens.MainMenuModel
	secretListScreen screens.SecretListModel
	secretViewScreen screens.SecretViewModel
	secretEditScreen screens.SecretEditModel
	syncScreen       screens.SyncStatusModel

	// Состояние
	statusMessage string
	err           error
}

// New создаёт корневую модель TUI.
func New(services Services) Model {
	return Model{
		services:     services,
		keys:         DefaultKeyMap(),
		activeScreen: ScreenLogin,
		loginScreen:  screens.NewLoginModel(),
	}
}

// Init инициализирует bubbletea-программу.
func (m Model) Init() tea.Cmd {
	return m.loginScreen.Init()
}

// Update обрабатывает сообщения.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case screens.NavigateMsg:
		return m.navigate(msg)

	case screens.ErrorMsg:
		m.err = msg.Err
		return m, nil

	case screens.StatusMsg:
		m.statusMessage = msg.Text
		return m, nil

	case screens.DeleteSecretMsg:
		if err := m.services.Secrets.Delete(msg.SecretID); err != nil {
			m.err = err
			return m, nil
		}

		m.statusMessage = "Секрет удалён"

		// Обновляем список
		secrets, err := m.services.Secrets.List()
		if err != nil {
			m.err = err
			return m, nil
		}
		m.secretListScreen = screens.NewSecretListModel(secrets, m.secretListScreen.GetFilter())
		return m, m.secretListScreen.Init()

	// Logout — очищаем токены и возвращаемся на логин
	case screens.LogoutMsg:
		_ = m.services.Auth.Logout()

		m.services.Secrets = secretsDomain.NewService(
			m.services.Secrets.GetRepo(),
			"",
		)

		m.activeScreen = ScreenLogin
		m.loginScreen = screens.NewLoginModel()

		switch msg.Reason {
		case screens.LogoutReasonUser:
			m.statusMessage = msg.Message
		case screens.LogoutReasonTokenExpired:
			m.err = fmt.Errorf("сессия истекла: %s", msg.Message)
		case screens.LogoutReasonServerError:
			m.err = fmt.Errorf("ошибка сервера: %s", msg.Message)
		}

		return m, m.loginScreen.Init()
	}
	return m.updateActiveScreen(msg)
}

// View отрисовывает интерфейс.
func (m Model) View() string {
	var content string

	switch m.activeScreen {
	case ScreenLogin:
		content = m.loginScreen.View()
	case ScreenMainMenu:
		content = m.mainMenuScreen.View()
	case ScreenSecretList:
		content = m.secretListScreen.View()
	case ScreenSecretView:
		content = m.secretViewScreen.View()
	case ScreenSecretEdit:
		content = m.secretEditScreen.View()
	case ScreenSyncStatus:
		content = m.syncScreen.View()
	}

	// Строка статуса
	status := m.renderStatusBar()

	return fmt.Sprintf("%s\n%s", content, status)
}

// navigate переключает экраны.
func (m Model) navigate(msg screens.NavigateMsg) (tea.Model, tea.Cmd) {
	m.err = nil
	m.statusMessage = ""

	// Если пришёл мастер-пароль — пересоздаём secrets-сервис
	if msg.MasterPass != "" {
		m.services.Secrets = secretsDomain.NewService(
			m.services.Secrets.GetRepo(),
			msg.MasterPass,
		)
	}

	switch msg.Screen {
	case int(ScreenMainMenu):
		m.activeScreen = ScreenMainMenu
		m.mainMenuScreen = screens.NewMainMenuModel()
		return m, m.mainMenuScreen.Init()

	case int(ScreenSecretList):
		m.activeScreen = ScreenSecretList
		secrets, err := m.services.Secrets.List()
		if err != nil {
			m.err = err
			return m, nil
		}
		m.secretListScreen = screens.NewSecretListModel(secrets, msg.Filter)
		return m, m.secretListScreen.Init()

	case int(ScreenSecretView):
		m.activeScreen = ScreenSecretView
		m.secretViewScreen = screens.NewSecretViewModel(msg.Secret, m.services.Secrets)
		return m, m.secretViewScreen.Init()

	case int(ScreenSecretEdit):
		m.activeScreen = ScreenSecretEdit
		m.secretEditScreen = screens.NewSecretEditModel(msg.Secret, msg.SecretType)
		if msg.Secret != nil {
			m.secretEditScreen.PreloadFields(m.services.Secrets)
		}
		return m, m.secretEditScreen.Init()

	case int(ScreenSyncStatus):
		m.activeScreen = ScreenSyncStatus
		m.syncScreen = screens.NewSyncStatusModel(m.services.Sync)
		return m, m.syncScreen.Init()

	case int(ScreenLogin):
		m.activeScreen = ScreenLogin
		m.loginScreen = screens.NewLoginModel()
		return m, m.loginScreen.Init()
	}

	return m, nil
}

// updateActiveScreen делегирует Update активному экрану.
func (m Model) updateActiveScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.activeScreen {
	case ScreenLogin:
		m.loginScreen, cmd = m.loginScreen.Update(msg, m.services.Auth)
	case ScreenMainMenu:
		m.mainMenuScreen, cmd = m.mainMenuScreen.Update(msg)
	case ScreenSecretList:
		m.secretListScreen, cmd = m.secretListScreen.Update(msg)
	case ScreenSecretView:
		m.secretViewScreen, cmd = m.secretViewScreen.Update(msg)
	case ScreenSecretEdit:
		m.secretEditScreen, cmd = m.secretEditScreen.Update(msg, m.services.Secrets)
	case ScreenSyncStatus:
		m.syncScreen, cmd = m.syncScreen.Update(msg)
	}

	return m, cmd
}

// renderStatusBar отрисовывает нижнюю строку статуса.
func (m Model) renderStatusBar() string {
	left := statusBarStyle.Render(buildinfo.String())

	var right string
	if m.err != nil {
		right = errorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error()))
	} else if m.statusMessage != "" {
		right = successStyle.Render(fmt.Sprintf("✓ %s", m.statusMessage))
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + statusBarStyle.Render(fmt.Sprintf("%*s", gap, "")) + right
}
