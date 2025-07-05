package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	logger "github.com/anshul-kh/zep-core/internal/logger"
	"github.com/anshul-kh/zep-core/internal/master"
	dm "github.com/sevlyar/go-daemon"
)

// * file and folder paths used by the daemon
var (
	logDir        = "/var/log/zep"         // log dir for all processes
	pidDir        = "/var/run/zep"         // pid file dir for all processes
	logFilePath   = "/var/log/zep/zep.log" // core daemon log file
	pidFilePath   = "/var/run/zep/zep.pid" // core daemon pid file
	masterLogFile = "master.log"           // master process log file (*dir path same as $logDir)
)

// daemon default commands
const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
	STATUS  = "status"
)

// * setup dirs to be used by daemon on the system
func prepDirs() {
	err := os.MkdirAll(pidDir, 0755)
	if err != nil {
		log.Fatalf("failed to pid directory:%v", err)
	}

	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("failed to create log dir:%v", err)
	}
}

/**
* start daemon (run with ```start``` command)
* created using ```go-daemon``` lib
 */
func startDaemon() error {
	prepDirs()

	p, err := findDaemonProcess()
	if err == nil {
		return fmt.Errorf("daemon is already running with pid:%d", p.Pid)
	}
	// cntx := &dm.Context{
	// 	LogFileName: logFilePath,
	// 	LogFilePerm: 0664,
	// 	PidFileName: pidFilePath,
	// 	PidFilePerm: 0664,
	// 	Umask:       027,
	// 	WorkDir:     "/",
	// }

	// d, err := cntx.Reborn()
	// if err != nil {
	// 	return fmt.Errorf("failed to create daemon:%v", err)
	// }

	// if d != nil {
	// 	fmt.Println("Daemon process started successfully.")
	// 	return nil
	// }

	// defer cntx.Release()

	logger, err := logger.NewLogger(masterLogFile)

	if err != nil {
		return fmt.Errorf("failed to start logger service:-\n=====\n%v\n====\n", err)
	}

	// setting up master process
	m := master.NewMaster(logger)

	if !m.IsRunning {
		go m.StartMaster()
	}

	err = os.WriteFile(pidFilePath, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PID file: %v", err)
	}

	fmt.Printf("daemon started with pid:%d", os.Getpid())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	log.Printf("recieved terminating signal:%v", sig)
	m.StopMaster()
	return nil
}

/**
* stop daemon (run with ```stop``` command)
* ! all active processes dies and their log files are deleted
* ! killing processes handled through core pkg
 */
func stopDaemon() error {
	p, err := findDaemonProcess()
	if err != nil {
		return err
	}
	err = p.Signal(syscall.SIGTERM)

	if err != nil {
		return fmt.Errorf("failed to stop daemon: %v", err)
	}

	err = os.Remove(pidFilePath)
	if err != nil {
		return fmt.Errorf("failed to stop daemon properly")
	}

	log.Print("daemon stopped")
	return nil
}

/**
* restart daemon (run with ```restart``` command)
 */
func restartDaemon() {
	err := stopDaemon()
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	// start a fresh process
	attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	process, err := os.StartProcess(os.Args[0], []string{os.Args[0], START}, attr)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Printf("daemon restarted with pid %d\n", process.Pid)
}

/**
* status of daemon (run with ```status``` command)
* returns boolean as status
 */
func statusOfDaemon() bool {
	d := &dm.Context{
		PidFileName: pidFilePath,
	}

	p, err := d.Search()

	if err != nil {
		log.Fatalf("daemon not found,(may be not running)")
		return false
	}

	log.Printf("daemon running..., pid:%d", p.Pid)
	return true
}

func findDaemonProcess() (*os.Process, error) {
	data, err := os.ReadFile(pidFilePath)

	if err != nil {
		return nil, fmt.Errorf("failed to stop daemon,(may be not running)")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("invalid pid in file: %w", err)
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to stop daemon,(may be not running)")
	}

	return p, nil
}

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Printf("usage: zep-daemon start|stop|restart|status\n")
		return
	}

	switch args[1] {
	case START:
		err := startDaemon()
		if err != nil {
			fmt.Print(err.Error())
			return
		}
		return
	case STOP:
		err := stopDaemon()
		if err != nil {
			fmt.Print(err.Error())
		}
		return
	case RESTART:
		restartDaemon()
		return
	case STATUS:
		statusOfDaemon()
		return
	default:
		fmt.Printf("usage: zep-daemon start|stop|restart|status")
	}
}
