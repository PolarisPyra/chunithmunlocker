package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func countSpecificXMLFiles(dir string, filenames []string) (map[string]int, error) {
	counts := make(map[string]int)
	for _, filename := range filenames {
		counts[filename] = 0
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			for _, filename := range filenames {
				if info.Name() == filename {
					counts[filename]++
				}
			}
		}
		return nil
	})
	return counts, err
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table      table.Model
	fileCounts map[string]int  // Stores the counts of all files
	changes    []string        // Stores all the changes made for the selected file
	dirInput   textinput.Model // Input field for directory path
	dir        string          // Stores the directory path
	view       string          // Tracks the current view
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter directory path"
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 50

	return model{
		dirInput: ti,
		view:     "input",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.view {
	case "input":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				m.dir = m.dirInput.Value()
				if m.dir == "" {
					return m, nil
				}
				m.view = "main"

				filenames := []string{"Music.xml", "Event.xml", "Chara.xml", "NamePlate.xml", "AvatarAccessory.xml", "Trophy.xml", "MapIcon.xml"}
				fileCounts, err := countSpecificXMLFiles(m.dir, filenames)
				if err != nil {
					log.Printf("Error scanning directory: %v\n", err)
					return m, tea.Quit
				}
				m.fileCounts = fileCounts

				columns := []table.Column{
					{Title: "Option", Width: 10},
					{Title: "File Name", Width: 20},
					{Title: "Total Number", Width: 15},
				}

				rows := []table.Row{
					{"1", "Music.xml", fmt.Sprintf("%d", fileCounts["Music.xml"])},
					{"2", "Event.xml", fmt.Sprintf("%d", fileCounts["Event.xml"])},
					{"3", "Chara.xml", fmt.Sprintf("%d", fileCounts["Chara.xml"])},
					{"4", "NamePlate.xml", fmt.Sprintf("%d", fileCounts["NamePlate.xml"])},
					{"5", "AvatarAccessory.xml", fmt.Sprintf("%d", fileCounts["AvatarAccessory.xml"])},
					{"6", "Trophy.xml", fmt.Sprintf("%d", fileCounts["Trophy.xml"])},
					{"7", "MapIcon.xml", fmt.Sprintf("%d", fileCounts["MapIcon.xml"])},
					{"8", "Unlock all", "All files"},
					{"9", "Relock all", "All files"},
				}

				t := table.New(
					table.WithColumns(columns),
					table.WithRows(rows),
					table.WithFocused(true),
					table.WithHeight(10),
				)

				s := table.DefaultStyles()
				s.Header = s.Header.
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(lipgloss.Color("240")).
					BorderBottom(true).
					Bold(false)
				s.Selected = s.Selected.
					Foreground(lipgloss.Color("229")).
					Background(lipgloss.Color("57")).
					Bold(false)
				t.SetStyles(s)

				m.table = t
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit

			}

		}

		m.dirInput, cmd = m.dirInput.Update(msg)
		return m, cmd

	case "main":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				if m.table.Focused() {
					m.table.Blur()
				} else {
					m.table.Focus()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "1":
				m.table.SetCursor(0)
			case "2":
				m.table.SetCursor(1)
			case "3":
				m.table.SetCursor(2)
			case "4":
				m.table.SetCursor(3)
			case "5":
				m.table.SetCursor(4)
			case "6":
				m.table.SetCursor(5)
			case "7":
				m.table.SetCursor(6)
			case "8":
				m.table.SetCursor(7)
			case "9":
				m.table.SetCursor(8)
			case "enter":
				m.changes = nil

				selectedRow := m.table.SelectedRow()
				if len(selectedRow) == 0 {
					return m, nil
				}

				// If "UNLOCK ALL" is selected
				isUnlockAll := selectedRow[0] == "8"
				// If "RELOCK ALL" is selected
				isRelockAll := selectedRow[0] == "9"

				// Define which files to process
				var filesToProcess []string
				if isUnlockAll || isRelockAll {
					filesToProcess = []string{"Music.xml", "Event.xml", "Chara.xml", "NamePlate.xml", "AvatarAccessory.xml", "Trophy.xml", "MapIcon.xml"}
				} else {
					filesToProcess = []string{selectedRow[1]}
				}

				for _, fileToProcess := range filesToProcess {
					err := filepath.Walk(m.dir, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}

						if !info.IsDir() && info.Name() == fileToProcess {
							data, err := os.ReadFile(path)
							if err != nil {
								log.Printf("Error reading file %s: %v\n", path, err)
								return nil
							}

							content := string(data)
							var updatedContent string
							var change string

							switch fileToProcess {
							case "Music.xml":
								if isRelockAll {
									if strings.Contains(content, "<firstLock>false</firstLock>") {
										updatedContent = strings.Replace(content, "<firstLock>false</firstLock>", "<firstLock>true</firstLock>", -1)
										change = fmt.Sprintf("Updated %s: Changed <firstLock>false</firstLock> to <firstLock>true</firstLock>", path)
									} else {
										change = fmt.Sprintf("Skipped %s: Already has <firstLock>true</firstLock>", path)
									}
								} else {
									if strings.Contains(content, "<firstLock>true</firstLock>") {
										updatedContent = strings.Replace(content, "<firstLock>true</firstLock>", "<firstLock>false</firstLock>", -1)
										change = fmt.Sprintf("Updated %s: Changed <firstLock>true</firstLock> to <firstLock>false</firstLock>", path)
									} else {
										change = fmt.Sprintf("Skipped %s: Already has <firstLock>false</firstLock>", path)
									}
								}
							case "Event.xml":
								if isRelockAll {
									if strings.Contains(content, "<alwaysOpen>true</alwaysOpen>") {
										updatedContent = strings.Replace(content, "<alwaysOpen>true</alwaysOpen>", "<alwaysOpen>false</alwaysOpen>", -1)
										change = fmt.Sprintf("Updated %s: Changed <alwaysOpen>true</alwaysOpen> to <alwaysOpen>false</alwaysOpen>", path)
									} else {
										change = fmt.Sprintf("Skipped %s: Already has <alwaysOpen>false</alwaysOpen>", path)
									}
								} else {
									if strings.Contains(content, "<alwaysOpen>false</alwaysOpen>") {
										updatedContent = strings.Replace(content, "<alwaysOpen>false</alwaysOpen>", "<alwaysOpen>true</alwaysOpen>", -1)
										change = fmt.Sprintf("Updated %s: Changed <alwaysOpen>false</alwaysOpen> to <alwaysOpen>true</alwaysOpen>", path)
									} else {
										change = fmt.Sprintf("Skipped %s: Already has <alwaysOpen>true</alwaysOpen>", path)
									}
								}
							case "Chara.xml", "NamePlate.xml", "AvatarAccessory.xml", "Trophy.xml", "MapIcon.xml":
								if isRelockAll {
									if strings.Contains(content, "<defaultHave>true</defaultHave>") {
										updatedContent = strings.Replace(content, "<defaultHave>true</defaultHave>", "<defaultHave>false</defaultHave>", -1)
										change = fmt.Sprintf("Updated %s: Changed <defaultHave>true</defaultHave> to <defaultHave>false</defaultHave>", path)
									} else {
										change = fmt.Sprintf("Skipped %s: Already has <defaultHave>false</defaultHave>", path)
									}
								} else {
									if strings.Contains(content, "<defaultHave>false</defaultHave>") {
										updatedContent = strings.Replace(content, "<defaultHave>false</defaultHave>", "<defaultHave>true</defaultHave>", -1)
										change = fmt.Sprintf("Updated %s: Changed <defaultHave>false</defaultHave> to <defaultHave>true</defaultHave>", path)
									} else {
										change = fmt.Sprintf("Skipped %s: Already has <defaultHave>true</defaultHave>", path)
									}
								}
							default:
								return nil
							}

							if updatedContent != "" {
								err = os.WriteFile(path, []byte(updatedContent), 0644)
								if err != nil {
									log.Printf("Error writing modified XML to file %s: %v\n", path, err)
									return nil
								}
							}

							m.changes = append(m.changes, change)
						}
						return nil
					})
					if err != nil {
						log.Printf("Error walking directory: %v\n", err)
					}
				}
				m.view = "success"
			}
		}

		m.table, cmd = m.table.Update(msg)
		return m, cmd

	case "success":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "b":
				m.view = "main"
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	switch m.view {
	case "input":
		return fmt.Sprintf(
			"Enter the directory path:\n\n%s\n\nPress Enter to continue, or 'q' to quit.\n",
			m.dirInput.View(),
		)

	case "main":
		return baseStyle.Render(m.table.View()) + "\nPress '1'-'9' to select, 'enter' to modify selected file(s), 'q' to quit.\n"

	case "success":
		view := "Changes:\n"
		for _, change := range m.changes {
			view += change + "\n"
		}
		view += "\nPress 'b' to return to main view, 'q' to quit.\n"
		return view
	}

	return ""
}

func main() {
	m := initialModel()

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
