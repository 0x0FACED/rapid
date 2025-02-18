package app

import (
	"context"
	"log"
	"strconv"
	"sync"

	"github.com/0x0FACED/rapid/internal/app/filepick"
	"github.com/0x0FACED/rapid/internal/app/tables"
	"github.com/0x0FACED/rapid/internal/lan/client"
	"github.com/0x0FACED/rapid/internal/lan/server"
	mod "github.com/0x0FACED/rapid/internal/model"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keymap struct {
	exit    tea.Key
	scan    tea.Key
	options tea.Key
	share   tea.Key
	choose  tea.Key
}

const (
	selectedAddrs int = iota
	selectedFiles
	selectedSharedFiles
	selectedFilePicker
)

type App struct {
	selected        int
	lastSelected    int
	tables          tables.TablesModel
	filePicker      filepick.FilePickerModel
	client          *client.LANClient
	server          *server.LANServer
	keymap          keymap
	currAddr        string
	currRecvFiles   []mod.File
	currSharedFiles []mod.File
}

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func NewApp(c *client.LANClient, s *server.LANServer) App {
	return App{
		tables:          tables.NewTablesModel(),
		currRecvFiles:   make([]mod.File, 0),
		currSharedFiles: make([]mod.File, 0),
		filePicker:      filepick.NewFilePickerModel(),
		client:          c,
		server:          s,
		currAddr:        "/",
		keymap: keymap{
			scan:    tea.Key{Type: tea.KeyRunes, Runes: []rune("r")},
			share:   tea.Key{Type: tea.KeyRunes, Runes: []rune("s")},
			choose:  tea.Key{Type: tea.KeyEnter},
			options: tea.Key{Type: tea.KeyRunes, Runes: []rune("N")},
			exit:    tea.Key{Type: tea.KeyRunes, Runes: []rune("q")},
		},
	}
}

func (m App) Init() tea.Cmd {
	return m.filePicker.Init()
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "ctrl+q":
			if m.selected == selectedFilePicker {
				m.selected = m.lastSelected
			}
		case "down", "up":
			if m.selected == selectedAddrs {
				cmds = append(cmds, m.tables.Update(0, msg))
				return m, tea.Batch(cmds...)
			} else if m.selected == selectedFiles {
				cmds = append(cmds, m.tables.Update(1, msg))
				return m, tea.Batch(cmds...)
			} else if m.selected == selectedSharedFiles {
				cmds = append(cmds, m.tables.Update(2, msg))
				return m, tea.Batch(cmds...)
			}
		case "tab":
			if m.selected == selectedAddrs {
				m.selected = selectedFiles
				m.lastSelected = m.selected
				m.tables.NextTable(selectedAddrs, selectedFiles)
				return m, tea.Batch(cmds...)
			} else if m.selected == selectedFiles {
				m.selected = selectedSharedFiles
				m.lastSelected = m.selected
				m.tables.NextTable(selectedFiles, selectedSharedFiles)
				return m, tea.Batch(cmds...)
			} else {
				m.selected = selectedAddrs
				m.lastSelected = m.selected
				m.tables.NextTable(selectedSharedFiles, selectedAddrs)
				return m, tea.Batch(cmds...)
			}
		case "f":
			m.selected = selectedFilePicker
		case "enter":
			if m.selected == selectedAddrs {
				row := m.tables.Content[0].SelectedRow()
				if len(row) == 0 {
					return m, nil
				}
				addr := string(row[2])
				m.currAddr = addr
				return m, m.fetchFiles(addr)
			} else if m.selected == selectedFiles {
				row := m.tables.Content[1].SelectedRow()
				if len(row) == 0 {
					return m, nil
				}
				id := string(row[1])
				name := string(row[2])
				return m, m.downloadFile(m.currAddr, id, name)
			} else if m.selected == selectedSharedFiles {
				row := m.tables.Content[2].SelectedRow()
				if len(row) == 0 {
					return m, nil
				}
				id := string(row[1])
				name := string(row[2])
				_ = id
				_ = name

				// TODO: add delete command
				return m, nil
			}
		case "r":
			return m, m.updateServers()
		}
	case serversUpdateMsg:
		m.tables.SetAddressRows(msg.rows)
	}

	m.filePicker, cmd = m.filePicker.Update(msg)
	cmds = append(cmds, cmd)
	if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
		m.filePicker.SelectedFile = path
		m.selected = selectedAddrs
		cmds = append(cmds, m.shareFile())
	}

	return m, tea.Batch(cmds...)
}

func (m App) View() string {
	var fullView string
	if m.selected == selectedFilePicker {
		return m.filePicker.View()
	}

	fullView += m.tables.View()

	return fullView
}

func (m App) shareFile() tea.Cmd {
	file, err := m.server.ShareLocal(m.filePicker.SelectedFile)
	if err != nil {
		log.Println("failed share file:", err)
		return nil
	}

	rows := []table.Row{}
	var i int
	if len(m.currSharedFiles) != 0 {
		for i, f := range m.currRecvFiles {
			rows = append(rows, []string{strconv.Itoa(i + 1), f.ID, f.Name, strconv.FormatInt(f.Size, 10) + " bytes"})
		}
	}

	m.currSharedFiles = append(m.currSharedFiles, mod.File{ID: file.ID, Name: file.Name, Path: file.Path, Size: file.Size})
	rows = append(rows, []string{strconv.Itoa(i + 1), file.ID, file.Name, strconv.FormatInt(file.Size, 10) + " bytes"})
	m.tables.SetSharedFilesRows(rows)
	return nil
}

func (m App) fetchFiles(serverAddr string) tea.Cmd {
	files, err := m.client.GetFiles(serverAddr)
	if err != nil {
		log.Println("failed get files:", err)
		return nil
	}

	rows := []table.Row{}
	m.currRecvFiles = make([]mod.File, len(files))
	for i, f := range files {
		m.currRecvFiles[i] = mod.File{ID: f.ID, Name: f.Name, Path: f.Path, Size: f.Size}
		rows = append(rows, []string{strconv.Itoa(i + 1), f.ID, f.Name, strconv.FormatInt(f.Size, 10) + " bytes"})
	}

	m.tables.SetFileRows(rows)
	return nil
}

var wg sync.WaitGroup

type serversUpdateMsg struct {
	rows []table.Row
}

func (m *App) updateServers() tea.Cmd {
	ch := make(chan mod.ServiceInstance, 10)
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(ch)
		defer cancel()

		m.client.DiscoverServers(ctx, ch)
	}()

	return func() tea.Msg {
		servers := make(map[string]string)

		for e := range ch {
			servers[e.InstanceName] = e.IPv4
		}

		rows := []table.Row{}
		var c int = 1
		for name, addr := range servers {
			rows = append(rows, table.Row{strconv.Itoa(c), name, addr})
			c++
		}

		wg.Wait()

		return serversUpdateMsg{rows: rows}
	}
}

func (m App) downloadFile(addr, id, filename string) tea.Cmd {
	err := m.client.DownloadFile(addr, id, filename, ".")
	if err != nil {
		log.Println("failed download file:", err)
		return nil
	}

	return nil
}
