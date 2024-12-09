package json

import "encoding/json"

type Function struct {
	Name            string          `json:"name"`
	Params          []FunctionParam `json:"params"`
	VariadicParam   *FunctionParam  `json:"variadic_param,omitempty"`
	ReturnType      json.RawMessage `json:"return_type"`
	Description     string          `json:"description,omitempty"`
	DescriptionKind string          `json:"description_kind,omitempty"`
}

type FunctionParam struct {
	Name            string          `json:"name"`
	Type            json.RawMessage `json:"type"`
	Description     string          `json:"description,omitempty"`
	DescriptionKind string          `json:"description_kind,omitempty"`
}
