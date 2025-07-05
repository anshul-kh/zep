package cmd

type MasterResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
