package main

import (
	"database/sql"
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

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./todo.db")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS projects (ID INTEGER PRIMARY KEY AUTOINCREMENT, projectName TEXT NOT NULL);")
	if err != nil {
		panic(err)
	}

	//Test Code, delete db before running
	// insertQuery := "INSERT INTO projects(projectName) VALUES (?)"
	// _, err = db.Exec(insertQuery, "Getting Started")
	// if err != nil {
	// 	panic(err)
	// }

	initialItems := InitProjectList(db)

	p := tea.NewProgram(initialModel(initialItems), tea.WithAltScreen())

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
	title             = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#d4c6a9")).MarginLeft(2)
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
	textInput     textinput.Model
	taskInput     textinput.Model
	list          list.Model
	taskList      list.Model
	items         []list.Item
	taskItems     []list.Item
	tasks         []string
	choice        string
	done          []bool
	addingProject bool
	taskView      bool
	addingTask    bool
}

func InitProjectList(db *sql.DB) []list.Item {
	rows, err := db.Query("SELECT ID, projectName from projects")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var items []list.Item

	for rows.Next() {
		var ID string
		var projName string
		err = rows.Scan(&ID, &projName)
		if err != nil {
			panic(err)
		}
		items = append(items, item(projName))
	}

	return items
}

func initialModel(initialItems []list.Item) model {
	ti := textinput.New()
	ti.Placeholder = "Type something"
	ti.CharLimit = 100
	ti.Width = 20

	taskIn := textinput.New()
	taskIn.Placeholder = "Enter a task"
	taskIn.CharLimit = 100
	taskIn.Width = 40

	items := initialItems

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Project List"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = title
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return model{
		tasks:         []string{"Test"},
		addingProject: false,
		textInput:     ti,
		list:          l,
		items:         items,
		done:          []bool{true},
		taskView:      false,
		choice:        "",
		addingTask:    false,
		taskInput:     taskIn,
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
		//This is for adding a new project
		if msg.String() == "ctrl+a" && !m.addingProject && !m.taskView {
			m.textInput.Focus()
			m.addingProject = true
			return m, textinput.Blink
		}
		if msg.String() == "ctrl+a" && !m.addingProject && m.taskView {
			m.taskInput.Focus()
			m.addingTask = true
			return m, textinput.Blink
		}

		if m.taskView && !m.addingTask {
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

			case "delete":
				idx := m.list.Index()
				m.list.RemoveItem(idx)
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		if m.addingTask && m.taskView {
			switch msg.String() {
			case "enter":
				task := strings.TrimSpace(m.taskInput.Value())
				if task != "" {
					m.tasks = append(m.tasks, task)
					m.items = append(m.items, item(task))
					m.list.InsertItem(len(m.items)+1, item(task))
				}
				// Clear input and return to normal mode
				m.taskInput.Reset()
				m.addingTask = false

			case "tab":
				m.taskInput.Reset()
				m.addingTask = false
			}
			m.taskInput, cmd = m.taskInput.Update(msg)
			return m, cmd
		}

		if m.addingProject {
			switch msg.String() {
			case "enter":
				task := strings.TrimSpace(m.textInput.Value())

				insertQuery := "INSERT INTO projects(projectName) VALUES (?)"
				_, err := db.Exec(insertQuery, task)
				if err != nil {
					panic(err)
				}

				if task != "" {
					m.tasks = append(m.tasks, task)
					m.items = append(m.items, item(task))
					m.list.InsertItem(len(m.items)+1, item(task))
				}
				// Clear input and return to normal mode
				m.textInput.Reset()
				m.addingProject = false

			case "tab":
				m.textInput.Reset()
				m.addingProject = false
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		if !m.addingProject && !m.taskView && !m.addingTask {
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

		if (msg.String() == "q" || msg.String() == "ctrl+c") && !m.addingProject {
			return m, tea.Quit
		}

	}
	return m, nil
}

func (m model) View() string {
	if m.addingProject {
		return fmt.Sprintf("Add a new project: %s\n\nPress Enter to submit or Tab to cancel.", m.textInput.View())
	}

	if m.addingTask {
		return fmt.Sprintf("Add a new task: %s\n\nPress Enter to submit or Tab to cancel.", m.taskInput.View())
	}

	// if m.taskView {
	// 	return "\n\n" + m.taskList.View()
	// }

	//var help string
	help := helpStyle.Render("Ctrl+A to add new task")
	//listView := backgroundStyle.Render(m.list.View()) experiment
	//listView := centered.Render(m.list.View()) //Uncomment for center

	if !m.addingProject {
		return "\n\n" + m.list.View() + "\n\n" + help
	}

	return "Welp"
}
