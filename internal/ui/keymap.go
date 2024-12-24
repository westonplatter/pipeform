package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
)

type KeyMap struct {
	TableKeyMap table.KeyMap
	Follow      key.Binding
	Quit        key.Binding
	Copy        key.Binding
	Help        key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	tableHelp := k.TableKeyMap.ShortHelp()
	return append([]key.Binding{k.Follow, k.Quit, k.Copy, k.Help}, tableHelp...)
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	tableHelp := k.TableKeyMap.FullHelp()
	return append([][]key.Binding{{k.Follow, k.Quit, k.Copy, k.Help}}, tableHelp...)
}

func NewKeyMap(clipboardEnabled bool) KeyMap {
	keymap := KeyMap{
		TableKeyMap: table.DefaultKeyMap(),
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
