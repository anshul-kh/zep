package master

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"message"`
}

type MalformedRequest struct {
	Status int
	Msg    string
}

func (m *MalformedRequest) Error() string {
	return m.Msg
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}, headers http.Header) error {
	for key, vals := range headers {
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	_, err = w.Write(data)
	return err
}

func WriteError(w http.ResponseWriter, status int, msg string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	data := ErrorResponse{
		Success: false,
		Msg:     msg,
	}

	err := WriteJSON(w, status, data, nil)
	return err
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	ct := r.Header.Get("Content-Type")

	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			return &MalformedRequest{
				Status: http.StatusUnsupportedMediaType,
				Msg:    "Content-Type header is not set to application/json",
			}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("badly formed json at potistion:%d", syntaxError.Offset)
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: msg}
		case errors.Is(err, io.ErrUnexpectedEOF):
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: "badly formed json body"}
		case errors.As(err, &unmarshalError):
			msg := fmt.Sprintf("request body conatins invalid field value : %q at postion %d", unmarshalError.Field, unmarshalError.Offset)
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: msg}
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("body contains unknown field: %s", fieldName)
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: msg}
		case errors.Is(err, io.EOF):
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: "body must not be empty"}
		case err.Error() == "http: request body too large":
			return &MalformedRequest{Status: http.StatusBadRequest, Msg: "body size is too large"}
		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return &MalformedRequest{Status: http.StatusBadRequest, Msg: "body must contain a single json"}
	}

	return nil
}
