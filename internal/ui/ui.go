package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/magodo/pipeform/internal/log"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/magodo/pipeform/internal/reader"
	"github.com/magodo/pipeform/internal/terraform/views"
	"github.com/magodo/pipeform/internal/terraform/views/json"
)

const (
	padding = 2
)

type versionInfo struct {
	terraform string
	ui        string
}

type UIModel struct {
	logger    *log.Logger
	reader    reader.Reader
	teeWriter io.Writer

	viewState ViewState
	isEOF     bool

	diags Diags

	resourceInfos ResourceInfos

	version *versionInfo

	// These are read from the ChangeSummaryMsg
	operation json.Operation
	totalCnt  int

	doneCnt int

	spinner  spinner.Model
	table    table.Model
	progress progress.Model
}

func NewRuntimeModel(logger *log.Logger, reader reader.Reader) UIModel {
	t := table.New(
		table.WithColumns(TableColumn(30)),
		table.WithFocused(true),
	)
	t.SetStyles(StyleTableFunc())

	model := UIModel{
		logger:        logger,
		reader:        reader,
		viewState:     ViewStateIdle,
		resourceInfos: ResourceInfos{},
		spinner:       spinner.New(),
		table:         t,
		progress:      progress.New(),
	}

	return model
}

func (m UIModel) Diags() Diags {
	return m.diags
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
	return tea.Batch(m.nextMessage, m.spinner.Tick)
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.logger.Trace("Message received", "type", fmt.Sprintf("%T", msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.logger.Warn("Interrupt key received, quit the program")
			return m, tea.Quit
		default:
			table, cmd := m.table.Update(msg)
			m.table = table
			return m, cmd
		}
	case tea.WindowSizeMsg:
		width := msg.Width - padding*2 - 8
		height := msg.Height - padding*2 - 20

		m.progress.Width = width

		m.table.SetColumns(TableColumn(width))
		m.table.SetWidth(width)
		m.table.SetHeight(height)

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

	// Log the receiver error message
	case receiverErrorMsg:
		m.logger.Error("Receiver error", "error", msg.Error())
		return m, m.nextMessage

	case receiverEOFMsg:
		m.logger.Info("Receiver reaches EOF")
		m.isEOF = true
		return m, nil

	case receiverMsg:
		m.logger.Debug("Message receiverMsg received", "type", fmt.Sprintf("%T", msg.msg))

		cmds := []tea.Cmd{m.nextMessage}

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

		case views.OutputMsg:
			// TODO: How to show output?

		case views.HookMsg:
			m.logger.Debug("Hook message", "type", fmt.Sprintf("%T", msg.Hook))
			switch hooker := msg.Hook.(type) {
			case json.OperationStart:
				res := &ResourceInfo{
					Loc: ResourceInfoLocator{
						Module:       hooker.Resource.Module,
						ResourceAddr: hooker.Resource.Addr,
						Action:       hooker.Action,
					},
					Status:    ResourceStatusStart,
					StartTime: msg.TimeStamp,
				}
				m.resourceInfos = append(m.resourceInfos, res)
			case json.OperationProgress:
				loc := ResourceInfoLocator{
					Module:       hooker.Resource.Module,
					ResourceAddr: hooker.Resource.Addr,
					Action:       hooker.Action,
				}
				status := ResourceStatusProgress
				update := ResourceInfoUpdate{
					Status: &status,
				}
				if !m.resourceInfos.Update(loc, update) {
					m.logger.Error("OperationProgress hooker can't find the resource info", "module", hooker.Resource.Module, "addr", hooker.Resource.Addr, "action", hooker.Action)
					break
				}

			case json.OperationComplete:
				loc := ResourceInfoLocator{
					Module:       hooker.Resource.Module,
					ResourceAddr: hooker.Resource.Addr,
					Action:       hooker.Action,
				}
				status := ResourceStatusComplete
				update := ResourceInfoUpdate{
					Status:  &status,
					Endtime: &msg.TimeStamp,
				}
				if !m.resourceInfos.Update(loc, update) {
					m.logger.Error("OperationComplete hooker can't find the resource info", "module", hooker.Resource.Module, "addr", hooker.Resource.Addr, "action", hooker.Action)
					break
				}

				m.doneCnt += 1
				percentage := float64(m.doneCnt) / float64(m.totalCnt)
				cmds = append(cmds, m.progress.SetPercent(percentage))

			case json.OperationErrored:
				loc := ResourceInfoLocator{
					Module:       hooker.Resource.Module,
					ResourceAddr: hooker.Resource.Addr,
					Action:       hooker.Action,
				}
				status := ResourceStatusErrored
				update := ResourceInfoUpdate{
					Status:  &status,
					Endtime: &msg.TimeStamp,
				}
				if !m.resourceInfos.Update(loc, update) {
					m.logger.Error("OperationErrored hooker can't find the resource info", "module", hooker.Resource.Module, "addr", hooker.Resource.Addr, "action", hooker.Action)
					break
				}

				m.doneCnt += 1
				percentage := float64(m.doneCnt) / float64(m.totalCnt)
				cmds = append(cmds, m.progress.SetPercent(percentage))

			case json.ProvisionStart:
			case json.ProvisionProgress:
			case json.ProvisionComplete:
			case json.ProvisionErrored:
			case json.RefreshStart:
			case json.RefreshComplete:
			default:
			}
		default:
			panic(fmt.Sprintf("unknown message type: %T", msg))
		}

		// Update viewState
		m.viewState = m.viewState.NextState(msg.msg)

		// Update table rows
		m.table.SetRows(m.resourceInfos.ToRows(m.totalCnt))

		return m, tea.Batch(cmds...)

	default:
		return m, nil
	}
}

func (m UIModel) View() string {
	s := "\n" + m.logoView()

	s += "\n\n" + m.stateView()

	s += "\n\n" + StyleTableBase.Render(m.table.View())

	s += "\n\n" + m.progress.View()

	return s
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

	return prefix + " " + m.viewState.String()
}
