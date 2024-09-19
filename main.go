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
	DB_PATH := "./todo.db"

	var err error
	db, err = sql.Open("sqlite3", DB_PATH)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	InitDbSchema(db)

	//Test Code, delete db before running
	// insertQuery := "INSERT INTO projects(projectName) VALUES (?)"
	// _, err = db.Exec(insertQuery, "Getting Started")
	// if err != nil {
	// 	panic(err)
	// }

	initialItems, projectIDs := InitProjectList(db)

	p := tea.NewProgram(initialModel(initialItems, projectIDs), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

func InitDbSchema(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS projects (ID INTEGER PRIMARY KEY AUTOINCREMENT, projectName TEXT NOT NULL UNIQUE);")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS tasks ( taskID INTEGER PRIMARY KEY AUTOINCREMENT, Task TEXT NOT NULL, projectID INTEGER, FOREIGN KEY (projectID) REFERENCES projects(ID));")
	if err != nil {
		panic(err)
	}
}

const listHeight = 12

// Lipgloss styles cuz brat summer
var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#d4c6a9"))
	selectedItemStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("#1f1d19")).Background(lipgloss.Color("#d4c6a9")).PaddingRight(1)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	doneStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#70695a")).Strikethrough(true).PaddingLeft(4)
	title             = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#d4c6a9")).MarginLeft(0)
	addingStyle       = lipgloss.NewStyle().MarginLeft(4).MarginTop(1).Foreground(lipgloss.Color("#d4c6a9"))
	hotKeyStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#70695a")).MarginBottom(1).MarginLeft(2)
	creditStyle       = lipgloss.NewStyle().MarginBottom(1).MarginLeft(2)
	credInfoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#70695a"))
	//Credits Styles below
	// Main Gruvbox colors
	gruvboxYellow = "#fabd2f"
	gruvboxRed    = "#fb4934"
	gruvboxGreen  = "#b8bb26"
	gruvboxBlue   = "#83a598"
	gruvboxPurple = "#d3869b"
	gruvboxAqua   = "#8ec07c"

	// Styles with foreground and background having the same color
	StyleYellow = lipgloss.NewStyle().
			Foreground(lipgloss.Color(gruvboxYellow)).
			Background(lipgloss.Color(gruvboxYellow))

	StyleRed = lipgloss.NewStyle().
			Foreground(lipgloss.Color(gruvboxRed)).
			Background(lipgloss.Color(gruvboxRed))

	StyleGreen = lipgloss.NewStyle().
			Foreground(lipgloss.Color(gruvboxGreen)).
			Background(lipgloss.Color(gruvboxGreen))

	StyleBlue = lipgloss.NewStyle().
			Foreground(lipgloss.Color(gruvboxBlue)).
			Background(lipgloss.Color(gruvboxBlue))

	StylePurple = lipgloss.NewStyle().
			Foreground(lipgloss.Color(gruvboxPurple)).
			Background(lipgloss.Color(gruvboxPurple))

	StyleAqua = lipgloss.NewStyle().
			Foreground(lipgloss.Color(gruvboxAqua)).
			Background(lipgloss.Color(gruvboxAqua))
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

	// if strings.Contains(str, "X") {
	// 	fn = func(s ...string) string {
	// 		return doneStyle.Render(strings.Join(s, " "))
	// 	}
	// }

	if string(str[3]) == "X" {
		fn = func(s ...string) string {
			return doneStyle.Render(strings.Join(s, " "))
		}
	}
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
	projectID     []string
	choice        string
	currentProjID string
	done          []bool
	addingProject bool
	taskView      bool
	addingTask    bool
}

func InitProjectList(db *sql.DB) ([]list.Item, []string) {
	rows, err := db.Query("SELECT ID, projectName from projects")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var items []list.Item
	var projectIDs []string

	for rows.Next() {
		var ID string
		var projName string
		err = rows.Scan(&ID, &projName)
		if err != nil {
			panic(err)
		}
		items = append(items, item(projName))
		projectIDs = append(projectIDs, ID)
	}

	return items, projectIDs
}

