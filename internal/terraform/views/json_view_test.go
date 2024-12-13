package views_test

import (
	"testing"
	"time"

	gojson "encoding/json"

	"github.com/magodo/pipeform/internal/terraform/views"
	"github.com/magodo/pipeform/internal/terraform/views/json"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

var timE, _ = time.Parse(time.RFC3339, "2024-12-09T10:25:00Z")

var resourceAddr = json.ResourceAddr{
	Addr:            "random_pet.animal",
	Module:          "",
	Resource:        "random_pet.animal",
	ImpliedProvider: "random",
	ResourceType:    "random_pet",
	ResourceName:    "animal",
	ResourceKey:     ctyjson.SimpleJSONValue{Value: cty.NullVal(cty.DynamicPseudoType)},
}

var change = json.ResourceInstanceChange{
	Resource:         resourceAddr,
	PreviousResource: nil,
	Action:           json.ActionCreate,
	Reason:           json.ReasonRequested,
	Importing:        nil,
	GeneratedConfig:  "",
}

var changeSummary = json.ChangeSummary{
	Add:       1,
	Change:    2,
	Import:    3,
	Remove:    4,
	Operation: json.OperationApplied,
}

var output = json.Output{
	Sensitive: true,
	Type:      []byte("123"),
	Value:     []byte("321"),
	Action:    json.ActionCreate,
}

func newBaseMsg(typ json.MessageType) views.BaseMsg {
	return views.BaseMsg{
		Level:     "info",
		Message:   "base message",
		Module:    "terraform.ui",
		TimeStamp: timE,
		Type:      typ,
	}
}

func TestMarshal(t *testing.T) {
	cases := []struct {
		name   string
		msg    views.Message
		expect string
	}{
		{
			name: "Version Message",
			msg: views.VersionMsg{
				BaseMsg:   newBaseMsg(json.MessageVersion),
				Terraform: "1.10.0",
				UI:        "0.1.0",
			},
			expect: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "version",
  "terraform": "1.10.0",
  "ui": "0.1.0"
}
`,
		},
		{
			name: "Log Message",
			msg: views.LogMsg{
				BaseMsg: newBaseMsg(json.MessageLog),
				KVs:     map[string]interface{}{"k1": "v1"},
			},
			expect: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "log",
  "k1": "v1"
}
`,
		},
		{
			name: "Diagnostic Message",
			msg: views.DiagnosticsMsg{
				BaseMsg: newBaseMsg(json.MessageDiagnostic),
				Diagnostic: &json.Diagnostic{
					Severity: "sev1",
					Summary:  "summary1",
					Detail:   "detail1",
					Address:  "foo.bar",
					Range: &json.DiagnosticRange{
						Filename: "file.tf",
						Start: json.Pos{
							Line:   1,
							Column: 1,
							Byte:   1,
						},
						End: json.Pos{
							Line:   1,
							Column: 1,
							Byte:   1,
						},
					},
				},
			},
			expect: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "diagnostic",
  "diagnostic": {
    "address": "foo.bar",
	"detail": "detail1",
	"range": {
	  "filename": "file.tf",
	  "start": {
	    "line": 1,
		"column": 1,
		"byte": 1
	  },
	  "end": {
	    "line": 1,
		"column": 1,
		"byte": 1
	  }
	},
	"severity": "sev1",
	"summary": "summary1"
  }
}
`,
		},
		{
			name: "Planned Change Message",
			msg: views.PlannedChangeMsg{
				BaseMsg: newBaseMsg(json.MessagePlannedChange),
				Change:  &change,
			},
			expect: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "planned_change",
  "change": {
	  "resource": {
		"addr": "random_pet.animal",
		"implied_provider": "random",
		"module": "",
		"resource": "random_pet.animal",
		"resource_key": null,
		"resource_type": "random_pet",
		"resource_name": "animal"
	  },
	  "action": "create",
	  "reason": "requested"
  }
}
`,
		},
		{
			name: "Resource Drift Message",
			msg: views.ResourceDriftMsg{
				BaseMsg: newBaseMsg(json.MessageResourceDrift),
				Change:  &change,
			},
			expect: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "resource_drift",
  "change": {
	  "resource": {
		"addr": "random_pet.animal",
		"implied_provider": "random",
		"module": "",
		"resource": "random_pet.animal",
		"resource_key": null,
		"resource_type": "random_pet",
		"resource_name": "animal"
	  },
	  "action": "create",
	  "reason": "requested"
  }
}
`,
		},
		{
			name: "Change Summary Message",
			msg: views.ChangeSummaryMsg{
				BaseMsg: newBaseMsg(json.MessageChangeSummary),
				Changes: &changeSummary,
			},
			expect: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "change_summary",
  "changes": {
    "add": 1,
	"change": 2,
	"import": 3,
	"remove": 4,
	"operation": "apply"
  }
}
`,
		},
		{
			name: "Output Message",
			msg: views.OutputMsg{
				BaseMsg: newBaseMsg(json.MessageOutputs),
				Outputs: &output,
			},
			expect: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "outputs",
  "outputs": {
	"action": "create",
	"sensitive": true,
	"type": 123,
	"value": 321
  }
}
`,
		},
		{
			name: "Hook Message (Operation Start)",
			msg: views.HookMsg{
				BaseMsg: newBaseMsg(json.MessageApplyStart),
				Hook: json.OperationStart{
					Resource: resourceAddr,
					Action:   json.ActionCreate,
					IDKey:    "id",
					IDValue:  "/foo/bar",
				},
			},
			expect: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "apply_start",
  "hook": {
	  "resource": {
		"addr": "random_pet.animal",
		"implied_provider": "random",
		"module": "",
		"resource": "random_pet.animal",
		"resource_key": null,
		"resource_type": "random_pet",
		"resource_name": "animal"
	  },
	  "action": "create",
	  "id_key": "id",
	  "id_value": "/foo/bar"
  }
}
`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			b, err := gojson.Marshal(tt.msg)
			require.NoError(t, err)
			require.JSONEq(t, tt.expect, string(b))
		})
	}
}

