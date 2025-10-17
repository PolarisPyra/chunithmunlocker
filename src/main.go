package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	viewInput      = "input"
	viewMain       = "main"
	viewProcessing = "processing"
	viewSuccess    = "success"

	progressBarWidth  = 60
	inputWidth        = 50
	inputCharLimit    = 200
	tableHeight       = 12
	maxRecentChanges  = 10
	maxSuccessChanges = 15

	msgUpdated = "[+]"
	msgSkipped = "[-]"
	msgError   = "[ERROR]"
)

var xmlFileNames = []string{
	"Music.xml",
	"Event.xml",
	"Chara.xml",
	"NamePlate.xml",
	"AvatarAccessory.xml",
	"Trophy.xml",
	"MapIcon.xml",
	"SystemVoice.xml",
	"Stage.xml",
}

type xmlTransformRule struct {
	lockedTag   string
	unlockedTag string
}

func getTransformRules() map[string]xmlTransformRule {
	return map[string]xmlTransformRule{
		"Music.xml": {
			lockedTag:   "<firstLock>true</firstLock>",
			unlockedTag: "<firstLock>false</firstLock>",
		},
		"Event.xml": {
			lockedTag:   "<alwaysOpen>false</alwaysOpen>",
			unlockedTag: "<alwaysOpen>true</alwaysOpen>",
		},
	}
}

var defaultTransformRule = xmlTransformRule{
	lockedTag:   "<defaultHave>false</defaultHave>",
	unlockedTag: "<defaultHave>true</defaultHave>",
}

// Messages for Bubble Tea event system
type fileChangeMsg string
type processingDoneMsg struct{}
type startProcessingMsg struct {
	messageChan <-chan fileChangeMsg
	totalFiles  int
}

type model struct {
	// UI components
	table    table.Model
	dirInput textinput.Model
	progress progress.Model

	// State
	view           string
	dir            string
	fileCounts     map[string]int
	changes        []string
	messageChan    <-chan fileChangeMsg
	totalFiles     int
	processedFiles int
}

func newModel() model {
	input := textinput.New()
	input.Placeholder = "Enter directory path"
	input.Focus()
	input.CharLimit = inputCharLimit
	input.Width = inputWidth

	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = progressBarWidth

	return model{
		dirInput: input,
		view:     viewInput,
		progress: prog,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startProcessingMsg:
		m.messageChan = msg.messageChan
		m.totalFiles = msg.totalFiles
		m.processedFiles = 0
		m.view = viewProcessing
		return m, listenForMessages(m.messageChan)

	case fileChangeMsg:
		m.changes = append(m.changes, string(msg))
		m.processedFiles++
		return m, listenForMessages(m.messageChan)

	case processingDoneMsg:
		m.view = viewSuccess
		m.messageChan = nil
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m.updateSubComponents(msg)
}

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.view {
	case viewInput:
		return m.handleInputKeys(msg)
	case viewMain:
		return m.handleMainKeys(msg)
	case viewSuccess, viewProcessing:
		return m.handleSuccessKeys(msg)
	}
	return m, nil
}

func (m model) updateSubComponents(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.view {
	case viewInput:
		m.dirInput, cmd = m.dirInput.Update(msg)
	case viewMain:
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		dir := m.dirInput.Value()
		if dir == "" {
			return m, nil
		}

		m.dir = dir
		counts, err := countXMLFiles(m.dir, xmlFileNames)
		if err != nil {
			return m, tea.Quit
		}

		m.fileCounts = counts
		m.table = createSelectionTable(counts)
		m.view = viewMain
		return m, nil

	case "q", "ctrl+c":
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.dirInput, cmd = m.dirInput.Update(msg)
	return m, cmd
}

func (m model) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.table.SetCursor(int(key[0] - '1'))
	case "10":
		m.table.SetCursor(9)
	case "11":
		m.table.SetCursor(10)
	case "enter":
		return m.startProcessing()
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) handleSuccessKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b":
		m.view = viewMain
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) startProcessing() (tea.Model, tea.Cmd) {
	selectedRow := m.table.SelectedRow()
	if len(selectedRow) == 0 {
		return m, nil
	}

	option := selectedRow[0]
	shouldRelock := option == "11"

	var filesToProcess map[string]bool
	if option == "10" || option == "11" {
		filesToProcess = makeFileSet(xmlFileNames)
	} else {
		filesToProcess = map[string]bool{selectedRow[1]: true}
	}

	m.changes = nil
	return m, processFiles(m.dir, filesToProcess, shouldRelock)
}

func (m model) View() string {
	switch m.view {
	case viewInput:
		return m.renderInputView()
	case viewMain:
		return m.renderMainView()
	case viewProcessing:
		return m.renderProcessingView()
	case viewSuccess:
		return m.renderSuccessView()
	}
	return ""
}

func (m model) renderInputView() string {
	return fmt.Sprintf(
		"Enter the directory path:\n\n%s\n\nPress Enter to continue, or 'q' to quit.\n",
		m.dirInput.View(),
	)
}

