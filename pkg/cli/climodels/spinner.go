package climodels

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SpinnerDoneMsg struct {
	Error error
}

type model struct {
	Spinner  spinner.Model
	quitting bool
	err      error
}

func NewSpinner() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{Spinner: s}
}

func (m model) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case SpinnerDoneMsg:
		m.quitting = true
		m.err = msg.Error
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", m.Spinner.View())
	if m.quitting {
		return str + "\n"
	}
	return str
}
