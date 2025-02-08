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

const historyHeight = 10

var (
	historyItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	historySelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

type historyItemDelegate struct{}

func (d historyItemDelegate) Height() int                             { return 1 }
func (d historyItemDelegate) Spacing() int                            { return 0 }
func (d historyItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d historyItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	request, ok := listItem.(app.Request)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, request)

	fn := historyItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return historySelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func mapHistory(history app.History) []list.Item {
	items := []list.Item{}
	for i := len(history) - 1; i >= 0; i-- {
		items = append(items, history[i])
	}
	return items
}

func newHistoryList(history app.History) list.Model {
	l := list.New(mapHistory(history), historyItemDelegate{}, 0, historyHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	return l
}
