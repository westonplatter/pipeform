package json

import "encoding/json"

type Output struct {
	Sensitive bool            `json:"sensitive"`
	Type      json.RawMessage `json:"type,omitempty"`
	Value     json.RawMessage `json:"value,omitempty"`
	Action    ChangeAction    `json:"action,omitempty"`
}

type Outputs map[string]Output
