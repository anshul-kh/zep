package master

import (
	"fmt"
	"net/http"

	"github.com/anshul-kh/zep-core/internal/core"
	logger "github.com/anshul-kh/zep-core/internal/logger"
)

const (
	host = "127.0.0.1"
	port = "3746"
)

type Master struct {
	IsRunning bool
	log       *logger.Logger
	core      *core.ZepCore
}

func NewMaster(logger *logger.Logger) *Master {
	z := core.NewCore()
	return &Master{
		IsRunning: false,
		log:       logger,
		core:      z,
	}
}

func (m *Master) StartMaster() error {

	if m.IsRunning {
		return fmt.Errorf("master node is already running...")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		m.log.Info("server up and running")
		fmt.Fprint(w, "server up and running")
	})

	err := http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), mux)

	if err != nil {
		return fmt.Errorf("failed to start master process: %v", err)
	}

	m.IsRunning = true

	m.core.MonitorProcesses()

	return nil
}
