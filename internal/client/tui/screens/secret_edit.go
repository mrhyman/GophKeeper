// internal/client/tui/screens/secret_edit.go
package screens

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	secretsDomain "gophkeeper/internal/client/domain/secrets"
	"gophkeeper/pkg/models"
)

var (
	editTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	editBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 3).
			Width(60)

	editLabelFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	editLabelBlurredStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C6C6C"))

	editSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)

	editButtonFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 2).
				MarginTop(1)

	editButtonBlurredStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C6C6C")).
				Padding(0, 2).
				MarginTop(1)

	editButtonDangerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#FF4672")).
				Padding(0, 2).
				MarginTop(1)

	editErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4672")).
			MarginTop(1)

	editHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C6C6C")).
			MarginTop(1)
)

// editField описывает одно поле формы.
type editField struct {
	label       string
	input       textinput.Model
	isTextArea  bool
	textArea    textarea.Model
	isSensitive bool
}

// SecretEditModel — модель экрана создания/редактирования секрета.
type SecretEditModel struct {
	secret     *models.Secret // nil при создании
	secretType models.SecretType
	isNew      bool

	// Общие поля
	nameInput  textinput.Model
	metaInputs []metaField

	// Типизированные поля
	fields []editField

	// Навигация
	focusIndex int
	totalItems int // поля + кнопки

	// Состояние
	loading bool
	err     error
}

// metaField — пара ключ-значение метаданных.
type metaField struct {
	keyInput   textinput.Model
	valueInput textinput.Model
}

// NewSecretEditModel создаёт модель редактирования.
// secret == nil — создание нового, иначе — редактирование.
func NewSecretEditModel(secret *models.Secret, secretType models.SecretType) SecretEditModel {
	m := &SecretEditModel{
		secret: secret,
		isNew:  secret == nil,
	}

	if secret != nil {
		m.secretType = secret.Type
	} else {
		m.secretType = secretType
	}

	m.initNameInput()
	m.initTypedFields()
	m.initMetaInputs()
	m.calculateTotalItems()
	m.focusFirst()

	return *m
}

// --- Инициализация полей ---

func (m *SecretEditModel) initNameInput() {
	m.nameInput = textinput.New()
	m.nameInput.Placeholder = "название секрета"
	m.nameInput.CharLimit = 128
	m.nameInput.Width = 40

	if m.secret != nil {
		m.nameInput.SetValue(m.secret.Name)
	}
}

func (m *SecretEditModel) initTypedFields() {
	switch m.secretType {
	case models.SecretTypeLoginPassword:
		m.fields = []editField{
			m.newTextField("Логин:", "логин", 128, false),
			m.newTextField("Пароль:", "пароль", 128, true),
			m.newTextField("URL:", "https://example.com", 256, false),
		}

	case models.SecretTypeTextData:
		m.fields = []editField{
			m.newTextAreaField("Текст:", "введите текст..."),
		}

	case models.SecretTypeBinaryData:
		m.fields = []editField{
			m.newTextField("Имя файла:", "document.pdf", 256, false),
			m.newTextField("Путь к файлу:", "/path/to/file", 512, false),
		}

	case models.SecretTypeBankCard:
		m.fields = []editField{
			m.newTextField("Номер карты:", "4276380012345678", 19, false),
			m.newTextField("Владелец:", "IVAN PETROV", 64, false),
			m.newTextField("Месяц (MM):", "01", 2, false),
			m.newTextField("Год (YYYY):", "2027", 4, false),
			m.newTextField("CVV:", "123", 4, true),
		}
	}
}

func (m *SecretEditModel) initMetaInputs() {
	if m.secret != nil && len(m.secret.Metadata) > 0 {
		for k, v := range m.secret.Metadata {
			m.metaInputs = append(m.metaInputs, m.newMetaField(k, v))
		}
	}
	// Пустое поле для добавления новых
	m.metaInputs = append(m.metaInputs, m.newMetaField("", ""))
}

