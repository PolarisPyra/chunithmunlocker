package main

import (
	"fmt"
	"strings"
)

func (m model) View() string {
	switch m.view {
	case "input":
		return fmt.Sprintf(
			"Enter the directory path:\n\n%s\n\nPress Enter to continue, or 'q' to quit.\n",
			m.dirInput.View(),
		)

	case "main":
		if m.showUpdated {
			if len(m.changes) == 0 {
				return baseStyle.Render("No changes made.") + "\nPress 'b' to go back.\n"
			}

			var changesText strings.Builder
			for i, change := range m.changes {
				if i == m.highlightPos {
					changesText.WriteString(currentHighlightStyle.Render(change) + "\n")
				} else {
					changesText.WriteString(change + "\n")
				}
			}

			return baseStyle.Render(changesText.String()) + "\nPress '↑' and '↓' to move highlight, 'b' to go back.\n"
		}
		return baseStyle.Render(m.table.View()) + "\nPress '1'-'5' to select, 'enter' to modify selected file, 'q' to quit.\n"
	}

	return ""
}
