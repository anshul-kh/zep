package master

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

/*
* * HTTP Handlers to execute http request
* * calls core functions of ZepCore
 */

type NewProcRequest struct {
	BinaryPath string   `json:"binaryPath"`
	Args       []string `json:"args"`
	Name       string   `json:"name"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (m *Master) listProcs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	data := m.core.ListProcs()
	m.log.Debug(data, "es")
	WriteJSON(w, http.StatusOK, SuccessResponse{Success: true, Msg: "List Fetched Successfully!!", Data: data}, nil)
}

func (m *Master) newProc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req NewProcRequest

	if err := DecodeJSONBody(w, r, &req); err != nil {
		if mr, ok := err.(*MalformedRequest); ok {
			WriteError(w, mr.Status, mr.Msg)
		} else {
			WriteError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		return
	}

	err := m.core.SpawnProcess(req.BinaryPath, -1, req.Name, req.Args...)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = WriteJSON(w, http.StatusCreated, SuccessResponse{Success: true, Msg: "process created successfully!!!"}, nil)
	if err != nil {
		m.log.Error(err.Error())
		WriteError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func (m *Master) killProc(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query().Get("q")
	typ := r.URL.Query().Get("type")

	if q == "" || typ == "" {
		WriteError(w, http.StatusBadRequest, "Invalid/Missing Parameters q or type")
		return
	}

	var err error

	switch typ {
	case "name":
		q = strings.TrimSpace(q)
		err = m.core.KillProcessByName(q)
	case "id":
		zid, err := strconv.Atoi(q)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if zid <= 0 {
			WriteError(w, http.StatusInternalServerError, "Invalid ID")
			return
		}

		err = m.core.KillProcess(zid)
	default:
		WriteError(w, http.StatusBadRequest, "Missing/Invalid Parameters q or type")
		return
	}

	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, SuccessResponse{Success: true, Msg: "process killed successfully!!"}, nil)
}

func (m *Master) showStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query().Get("q")
	typ := r.URL.Query().Get("type")

	if q == "" || typ == "" {
		WriteError(w, http.StatusBadRequest, "Invalid/Missing Parameters q or type")
		return
	}

	var stats map[string]interface{}
	var err error

	switch typ {
	case "name":
		q = strings.TrimSpace(q)
		stats, err = m.core.ShowStatsByName(q)

		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "id":
		zid, err := strconv.Atoi(q)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		if zid <= 0 {
			WriteError(w, http.StatusBadRequest, "Invalid ID")
			return
		}

		stats, err = m.core.ShowStats(zid)

		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

	default:
		WriteError(w, http.StatusBadRequest, "Missing/Invalid Parameters q or type")
		return
	}

	data := make([]any, 1)
	data[0] = stats

	WriteJSON(w, http.StatusOK, SuccessResponse{Success: true, Msg: "Stats Fetched Successfully!!", Data: data}, nil)
}

func (m *Master) watchProc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		WriteError(w, http.StatusMethodNotAllowed, ",method not allowed")
		return
	}

	q := r.URL.Query().Get("q")
	typ := r.URL.Query().Get("type")

	if q == "" || typ == "" {
		WriteError(w, http.StatusBadRequest, "Invalid/Missing Parameters q or type")
		return
	}

	var lines <-chan string
	var err error

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	switch typ {
	case "name":
		q = strings.TrimSpace(q)
		lines, err = m.core.WatchProcessByName(q)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "id":
		zid, err := strconv.Atoi(q)
		if err != nil || zid <= 0 {
			WriteError(w, http.StatusInternalServerError, "Invalid ID")
			return
		}

		lines, err = m.core.WatchProcess(zid)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		WriteError(w, http.StatusBadRequest, "Missing/Invalid Parameters q or type")
		return
	}

	for line := range lines {
		fmt.Fprintf(w, "%s", line)
		flusher.Flush()
	}
}
