package tui

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	helpHeight = 5
)

var Gogetter app.Gogetter

type keymap = struct {
	next, prev, execute, toggleHistory, quit, enter key.Binding
}

type focusedArea int

const (
	RequestArea focusedArea = iota
	ResponseArea
	HistoryArea
)

type model struct {
	width            int
	height           int
	keymap           keymap
	help             help.Model
	requestTextarea  textarea.Model
	responseTextarea textarea.Model
	history          list.Model
	focusedArea      focusedArea
	ongoingRequest   bool
	displayHistory   bool
}

func NewModel(gogetter app.Gogetter) model {
	Gogetter = gogetter
	requestTextarea := newTextarea()
	requestTextarea.Placeholder = "Type your request"
	l := newHistoryList(gogetter.History())
	m := model{
		requestTextarea:  requestTextarea,
		responseTextarea: newTextarea(),
		help:             help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			toggleHistory: key.NewBinding(
				key.WithKeys("alt+h"),
				key.WithHelp("alt+h", "toggle history"),
			),
			execute: key.NewBinding(
				key.WithKeys("alt+enter"),
				key.WithHelp("alt+enter", "execute"),
			),
			enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "enter"),
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
		displayHistory: false,
		history:        l,
	}

	m.requestTextarea.Focus()
	m.focusedArea = RequestArea
	return m
}

type newRequestMsg struct{}

type responseMsg struct {
	response string
	request  app.Request
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
		// TODO: Extract and extend the parsing
		input := strings.Split(m.requestTextarea.Value(), " ")
		if len(input) < 2 {
			return nil
		}
		method := input[0]
		url := input[1]
		var resp *http.Response
		var err error
		var request app.Request
		Gogetter, request, resp, err = Gogetter.Execute(method, url)
		var response []byte

		if resp != nil {
			defer resp.Body.Close()
			var err error
			response, err = io.ReadAll(resp.Body)
			if err != nil {
				return responseMsg{err: err, request: request, response: ""}
			}
		}

		return responseMsg{err: err, request: request, response: string(response)}
	}}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) SwitchFocus(next bool) (model, []tea.Cmd) {
	if next {
		return m.FocusNextTab()
	}
	return m.FocusPreviousTab()
}

func (m model) FocusPreviousTab() (model, []tea.Cmd) {
	if m.focusedArea == ResponseArea {
		m.focusedArea = RequestArea
		m.responseTextarea.Blur()
		return m, []tea.Cmd{m.requestTextarea.Focus()}
	}

	if m.focusedArea == RequestArea {
		m.requestTextarea.Blur()
		if m.displayHistory {
			m.focusedArea = HistoryArea
			return m, []tea.Cmd{}
		}

		m.focusedArea = ResponseArea
		return m, []tea.Cmd{m.responseTextarea.Focus()}
	}

	m.focusedArea = ResponseArea
	return m, []tea.Cmd{m.responseTextarea.Focus()}
}

func (m model) FocusNextTab() (model, []tea.Cmd) {
	if m.focusedArea == RequestArea {
		m.focusedArea = ResponseArea
		m.requestTextarea.Blur()
		return m, []tea.Cmd{m.responseTextarea.Focus()}
	}

	if m.focusedArea == ResponseArea {
		m.responseTextarea.Blur()
		if m.displayHistory {
			m.focusedArea = HistoryArea
			return m, []tea.Cmd{}
		}

		m.focusedArea = RequestArea
		return m, []tea.Cmd{m.requestTextarea.Focus()}
	}

	m.focusedArea = RequestArea
	return m, []tea.Cmd{m.requestTextarea.Focus()}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.toggleHistory):
			m.displayHistory = !m.displayHistory
			if m.displayHistory {
				m.focusedArea = HistoryArea
				m.responseTextarea.Blur()
				m.requestTextarea.Blur()
				return m, nil
			}
			if m.focusedArea == HistoryArea {
				m.focusedArea = RequestArea
				return m, m.requestTextarea.Focus()
			}
			return m, nil
		case key.Matches(msg, m.keymap.execute):
			var executeRequestCommands []tea.Cmd
			m, executeRequestCommands = m.newRequest()
			cmds = append(cmds, executeRequestCommands...)
		case key.Matches(msg, m.keymap.next):
			var focusCmds []tea.Cmd
			m, focusCmds = m.SwitchFocus(true)
			cmds = append(cmds, focusCmds...)
		case key.Matches(msg, m.keymap.prev):
			var focusCmds []tea.Cmd
			m, focusCmds = m.SwitchFocus(false)
			cmds = append(cmds, focusCmds...)
		case key.Matches(msg, m.keymap.enter):
			if m.focusedArea != HistoryArea {
				break
			}
			selectedHistoryEntry, ok := m.history.SelectedItem().(app.Request)
			if !ok {
				return m, nil
			}
			m.requestTextarea.SetValue(fmt.Sprintf("%s %s", selectedHistoryEntry.Method, selectedHistoryEntry.Url))
			m.focusedArea = RequestArea
			return m, m.requestTextarea.Focus()
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
		newItems := append([]list.Item{response.request}, m.history.Items()...)
		cmds = append(cmds, m.history.SetItems(newItems))
	}

	m.sizeInputs()

	if !m.ongoingRequest {
		newModel, cmd := m.requestTextarea.Update(msg)
		m.requestTextarea = newModel
		cmds = append(cmds, cmd)
	}
	if m.focusedArea == HistoryArea {
		newModel, cmd := m.history.Update(msg)
		m.history = newModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	height := m.height - helpHeight
	if m.displayHistory {
		height -= historyHeight
		m.history.SetWidth(m.width)
		m.history.SetHeight(historyHeight)
	}
	m.requestTextarea.SetWidth(m.width / 2)
	m.requestTextarea.SetHeight(height)
	m.responseTextarea.SetWidth(m.width / 2)
	m.responseTextarea.SetHeight(height)
}

func (m model) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.execute,
		m.keymap.next,
		m.keymap.prev,
		m.keymap.toggleHistory,
		m.keymap.quit,
	})

	var views []string
	views = append(views, m.requestTextarea.View())
	views = append(views, m.responseTextarea.View())

	view := lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n"
	if m.displayHistory {
		view += m.history.View() + "\n\n"
	}

	return view + help
}
