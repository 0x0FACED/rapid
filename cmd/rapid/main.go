package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/0x0FACED/rapid/configs"
	"github.com/0x0FACED/rapid/internal/lan/client"
	"github.com/0x0FACED/rapid/internal/lan/mdnss"
	"github.com/0x0FACED/rapid/internal/lan/server"
	mod "github.com/0x0FACED/rapid/internal/model"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type keymap struct {
	exit    key.Binding
	scan    key.Binding
	options key.Binding
	share   key.Binding
	choose  key.Binding
}

const (
	selectedAddrs int = iota
	selectedFiles
)

type model struct {
	selectedTable int
	content       []table.Model
	logs          viewport.Model
	client        *client.LANClient
	server        *server.LANServer
	keymap        keymap
	help          help.Model
	currAddr      string
}

var (
	addrsStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	filesStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
)

func newModel(c *client.LANClient, addrs table.Model, files table.Model) model {
	cont := make([]table.Model, 2)
	cont[0] = addrs
	cont[1] = files
	return model{
		content: cont,
		logs:    viewport.New(80, 5),
		client:  c,
		help:    help.New(),
		keymap: keymap{
			scan: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "scan"),
			),
			share: key.NewBinding(
				key.WithKeys("shift+s"),
				key.WithHelp("shift+s", "choose file to share"),
			),
			choose: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "choose file"),
			),
			options: key.NewBinding(
				key.WithKeys("shift+n"),
				key.WithHelp("shift+n", "open options"),
			),
			exit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "exit"),
			),
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "tab":
			if m.selectedTable == selectedAddrs {
				m.selectedTable = 1
				m.content[0].Blur()
				m.content[0].SetStyles(headerSyle())
				m.content[1].Focus()
				m.content[1].SetStyles(selectedStyle())
			} else {
				m.selectedTable = 0
				m.content[1].Blur()
				m.content[1].SetStyles(headerSyle())
				m.content[0].Focus()
				m.content[0].SetStyles(selectedStyle())
			}
		case "enter":
			if m.selectedTable == selectedAddrs {
				row := m.content[0].SelectedRow()
				addr := string(row[2])
				m.currAddr = addr
				return m, m.fetchFiles(addr)
			} else if m.selectedTable == selectedFiles {
				row := m.content[1].SelectedRow()
				id := string(row[1])
				name := string(row[2])
				return m, m.downloadFile(m.currAddr, id, name)
			}
		case "r":
			return m, m.updateServers()
		}
	}

	m.content[0], cmd = m.content[0].Update(msg)
	cmds = append(cmds, cmd)
	m.content[1], cmd = m.content[1].Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return m.render()
}

func (m model) render() string {
	addrs := lipgloss.JoinHorizontal(lipgloss.Center, addrsStyle.Render(m.content[0].View()+"\n"), addrsStyle.Render(m.content[1].View()+"\n"))
	return addrs
}

func (m *model) updateServers() tea.Cmd {
	ch := make(chan mod.ServiceInstance, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	m.client.DiscoverServers(ctx, ch)
	servers := make(map[string]string)

	select {
	case e := <-ch:
		servers[e.InstanceName] = e.IPv4
	}

	rows := []table.Row{}
	var c int = 1
	for name, addr := range servers {
		rows = append(rows, table.Row{strconv.Itoa(c), name, addr})
		c++
	}

	m.content[0].SetRows(rows)

	m.content[0].Focus()

	return nil
}

func (m model) downloadFile(addr, id, filename string) tea.Cmd {
	err := m.client.DownloadFile(addr, id, filename, ".")
	if err != nil {
		log.Println("failed download file:", err)
		return nil
	}

	return nil
}

func (m model) fetchFiles(serverAddr string) tea.Cmd {
	files, err := m.client.GetFiles(serverAddr)
	if err != nil {
		log.Println("failed get files:", err)
		return nil
	}
	rows := []table.Row{}
	for i, f := range files {
		rows = append(rows, []string{strconv.Itoa(i + 1), f.ID, f.Name, f.Path, strconv.FormatInt(f.Size, 10) + " bytes"})
	}
	m.content[1].SetRows(rows)
	return nil
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

func headerSyle() table.Styles {
	tableStyles := table.DefaultStyles()

	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)

	return tableStyles
}
func initAddrsTable() table.Model {
	columns := []table.Column{
		{Title: "№", Width: 3},
		{Title: "Name", Width: 40},
		{Title: "Address", Width: 20},
	}
	rows := []table.Row{}
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
		{Title: "ID", Width: 10},
		{Title: "Filename", Width: 40},
		{Title: "Path", Width: 40},
		{Title: "Size", Width: 10},
	}
	rows := []table.Row{}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(10),
		table.WithFocused(false),
	)

	tableStyles := table.DefaultStyles()

	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)

	t.SetStyles(tableStyles)

	t.Blur()
	return t
}
func main() {
	_uuid := uuid.New()
	mdnss, err := mdnss.New(_uuid.String(), 8080)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	addrs := initAddrsTable()
	files := initFilesTable()
	c := client.New(mdnss)
	s := server.New(configs.LANServerConfig{DownloadsDir: "./test-dir"})
	go s.Start()

	p := tea.NewProgram(newModel(c, addrs, files), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}
}
