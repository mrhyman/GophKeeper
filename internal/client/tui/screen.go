package tui

// Screen определяет идентификатор экрана.
type Screen int

const (
	ScreenLogin Screen = iota
	ScreenMainMenu
	ScreenSecretList
	ScreenSecretView
	ScreenSecretEdit
	ScreenSyncStatus
)
