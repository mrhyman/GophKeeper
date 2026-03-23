// internal/client/tui/screens/sync_status.go
package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	syncDomain "gophkeeper/internal/client/domain/sync"
)

var (
	syncTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	syncBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 3).
			Width(55)

	syncSpinnerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4"))

	syncProgressStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C6C6C"))

	syncSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575")).
				Bold(true)

	syncErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4672")).
			Bold(true)

	syncDetailLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C6C6C")).
				Width(22)

	syncDetailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	syncButtonFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 2).
				MarginTop(1)

	syncButtonBlurredStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C6C6C")).
				Padding(0, 2).
				MarginTop(1)

	syncHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			MarginTop(1)

	syncStepDoneStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575"))

	syncStepActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4"))

	syncStepPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C6C6C"))
)

// SyncPhase описывает текущую фазу синхронизации.
type SyncPhase int

const (
	SyncPhaseIdle SyncPhase = iota
	SyncPhaseConnecting
	SyncPhaseSendingLocal
	SyncPhaseReceivingServer
	SyncPhaseApplying
	SyncPhaseDone
	SyncPhaseFailed
)

// syncResultMsg — результат операции синхронизации.
type syncResultMsg struct {
	uploaded   int
	downloaded int
	duration   time.Duration
	err        error
}

// SyncStatusModel — модель экрана синхронизации.
type SyncStatusModel struct {
	syncService *syncDomain.Service
	spinner     spinner.Model

	// Состояние
	phase      SyncPhase
	startTime  time.Time
	uploaded   int
	downloaded int
	duration   time.Duration
	err        error

	// Кнопки
	buttonCursor int // 0 = повторить, 1 = назад
}

// NewSyncStatusModel создаёт новую модель экрана синхронизации.
func NewSyncStatusModel(syncService *syncDomain.Service) SyncStatusModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = syncSpinnerStyle

	return SyncStatusModel{
		syncService: syncService,
		spinner:     s,
		phase:       SyncPhaseIdle,
	}
}

// Init запускает синхронизацию сразу при открытии экрана.
func (m SyncStatusModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startSync(),
	)
}

// Update обрабатывает события.
func (m SyncStatusModel) Update(msg tea.Msg) (SyncStatusModel, tea.Cmd) {
	switch msg := msg.(type) {

	case spinner.TickMsg:
		if m.phase != SyncPhaseDone && m.phase != SyncPhaseFailed {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case syncPhaseMsg:
		m.phase = msg.phase
		return m, nil

	case syncResultMsg:
		if msg.err != nil {
			m.phase = SyncPhaseFailed
			m.err = msg.err
		} else {
			m.phase = SyncPhaseDone
			m.uploaded = msg.uploaded
			m.downloaded = msg.downloaded
			m.duration = msg.duration
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
			// Если идёт синхронизация — не уходим
			if m.isInProgress() {
				return m, nil
			}
			return m.goBack()

		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h", "right", "l", "tab"))):
			if !m.isInProgress() {
				m.buttonCursor = (m.buttonCursor + 1) % 2
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.isInProgress() {
				return m, nil
			}
			return m.handleButton()

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			if !m.isInProgress() {
				return m.retry()
			}
		}
	}

	return m, nil
}

// View отрисовывает экран.
func (m SyncStatusModel) View() string {
	var b strings.Builder

	b.WriteString(syncTitleStyle.Render("🔄 Синхронизация"))
	b.WriteString("\n\n")

	// Шаги
	b.WriteString(m.renderSteps())
	b.WriteString("\n")

	// Результат
	switch m.phase {
	case SyncPhaseDone:
		b.WriteString(m.renderSuccess())
	case SyncPhaseFailed:
		b.WriteString(m.renderError())
	default:
		// В процессе — спиннер
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s %s",
			m.spinner.View(),
			syncProgressStyle.Render(m.phaseText()),
		))
		b.WriteString("\n")
	}

	// Кнопки (только когда завершено)
	if !m.isInProgress() {
		b.WriteString("\n")
		b.WriteString(m.renderButtons())
	}

	// Подсказки
	b.WriteString("\n")
	if m.isInProgress() {
		b.WriteString(syncHintStyle.Render("Подождите, идёт синхронизация..."))
	} else {
		b.WriteString(syncHintStyle.Render("←→/hl: выбор • enter: выполнить • r: повторить • esc: назад"))
	}

	return syncBoxStyle.Render(b.String())
}

// --- Запуск синхронизации ---

// syncPhaseMsg — обновление фазы синхронизации.
type syncPhaseMsg struct {
	phase SyncPhase
}

func (m SyncStatusModel) startSync() tea.Cmd {
	syncService := m.syncService

	return func() tea.Msg {
		start := time.Now()

		result, err := syncService.Sync()
		duration := time.Since(start)

		if err != nil {
			return syncResultMsg{err: err, duration: duration}
		}

		return syncResultMsg{
			uploaded:   result.Uploaded,
			downloaded: result.Downloaded,
			duration:   duration,
		}
	}
}

