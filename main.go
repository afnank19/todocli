package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

const listHeight = 14

// Lipgloss styles cuz brat summer
var (
	//taskStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	doneStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#4a4a4a")).Strikethrough(true).PaddingLeft(4)
	//centered          = lipgloss.NewStyle().Align(lipgloss.Center) //This needs a .Width with terminal width to appropriately center
	//backgroundStyle   = lipgloss.NewStyle().Background(lipgloss.Color("201"))
	//quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	if strings.Contains(str, "[X]") {
		fn = func(s ...string) string {
			return doneStyle.Render(strings.Join(s, " "))
		}
	}
	// if index == 2 {
	// 	fn = func(s ...string) string {
	// 		return doneStyle.Render(strings.Join(s, " "))
	// 	}
	// }

	fmt.Fprint(w, fn(str))
}

type model struct {
	tasks      []string
	textInput  textinput.Model
	list       list.Model
	addingTask bool
	items      []list.Item
	done       []bool
	taskList   list.Model
	taskItems  []list.Item
	taskView   bool
	choice     string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something"
	ti.CharLimit = 100
	ti.Width = 20

	items := []list.Item{}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Project List"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return model{
		tasks:      []string{"Test"},
		addingTask: false,
		textInput:  ti,
		list:       l,
		items:      items,
		done:       []bool{true},
		taskView:   false,
		choice:     "",
	}
}

// Init is the function that runs when the program starts. We don’t need to do anything here.
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+a" && !m.addingTask {
			m.textInput.Focus()
			m.addingTask = true
			return m, textinput.Blink
		}

		if m.taskView {
			switch msg.String() {
			case "enter":
				idx := m.list.Index()
				i, ok := m.list.SelectedItem().(item)
				if !ok {
					panic("Some fucking thing went wrong :P")
				}
				it := i + item(" [X]")
				m.list.SetItem(idx, it)

			case "backspace":
				// refreshItems := []list.Item{}
				m.list.Title = "Project List"
				// m.list.SetItems(refreshItems)
				items := []list.Item{item("Project 1"), item("Lewis")} //Probably a database call to get the project list
				m.list.SetItems(items)
				m.taskView = false
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.addingTask {
			switch msg.String() {
			case "enter":
				task := strings.TrimSpace(m.textInput.Value())
				if task != "" {
					m.tasks = append(m.tasks, task)
					m.items = append(m.items, item(task))
					m.list.InsertItem(len(m.items)+1, item(task))
				}
				// Clear input and return to normal mode
				m.textInput.Reset()
				m.addingTask = false

			case "esc":
				m.textInput.Reset()
				m.addingTask = false
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		if !m.addingTask && !m.taskView {
			switch msg.String() {
			case "enter":
				//m.taskView = true
				i, ok := m.list.SelectedItem().(item)
				if !ok {
					panic("Some fucking thing went wrong :P")
				}

				//Simulating a database system or some sort of storage
				var simulatedItems []list.Item
				if string(i) == "Project 1" {
					simulatedItems = []list.Item{item("Doja Cat"), item("Sydney Sweeney"), item("Megan Fox")}
				} else {
					simulatedItems = []list.Item{item("Do your work"), item("Post a story"), item("dread existence")}
				}

				m.list.Title = string("$/" + "projects/" + i + "/task-list")
				m.list.SetItems(simulatedItems)
				m.taskView = true
				// i, ok := m.list.SelectedItem().(item)
				// idx := m.list.Index()
				// i, ok := m.list.SelectedItem().(item)
				// if !ok {
				// 	panic("Some fucking thing went wrong :P")
				// }
				// it := i + item(" [X]")
				// m.list.SetItem(idx, it)
			case "delete":
				idx := m.list.Index()
				m.list.RemoveItem(idx)
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if (msg.String() == "q" || msg.String() == "ctrl+c") && !m.addingTask {
			return m, tea.Quit
		}

	}
	return m, nil
}

func (m model) View() string {
	if m.addingTask {
		return fmt.Sprintf("Add a new task: %s\n\nPress Enter to submit or Esc to cancel.", m.textInput.View())
	}

	// if m.taskView {
	// 	return "\n\n" + m.taskList.View()
	// }

	//var help string
	help := helpStyle.Render("Ctrl+A to add new task")
	//listView := backgroundStyle.Render(m.list.View()) experiment
	//listView := centered.Render(m.list.View()) //Uncomment for center

	if !m.addingTask {
		return "\n\n" + m.list.View() + "\n\n" + help
	}

	return "Welp"
}
