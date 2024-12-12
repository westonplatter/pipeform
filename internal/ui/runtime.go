package ui

import (
	"fmt"
	"io"

	"github.com/magodo/pipeform/internal/log"

	prog "github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/magodo/pipeform/internal/reader"
	"github.com/magodo/pipeform/internal/terraform/views"
	"github.com/magodo/pipeform/internal/terraform/views/json"
)

const (
	padding  = 2
	maxWidth = 80
)

type versionInfo struct {
	terraform string
	ui        string
}

type runtimeModel struct {
	logger    *log.Logger
	reader    reader.Reader
	teeWriter io.Writer

	resourceInfos ResourceInfos

	// diags represent non-resource, non-provision diagnostics (as they are collected in the *Info)
	// E.g. this can be the provider diagnostic.
	diags []json.Diagnostic

	version *versionInfo

	// These are read from the ChangeSummaryMsg
	operation json.Operation
	totalCnt  int
	doneCnt   int

	progress prog.Model
}

func NewRuntimeModel(logger *log.Logger, reader reader.Reader) runtimeModel {
	model := runtimeModel{
		logger:        logger,
		reader:        reader,
		resourceInfos: ResourceInfos{},
		progress:      prog.New(),
	}

	return model
}

func (m runtimeModel) nextMessage() tea.Msg {
	msg, err := m.reader.Next()
	if err != nil {
		if err == io.EOF {
			return receiverEOFMsg{}
		}
		return receiverErrorMsg{err: err}
	}
	return receiverMsg{msg: msg}
}

func (m runtimeModel) Init() tea.Cmd {
	return m.nextMessage
}

func (m runtimeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.logger.Trace("Message received", "type", fmt.Sprintf("%T", msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.logger.Warn("Interrupt key received, quit the program")
			return m, tea.Quit
		default:
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	// FrameMsg is sent when the progress bar wants to animate itself
	case prog.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(prog.Model)
		return m, cmd

	// Log the receiver error message
	case receiverErrorMsg:
		m.logger.Error("Receiver error", "error", msg.Error())
		return m, m.nextMessage

	case receiverEOFMsg:
		m.logger.Info("Receiver reaches EOF")
		return m, tea.Quit

	case receiverMsg:
		m.logger.Debug("Message receiverMsg received", "type", fmt.Sprintf("%T", msg.msg))
		switch msg := msg.msg.(type) {
		case views.VersionMsg:
			m.version = &versionInfo{
				terraform: msg.Terraform,
				ui:        msg.UI,
			}
			return m, m.nextMessage

		case views.LogMsg:
			return m, m.nextMessage

		case views.DiagnosticsMsg:
			// TODO: Link resource related diag to the resource info
			m.diags = append(m.diags, *msg.Diagnostic)
			return m, m.nextMessage

		case views.ResourceDriftMsg:
			// There's no much useful information for now.
			return m, m.nextMessage

		case views.PlannedChangeMsg:
			// TODO: Consider record the planned change action.
			return m, m.nextMessage

		case views.ChangeSummaryMsg:
			changes := msg.Changes
			m.logger.Debug("Change summary", "add", changes.Add, "change", changes.Change, "import", changes.Import, "remove", changes.Remove)
			m.totalCnt = changes.Add + changes.Change + changes.Import + changes.Remove
			m.operation = changes.Operation
			return m, m.nextMessage

		case views.OutputMsg:
			// TODO: How to show output?
			return m, m.nextMessage

		case views.HookMsg:
			m.logger.Debug("Hook message", "type", fmt.Sprintf("%T", msg.Hooker))
			switch hooker := msg.Hooker.(type) {
			case json.OperationStart:
				res := ResourceInfo{
					ResourceAddr: hooker.Resource.Addr,
					Action:       hooker.Action,
					State:        ResourceStateStart,
					StartTime:    msg.TimeStamp,
				}
				r1, ok := m.resourceInfos[hooker.Resource.Addr]
				if !ok {
					r1 = map[json.ChangeAction]*ResourceInfo{}
					m.resourceInfos[hooker.Resource.Addr] = r1
				}
				r1[hooker.Action] = &res
				return m, m.nextMessage

			case json.OperationProgress:
				r1, ok := m.resourceInfos[hooker.Resource.Addr]
				if !ok {
					m.logger.Error("OperationProgress hooker can't find the resource info for addr", "addr", hooker.Resource.Addr)
					return m, m.nextMessage
				}
				r2, ok := r1[hooker.Action]
				if !ok {
					m.logger.Error("OperationProgress hooker can't find the resource info for action", "addr", hooker.Resource.Addr, "action", hooker.Action)
					return m, m.nextMessage
				}
				r2.State = ResourceStateProgress
				return m, m.nextMessage

			case json.OperationComplete:
				r1, ok := m.resourceInfos[hooker.Resource.Addr]
				if !ok {
					m.logger.Error("OperationComplete hooker can't find the resource info for addr", "addr", hooker.Resource.Addr)
					return m, m.nextMessage
				}
				r2, ok := r1[hooker.Action]
				if !ok {
					m.logger.Error("OperationComplete hooker can't find the resource info for action", "addr", hooker.Resource.Addr, "action", hooker.Action)
					return m, m.nextMessage
				}
				r2.State = ResourceStateComplete
				r2.EndTime = msg.TimeStamp

				m.doneCnt += 1

				cmd := m.progress.SetPercent(float64(m.doneCnt) / float64(m.totalCnt))
				cmds := tea.Batch(cmd, m.nextMessage)

				return m, cmds

			case json.OperationErrored:
				r1, ok := m.resourceInfos[hooker.Resource.Addr]
				if !ok {
					m.logger.Error("OperationErrored hooker can't find the resource info for addr", "addr", hooker.Resource.Addr)
					return m, m.nextMessage
				}
				r2, ok := r1[hooker.Action]
				if !ok {
					m.logger.Error("OperationErrored hooker can't find the resource info for action", "addr", hooker.Resource.Addr, "action", hooker.Action)
					return m, m.nextMessage
				}
				r2.State = ResourceStateErrored
				r2.EndTime = msg.TimeStamp

				m.doneCnt += 1

				cmd := m.progress.SetPercent(float64(m.doneCnt) / float64(m.totalCnt))
				cmds := tea.Batch(cmd, m.nextMessage)

				return m, cmds

			case json.ProvisionStart:
				return m, m.nextMessage
			case json.ProvisionProgress:
				return m, m.nextMessage
			case json.ProvisionComplete:
				return m, m.nextMessage
			case json.ProvisionErrored:
				return m, m.nextMessage
			case json.RefreshStart:
				return m, m.nextMessage
			case json.RefreshComplete:
				return m, m.nextMessage
			default:
				return m, m.nextMessage
			}
		default:
			panic(fmt.Sprintf("unknown message type: %T", msg))
		}

	default:
		return m, nil
	}
}

func (m runtimeModel) View() string {
	s := m.logoView()

	return s + "\n\n" + m.progress.View() + "\n\n"
}

func (m runtimeModel) logoView() string {
	msg := "Terraform"
	if m.version != nil {
		msg += fmt.Sprintf(" %s (%s)", m.version.terraform, m.version.ui)
	}
	return "\n" + StyleTitle.Render(" "+msg+" ") + "\n\n"
}
