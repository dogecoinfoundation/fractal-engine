package climodels

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var selectDocStyle = lipgloss.NewStyle().Margin(1, 2)

type SelectSimpleListItem struct {
	OfferId    string
	Name, Desc string
}

func (i SelectSimpleListItem) Title() string       { return i.Name }
func (i SelectSimpleListItem) Description() string { return i.Desc }
func (i SelectSimpleListItem) FilterValue() string { return i.Name }

type SelectSimpleListModel struct {
	List list.Model
}

func (m SelectSimpleListModel) Init() tea.Cmd {
	return nil
}

func (m SelectSimpleListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := selectDocStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m SelectSimpleListModel) View() string {
	return docStyle.Render(m.List.View())
}
