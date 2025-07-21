package climodels

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type SimpleListItem struct {
	Name, Desc string
}

func (i SimpleListItem) Title() string       { return i.Name }
func (i SimpleListItem) Description() string { return i.Desc }
func (i SimpleListItem) FilterValue() string { return i.Name }

type SimpleListModel struct {
	List list.Model
}

func (m SimpleListModel) Init() tea.Cmd {
	return nil
}

func (m SimpleListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m SimpleListModel) View() string {
	return docStyle.Render(m.List.View())
}
