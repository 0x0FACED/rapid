package filepick

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type FilePickerModel struct {
	CurrDir      string
	SelectedFile string
	InitDir      string
	Filepicker   filepicker.Model
}

func NewFilePickerModel() FilePickerModel {
	fp := filepicker.New()
	fp.CurrentDirectory, _ = os.UserHomeDir()
	return FilePickerModel{
		Filepicker:   fp,
		InitDir:      "/",
		CurrDir:      fp.CurrentDirectory,
		SelectedFile: "",
	}
}

func (f *FilePickerModel) Init() tea.Cmd {
	return f.Filepicker.Init()
}

func (f *FilePickerModel) Update(msg tea.Msg) (FilePickerModel, tea.Cmd) {
	var cmd tea.Cmd
	f.Filepicker, cmd = f.Filepicker.Update(msg)
	f.CurrDir = f.Filepicker.CurrentDirectory
	return *f, cmd
}

func (f *FilePickerModel) View() string {
	var s strings.Builder
	s.WriteString("\n  ")
	s.WriteString(f.Filepicker.Styles.Selected.Render(f.CurrDir))
	s.WriteString("\n\n" + f.Filepicker.View() + "\n")
	return s.String()
}

func (f *FilePickerModel) DidSelectFile(msg tea.Msg) (bool, string) {
	return f.Filepicker.DidSelectFile(msg)
}
