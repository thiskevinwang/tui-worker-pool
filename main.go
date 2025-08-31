package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"tui-worker-pool/logs"

	"github.com/chdb-io/chdb-go/chdb"
	slogmulti "github.com/samber/slog-multi"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Agent struct {
	model      spinner.Model
	shouldSpin bool
	message    string
}

type model struct {
	logger *slog.Logger

	// the program will send tasks to this channel
	// for external workers to consume.
	// Note: This is producer-consumer pattern, not publisher-subscriber.
	outbound chan string

	textInput textinput.Model

	a1 Agent
	a2 Agent
	a3 Agent
	a4 Agent
	a5 Agent
}

// MARK: Init
func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

// MARK: Update
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.logger.Info("Update", "tea_msg", msg)
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case WorkerResult:
		{
			switch msg.id {
			case 1:
				m.a1.message = msg.message
				m.a1.shouldSpin = msg.spinning
				m.a1.model, cmd = m.a1.model.Update(spinner.TickMsg{})
			case 2:
				m.a2.message = msg.message
				m.a2.shouldSpin = msg.spinning
				m.a2.model, cmd = m.a2.model.Update(spinner.TickMsg{})
			case 3:
				m.a3.message = msg.message
				m.a3.shouldSpin = msg.spinning
				m.a3.model, cmd = m.a3.model.Update(spinner.TickMsg{})
			case 4:
				m.a4.message = msg.message
				m.a4.shouldSpin = msg.spinning
				m.a4.model, cmd = m.a4.model.Update(spinner.TickMsg{})
			case 5:
				m.a5.message = msg.message
				m.a5.shouldSpin = msg.spinning
				m.a5.model, cmd = m.a5.model.Update(spinner.TickMsg{})
			}
			return m, cmd
		}

	// Handle manual key presses
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.outbound <- m.textInput.Value()
			m.textInput.Reset()
		}

	// Handle perpetual spinner
	case spinner.TickMsg:
		var cmd1 tea.Cmd
		var cmd2 tea.Cmd
		var cmd3 tea.Cmd
		var cmd4 tea.Cmd
		var cmd5 tea.Cmd
		if m.a1.shouldSpin {
			m.a1.model, cmd1 = m.a1.model.Update(msg)
		}
		if m.a2.shouldSpin {
			m.a2.model, cmd2 = m.a2.model.Update(msg)
		}
		if m.a3.shouldSpin {
			m.a3.model, cmd3 = m.a3.model.Update(msg)
		}
		if m.a4.shouldSpin {
			m.a4.model, cmd4 = m.a4.model.Update(msg)
		}
		if m.a5.shouldSpin {
			m.a5.model, cmd5 = m.a5.model.Update(msg)
		}
		return m, tea.Batch(cmd1, cmd2, cmd3, cmd4, cmd5)
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

var greyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render

// MARK: View
func (m *model) View() string {
	// This is human slop :P
	// TODO: is there a DRYer way to handle multiple TUI worker-representations?
	// ...and also a dynamic number of them too?
	worker1 := greyStyle("Agent 001:")
	worker1Msg := greyStyle(m.a1.message)
	if m.a1.shouldSpin {
		worker1 = "Agent 001: " + m.a1.model.View()
		worker1Msg = m.a1.message
	}

	worker2 := greyStyle("Agent 002:")
	worker2Msg := greyStyle(m.a2.message)
	if m.a2.shouldSpin {
		worker2 = "Agent 002: " + m.a2.model.View()
		worker2Msg = m.a2.message
	}

	worker3 := greyStyle("Agent 003:")
	worker3Msg := greyStyle(m.a3.message)
	if m.a3.shouldSpin {
		worker3 = "Agent 003: " + m.a3.model.View()
		worker3Msg = m.a3.message
	}

	worker4 := greyStyle("Agent 004:")
	worker4Msg := greyStyle(m.a4.message)
	if m.a4.shouldSpin {
		worker4 = "Agent 004: " + m.a4.model.View()
		worker4Msg = m.a4.message
	}

	worker5 := greyStyle("Agent 005:")
	worker5Msg := greyStyle(m.a5.message)
	if m.a5.shouldSpin {
		worker5 = "Agent 005: " + m.a5.model.View()
		worker5Msg = m.a5.message
	}

	return m.textInput.View() +
		"\n\n" +
		worker1 + " " + worker1Msg + "\n" +
		worker2 + " " + worker2Msg + "\n" +
		worker3 + " " + worker3Msg + "\n" +
		worker4 + " " + worker4Msg + "\n" +
		worker5 + " " + worker5Msg
}

// MARK: WorkerResult
type WorkerResult struct {
	id       int
	message  string
	spinning bool
}

func worker(id int, jobs <-chan string, p *tea.Program) {
	for j := range jobs {
		// notify the tea program
		p.Send(WorkerResult{
			id:       id,
			message:  fmt.Sprintf("Doing work... %q", j),
			spinning: true,
		})

		// simulate expensive work
		time.Sleep(5 * time.Second)

		// notify the tea program
		p.Send(WorkerResult{
			id:       id,
			message:  fmt.Sprintf("Finished: %q", j),
			spinning: false,
		})
	}
}

// Visualization for https://gobyexample.com/worker-pools
func main() {
	// Clear the terminal
	fmt.Print("\033[H\033[2J")

	// Create tmp dir if it doesn't exist
	if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
		log.Fatalf("error creating tmp dir: %v", err)
	}

	// fanout debug logs to in-process ClickHouse and local logs
	session, err := chdb.NewSession("./tmp/chdb")
	if err != nil {
		log.Fatalf("error creating chdb session: %v", err)
	}
	defer session.Close()
	chLogHandler := logs.NewClickHouseHandler(session)

	// Create a log file to dump debug logs to since the terminal
	// itself is reserved for the user interface.
	// If it already exists, truncate it first.
	f, err := os.OpenFile("tmp/log.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	textFileLogHandler := slog.NewTextHandler(f, nil)

	// Create a new log file for json formatted logs
	jsonFile, err := os.OpenFile("tmp/log.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer jsonFile.Close()
	jsonLogHandler := slog.NewJSONHandler(jsonFile, nil)

	// Create a channel for workers to consume tasks from
	const numJobs = 5
	jobs := make(chan string, numJobs)
	defer close(jobs)

	ti := textinput.New()
	ti.Placeholder = "Provide a task..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	// Initialize the TUI program
	p := tea.NewProgram(&model{
		textInput: ti,
		a1:        Agent{model: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("1"))))},
		a2:        Agent{model: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("2"))))},
		a3:        Agent{model: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("3"))))},
		a4:        Agent{model: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("4"))))},
		a5:        Agent{model: spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("5"))))},
		logger: slog.New(
			slogmulti.Fanout(
				textFileLogHandler,
				jsonLogHandler,
				chLogHandler,
			),
		),
		outbound: jobs,
	})

	// This starts up N workers, initially blocked because there are no jobs yet.
	for w := 1; w <= 5; w++ {
		go worker(w, jobs, p)
	}

	_, err = p.Run()
	if err != nil {
		log.Fatalf("error running program: %v", err)
	}
	defer p.Quit()
}
