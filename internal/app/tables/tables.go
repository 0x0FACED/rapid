package tables

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type TablesModel struct {
	Content []table.Model
}

func NewTablesModel() TablesModel {
	return TablesModel{
		Content: []table.Model{
			initAddrsTable(),
			initFilesTable(),
			initSharedFilesTable(),
		},
	}
}

func (t *TablesModel) SetAddressRows(rows []table.Row) {
	t.Content[0].SetRows(rows)
}

func (t *TablesModel) SetFileRows(rows []table.Row) {
	t.Content[1].SetRows(rows)
}

func (t *TablesModel) SetSharedFilesRows(rows []table.Row) {
	t.Content[2].SetRows(rows)
}

func (t *TablesModel) Update(index int, msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	t.Content[index], cmd = t.Content[index].Update(msg)
	return cmd
}

func (t TablesModel) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		defaultStyle.Render(t.Content[0].View()),
		defaultStyle.Render(t.Content[1].View()),
		defaultStyle.Render(t.Content[2].View()),
		"\n\n")
}

func (t TablesModel) NextTable(prev, next int) {
	t.Content[prev].Blur()
	t.Content[prev].SetStyles(headerStyle())
	t.Content[next].Focus()
	t.Content[next].SetStyles(selectedStyle())
}

func initAddrsTable() table.Model {
	columns := []table.Column{
		{Title: "№", Width: 3},
		{Title: "Name", Width: 15},
		{Title: "Address", Width: 15},
	}

	rows := []table.Row{
		{"1", "2", "3"},
		{"1", "2", "3"},
		{"1", "2", "3"},
		{"1", "2", "3"},
		{"1", "2", "3"},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	t.SetStyles(selectedStyle())
	return t
}

func initFilesTable() table.Model {
	columns := []table.Column{
		{Title: "№", Width: 3},
		{Title: "UUID", Width: 10},
		{Title: "Filename", Width: 20},
		{Title: "Size", Width: 5},
	}
	rows := []table.Row{
		{"1", "2", "3"},
		{"1", "2", "3"},
		{"1", "2", "3"},
		{"1", "2", "3"},
		{"1", "2", "3"},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(10),
		table.WithFocused(false),
	)
	t.SetStyles(headerStyle())
	t.Blur()
	return t
}

func initSharedFilesTable() table.Model {
	columns := []table.Column{
		{Title: "№", Width: 3},
		{Title: "UUID", Width: 10},
		{Title: "Filename", Width: 20},
		{Title: "Size", Width: 5},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithHeight(10),
		table.WithFocused(false),
	)
	t.SetStyles(headerStyle())
	t.Blur()
	return t
}

func selectedStyle() table.Styles {
	tableStyles := table.DefaultStyles()

	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)

	tableStyles.Selected = tableStyles.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	return tableStyles
}

func headerStyle() table.Styles {
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	return tableStyles
}
