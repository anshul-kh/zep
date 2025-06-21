package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	logger "github.com/anshul-kh/zep-core/internal/logger"
	"github.com/anshul-kh/zep-core/internal/master"
	dm "github.com/sevlyar/go-daemon"
)

var (
	logDir        = "/var/log/zep"
	pidDir        = "/var/run/zep"
	logFilePath   = "/var/log/zep/zep.log"
	pidFilePath   = "/var/run/zep/zep.pid"
	masterLogFile = "master.log"
)

const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
	STATUS  = "status"
)

func setupFiles() {
	err := os.MkdirAll(pidDir, 0755)
	if err != nil {
		log.Fatalf("failed to pid directory:%v", err)
	}

	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("failed to create log dir:%v", err)
	}
}

func startDaemon() {
	setupFiles()

	cntx := &dm.Context{
		LogFileName: logFilePath,
		LogFilePerm: 0664,
		PidFileName: pidFilePath,
		PidFilePerm: 0664,
		Umask:       027,
		WorkDir:     "/",
	}

	d, err := cntx.Reborn()
	if err != nil {
		log.Printf("failed to create daemon:%v", err)
	}

	if d != nil {
		return
	}

	defer cntx.Release()

	logger, err := logger.NewLogger(masterLogFile)

	if err != nil {
		log.Printf("failed to start logger service:-\n=====\n%v\n====\n", err)
	}

	m := master.NewMaster(logger)

	if !m.IsRunning {
		go m.StartMaster()
	}

	log.Printf("daemon started with pid:%d", os.Getpid())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	log.Printf("recieved terminating signal:%v", sig)
	m.StopMaster()

}

func stopDaemon() {
	d := &dm.Context{
		PidFileName: pidFilePath,
	}

	p, err := d.Search()

	if err != nil || p == nil {
		log.Fatalf("failed to stop daemon,(may be not running)")
	}

	err = p.Signal(syscall.SIGTERM)

	if err != nil {
		log.Fatalf("failed to stop daemon: %v", err)
	}

	err = os.Remove(pidFilePath)
	if err != nil {
		log.Fatal("failed to stop daemon properly")
	}

	log.Print("daemon stopped")
}

func restartDaemon() {
	stopDaemon()
	startDaemon()
}

func statusOfDaemon() {
	d := &dm.Context{
		PidFileName: pidFilePath,
	}

	p, err := d.Search()

	if err != nil {
		log.Fatalf("daemon not found,(may be not running)")
	}

	log.Printf("daemon running..., pid:%d", p.Pid)
}

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Printf("usage: zep start|stop|restart|status\n")
		return
	}

	switch args[1] {
	case START:
		startDaemon()
		return
	case STOP:
		stopDaemon()
		return
	case RESTART:
		restartDaemon()
		return
	case STATUS:
		statusOfDaemon()
		return
	default:
		fmt.Printf("usage: zep start|stop|restart|status")
	}
}
