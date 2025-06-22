package master

import (
	"fmt"
	"net/http"
	"strconv"
)

type NewProcRequest struct {
	BinaryPath string   `json:"binaryPath"`
	Args       []string `json:"args"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"message"`
	Data    any    `json:"data,omitempty"`
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

	err := m.core.SpawnNewProcess(req.BinaryPath, -1, req.Args...)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
	}

	err = WriteJSON(w, http.StatusCreated, SuccessResponse{Success: true, Msg: "process created successfully!!!"}, nil)
	if err != nil {
		m.log.Error(err.Error())
		WriteError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func (m *Master) killProc(w http.ResponseWriter, r *http.Request) {

	if r.Method != "DELETE" {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := r.PathValue("id")

	zid, err := strconv.Atoi(id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if zid <= 0 {
		WriteError(w, http.StatusInternalServerError, "Invalid ID")
	}

	err = m.core.KillProcess(zid)
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

	id := r.PathValue("id")

	zid, err := strconv.Atoi(id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if zid <= 0 {
		WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	stats, err := m.core.ShowStats(zid)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
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

	id := r.PathValue("id")
	zid, err := strconv.Atoi(id)
	if err != nil || zid <= 0 {
		WriteError(w, http.StatusInternalServerError, "Invalid ID")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	lines, err := m.core.WatchProcess(zid)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for line := range lines {
		fmt.Fprintf(w, "data: %s\n\n", line)
		flusher.Flush()
	}
}
