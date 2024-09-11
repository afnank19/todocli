package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// m := model{
	// 	count: 10,
	// 	tasks: []string{"Build Cli", "Go lang"},
	// }

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

var (
	style     = lipgloss.NewStyle().Foreground(lipgloss.Color("ffffff")).Bold(true)
	taskStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
)

type model struct {
	count      int
	tasks      []string
	textInput  textinput.Model
	addingTask bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something"
	ti.CharLimit = 100
	ti.Width = 20

	return model{
		count:      0,
		tasks:      []string{"Test"},
		addingTask: false,
		textInput:  ti,
	}
}

// Init is the function that runs when the program starts. We don’t need to do anything here.
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+a" && !m.addingTask {
			m.textInput.Focus()
			m.addingTask = true
			return m, textinput.Blink
		}

		if m.addingTask {
			switch msg.String() {
			case "enter":
				task := strings.TrimSpace(m.textInput.Value())
				if task != "" {
					m.tasks = append(m.tasks, task)
				}
				// Clear input and return to normal mode
				m.textInput.Reset()
				m.addingTask = false

			case "esc":
				m.textInput.Reset()
				m.addingTask = false
			}
		}

		if (msg.String() == "q" || msg.String() == "ctrl+c") && !m.addingTask {
			return m, tea.Quit
		}

		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd

	}
	return m, nil
}

func (m model) View() string {
	if m.addingTask {
		return fmt.Sprintf("Add a new task: %s\n\nPress Enter to submit or Esc to cancel.", m.textInput.View())
	}

	var s string
	s += fmt.Sprintf("Counter: %d\n", m.count)

	s += "Press ↑ to increase, ↓ to decrease, Ctrl + C to quit. \n"

	s = style.Render(s)

	var taskString string
	if len(m.tasks) > 0 {
		for _, task := range m.tasks {
			taskString += task + "\n"
		}
		taskString = taskStyle.Render(taskString)
	}

	return fmt.Sprintf("%s\n%s", s, taskString)
}
