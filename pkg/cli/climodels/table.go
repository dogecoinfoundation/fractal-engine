package climodels

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type CliTableModel struct {
	Table table.Model
}

func (m CliTableModel) Init() tea.Cmd { return nil }

func (m CliTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m CliTableModel) View() string {
	return baseStyle.Render(m.Table.View()) + "\n"
}
