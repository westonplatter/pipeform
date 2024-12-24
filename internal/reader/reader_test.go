package reader_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/magodo/pipeform/internal/reader"
	"github.com/magodo/pipeform/internal/terraform/views"
	vjson "github.com/magodo/pipeform/internal/terraform/views/json"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func mustUnmarshalTime(t *testing.T, v string) time.Time {
	var timE time.Time
	require.NoError(t, json.Unmarshal([]byte(v), &timE))
	return timE
}

func TestReader(t *testing.T) {
	inputs := []string{
		`{"@level":"info","@message":"Terraform 0.15.4","@module":"terraform.ui","@timestamp":"2021-05-25T13:32:41.275359-04:00","terraform":"0.15.4","type":"version","ui":"0.1.0"}`,
		`{"@level":"info","@message":"random_pet.animal: Plan to create","@module":"terraform.ui","@timestamp":"2021-05-25T13:32:41.705503-04:00","change":{"resource":{"addr":"random_pet.animal","module":"","resource":"random_pet.animal","implied_provider":"random","resource_type":"random_pet","resource_name":"animal","resource_key":null},"action":"create"},"type":"planned_change"}`,
		`{"@level":"info","@message":"Plan: 1 to add, 0 to change, 0 to destroy.","@module":"terraform.ui","@timestamp":"2021-05-25T13:32:41.705638-04:00","changes":{"add":1,"change":0,"remove":0,"operation":"plan"},"type":"change_summary"}`,
		`{"@level":"info","@message":"random_pet.animal: Creating...","@module":"terraform.ui","@timestamp":"2021-05-25T13:32:41.825308-04:00","hook":{"resource":{"addr":"random_pet.animal","module":"","resource":"random_pet.animal","implied_provider":"random","resource_type":"random_pet","resource_name":"animal","resource_key":null},"action":"create"},"type":"apply_start"}`,
		`{"@level":"info","@message":"random_pet.animal: Creation complete after 0s [id=smart-lizard]","@module":"terraform.ui","@timestamp":"2021-05-25T13:32:41.826179-04:00","hook":{"resource":{"addr":"random_pet.animal","module":"","resource":"random_pet.animal","implied_provider":"random","resource_type":"random_pet","resource_name":"animal","resource_key":null},"action":"create","id_key":"id","id_value":"smart-lizard","elapsed_seconds":0},"type":"apply_complete"}`,
		`{"@level":"info","@message":"Apply complete! Resources: 1 added, 0 changed, 0 destroyed.","@module":"terraform.ui","@timestamp":"2021-05-25T13:32:41.869168-04:00","changes":{"add":1,"change":0,"remove":0,"operation":"apply"},"type":"change_summary"}`,
		`{"@level":"info","@message":"Outputs: 1","@module":"terraform.ui","@timestamp":"2021-05-25T13:32:41.869280-04:00","outputs":{"pets":{"sensitive":false,"type":"string","value":"smart-lizard"}},"type":"outputs"}`,
	}

	expects := []views.Message{
		views.VersionMsg{
			BaseMsg: views.BaseMsg{
				Level:     "info",
				Message:   "Terraform 0.15.4",
				Type:      "version",
				Module:    "terraform.ui",
				TimeStamp: mustUnmarshalTime(t, `"2021-05-25T13:32:41.275359-04:00"`),
			},
			UI:        "0.1.0",
			Terraform: "0.15.4",
		},
		views.PlannedChangeMsg{
			BaseMsg: views.BaseMsg{
				Level:     "info",
				Message:   "random_pet.animal: Plan to create",
				Type:      "planned_change",
				Module:    "terraform.ui",
				TimeStamp: mustUnmarshalTime(t, `"2021-05-25T13:32:41.705503-04:00"`),
			},
			Change: &vjson.ResourceInstanceChange{
				Resource: vjson.ResourceAddr{
					Addr:            "random_pet.animal",
					Module:          "",
					Resource:        "random_pet.animal",
					ImpliedProvider: "random",
					ResourceType:    "random_pet",
					ResourceName:    "animal",
					ResourceKey:     ctyjson.SimpleJSONValue{Value: cty.NullVal(cty.DynamicPseudoType)},
				},
				Action: "create",
			},
		},
		views.ChangeSummaryMsg{
			BaseMsg: views.BaseMsg{
				Level:     "info",
				Message:   "Plan: 1 to add, 0 to change, 0 to destroy.",
				Type:      "change_summary",
				Module:    "terraform.ui",
				TimeStamp: mustUnmarshalTime(t, `"2021-05-25T13:32:41.705638-04:00"`),
			},
			Changes: &vjson.ChangeSummary{
				Add:       1,
				Operation: "plan",
			},
		},
		views.HookMsg{
			BaseMsg: views.BaseMsg{
				Level:     "info",
				Message:   "random_pet.animal: Creating...",
				Type:      "apply_start",
				Module:    "terraform.ui",
				TimeStamp: mustUnmarshalTime(t, `"2021-05-25T13:32:41.825308-04:00"`),
			},
			Hook: vjson.OperationStart{
				Resource: vjson.ResourceAddr{
					Addr:            "random_pet.animal",
					Module:          "",
					Resource:        "random_pet.animal",
					ImpliedProvider: "random",
					ResourceType:    "random_pet",
					ResourceName:    "animal",
					ResourceKey:     ctyjson.SimpleJSONValue{Value: cty.NullVal(cty.DynamicPseudoType)},
				},
				Action: "create",
			},
		},
		views.HookMsg{
			BaseMsg: views.BaseMsg{
				Level:     "info",
				Message:   "random_pet.animal: Creation complete after 0s [id=smart-lizard]",
				Type:      "apply_complete",
				Module:    "terraform.ui",
				TimeStamp: mustUnmarshalTime(t, `"2021-05-25T13:32:41.826179-04:00"`),
			},
			Hook: vjson.OperationComplete{
				Resource: vjson.ResourceAddr{
					Addr:            "random_pet.animal",
					Module:          "",
					Resource:        "random_pet.animal",
					ImpliedProvider: "random",
					ResourceType:    "random_pet",
					ResourceName:    "animal",
					ResourceKey:     ctyjson.SimpleJSONValue{Value: cty.NullVal(cty.DynamicPseudoType)},
				},
				Action:  "create",
				IDKey:   "id",
				IDValue: "smart-lizard",
				Elapsed: 0,
			},
		},
		views.ChangeSummaryMsg{
			BaseMsg: views.BaseMsg{
				Level:     "info",
				Message:   "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.",
				Type:      "change_summary",
				Module:    "terraform.ui",
				TimeStamp: mustUnmarshalTime(t, `"2021-05-25T13:32:41.869168-04:00"`),
			},
			Changes: &vjson.ChangeSummary{
				Add:       1,
				Operation: vjson.OperationApplied,
			},
		},
		views.OutputMsg{
			BaseMsg: views.BaseMsg{
				Level:     "info",
				Message:   "Outputs: 1",
				Type:      "outputs",
				Module:    "terraform.ui",
				TimeStamp: mustUnmarshalTime(t, `"2021-05-25T13:32:41.869280-04:00"`),
			},
			Outputs: vjson.Outputs{
				"pets": {
					Sensitive: false,
					Type:      "string",
					Value:     json.RawMessage([]byte(`"smart-lizard"`)),
				},
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	for _, input := range inputs {
		_, err := buf.WriteString(input)
		buf.WriteString("\n")
		require.NoError(t, err)
	}
	reader := reader.NewReader(buf, io.Discard)

	for i := 0; i < len(inputs); i++ {
		msg, err := reader.Next()
		require.NoError(t, err)
		require.Equal(t, expects[i], msg)
	}

	_, err := reader.Next()
	require.Equal(t, io.EOF, err)

	_, err = reader.Next()
	require.Equal(t, io.EOF, err)
}
