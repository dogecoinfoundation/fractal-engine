package climodels

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var selectBaseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type CliSelectTableModel struct {
	Table table.Model
}

func (m CliSelectTableModel) Init() tea.Cmd {
	return nil
}

func (m CliSelectTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.Table.Focused() {
				m.Table.Blur()
			} else {
				m.Table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}
	}
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

func (m CliSelectTableModel) View() string {
	return baseStyle.Render(m.Table.View()) + "\n"
}
