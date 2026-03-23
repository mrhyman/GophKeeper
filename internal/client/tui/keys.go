package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap описывает глобальные привязки клавиш.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
	Tab    key.Binding
	Help   key.Binding
	Sync   key.Binding
	Delete key.Binding
	New    key.Binding
}

// DefaultKeyMap возвращает привязки по умолчанию.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "вверх"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "вниз"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "выбрать"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "назад"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "выход"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "переключить"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "помощь"),
		),
		Sync: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "синхронизация"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "удалить"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "создать"),
		),
	}
}
