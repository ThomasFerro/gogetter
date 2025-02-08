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
	helpHeight       = 5
	bottomListHeight = 10
)

var Gogetter app.Gogetter

type keymap = struct {
	next, prev, execute, save, toggleHistory, toggleSavedRequests, quit, enter key.Binding
}

type focusedArea int

const (
	RequestArea focusedArea = iota
	ResponseArea
	BottomListArea
)

type bottomList int

const (
	HistoryBottomList bottomList = iota
	SavedRequestsBottomList
)

type model struct {
	width             int
	height            int
	keymap            keymap
	help              help.Model
	requestTextarea   textarea.Model
	responseTextarea  textarea.Model
	history           list.Model
	savedRequests     list.Model
	focusedArea       focusedArea
	ongoingRequest    bool
	displayBottomList bool
	bottomList        bottomList
}

func NewModel(gogetter app.Gogetter) model {
	Gogetter = gogetter
	requestTextarea := newTextarea()
	requestTextarea.Placeholder = "Type your request"
	history := newHistoryList(gogetter.History())
	savedRequests := newSavedRequestsList(gogetter.SavedRequests())
	m := model{
		requestTextarea:  requestTextarea,
		responseTextarea: newTextarea(),
		help:             help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			save: key.NewBinding(
				key.WithKeys("alt+s"),
				key.WithHelp("alt+s", "save request"),
			),
			toggleHistory: key.NewBinding(
				key.WithKeys("alt+h"),
				key.WithHelp("alt+h", "toggle history"),
			),
			toggleSavedRequests: key.NewBinding(
				key.WithKeys("alt+r"),
				key.WithHelp("alt+r", "toggle saved requests"),
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
		displayBottomList: false,
		history:           history,
		savedRequests:     savedRequests,
	}

	m.requestTextarea.Focus()
	m.focusedArea = RequestArea
	return m
}

type newRequestMsg struct{}

type responseMsg struct {
	responseBody       string
	requestAndResponse app.RequestAndResponse
	err                error
}

func (m model) newRequest() (model, []tea.Cmd) {
	return m, []tea.Cmd{func() tea.Msg {
		return newRequestMsg{}
	}}
}

func (m model) currentRequest() (app.Request, bool) {
	// TODO: Extract and extend the parsing
	input := strings.Split(m.requestTextarea.Value(), " ")
	if len(input) < 2 {
		return app.Request{}, false
	}
	method := input[0]
	url := input[1]
	return app.Request{Method: method, Url: url}, true
}

func (m model) executeRequest() (model, []tea.Cmd) {
	if m.ongoingRequest {
		return m, []tea.Cmd{}
	}
	m.ongoingRequest = true
	return m, []tea.Cmd{func() tea.Msg {
		request, ok := m.currentRequest()
		if !ok {
			return nil
		}
		var resp *http.Response
		var err error
		var requestAndResponse app.RequestAndResponse
		Gogetter, requestAndResponse, resp, err = Gogetter.Execute(request.Method, request.Url)
		var response []byte

		if resp != nil {
			defer resp.Body.Close()
			response, err = io.ReadAll(resp.Body)
			if err != nil {
				return responseMsg{err: err, requestAndResponse: requestAndResponse, responseBody: ""}
			}
		}

		return responseMsg{err: err, requestAndResponse: requestAndResponse, responseBody: string(response)}
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
		if m.displayBottomList {
			m.focusedArea = BottomListArea
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
		if m.displayBottomList {
			m.focusedArea = BottomListArea
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
			if m.displayBottomList {
				m.displayBottomList = m.bottomList != HistoryBottomList
			} else {
				m.displayBottomList = true
			}

			m.bottomList = HistoryBottomList
			if m.displayBottomList {
				m.focusedArea = BottomListArea
				m.responseTextarea.Blur()
				m.requestTextarea.Blur()
				return m, nil
			}
			if m.focusedArea == BottomListArea {
				m.focusedArea = RequestArea
				return m, m.requestTextarea.Focus()
			}
			return m, nil

		case key.Matches(msg, m.keymap.toggleSavedRequests):
			if m.displayBottomList {
				m.displayBottomList = m.bottomList != SavedRequestsBottomList
			} else {
				m.displayBottomList = true
			}

			m.bottomList = SavedRequestsBottomList
			if m.displayBottomList {
				m.focusedArea = BottomListArea
				m.responseTextarea.Blur()
				m.requestTextarea.Blur()
				return m, nil
			}
			if m.focusedArea == BottomListArea {
				m.focusedArea = RequestArea
				return m, m.requestTextarea.Focus()
			}
			return m, nil

		case key.Matches(msg, m.keymap.execute):
			var executeRequestCommands []tea.Cmd
			m, executeRequestCommands = m.newRequest()
			cmds = append(cmds, executeRequestCommands...)

		case key.Matches(msg, m.keymap.save):
			request, ok := m.currentRequest()
			if !ok {
				break
			}
			var err error
			Gogetter, err = Gogetter.SaveRequest(request)
			if err != nil {
				m.responseTextarea.SetValue(fmt.Sprintf("request saving error: %w", err))
			}
			newItems := append([]list.Item{request}, m.savedRequests.Items()...)
			cmds = append(cmds, m.savedRequests.SetItems(newItems))

			return m, nil

		case key.Matches(msg, m.keymap.next):
			var focusCmds []tea.Cmd
			m, focusCmds = m.SwitchFocus(true)
			cmds = append(cmds, focusCmds...)
		case key.Matches(msg, m.keymap.prev):
			var focusCmds []tea.Cmd
			m, focusCmds = m.SwitchFocus(false)
			cmds = append(cmds, focusCmds...)
		case key.Matches(msg, m.keymap.enter):
			if m.focusedArea != BottomListArea {
				break
			}
			if m.bottomList == HistoryBottomList {
				selectedHistoryEntry, ok := m.history.SelectedItem().(app.RequestAndResponse)
				if !ok {
					return m, nil
				}
				m.requestTextarea.SetValue(fmt.Sprintf("%s %s", selectedHistoryEntry.Method, selectedHistoryEntry.Url))
			}

			if m.bottomList == SavedRequestsBottomList {
				selectedSavedRequest, ok := m.savedRequests.SelectedItem().(app.Request)
				if !ok {
					return m, nil
				}
				m.requestTextarea.SetValue(fmt.Sprintf("%s %s", selectedSavedRequest.Method, selectedSavedRequest.Url))
			}
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
		responseTextareaValue = fmt.Sprintf("%v%v", responseTextareaValue, response.responseBody)
		m.responseTextarea.SetValue(responseTextareaValue)
		m.ongoingRequest = false
		if response.err == nil {
			newItems := append([]list.Item{response.requestAndResponse}, m.history.Items()...)
			cmds = append(cmds, m.history.SetItems(newItems))
		}
	}

	m.sizeInputs()

	if !m.ongoingRequest {
		newModel, cmd := m.requestTextarea.Update(msg)
		m.requestTextarea = newModel
		cmds = append(cmds, cmd)
	}
	if m.focusedArea == BottomListArea {
		if m.bottomList == HistoryBottomList {
			newModel, cmd := m.history.Update(msg)
			m.history = newModel
			cmds = append(cmds, cmd)
		}
		if m.bottomList == SavedRequestsBottomList {
			newModel, cmd := m.savedRequests.Update(msg)
			m.savedRequests = newModel
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	height := m.height - helpHeight
	if m.displayBottomList {
		height -= bottomListHeight
		m.history.SetWidth(m.width)
		m.savedRequests.SetWidth(m.width)
	}
	m.requestTextarea.SetWidth(m.width / 2)
	m.requestTextarea.SetHeight(height)
	m.responseTextarea.SetWidth(m.width / 2)
	m.responseTextarea.SetHeight(height)
}

func (m model) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.execute,
		m.keymap.save,
		m.keymap.next,
		m.keymap.prev,
		m.keymap.toggleHistory,
		m.keymap.toggleSavedRequests,
	})

	var views []string
	views = append(views, m.requestTextarea.View())
	views = append(views, m.responseTextarea.View())

	view := lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n"
	if m.displayBottomList {
		if m.bottomList == HistoryBottomList {
			view += m.history.View() + "\n\n"
		}
		if m.bottomList == SavedRequestsBottomList {
			view += m.savedRequests.View() + "\n\n"
		}
	}

	return view + help
}