func initialModel(initialItems []list.Item, projectIDs []string) model {
	ti := textinput.New()
	ti.Placeholder = "Type something"
	ti.CharLimit = 100
	ti.Width = 20

	taskIn := textinput.New()
	taskIn.Placeholder = "Enter a task"
	taskIn.CharLimit = 100
	taskIn.Width = 40

	items := initialItems
	projIDs := projectIDs

	const defaultWidth = 30

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "/projects"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
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
		projectID:     projIDs,
		currentProjID: "",
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

				selectedItem := string(i)
				if !strings.Contains(selectedItem, "X") {
					it := item("X ") + i
					CheckOffTask(db, selectedItem, string(it))
					m.list.SetItem(idx, it)
				}

			case "backspace":
				// refreshItems := []list.Item{}
				m.list.Title = "Project List"
				// m.list.SetItems(refreshItems)
				items, projIDs := InitProjectList(db) //Probably a database call to get the project list
				m.projectID = projIDs
				m.list.SetItems(items)
				m.taskView = false

			case "delete":
				idx := m.list.Index()
				currentTask := string(m.list.SelectedItem().(item))
				DeleteTask(db, currentTask)
				taskList := GetTasks(db, m.choice)
				m.list.RemoveItem(idx)
				m.list.SetItems(taskList)
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
					AddTask(db, task, m.currentProjID)
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
				//task here refers to project name, to be refactored
				task := strings.TrimSpace(m.textInput.Value())

				insertQuery := "INSERT INTO projects(projectName) VALUES (?);"
				res, err := db.Exec(insertQuery, task)
				if err != nil {
					panic(err)
				}
				lastInsertID, err := res.LastInsertId()
				if err != nil {
					panic("Last Insert ID")
				}
				m.projectID = append(m.projectID, string(lastInsertID))

				if task != "" {
					projects, projIDs := InitProjectList(db)
					m.projectID = projIDs
					m.list.SetItems(projects)
					m.tasks = append(m.tasks, task)
					m.items = append(m.items, item(task))
					//m.list.InsertItem(len(m.items)+1, item(task))
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

				m.choice = string(i) //Set the chosen project so it can be referenced later
				m.currentProjID = m.projectID[m.list.Index()]

				var taskList []list.Item
				taskList = GetTasks(db, string(i))

				m.list.Title = string("$/" + "projects/" + i + "/task-list")
				m.list.SetItems(taskList)
				m.taskView = true
			case "delete":
				idx := m.list.Index()
				m.currentProjID = m.projectID[idx]
				DeleteProject(db, m.currentProjID)
				m.list.RemoveItem(idx)
				items, projIDs := InitProjectList(db)
				m.projectID = projIDs
				m.list.SetItems(items)
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

func AddProject(db *sql.DB, projectName string) (int64, error) {
	insertQuery := "INSERT INTO projects(projectName) VALUES (?);"
	res, err := db.Exec(insertQuery, projectName)
	if err != nil {
		panic(err)
	}

	return res.LastInsertId()
}

func DeleteProject(db *sql.DB, projectID string) {
	taskDeletionQuery := "DELETE FROM tasks WHERE projectID = ?"

	projectDeletionQuery := "DELETE FROM projects WHERE ID = ?"

	_, err := db.Exec(taskDeletionQuery, projectID)
	if err != nil {
		panic("In deletion of all tasks related to current project")
	}

	_, err = db.Exec(projectDeletionQuery, projectID)
	if err != nil {
		panic("In deletion of project")
	}
}

func GetTasks(db *sql.DB, projectName string) []list.Item {
	taskQuery := "SELECT Task from tasks WHERE projectID = (SELECT ID FROM projects WHERE projectName = ?)"
	rows, err := db.Query(taskQuery, projectName)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var items []list.Item

	for rows.Next() {
		var task string
		err = rows.Scan(&task)
		if err != nil {
			panic(err)
		}
		items = append(items, item(task))
	}

	return items
}

func AddTask(db *sql.DB, task string, projectID string) {
	insertQuery := "INSERT INTO tasks(Task, projectID) VALUES (?, ?);"
	_, err := db.Exec(insertQuery, task, projectID)
	if err != nil {
		panic(err)
	}
}

func DeleteTask(db *sql.DB, task string) {
	query := "DELETE FROM tasks WHERE Task = ?"
	_, err := db.Exec(query, task)
	if err != nil {
		panic("Error while deleting")
	}
}

func CheckOffTask(db *sql.DB, task string, newTask string) {
	query := "UPDATE tasks SET Task = ? WHERE Task = ?"
	_, err := db.Exec(query, newTask, task)
	if err != nil {
		panic("Error checking the box")
	}
}

func CreditStringBuilder() string {
	green := StyleGreen.Render("__")
	blue := StyleBlue.Render("__")
	yellow := StyleYellow.Render("__")
	red := StyleRed.Render("__")
	aqua := StyleAqua.Render("__")
	credInfo := credInfoStyle.Render(" Todo-TUI v0.1.0 //afn")

	return creditStyle.Render("\n" + green + blue + yellow + red + aqua + credInfo)
}

func (m model) View() string {
	if m.addingProject {
		return addingStyle.Render(fmt.Sprintf("Add a new project: %s\n\nPress Enter to submit or Tab to cancel.\n\nWARNING: No duplicate project names", m.textInput.View()))
	}

	if m.addingTask {
		return addingStyle.Render(fmt.Sprintf("Add a new task: %s\n\nPress Enter to submit or Tab to cancel.", m.taskInput.View()))
	}

	//var help string
	help := hotKeyStyle.Render("ctrl+a: add task/proj  del: delete task/proj  esc/q/ctrl+c: quit  enter: open proj/complete task")
	var credits string = CreditStringBuilder()

	if !m.addingProject {
		return "\n\n" + m.list.View() + "\n\n" + help + credits
	}

	return "Welp"
}
