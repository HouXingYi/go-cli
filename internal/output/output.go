package output

import (
	"encoding/json"
	"fmt"
	"io"

	"ziniao/internal/apperr"
	"ziniao/internal/config"
)

type Renderer struct {
	format string
	out    io.Writer
	errOut io.Writer
}

type envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *errorBody  `json:"error,omitempty"`
}

type errorBody struct {
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
	Kind    string `json:"kind,omitempty"`
}

func New(format string, out, errOut io.Writer) Renderer {
	if format == "" {
		format = config.DefaultOutput
	}
	return Renderer{
		format: format,
		out:    out,
		errOut: errOut,
	}
}

func (r Renderer) Success(message string, data interface{}) error {
	if r.format == config.OutputJSON {
		return writeJSON(r.out, envelope{
			Success: true,
			Data:    data,
		})
	}

	_, err := fmt.Fprintln(r.out, message)
	if err != nil {
		return apperr.Wrap(apperr.KindOutput, "failed to write output", "check whether stdout is available.", err)
	}
	return nil
}

func (r Renderer) Error(err error) error {
	if err == nil {
		return nil
	}

	if r.format == config.OutputJSON {
		writeErr := writeJSON(r.errOut, envelope{
			Success: false,
			Error: &errorBody{
				Message: err.Error(),
				Hint:    apperr.Hint(err),
				Kind:    apperr.Kind(err),
			},
		})
		if writeErr != nil {
			return writeErr
		}
		return nil
	}

	if _, writeErr := fmt.Fprintf(r.errOut, "Error: %s\n", err.Error()); writeErr != nil {
		return apperr.Wrap(apperr.KindOutput, "failed to write error output", "check whether stderr is available.", writeErr)
	}
	if hint := apperr.Hint(err); hint != "" {
		if _, writeErr := fmt.Fprintf(r.errOut, "Hint: %s\n", hint); writeErr != nil {
			return apperr.Wrap(apperr.KindOutput, "failed to write error output", "check whether stderr is available.", writeErr)
		}
	}
	return nil
}

func writeJSON(w io.Writer, value interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return apperr.Wrap(apperr.KindOutput, "failed to write json output", "check whether stdout or stderr is available.", err)
	}
	return nil
}
