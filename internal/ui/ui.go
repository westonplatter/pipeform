package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/magodo/pipeform/internal/clipboard"
	"github.com/magodo/pipeform/internal/log"
	"github.com/muesli/reflow/indent"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/magodo/pipeform/internal/reader"
	"github.com/magodo/pipeform/internal/terraform/views"
	"github.com/magodo/pipeform/internal/terraform/views/json"
)

const (
	padding     = 2
	indentLevel = 2
)

type versionInfo struct {
	terraform string
	ui        string
}

type UIModel struct {
	logger    *log.Logger
	reader    reader.Reader
	teeWriter io.Writer

	viewState         ViewState
	lastLog           string
	userOperationInfo string

	isEOF bool

	diags Diags

	refreshInfos ResourceInfos
	applyInfos   ResourceInfos

	outputInfos OutputInfos

	version *versionInfo

	// These are read from the ChangeSummaryMsg
	operation json.Operation
	totalCnt  int

	doneCnt int

	keymap KeyMap

	help     help.Model
	spinner  spinner.Model
	table    table.Model
	progress progress.Model

	cp clipboard.Clipboard

	followed bool
}

func NewRuntimeModel(logger *log.Logger, reader reader.Reader) UIModel {
	t := table.New(table.WithFocused(true))
	t.SetStyles(StyleTableFunc())

	cp := clipboard.NewClipboard()

	model := UIModel{
		logger:    logger,
		reader:    reader,
		viewState: ViewStateIdle,
		keymap:    NewKeyMap(cp.Enabled()),
		help:      help.New(),
		spinner:   spinner.New(),
		table:     t,
		progress:  progress.New(),
		cp:        cp,
	}

	return model
}

func (m UIModel) Diags() Diags {
	return m.diags
}

func (m UIModel) IsEOF() bool {
	return m.isEOF
}

func (m UIModel) nextMessage() tea.Msg {
	msg, err := m.reader.Next()
	if err != nil {
		if err == io.EOF {
			return receiverEOFMsg{}
		}
		return receiverErrorMsg{err: err}
	}
	return receiverMsg{msg: msg}
}

