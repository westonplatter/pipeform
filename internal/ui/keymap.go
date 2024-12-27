package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
)

type KeyMap struct {
	TableKeyMap  table.KeyMap
	PaginatorMap paginator.KeyMap

	Follow key.Binding
	Quit   key.Binding
	Copy   key.Binding

	Help key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	tableHelp := k.TableKeyMap.ShortHelp()
	return append([]key.Binding{k.Follow, k.Quit, k.Copy, k.PaginatorMap.PrevPage, k.PaginatorMap.NextPage, k.Help}, tableHelp...)
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	tableHelp := k.TableKeyMap.FullHelp()
	return append([][]key.Binding{{k.Follow, k.Quit, k.Copy, k.Help, k.PaginatorMap.PrevPage, k.PaginatorMap.NextPage}}, tableHelp...)
}

func NewKeyMap(clipboardEnabled bool) KeyMap {
	keymap := KeyMap{
		TableKeyMap: table.DefaultKeyMap(),
		PaginatorMap: paginator.KeyMap{
			PrevPage: key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("← / h", "left page"), key.WithDisabled()),
			NextPage: key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→ / l", "right page"), key.WithDisabled()),
		},
		Follow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "follow"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}

	if !clipboardEnabled {
		keymap.Copy.SetEnabled(false)
	}

	return keymap
}

func (km *KeyMap) EnablePaginator() {
	km.PaginatorMap.PrevPage.SetEnabled(true)
	km.PaginatorMap.NextPage.SetEnabled(true)
}
