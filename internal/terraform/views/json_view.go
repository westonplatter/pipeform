package views

import (
	"fmt"
	"time"

	gojson "encoding/json"

	"github.com/magodo/pipeform/internal/terraform/views/json"
)

type Message interface {
	BaseMessage() BaseMsg
}

// This file define structures corresponding to the different logs defined in:
// terraform/internal/command/views/json_view.go

type BaseMsg struct {
	Level     string           `json:"level"`
	Message   string           `json:"message"`
	Module    string           `json:"module"`
	TimeStamp time.Time        `json:"time_stamp"`
	Type      json.MessageType `json:"type"`
}

func (m BaseMsg) BaseMessage() BaseMsg {
	return m
}

type VersionMsg struct {
	BaseMsg
	Terraform string `json:"terraform"`
	UI        string `json:"ui"`
}

func (m VersionMsg) BaseMessage() BaseMsg {
	return m.BaseMsg
}

type LogMsg struct {
	BaseMsg
	KVs map[string]interface{}
}

func (m LogMsg) BaseMessage() BaseMsg {
	return m.BaseMsg
}

func (v LogMsg) MarshalJSON() ([]byte, error) {
	b, err := gojson.Marshal(v.BaseMsg)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := gojson.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	for k, v := range v.KVs {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}

	return gojson.Marshal(m)
}

type DiagnosticsMsg struct {
	BaseMsg
	Diagnostic *json.Diagnostic `json:"diagnostic"`
}

func (m DiagnosticsMsg) BaseMessage() BaseMsg {
	return m.BaseMsg
}

type PlannedChangeMsg struct {
	BaseMsg
	Change *json.ResourceInstanceChange `json:"change"`
}

func (m PlannedChangeMsg) BaseMessage() BaseMsg {
	return m.BaseMsg
}

type ResourceDriftMsg struct {
	BaseMsg
	Change *json.ResourceInstanceChange `json:"change"`
}

func (m ResourceDriftMsg) BaseMessage() BaseMsg {
	return m.BaseMsg
}

type ChangeSummaryMsg struct {
	BaseMsg
	Changes *json.ChangeSummary `json:"changes"`
}

func (m ChangeSummaryMsg) BaseMessage() BaseMsg {
	return m.BaseMsg
}

type OutputMsg struct {
	BaseMsg
	Outputs *json.Output `json:"outputs"`
}

func (m OutputMsg) BaseMessage() BaseMsg {
	return m.BaseMsg
}

type HookMsg struct {
	BaseMsg
	json.Hooker
}

func (m HookMsg) BaseMessage() BaseMsg {
	return m.BaseMsg
}

func (m HookMsg) MarshalJSON() ([]byte, error) {
	switch hooker := m.Hooker.(type) {
	case json.OperationStart:
		mm := struct {
			BaseMsg
			json.OperationStart
		}{
			BaseMsg:        m.BaseMsg,
			OperationStart: hooker,
		}
		return gojson.Marshal(mm)
	case json.OperationProgress:
		mm := struct {
			BaseMsg
			json.OperationProgress
		}{
			BaseMsg:           m.BaseMsg,
			OperationProgress: hooker,
		}
		return gojson.Marshal(mm)
	case json.OperationComplete:
		mm := struct {
			BaseMsg
			json.OperationComplete
		}{
			BaseMsg:           m.BaseMsg,
			OperationComplete: hooker,
		}
		return gojson.Marshal(mm)
	case json.OperationErrored:
		mm := struct {
			BaseMsg
			json.OperationErrored
		}{
			BaseMsg:          m.BaseMsg,
			OperationErrored: hooker,
		}
		return gojson.Marshal(mm)
	case json.ProvisionStart:
		mm := struct {
			BaseMsg
			json.ProvisionStart
		}{
			BaseMsg:        m.BaseMsg,
			ProvisionStart: hooker,
		}
		return gojson.Marshal(mm)
	case json.ProvisionProgress:
		mm := struct {
			BaseMsg
			json.ProvisionProgress
		}{
			BaseMsg:           m.BaseMsg,
			ProvisionProgress: hooker,
		}
		return gojson.Marshal(mm)
	case json.ProvisionComplete:
		mm := struct {
			BaseMsg
			json.ProvisionComplete
		}{
			BaseMsg:           m.BaseMsg,
			ProvisionComplete: hooker,
		}
		return gojson.Marshal(mm)
	case json.ProvisionErrored:
		mm := struct {
			BaseMsg
			json.ProvisionErrored
		}{
			BaseMsg:          m.BaseMsg,
			ProvisionErrored: hooker,
		}
		return gojson.Marshal(mm)
	case json.RefreshStart:
		mm := struct {
			BaseMsg
			json.RefreshStart
		}{
			BaseMsg:      m.BaseMsg,
			RefreshStart: hooker,
		}
		return gojson.Marshal(mm)
	case json.RefreshComplete:
		mm := struct {
			BaseMsg
			json.RefreshComplete
		}{
			BaseMsg:         m.BaseMsg,
			RefreshComplete: hooker,
		}
		return gojson.Marshal(mm)
	default:
		return nil, fmt.Errorf("unknown hook type: %T", m.Hooker)
	}
}