func TestUnmarshal(t *testing.T) {
	cases := []struct {
		name  string
		input string
		msg   views.Message
	}{
		{
			name: "Version Message",
			input: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "version",
  "terraform": "1.10.0",
  "ui": "0.1.0"
}
`,
			msg: views.VersionMsg{
				BaseMsg:   newBaseMsg(json.MessageVersion),
				Terraform: "1.10.0",
				UI:        "0.1.0",
			},
		},

		{
			name: "Log Message",
			input: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "log",
  "k1": "v1"
}
`,
			msg: views.LogMsg{
				BaseMsg: newBaseMsg(json.MessageLog),
				KVs:     map[string]interface{}{"k1": "v1"},
			},
		},
		{
			name: "Diagnostic Message",
			input: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "diagnostic",
  "diagnostic": {
    "address": "foo.bar",
	"detail": "detail1",
	"range": {
	  "filename": "file.tf",
	  "start": {
	    "line": 1,
		"column": 1,
		"byte": 1
	  },
	  "end": {
	    "line": 1,
		"column": 1,
		"byte": 1
	  }
	},
	"severity": "sev1",
	"summary": "summary1"
  }
}
`,
			msg: views.DiagnosticsMsg{
				BaseMsg: newBaseMsg(json.MessageDiagnostic),
				Diagnostic: &json.Diagnostic{
					Severity: "sev1",
					Summary:  "summary1",
					Detail:   "detail1",
					Address:  "foo.bar",
					Range: &json.DiagnosticRange{
						Filename: "file.tf",
						Start: json.Pos{
							Line:   1,
							Column: 1,
							Byte:   1,
						},
						End: json.Pos{
							Line:   1,
							Column: 1,
							Byte:   1,
						},
					},
				},
			},
		},
		{
			name: "Planned Change Message",
			input: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "planned_change",
  "change": {
	  "resource": {
		"addr": "random_pet.animal",
		"implied_provider": "random",
		"module": "",
		"resource": "random_pet.animal",
		"resource_key": null,
		"resource_type": "random_pet",
		"resource_name": "animal"
	  },
	  "action": "create",
	  "reason": "requested"
  }
}
`,
			msg: views.PlannedChangeMsg{
				BaseMsg: newBaseMsg(json.MessagePlannedChange),
				Change:  &change,
			},
		},
		{
			name: "Resource Drift Message",
			input: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "resource_drift",
  "change": {
	  "resource": {
		"addr": "random_pet.animal",
		"implied_provider": "random",
		"module": "",
		"resource": "random_pet.animal",
		"resource_key": null,
		"resource_type": "random_pet",
		"resource_name": "animal"
	  },
	  "action": "create",
	  "reason": "requested"
  }
}
`,
			msg: views.ResourceDriftMsg{
				BaseMsg: newBaseMsg(json.MessageResourceDrift),
				Change:  &change,
			},
		},
		{
			name: "Change Summary Message",
			input: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "change_summary",
  "changes": {
    "add": 1,
	"change": 2,
	"import": 3,
	"remove": 4,
	"operation": "apply"
  }
}
`,
			msg: views.ChangeSummaryMsg{
				BaseMsg: newBaseMsg(json.MessageChangeSummary),
				Changes: &changeSummary,
			},
		},
		{
			name: "Output Message",
			input: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "outputs",
  "outputs": {
	"action": "create",
	"sensitive": true,
	"type": 123,
	"value": 321
  }
}
`,
			msg: views.OutputMsg{
				BaseMsg: newBaseMsg(json.MessageOutputs),
				Outputs: &output,
			},
		},
		{
			name: "Hook Message (Operation Start)",
			input: `
{
  "@level": "info",
  "@message": "base message",
  "@module": "terraform.ui",
  "@timestamp": "2024-12-09T10:25:00Z",
  "type": "apply_start",
  "hook": {
	  "resource": {
		"addr": "random_pet.animal",
		"implied_provider": "random",
		"module": "",
		"resource": "random_pet.animal",
		"resource_key": null,
		"resource_type": "random_pet",
		"resource_name": "animal"
	  },
	  "action": "create",
	  "id_key": "id",
	  "id_value": "/foo/bar"
  }
}
`,
			msg: views.HookMsg{
				BaseMsg: newBaseMsg(json.MessageApplyStart),
				Hook: json.OperationStart{
					Resource: resourceAddr,
					Action:   json.ActionCreate,
					IDKey:    "id",
					IDValue:  "/foo/bar",
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := views.UnmarshalMessage([]byte(tt.input))
			require.NoError(t, err)
			require.Equal(t, tt.msg, msg)
		})
	}
}