func (m *SecretEditModel) newTextField(label, placeholder string, charLimit int, sensitive bool) editField {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = charLimit
	input.Width = 40

	if sensitive {
		input.EchoMode = textinput.EchoPassword
		input.EchoCharacter = '•'
	}

	return editField{
		label:       label,
		input:       input,
		isSensitive: sensitive,
	}
}

func (m *SecretEditModel) newTextAreaField(label, placeholder string) editField {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.SetWidth(40)
	ta.SetHeight(5)
	ta.CharLimit = 4096

	return editField{
		label:      label,
		isTextArea: true,
		textArea:   ta,
	}
}

func (m *SecretEditModel) newMetaField(k, v string) metaField {
	keyInput := textinput.New()
	keyInput.Placeholder = "ключ"
	keyInput.CharLimit = 64
	keyInput.Width = 15

	if k != "" {
		keyInput.SetValue(k)
	}

	valueInput := textinput.New()
	valueInput.Placeholder = "значение"
	valueInput.CharLimit = 256
	valueInput.Width = 22

	if v != "" {
		valueInput.SetValue(v)
	}

	return metaField{keyInput: keyInput, valueInput: valueInput}
}

// calculateTotalItems подсчитывает общее количество фокусируемых элементов.
// name + typed fields + meta (key + value каждый) + кнопки (save, cancel, add meta)
func (m *SecretEditModel) calculateTotalItems() {
	count := 1                     // name
	count += len(m.fields)         // typed fields
	count += len(m.metaInputs) * 2 // key + value
	count += 3                     // add meta, save, cancel

	m.totalItems = count
}

// --- Индексация элементов ---

func (m SecretEditModel) fieldCount() int {
	return 1 + len(m.fields) // name + typed
}

func (m SecretEditModel) metaStartIndex() int {
	return m.fieldCount()
}

func (m SecretEditModel) buttonStartIndex() int {
	return m.fieldCount() + len(m.metaInputs)*2
}

func (m SecretEditModel) isOnName() bool {
	return m.focusIndex == 0
}

func (m SecretEditModel) typedFieldIndex() (int, bool) {
	idx := m.focusIndex - 1
	if idx >= 0 && idx < len(m.fields) {
		return idx, true
	}
	return 0, false
}

func (m SecretEditModel) metaFieldIndex() (int, bool, bool) {
	start := m.metaStartIndex()
	rel := m.focusIndex - start
	if rel >= 0 && rel < len(m.metaInputs)*2 {
		fieldIdx := rel / 2
		isValue := rel%2 == 1
		return fieldIdx, isValue, true
	}
	return 0, false, false
}

func (m SecretEditModel) buttonIndex() (int, bool) {
	start := m.buttonStartIndex()
	rel := m.focusIndex - start
	if rel >= 0 && rel < 3 {
		return rel, true // 0=add meta, 1=save, 2=cancel
	}
	return 0, false
}

func (m *SecretEditModel) focusFirst() {
	m.focusIndex = 0
	m.blurAll()
	m.nameInput.Focus()
}

// Init инициализирует экран.
func (m SecretEditModel) Init() tea.Cmd {
	return textinput.Blink
}

// saveResultMsg — результат сохранения.
type saveResultMsg struct {
	secret *models.Secret
	err    error
}

// Update обрабатывает события.
func (m SecretEditModel) Update(msg tea.Msg, secretsService *secretsDomain.Service) (SecretEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case saveResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: 3, // ScreenSecretView
				Secret: msg.secret,
			}
		}

	case tea.KeyMsg:
		m.err = nil

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m.cancel()

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "down"))):
			return m.nextFocus()

		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab", "up"))):
			return m.prevFocus()

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m.handleEnter(secretsService)

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
			return m.save(secretsService)
		}
	}

	return m.updateFocusedInput(msg)
}