func (m SyncStatusModel) retry() (SyncStatusModel, tea.Cmd) {
	m.phase = SyncPhaseIdle
	m.err = nil
	m.uploaded = 0
	m.downloaded = 0
	m.duration = 0
	m.buttonCursor = 0

	return m, tea.Batch(
		m.spinner.Tick,
		m.startSync(),
	)
}

// --- Рендеринг шагов ---

func (m SyncStatusModel) renderSteps() string {
	type step struct {
		label string
		phase SyncPhase
	}

	steps := []step{
		{"Подключение к серверу", SyncPhaseConnecting},
		{"Отправка локальных изменений", SyncPhaseSendingLocal},
		{"Получение серверных изменений", SyncPhaseReceivingServer},
		{"Применение изменений", SyncPhaseApplying},
	}

	var b strings.Builder

	for _, s := range steps {
		var icon string
		var style lipgloss.Style

		switch {
		case m.phase == SyncPhaseFailed && s.phase > m.phase:
			icon = "  ○"
			style = syncStepPendingStyle
		case m.phase == SyncPhaseDone || s.phase < m.phase:
			icon = "  ✓"
			style = syncStepDoneStyle
		case s.phase == m.phase:
			icon = "  ●"
			style = syncStepActiveStyle
		default:
			icon = "  ○"
			style = syncStepPendingStyle
		}

		b.WriteString(style.Render(fmt.Sprintf("%s %s", icon, s.label)))
		b.WriteString("\n")
	}

	return b.String()
}

// --- Рендеринг результатов ---

func (m SyncStatusModel) renderSuccess() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(syncSuccessStyle.Render("  ✓ Синхронизация завершена успешно"))
	b.WriteString("\n\n")

	b.WriteString(m.renderDetail("Время:", m.formatDuration()))

	if m.uploaded > 0 {
		b.WriteString(m.renderDetail("Отправлено:", fmt.Sprintf("%d", m.uploaded)))
	}
	if m.downloaded > 0 {
		b.WriteString(m.renderDetail("Получено:", fmt.Sprintf("%d", m.downloaded)))
	}
	if m.uploaded == 0 && m.downloaded == 0 {
		b.WriteString(m.renderDetail("Статус:", "Всё актуально, изменений нет"))
	}

	return b.String()
}

func (m SyncStatusModel) renderError() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(syncErrorStyle.Render("  ✗ Ошибка синхронизации"))
	b.WriteString("\n\n")

	b.WriteString(m.renderDetail("Время:", m.formatDuration()))
	b.WriteString(m.renderDetail("Ошибка:", m.err.Error()))

	return b.String()
}

func (m SyncStatusModel) renderDetail(label, value string) string {
	l := syncDetailLabelStyle.Render("  " + label)
	v := syncDetailValueStyle.Render(value)
	return fmt.Sprintf("%s %s\n", l, v)
}

// --- Кнопки ---

func (m SyncStatusModel) renderButtons() string {
	retryLabel := "🔄 Повторить"
	backLabel := "← Назад"

	var retry, back string

	if m.buttonCursor == 0 {
		retry = syncButtonFocusedStyle.Render(retryLabel)
	} else {
		retry = syncButtonBlurredStyle.Render(retryLabel)
	}

	if m.buttonCursor == 1 {
		back = syncButtonFocusedStyle.Render(backLabel)
	} else {
		back = syncButtonBlurredStyle.Render(backLabel)
	}

	return fmt.Sprintf("  %s  %s", retry, back)
}

func (m SyncStatusModel) handleButton() (SyncStatusModel, tea.Cmd) {
	switch m.buttonCursor {
	case 0:
		return m.retry()
	case 1:
		return m.goBack()
	}
	return m, nil
}

func (m SyncStatusModel) goBack() (SyncStatusModel, tea.Cmd) {
	return m, func() tea.Msg {
		return NavigateMsg{Screen: 1} // ScreenMainMenu
	}
}

// --- Утилиты ---

func (m SyncStatusModel) isInProgress() bool {
	return m.phase != SyncPhaseIdle &&
		m.phase != SyncPhaseDone &&
		m.phase != SyncPhaseFailed
}

func (m SyncStatusModel) phaseText() string {
	switch m.phase {
	case SyncPhaseIdle:
		return "Подготовка..."
	case SyncPhaseConnecting:
		return "Подключение к серверу..."
	case SyncPhaseSendingLocal:
		return "Отправка локальных изменений..."
	case SyncPhaseReceivingServer:
		return "Получение серверных изменений..."
	case SyncPhaseApplying:
		return "Применение изменений..."
	default:
		return ""
	}
}

func (m SyncStatusModel) formatDuration() string {
	d := m.duration

	switch {
	case d < time.Second:
		return fmt.Sprintf("%d мс", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.1f сек", d.Seconds())
	default:
		return fmt.Sprintf("%d мин %d сек", int(d.Minutes()), int(d.Seconds())%60)
	}
}