func (m model) renderMainView() string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return style.Render(m.table.View()) +
		"\nPress '1'-'11' to select, 'enter' to modify, 'q' to quit.\n"
}

func (m model) renderProcessingView() string {
	progressPercent := 0.0
	if m.totalFiles > 0 {
		progressPercent = float64(m.processedFiles) / float64(m.totalFiles)
	}

	view := fmt.Sprintf("Processing files... %d/%d\n\n", m.processedFiles, m.totalFiles)
	view += m.progress.ViewAs(progressPercent) + "\n\n"
	view += renderRecentChanges(m.changes, maxRecentChanges)

	if len(m.changes) == 0 {
		view += "Starting...\n"
	}

	return view
}

func (m model) renderSuccessView() string {
	view := fmt.Sprintf("Complete! Processed %d files\n\n", len(m.changes))
	view += renderRecentChanges(m.changes, maxSuccessChanges)
	view += "\nPress 'b' to return to main view, 'q' to quit.\n"
	return view
}

func renderRecentChanges(changes []string, maxVisible int) string {
	if len(changes) == 0 {
		return ""
	}

	startIdx := 0
	result := ""

	if len(changes) > maxVisible {
		startIdx = len(changes) - maxVisible
		result += fmt.Sprintf("... (%d more above)\n", startIdx)
	}

	for i := startIdx; i < len(changes); i++ {
		result += changes[i] + "\n"
	}

	return result
}

func countXMLFiles(dir string, filenames []string) (map[string]int, error) {
	counts := make(map[string]int, len(filenames))
	for _, filename := range filenames {
		counts[filename] = 0
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		for _, filename := range filenames {
			if info.Name() == filename {
				counts[filename]++
				break
			}
		}
		return nil
	})

	return counts, err
}

func createSelectionTable(counts map[string]int) table.Model {
	columns := []table.Column{
		{Title: "Option", Width: 10},
		{Title: "File Name", Width: 20},
		{Title: "Total Number", Width: 15},
	}

	rows := make([]table.Row, 0, len(xmlFileNames)+2)
	for i, filename := range xmlFileNames {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			filename,
			fmt.Sprintf("%d", counts[filename]),
		})
	}

	rows = append(rows,
		table.Row{"10", "Unlock all", "All files"},
		table.Row{"11", "Relock all", "All files"},
	)

	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	styles.Selected = styles.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	tbl.SetStyles(styles)

	return tbl
}

func processFiles(dir string, filesToProcess map[string]bool, shouldRelock bool) tea.Cmd {
	return func() tea.Msg {
		totalFiles := countFilesToProcess(dir, filesToProcess)
		messageChan := make(chan fileChangeMsg, 100)

		go processFilesAsync(dir, filesToProcess, shouldRelock, messageChan)

		return startProcessingMsg{
			messageChan: messageChan,
			totalFiles:  totalFiles,
		}
	}
}

func countFilesToProcess(dir string, filesToProcess map[string]bool) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filesToProcess[info.Name()] {
			count++
		}
		return nil
	})
	return count
}

func processFilesAsync(dir string, filesToProcess map[string]bool, shouldRelock bool, messageChan chan<- fileChangeMsg) {
	defer close(messageChan)

	rules := getTransformRules()

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !filesToProcess[info.Name()] {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			messageChan <- fileChangeMsg(fmt.Sprintf("%s reading %s", msgError, path))
			return nil
		}

		updated, message := transformXMLContent(
			info.Name(),
			string(content),
			path,
			shouldRelock,
			rules,
		)

		if updated != "" {
			if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
				messageChan <- fileChangeMsg(fmt.Sprintf("%s writing %s", msgError, path))
				return nil
			}
		}

		messageChan <- fileChangeMsg(message)
		return nil
	})
}

func transformXMLContent(filename, content, path string, shouldRelock bool, rules map[string]xmlTransformRule) (updated, message string) {
	rule, exists := rules[filename]
	if !exists {
		rule = defaultTransformRule
	}

	var oldTag, newTag string
	if shouldRelock {
		oldTag, newTag = rule.unlockedTag, rule.lockedTag
	} else {
		oldTag, newTag = rule.lockedTag, rule.unlockedTag
	}

	if strings.Contains(content, oldTag) {
		updated = strings.ReplaceAll(content, oldTag, newTag)
		message = fmt.Sprintf("%s %s: %s -> %s", msgUpdated, path, oldTag, newTag)
	} else {
		message = fmt.Sprintf("%s %s: Already has %s", msgSkipped, path, newTag)
	}

	return updated, message
}

func listenForMessages(messageChan <-chan fileChangeMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-messageChan
		if !ok {
			return processingDoneMsg{}
		}
		return msg
	}
}

func makeFileSet(files []string) map[string]bool {
	set := make(map[string]bool, len(files))
	for _, file := range files {
		set[file] = true
	}
	return set
}

func main() {
	program := tea.NewProgram(newModel())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