// View отрисовывает экран.
func (m SecretEditModel) View() string {
	var b strings.Builder

	// Заголовок
	icon := secretTypeIcon(m.secretType)
	action := "Редактирование"
	if m.isNew {
		action = "Создание"
	}
	typeName := secretTypeName(m.secretType)
	b.WriteString(editTitleStyle.Render(fmt.Sprintf("%s %s — %s", icon, action, typeName)))
	b.WriteString("\n\n")

	// Название
	b.WriteString(m.renderLabel("Название:", m.isOnName()))
	b.WriteString(m.nameInput.View())
	b.WriteString("\n\n")

	// Типизированные поля
	b.WriteString(editSectionStyle.Render("Данные"))
	b.WriteString("\n")

	for i, f := range m.fields {
		_, isTyped := m.typedFieldIndex()
		focused := isTyped && m.focusIndex == i+1

		b.WriteString(m.renderLabel(f.label, focused))
		if f.isTextArea {
			b.WriteString(f.textArea.View())
		} else {
			b.WriteString(f.input.View())
		}
		b.WriteString("\n\n")
	}

	// Метаданные
	b.WriteString(editSectionStyle.Render("Метаданные"))
	b.WriteString("\n")

	for i, mf := range m.metaInputs {
		metaIdx, isValue, isMeta := m.metaFieldIndex()
		keyFocused := isMeta && metaIdx == i && !isValue
		valFocused := isMeta && metaIdx == i && isValue

		keyLabel := "  ключ:"
		if keyFocused {
			keyLabel = editLabelFocusedStyle.Render(keyLabel)
		} else {
			keyLabel = editLabelBlurredStyle.Render(keyLabel)
		}

		valLabel := "значение:"
		if valFocused {
			valLabel = editLabelFocusedStyle.Render(valLabel)
		} else {
			valLabel = editLabelBlurredStyle.Render(valLabel)
		}

		fmt.Fprintf(&b, "%s %s  %s %s\n",
			keyLabel, mf.keyInput.View(),
			valLabel, mf.valueInput.View(),
		)
	}

	// Кнопки
	b.WriteString("\n")
	btnIdx, isBtn := m.buttonIndex()

	addMetaLabel := "+ Добавить метаданные"
	if isBtn && btnIdx == 0 {
		b.WriteString(editButtonFocusedStyle.Render(addMetaLabel))
	} else {
		b.WriteString(editButtonBlurredStyle.Render(addMetaLabel))
	}
	b.WriteString("  ")

	saveLabel := "💾 Сохранить"
	if m.loading {
		saveLabel = "⏳ Сохранение..."
	}
	if isBtn && btnIdx == 1 {
		b.WriteString(editButtonFocusedStyle.Render(saveLabel))
	} else {
		b.WriteString(editButtonBlurredStyle.Render(saveLabel))
	}
	b.WriteString("  ")

	cancelLabel := "✗ Отмена"
	if isBtn && btnIdx == 2 {
		b.WriteString(editButtonDangerStyle.Render(cancelLabel))
	} else {
		b.WriteString(editButtonBlurredStyle.Render(cancelLabel))
	}

	// Ошибка
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(editErrorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error())))
	}

	// Подсказки
	b.WriteString("\n")
	b.WriteString(editHintStyle.Render("tab/↓: далее • shift+tab/↑: назад • ctrl+s: сохранить • esc: отмена"))

	return editBoxStyle.Render(b.String())
}

// --- Навигация фокуса ---

func (m SecretEditModel) nextFocus() (SecretEditModel, tea.Cmd) {
	m.blurAll()
	m.focusIndex++
	if m.focusIndex >= m.totalItems {
		m.focusIndex = 0
	}
	return m, m.applyFocus()
}

func (m SecretEditModel) prevFocus() (SecretEditModel, tea.Cmd) {
	m.blurAll()
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = m.totalItems - 1
	}
	return m, m.applyFocus()
}

func (m *SecretEditModel) blurAll() {
	m.nameInput.Blur()

	for i := range m.fields {
		if m.fields[i].isTextArea {
			m.fields[i].textArea.Blur()
		} else {
			m.fields[i].input.Blur()
		}
	}

	for i := range m.metaInputs {
		m.metaInputs[i].keyInput.Blur()
		m.metaInputs[i].valueInput.Blur()
	}
}

