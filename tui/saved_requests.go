package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/ThomasFerro/gogetter/app"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	savedRequestsItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	savedRequestsSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

type savedRequestsItemDelegate struct{}

func (d savedRequestsItemDelegate) Height() int                             { return 1 }
func (d savedRequestsItemDelegate) Spacing() int                            { return 0 }
func (d savedRequestsItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d savedRequestsItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	request, ok := listItem.(app.Request)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, request)

	fn := savedRequestsItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return savedRequestsSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func mapSavedRequests(savedRequests app.SavedRequests) []list.Item {
	items := []list.Item{}
	for i := len(savedRequests) - 1; i >= 0; i-- {
		items = append(items, savedRequests[i])
	}
	return items
}

func newSavedRequestsList(savedRequests app.SavedRequests) list.Model {
	l := list.New(mapSavedRequests(savedRequests), savedRequestsItemDelegate{}, 0, bottomListHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Title = "Saved requests"
	l.SetShowTitle(true)
	return l
}
