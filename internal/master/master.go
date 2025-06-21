package master

import (
	"fmt"
	"net/http"

	"github.com/anshul-kh/zep-core/internal/core"
	logger "github.com/anshul-kh/zep-core/internal/logger"
	"github.com/anshul-kh/zep-core/internal/network"
)

const (
	host = "127.0.0.1"
	port = "3746"
)

type Master struct {
	IsRunning bool
	log       *logger.Logger
	core      *core.ZepCore
	br        *network.Bridge
}

func NewMaster(logger *logger.Logger) *Master {
	z := core.NewCore()
	br := network.NewBridge()

	return &Master{
		IsRunning: false,
		log:       logger,
		core:      z,
		br:        br,
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

	err := m.br.SetUpBridge()
	if err != nil {
		m.log.Error(err)
	}

	err = http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), mux)

	if err != nil {
		return fmt.Errorf("failed to start master process: %v", err)
	}

	m.IsRunning = true

	m.core.MonitorProcesses()

	return nil
}

func (m *Master) StopMaster() {
	m.log.Warn("Stopping master process...")

	err := m.br.StopBridge()
	if err != nil {
		m.log.Warn(err)
	}

	m.IsRunning = false

	m.core.StopCore()
	m.log.Warn("Master Process Stopped!!")
}