func (m *SecretEditModel) applyFocus() tea.Cmd {
	if m.isOnName() {
		return m.nameInput.Focus()
	}

	if idx, ok := m.typedFieldIndex(); ok {
		if m.fields[idx].isTextArea {
			m.fields[idx].textArea.Focus()
			return nil
		}
		return m.fields[idx].input.Focus()
	}

	if idx, isValue, ok := m.metaFieldIndex(); ok {
		if isValue {
			return m.metaInputs[idx].valueInput.Focus()
		}
		return m.metaInputs[idx].keyInput.Focus()
	}

	return nil
}

// --- Обработка Enter ---

func (m SecretEditModel) handleEnter(secretsService *secretsDomain.Service) (SecretEditModel, tea.Cmd) {
	if btnIdx, isBtn := m.buttonIndex(); isBtn {
		switch btnIdx {
		case 0: // Add meta
			return m.addMetaField()
		case 1: // Save
			return m.save(secretsService)
		case 2: // Cancel
			return m.cancel()
		}
	}

	// На полях ввода — переход к следующему
	return m.nextFocus()
}

func (m SecretEditModel) addMetaField() (SecretEditModel, tea.Cmd) {
	m.metaInputs = append(m.metaInputs, m.newMetaField("", ""))
	m.calculateTotalItems()
	return m, nil
}

func (m SecretEditModel) cancel() (SecretEditModel, tea.Cmd) {
	if m.secret != nil {
		// Возвращаемся к просмотру
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: 3, // ScreenSecretView
				Secret: m.secret,
			}
		}
	}
	// Новый секрет — возвращаемся в список
	return m, func() tea.Msg {
		return NavigateMsg{Screen: 2} // ScreenSecretList
	}
}

// --- Сохранение ---

func (m SecretEditModel) save(secretsService *secretsDomain.Service) (SecretEditModel, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	if name == "" {
		m.err = fmt.Errorf("введите название")
		return m, nil
	}

	payload, err := m.buildPayload()
	if err != nil {
		m.err = err
		return m, nil
	}

	metadata := m.buildMetadata()

	m.loading = true
	isNew := m.isNew
	secret := m.secret
	secretType := m.secretType

	return m, func() tea.Msg {
		var (
			result *models.Secret
			err    error
		)

		if isNew {
			result, err = secretsService.Create(name, secretType, payload, metadata)
		} else {
			result, err = secretsService.Update(secret.ID, name, payload, metadata)
		}

		return saveResultMsg{secret: result, err: err}
	}
}

func (m SecretEditModel) buildPayload() (interface{}, error) {
	switch m.secretType {
	case models.SecretTypeLoginPassword:
		login := strings.TrimSpace(m.fields[0].input.Value())
		password := m.fields[1].input.Value()
		if login == "" {
			return nil, fmt.Errorf("введите логин")
		}
		if password == "" {
			return nil, fmt.Errorf("введите пароль")
		}
		return &models.LoginPassword{
			Login:    login,
			Password: password,
			URI:      strings.TrimSpace(m.fields[2].input.Value()),
		}, nil

	case models.SecretTypeTextData:
		text := strings.TrimSpace(m.fields[0].textArea.Value())
		if text == "" {
			return nil, fmt.Errorf("введите текст")
		}
		return &models.TextData{Text: text}, nil

	case models.SecretTypeBinaryData:
		fileName := strings.TrimSpace(m.fields[0].input.Value())
		filePath := strings.TrimSpace(m.fields[1].input.Value())
		if fileName == "" {
			return nil, fmt.Errorf("введите имя файла")
		}
		if filePath == "" {
			return nil, fmt.Errorf("введите путь к файлу")
		}
		// Чтение файла будет в domain-сервисе; пока передаём путь
		return &models.BinaryData{
			FileName: fileName,
			Data:     []byte(filePath), // placeholder — реальное чтение в сервисе
		}, nil

	case models.SecretTypeBankCard:
		number := strings.TrimSpace(m.fields[0].input.Value())
		holder := strings.TrimSpace(m.fields[1].input.Value())
		monthStr := strings.TrimSpace(m.fields[2].input.Value())
		yearStr := strings.TrimSpace(m.fields[3].input.Value())
		cvv := strings.TrimSpace(m.fields[4].input.Value())

		if number == "" {
			return nil, fmt.Errorf("введите номер карты")
		}
		if holder == "" {
			return nil, fmt.Errorf("введите владельца")
		}

		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			return nil, fmt.Errorf("некорректный месяц (01-12)")
		}

		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 2024 {
			return nil, fmt.Errorf("некорректный год")
		}

		if len(cvv) < 3 {
			return nil, fmt.Errorf("CVV должен содержать минимум 3 цифры")
		}

		return &models.BankCard{
			Number:   number,
			Holder:   holder,
			ExpMonth: month,
			ExpYear:  year,
			CVV:      cvv,
		}, nil
	}

	return nil, fmt.Errorf("неизвестный тип секрета")
}

