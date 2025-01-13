package plainui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/magodo/pipeform/internal/csv"
	"github.com/magodo/pipeform/internal/log"
	"github.com/magodo/pipeform/internal/reader"
	"github.com/magodo/pipeform/internal/state"
	"github.com/magodo/pipeform/internal/terraform/views"
	"github.com/magodo/pipeform/internal/terraform/views/json"
)

type UIModel struct {
	startTime time.Time
	logger    *log.Logger
	reader    reader.Reader
	writer    io.Writer

	refreshInfos state.ResourceInfos
	applyInfos   state.ResourceInfos

	totalCnt int
	doneCnt  int

	isEOF bool
}

func NewRuntimeModel(logger *log.Logger, reader reader.Reader, writer io.Writer, startTime time.Time) UIModel {
	model := UIModel{
		startTime: startTime,
		logger:    logger,
		reader:    reader,
		writer:    writer,
	}

	return model
}

func (m *UIModel) Run() error {
	for {
		msg, err := m.reader.Next()
		if err != nil {
			if err == io.EOF {
				m.isEOF = true
				return nil
			}
			return err
		}

		var msgstr string
		switch msg := msg.(type) {
		case views.VersionMsg:
			msgstr = msg.Message
		case views.LogMsg:
			kvs := []string{}
			for k, v := range msg.KVs {
				kvs = append(kvs, fmt.Sprintf("%s=%v", k, v))
			}
			msgstr = fmt.Sprintf("%s. %s", msg.Message, strings.Join(kvs, " "))
		case views.DiagnosticsMsg:
			msgstr = fmt.Sprintf("Summary: %s.", msg.Diagnostic.Summary)
			if msg.Diagnostic.Detail != "" {
				msgstr += fmt.Sprintf(" Detail: %s", msg.Diagnostic.Detail)
			}
			if msg.Level != "info" {
				msgstr = fmt.Sprintf("[%s] %s", strings.ToUpper(msg.Level), msgstr)
			}
		case views.ResourceDriftMsg:
			msgstr = msg.Message
		case views.PlannedChangeMsg:
			// Normally, we don't need to handle the PlannedChangeMsg here, as the ChangeSummaryMsg has all these information.
			// The exception is that when apply with a plan file, there is no ChangeSummaryMsg sent from Terraform at this moment.
			// (see: https://github.com/magodo/pipeform/issues/1)
			// The counting here is a fallback logic to cover the case above. Otherwise, it will just be overwritten by ChangeSummaryMsg.
			//
			// TODO: Once https://github.com/hashicorp/terraform/pull/36245 merged, remove this part.
			//
			// Referencing the logic of terraform: internal/command/views/operation.go
			// But we also count the "import"
			switch msg.Change.Action {
			case json.ActionCreate:
				m.totalCnt++
			case json.ActionDelete:
				m.totalCnt++
			case json.ActionUpdate:
				m.totalCnt++
			case json.ActionReplace:
				m.totalCnt += 2
			case json.ActionImport:
				m.totalCnt++
			}
			msgstr = msg.Message

		case views.ChangeSummaryMsg:
			changes := msg.Changes
			m.totalCnt = changes.Add + changes.Change + changes.Import + changes.Remove
			msgstr = msg.Message

		case views.OutputMsg:
			outputs := []string{}
			for name, o := range msg.Outputs {
				if o.Action != "" {
					continue
				}
				output := fmt.Sprintf("%s=%s", name, string(o.Value))
				if o.Sensitive {
					output += " (sensitive)"
				}
				outputs = append(outputs, output)
			}
			msgstr = fmt.Sprintf("%s. %s", msg.Message, strings.Join(outputs, " "))

		case views.HookMsg:
			switch hook := msg.Hook.(type) {
			case json.RefreshStart:
				res := &state.ResourceInfo{
					Idx:             len(m.refreshInfos) + 1,
					RawResourceAddr: hook.Resource,
					Loc: state.ResourceInfoLocator{
						Module:       hook.Resource.Module,
						ResourceAddr: hook.Resource.Addr,
						Action:       "refresh",
					},
					Status:    state.ResourceStatusStart,
					StartTime: msg.TimeStamp,
				}
				m.refreshInfos = append(m.refreshInfos, res)
				msgstr = msg.Message

			case json.RefreshComplete:
				loc := state.ResourceInfoLocator{
					Module:       hook.Resource.Module,
					ResourceAddr: hook.Resource.Addr,
					Action:       "refresh",
				}
				status := state.ResourceStatusComplete
				update := state.ResourceInfoUpdate{
					Status:  &status,
					Endtime: &msg.TimeStamp,
				}
				if m.refreshInfos.Update(loc, update) == nil {
					m.logger.Error("RefreshComplete hook can't find the resource info", "module", hook.Resource.Module, "addr", hook.Resource.Addr, "action", "refresh")
					break
				}
				msgstr = msg.Message

			case json.OperationStart:
				info := &state.ResourceInfo{
					Idx:             len(m.applyInfos) + 1,
					RawResourceAddr: hook.Resource,
					Loc: state.ResourceInfoLocator{
						Module:       hook.Resource.Module,
						ResourceAddr: hook.Resource.Addr,
						Action:       string(hook.Action),
					},
					Status:    state.ResourceStatusStart,
					StartTime: msg.TimeStamp,
				}
				m.applyInfos = append(m.applyInfos, info)

				w := width(m.totalCnt)
				msgstr = fmt.Sprintf("[%*d/%*d] %s", w, info.Idx, w, m.totalCnt, msg.Message)

			case json.OperationProgress:
				loc := state.ResourceInfoLocator{
					Module:       hook.Resource.Module,
					ResourceAddr: hook.Resource.Addr,
					Action:       string(hook.Action),
				}
				info := m.applyInfos.Find(loc)
				if info == nil {
					m.logger.Error("OperationProgress hook can't find the resource info", "module", hook.Resource.Module, "addr", hook.Resource.Addr, "action", hook.Action)
					break
				}

				w := width(m.totalCnt)
				msgstr = fmt.Sprintf("[%*d/%*d] %s", w, info.Idx, w, m.totalCnt, msg.Message)

			case json.OperationComplete:
				loc := state.ResourceInfoLocator{
					Module:       hook.Resource.Module,
					ResourceAddr: hook.Resource.Addr,
					Action:       string(hook.Action),
				}
				status := state.ResourceStatusComplete
				update := state.ResourceInfoUpdate{
					Status:  &status,
					Endtime: &msg.TimeStamp,
				}
				info := m.applyInfos.Update(loc, update)
				if info == nil {
					m.logger.Error("OperationComplete hook can't find the resource info", "module", hook.Resource.Module, "addr", hook.Resource.Addr, "action", hook.Action)
					break
				}

				w := width(m.totalCnt)
				msgstr = fmt.Sprintf("[%*d/%*d] %s", w, info.Idx, w, m.totalCnt, msg.Message)

			case json.OperationErrored:
				loc := state.ResourceInfoLocator{
					Module:       hook.Resource.Module,
					ResourceAddr: hook.Resource.Addr,
					Action:       string(hook.Action),
				}
				status := state.ResourceStatusErrored
				update := state.ResourceInfoUpdate{
					Status:  &status,
					Endtime: &msg.TimeStamp,
				}
				info := m.applyInfos.Update(loc, update)
				if info == nil {
					m.logger.Error("OperationErrored hook can't find the resource info", "module", hook.Resource.Module, "addr", hook.Resource.Addr, "action", hook.Action)
					break
				}

				w := width(m.totalCnt)
				msgstr = fmt.Sprintf("[%*d/%*d] %s", w, info.Idx, w, m.totalCnt, msg.Message)

			case json.ProvisionStart:
				msgstr = msg.Message
			case json.ProvisionProgress:
				msgstr = msg.Message
			case json.ProvisionComplete:
				msgstr = msg.Message
			case json.ProvisionErrored:
				msgstr = msg.Message
			default:
				msgstr = msg.Message
			}
		}

		m.writer.Write([]byte(msgstr + "\n"))
	}
}

func (m UIModel) IsEOF() bool {
	return m.isEOF
}

func (m UIModel) ToCsv() []byte {
	return csv.ToCsv(csv.Input{
		RefreshInfos: m.refreshInfos,
		ApplyInfos:   m.applyInfos,
	})
}

func decorateMsg(level, msg string) string {
	return msg
}

func width(n int) int {
	var w int
	for n != 0 {
		n /= 10
		w += 1
	}
	return w
}
