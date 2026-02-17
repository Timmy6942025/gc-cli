package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type CLIError struct {
	Code           string `json:"code"`
	Reason         string `json:"reason"`
	ActionableHint string `json:"actionable_hint"`
	WebHandoffURL  string `json:"web_handoff_url,omitempty"`
}

func (e CLIError) Error() string {
	if e.ActionableHint == "" {
		return e.Reason
	}
	return fmt.Sprintf("%s (%s)", e.Reason, e.ActionableHint)
}

type Printer struct {
	JSON bool
	Out  io.Writer
}

func (p Printer) Print(v any) error {
	if p.JSON {
		enc := json.NewEncoder(p.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	_, err := fmt.Fprintln(p.Out, v)
	return err
}

func (p Printer) PrintJSON(v any) error {
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func (p Printer) PrintError(err error) error {
	if p.JSON {
		if cliErr, ok := err.(CLIError); ok {
			return p.PrintJSON(cliErr)
		}
		return p.PrintJSON(CLIError{Code: "internal", Reason: err.Error()})
	}
	_, writeErr := fmt.Fprintln(p.Out, "Error:", err.Error())
	return writeErr
}
