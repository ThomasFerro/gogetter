package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	helpHeight = 5
)

var gogetter app.Gogetter

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())
)

type keymap = struct {
	next, prev, execute, quit key.Binding
}

func requestTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Type your request"
	t.ShowLineNumbers = true
	t.Cursor.Style = cursorStyle
	t.FocusedStyle.Placeholder = focusedPlaceholderStyle
	t.BlurredStyle.Placeholder = placeholderStyle
	t.FocusedStyle.CursorLine = cursorLineStyle
	t.FocusedStyle.Base = focusedBorderStyle
	t.BlurredStyle.Base = blurredBorderStyle
	t.FocusedStyle.EndOfBuffer = endOfBufferStyle
	t.BlurredStyle.EndOfBuffer = endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.Blur()
	return t
}

func responseTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = ""
	t.ShowLineNumbers = true
	t.Cursor.Style = cursorStyle
	t.FocusedStyle.Placeholder = focusedPlaceholderStyle
	t.BlurredStyle.Placeholder = placeholderStyle
	t.FocusedStyle.CursorLine = cursorLineStyle
	t.FocusedStyle.Base = focusedBorderStyle
	t.BlurredStyle.Base = blurredBorderStyle
	t.FocusedStyle.EndOfBuffer = endOfBufferStyle
	t.BlurredStyle.EndOfBuffer = endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.Blur()
	return t
}

type focusedArea int

const (
	RequestArea focusedArea = iota
	ResponseArea
)

type model struct {
	width            int
	height           int
	keymap           keymap
	help             help.Model
	requestTextarea  textarea.Model
	responseTextarea textarea.Model
	focusedArea      focusedArea
	ongoingRequest   bool
}

func newModel() model {
	m := model{
		requestTextarea:  requestTextarea(),
		responseTextarea: responseTextarea(),
		help:             help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			execute: key.NewBinding(
				key.WithKeys("alt+enter"),
				key.WithHelp("alt+enter", "execute"),
			),

			prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}

	m.requestTextarea.Focus()
	m.focusedArea = RequestArea
	return m
}

type newRequestMsg struct{}

type responseMsg struct {
	response string
	err      error
}

func (m model) newRequest() (model, []tea.Cmd) {
	return m, []tea.Cmd{func() tea.Msg {
		return newRequestMsg{}
	}}
}

func (m model) executeRequest() (model, []tea.Cmd) {
	if m.ongoingRequest {
		return m, []tea.Cmd{}
	}
	m.ongoingRequest = true
	return m, []tea.Cmd{func() tea.Msg {
		input := strings.Split(m.requestTextarea.Value(), " ")
		if len(input) < 2 {
			return nil
		}
		method := input[0]
		url := input[1]
		var resp *http.Response
		var err error
		gogetter, resp, err = gogetter.Execute(method, url)
		var response []byte

		if resp != nil {
			defer resp.Body.Close()
			var err error
			response, err = io.ReadAll(resp.Body)
			if err != nil {
				return responseMsg{err: err, response: ""}
			}
		}

		return responseMsg{err: err, response: string(response)}
	}}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) SwitchFocus() (model, tea.Cmd) {
	if m.focusedArea == RequestArea {
		m.focusedArea = ResponseArea
		m.requestTextarea.Blur()
		return m, m.responseTextarea.Focus()
	}
	m.focusedArea = RequestArea
	m.responseTextarea.Blur()
	return m, m.requestTextarea.Focus()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.execute):
			var executeRequestCommands []tea.Cmd
			m, executeRequestCommands = m.newRequest()
			cmds = append(cmds, executeRequestCommands...)
		case key.Matches(msg, m.keymap.next):
			var focusCmd tea.Cmd
			m, focusCmd = m.SwitchFocus()
			cmds = append(cmds, focusCmd)
		case key.Matches(msg, m.keymap.prev):
			var focusCmd tea.Cmd
			m, focusCmd = m.SwitchFocus()
			cmds = append(cmds, focusCmd)
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case newRequestMsg:
		m.responseTextarea.SetValue("Pending request...")

		var executeRequestCommands []tea.Cmd
		m, executeRequestCommands = m.executeRequest()
		cmds = append(cmds, executeRequestCommands...)
	case responseMsg:
		responseTextareaValue := ""
		response := responseMsg(msg)
		if response.err != nil {
			responseTextareaValue = fmt.Sprintf("%v\n", response.err.Error())
		}
		responseTextareaValue = fmt.Sprintf("%v%v", responseTextareaValue, response.response)
		m.responseTextarea.SetValue(responseTextareaValue)
		m.ongoingRequest = false
	}

	m.sizeInputs()

	if !m.ongoingRequest {
		newModel, cmd := m.requestTextarea.Update(msg)
		m.requestTextarea = newModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	m.requestTextarea.SetWidth(m.width / 2)
	m.requestTextarea.SetHeight(m.height - helpHeight)
	m.responseTextarea.SetWidth(m.width / 2)
	m.responseTextarea.SetHeight(m.height - helpHeight)
}

func (m model) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.execute,
		m.keymap.next,
		m.keymap.prev,
		m.keymap.quit,
	})

	var views []string
	views = append(views, m.requestTextarea.View())
	views = append(views, m.responseTextarea.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help
}

func main() {
	historyFileReader, err := os.Open("gogetter_history.json")
	defer historyFileReader.Close()
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Error("error while opening history file", slog.Any("error", err))
			os.Exit(1)
		}
	}
	historyWritingFunc := func(toWrite []byte) error {
		historyFileWriter, err := os.Create("gogetter_history.json")
		if err != nil {
			return err
		}
		defer historyFileWriter.Close()
		_, err = historyFileWriter.Write(toWrite)
		return err
	}
	gogetter, err = app.NewGogetter(http.DefaultClient, historyFileReader, historyWritingFunc)
	if err != nil {
		slog.Error("error while creating new gogetter", slog.Any("error", err))
		os.Exit(1)
	}
	if _, err = tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		slog.Error("error while running program", slog.Any("error", err))
		os.Exit(1)
	}
}