func UnmarshalMessage(b []byte) (Message, error) {
	var baseMsg BaseMsg
	if err := gojson.Unmarshal(b, &baseMsg); err != nil {
		return nil, err
	}

	switch baseMsg.Type {
	case json.MessageVersion:
		var msg VersionMsg
		if err := gojson.Unmarshal(b, &msg); err != nil {
			return nil, err
		}
		return msg, nil

	case json.MessageLog:
		var m map[string]interface{}
		if err := gojson.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		var msg LogMsg
		if err := gojson.Unmarshal(b, &msg); err != nil {
			return nil, err
		}
		b, err := gojson.Marshal(msg)
		if err != nil {
			return nil, err
		}
		var m2 map[string]interface{}
		if err := gojson.Unmarshal(b, &m2); err != nil {
			return nil, err
		}

		msg.KVs = map[string]interface{}{}
		for k, v := range m {
			if _, ok := m2[k]; !ok {
				msg.KVs[k] = v
			}
		}

		return msg, nil

	case json.MessageDiagnostic:
		var msg DiagnosticsMsg
		if err := gojson.Unmarshal(b, &msg); err != nil {
			return nil, err
		}
		return msg, nil

	case json.MessageResourceDrift:
		var msg ResourceDriftMsg
		if err := gojson.Unmarshal(b, &msg); err != nil {
			return nil, err
		}
		return msg, nil

	case json.MessagePlannedChange:
		var msg PlannedChangeMsg
		if err := gojson.Unmarshal(b, &msg); err != nil {
			return nil, err
		}
		return msg, nil

	case json.MessageChangeSummary:
		var msg ChangeSummaryMsg
		if err := gojson.Unmarshal(b, &msg); err != nil {
			return nil, err
		}
		return msg, nil

	case json.MessageOutputs:
		var msg OutputMsg
		if err := gojson.Unmarshal(b, &msg); err != nil {
			return nil, err
		}
		return msg, nil

	case json.MessageApplyStart, json.MessageEphemeralOpStart:
		temp := struct {
			BaseMsg
			json.OperationStart
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.OperationStart,
		}, nil

	case json.MessageApplyProgress, json.MessageEphemeralOpProgress:
		temp := struct {
			BaseMsg
			json.OperationProgress
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.OperationProgress,
		}, nil

	case json.MessageApplyComplete, json.MessageEphemeralOpComplete:
		temp := struct {
			BaseMsg
			json.OperationComplete
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.OperationComplete,
		}, nil

	case json.MessageApplyErrored, json.MessageEphemeralOpErrored:
		temp := struct {
			BaseMsg
			json.OperationErrored
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.OperationErrored,
		}, nil

	case json.MessageProvisionStart:
		temp := struct {
			BaseMsg
			json.ProvisionStart
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.ProvisionStart,
		}, nil

	case json.MessageProvisionProgress:
		temp := struct {
			BaseMsg
			json.ProvisionProgress
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.ProvisionProgress,
		}, nil

	case json.MessageProvisionComplete:
		temp := struct {
			BaseMsg
			json.ProvisionComplete
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.ProvisionComplete,
		}, nil

	case json.MessageProvisionErrored:
		temp := struct {
			BaseMsg
			json.ProvisionErrored
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.ProvisionErrored,
		}, nil

	case json.MessageRefreshStart:
		temp := struct {
			BaseMsg
			json.RefreshStart
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.RefreshStart,
		}, nil

	case json.MessageRefreshComplete:
		temp := struct {
			BaseMsg
			json.RefreshComplete
		}{}
		if err := gojson.Unmarshal(b, &temp); err != nil {
			return nil, err
		}

		return HookMsg{
			BaseMsg: temp.BaseMsg,
			Hooker:  temp.RefreshComplete,
		}, nil

	default:
		return nil, fmt.Errorf("unhandled message type: %s", baseMsg.Type)
	}
}