func (m SecretEditModel) buildMetadata() map[string]string {
	meta := make(map[string]string)

	for _, mf := range m.metaInputs {
		k := strings.TrimSpace(mf.keyInput.Value())
		v := strings.TrimSpace(mf.valueInput.Value())
		if k != "" && v != "" {
			meta[k] = v
		}
	}

	return meta
}

// --- Обновление активного поля ---

func (m SecretEditModel) updateFocusedInput(msg tea.Msg) (SecretEditModel, tea.Cmd) {
	var cmd tea.Cmd

	if m.isOnName() {
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}

	if idx, ok := m.typedFieldIndex(); ok {
		if m.fields[idx].isTextArea {
			m.fields[idx].textArea, cmd = m.fields[idx].textArea.Update(msg)
		} else {
			m.fields[idx].input, cmd = m.fields[idx].input.Update(msg)
		}
		return m, cmd
	}

	if idx, isValue, ok := m.metaFieldIndex(); ok {
		if isValue {
			m.metaInputs[idx].valueInput, cmd = m.metaInputs[idx].valueInput.Update(msg)
		} else {
			m.metaInputs[idx].keyInput, cmd = m.metaInputs[idx].keyInput.Update(msg)
		}
		return m, cmd
	}

	return m, nil
}

// --- Хелперы ---

func (m SecretEditModel) renderLabel(label string, focused bool) string {
	if focused {
		return editLabelFocusedStyle.Render(label) + "\n"
	}
	return editLabelBlurredStyle.Render(label) + "\n"
}

// PreloadFields заполняет поля из существующего расшифрованного секрета.
func (m *SecretEditModel) PreloadFields(secretsService *secretsDomain.Service) {
	if m.secret == nil {
		return
	}

	switch m.secretType {
	case models.SecretTypeLoginPassword:
		var lp models.LoginPassword
		if err := secretsService.Decrypt(m.secret, &lp); err != nil {
			return
		}
		m.fields[0].input.SetValue(lp.Login)
		m.fields[1].input.SetValue(lp.Password)
		m.fields[2].input.SetValue(lp.URI)

	case models.SecretTypeTextData:
		var td models.TextData
		if err := secretsService.Decrypt(m.secret, &td); err != nil {
			return
		}
		m.fields[0].textArea.SetValue(td.Text)

	case models.SecretTypeBinaryData:
		var bd models.BinaryData
		if err := secretsService.Decrypt(m.secret, &bd); err != nil {
			return
		}
		m.fields[0].input.SetValue(bd.FileName)

	case models.SecretTypeBankCard:
		var bc models.BankCard
		if err := secretsService.Decrypt(m.secret, &bc); err != nil {
			return
		}
		m.fields[0].input.SetValue(bc.Number)
		m.fields[1].input.SetValue(bc.Holder)
		m.fields[2].input.SetValue(fmt.Sprintf("%02d", bc.ExpMonth))
		m.fields[3].input.SetValue(fmt.Sprintf("%d", bc.ExpYear))
		m.fields[4].input.SetValue(bc.CVV)
	}
}