func (m UIModel) Init() tea.Cmd {
	return tea.Batch(m.nextMessage, m.spinner.Tick, tickCmd())
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.logger.Trace("Message received", "type", fmt.Sprintf("%T", msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.userOperationInfo = ""
		switch {
		case key.Matches(msg, m.keymap.Quit):
			m.logger.Warn("Interrupt key received, quit the program")
			return m, tea.Quit
		case key.Matches(msg, m.keymap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keymap.Follow):
			m.followed = !m.followed
			return m, nil
		case key.Matches(msg, m.keymap.Copy):
			m.copyTableRow()
			return m, nil
		default:
			table, cmd := m.table.Update(msg)
			m.table = table
			return m, cmd
		}
	case tea.WindowSizeMsg:
		progressWidth := msg.Width - padding*2
		m.progress.Width = progressWidth

		tableWidth := msg.Width - padding*2 - 10
		tableHeight := msg.Height - padding*2 - 10
		m.setTableOutlook(tableWidth, tableHeight)
		m.setTableRows()

		return m, nil

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd

	case tickMsg:
		m.setTableRows()
		return m, tickCmd()

	// Log the receiver error message
	case receiverErrorMsg:
		m.logger.Error("Receiver error", "error", msg.Error())
		return m, m.nextMessage

	case receiverEOFMsg:
		m.logger.Info("Receiver reaches EOF")
		m.isEOF = true
		m.lastLog = ""
		return m, nil

	case receiverMsg:
		m.logger.Debug("Message receiverMsg received", "type", fmt.Sprintf("%T", msg.msg))

		cmds := []tea.Cmd{m.nextMessage}

		m.lastLog = msg.msg.BaseMessage().Message

		switch msg := msg.msg.(type) {
		case views.VersionMsg:
			m.version = &versionInfo{
				terraform: msg.Terraform,
				ui:        msg.UI,
			}

		case views.LogMsg:
			// There's no much useful information for now.
		case views.DiagnosticsMsg:
			// TODO: Link resource related diag to the resource info
			switch strings.ToLower(msg.Level) {
			case "warn":
				m.diags.Warns = append(m.diags.Warns, *msg.Diagnostic)
			case "error":
				m.diags.Errs = append(m.diags.Errs, *msg.Diagnostic)
			}

		case views.ResourceDriftMsg:
			// There's no much useful information for now.

		case views.PlannedChangeMsg:
			// TODO: Consider record the planned change action.

		case views.ChangeSummaryMsg:
			changes := msg.Changes
			m.logger.Debug("Change summary", "add", changes.Add, "change", changes.Change, "import", changes.Import, "remove", changes.Remove)
			m.totalCnt = changes.Add + changes.Change + changes.Import + changes.Remove
			m.operation = changes.Operation

			// Specifically, if the total count is 0, we update the progress bar directly as it is 100% anyway.
			if m.totalCnt == 0 {
				cmds = append(cmds, m.progress.SetPercent(1))
			}

		case views.OutputMsg:
			for name, o := range msg.Outputs {
				m.outputInfos = append(m.outputInfos, &OutputInfo{
					Name:      name,
					Sensitive: o.Sensitive,
					Type:      o.Type,
					ValueStr:  o.Value,
					Action:    o.Action,
				})
			}

		case views.HookMsg:
			m.logger.Debug("Hook message", "type", fmt.Sprintf("%T", msg.Hook))
			switch hook := msg.Hook.(type) {
			case json.RefreshStart:
				res := &ResourceInfo{
					Loc: ResourceInfoLocator{
						Module:       hook.Resource.Module,
						ResourceAddr: hook.Resource.Addr,
						Action:       "refresh",
					},
					Status:    ResourceStatusStart,
					StartTime: msg.TimeStamp,
				}
				m.refreshInfos = append(m.refreshInfos, res)

			case json.RefreshComplete:
				loc := ResourceInfoLocator{
					Module:       hook.Resource.Module,
					ResourceAddr: hook.Resource.Addr,
					Action:       "refresh",
				}
				status := ResourceStatusComplete
				update := ResourceInfoUpdate{
					Status:  &status,
					Endtime: &msg.TimeStamp,
				}
				if !m.refreshInfos.Update(loc, update) {
					m.logger.Error("RefreshComplete hook can't find the resource info", "module", hook.Resource.Module, "addr", hook.Resource.Addr, "action", "refresh")
					break
				}

			case json.OperationStart:
				res := &ResourceInfo{
					Loc: ResourceInfoLocator{
						Module:       hook.Resource.Module,
						ResourceAddr: hook.Resource.Addr,
						Action:       string(hook.Action),
					},
					Status:    ResourceStatusStart,
					StartTime: msg.TimeStamp,
				}
				m.applyInfos = append(m.applyInfos, res)

			case json.OperationProgress:
				// Ignore

			case json.OperationComplete:
				loc := ResourceInfoLocator{
					Module:       hook.Resource.Module,
					ResourceAddr: hook.Resource.Addr,
					Action:       string(hook.Action),
				}
				status := ResourceStatusComplete
				update := ResourceInfoUpdate{
					Status:  &status,
					Endtime: &msg.TimeStamp,
				}
				if !m.applyInfos.Update(loc, update) {
					m.logger.Error("OperationComplete hook can't find the resource info", "module", hook.Resource.Module, "addr", hook.Resource.Addr, "action", hook.Action)
					break
				}

				m.doneCnt += 1
				percentage := float64(m.doneCnt) / float64(m.totalCnt)
				cmds = append(cmds, m.progress.SetPercent(percentage))

			case json.OperationErrored:
				loc := ResourceInfoLocator{
					Module:       hook.Resource.Module,
					ResourceAddr: hook.Resource.Addr,
					Action:       string(hook.Action),
				}
				status := ResourceStatusErrored
				update := ResourceInfoUpdate{
					Status:  &status,
					Endtime: &msg.TimeStamp,
				}
				if !m.applyInfos.Update(loc, update) {
					m.logger.Error("OperationErrored hook can't find the resource info", "module", hook.Resource.Module, "addr", hook.Resource.Addr, "action", hook.Action)
					break
				}

				m.doneCnt += 1
				percentage := float64(m.doneCnt) / float64(m.totalCnt)
				cmds = append(cmds, m.progress.SetPercent(percentage))

			case json.ProvisionStart:
			case json.ProvisionProgress:
			case json.ProvisionComplete:
			case json.ProvisionErrored:
			default:
			}
		default:
			panic(fmt.Sprintf("unknown message type: %T", msg))
		}

		// Update viewState
		var change bool
		m.viewState, change = m.viewState.NextState(msg.msg)

		if change {
			cmds = append(cmds, tea.WindowSize())
		} else {
			m.setTableRows()
		}

		return m, tea.Batch(cmds...)

	default:
		return m, nil
	}
}

func (m *UIModel) setTableOutlook(width, height int) {
	m.table.SetWidth(width)
	m.table.SetHeight(height)

	// Clean up the rows before changing table columns, mainly to avoid
	// existing rows have more columns than the new columns, i.e. from
	// "apply" (6) to "summary" (5).
	m.table.SetRows(nil)

	switch m.viewState {
	case ViewStateRefresh:
		m.table.SetColumns(m.refreshInfos.ToColumns(width))
	case ViewStateApply:
		m.table.SetColumns(m.applyInfos.ToColumns(width))
	case ViewStateSummary:
		m.table.SetColumns(m.outputInfos.ToColumns(width))
	}
}

// setTableRows on a one second pace.
func (m *UIModel) setTableRows() {
	switch m.viewState {
	case ViewStateRefresh:
		m.table.SetRows(m.refreshInfos.ToRows(0))
	case ViewStateApply:
		m.table.SetRows(m.applyInfos.ToRows(m.totalCnt))
	case ViewStateSummary:
		m.table.SetRows(m.outputInfos.ToRows())
	}

	if m.followed {
		m.table.GotoBottom()
	}
}

func (m *UIModel) copyTableRow() {
	if !m.cp.Enabled() {
		return
	}

	switch m.viewState {
	case ViewStateRefresh:
		if row := m.table.SelectedRow(); len(row) > 4 {
			m.cp.Write([]byte(row[4]))
		}
	case ViewStateApply:
		if row := m.table.SelectedRow(); len(row) > 4 {
			m.cp.Write([]byte(row[4]))
		}
	case ViewStateSummary:
		if row := m.table.SelectedRow(); len(row) > 4 {
			m.cp.Write([]byte(row[4]))
		}
	}

	m.userOperationInfo = "Copied!"
}

func (m UIModel) logoView() string {
	msg := "pipeform"
	if m.version != nil {
		msg += fmt.Sprintf(" (terraform: %s)", m.version.terraform)
	}
	return StyleTitle.Render(" " + msg + " ")
}

func (m UIModel) stateView() string {
	prefix := m.spinner.View()
	if m.isEOF {
		if len(m.diags.Errs) == 0 {
			prefix = "✅"
		} else {
			prefix = "❌"
		}
	}

	s := prefix + " " + StyleSubtitle.Render(m.viewState.String())

	if m.followed {
		s += " [following]"
	}

	if m.lastLog != "" {
		s += "  " + StyleComment.Render(m.lastLog)
	}

	return s
}

func (m UIModel) View() string {
	s := "\n" + m.logoView()

	s += "\n\n" + m.stateView()

	s += "\n\n" + StyleTableBase.Render(m.table.View())

	var progressBar string
	if m.viewState >= ViewStateApply {
		progressBar = m.progress.View()
	} else {
		progressBar = "\n"
	}
	s += "\n\n" + progressBar

	var bottomLine string
	if m.userOperationInfo != "" {
		bottomLine = StyleComment.Render(m.userOperationInfo)
	} else {
		bottomLine = m.help.View(m.keymap)
	}
	s += "\n\n" + bottomLine

	return indent.String(s, indentLevel)
}
