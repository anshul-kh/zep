package core

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/anshul-kh/zep-core/internal/logger"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/shirou/gopsutil/v3/process"
)

/**
* ZepCore handles all the core functions like
* creating and managing lifespan of processes.
*
* It Uses A ID (used internally for management)
* We named this id as ```zid``` (int)
 */

// TODO: Create a bimap structure for bi-directional mapping of data
type ZepCore struct {
	zlog          map[int]*logger.Logger // * maps loggers to each of their processes
	pidStore      map[int]int            // * maps zid's to pid's
	logStore      map[int]string         // * maps id's to log file names
	freeIdSet     []int                  // * keeps track of freed(un-used) id's for re-use
	binPathStore  map[int]string         // * keeps track of binary path for each process (helps to restart)
	argsStore     map[int][]string       // * keeps track of args passed to each process initially
	nameStore     map[string]int         // * keeps track of name assigned to each process
	idToNameStore map[int]string         // * keeps track of id having some names
}

// * log dir path
const (
	logDir = "/var/log/zep"
)

// * default core instance
func NewCore() *ZepCore {
	return &ZepCore{
		pidStore:      make(map[int]int, 10),
		logStore:      make(map[int]string, 10),
		zlog:          make(map[int]*logger.Logger, 10),
		freeIdSet:     []int{},
		binPathStore:  make(map[int]string, 10),
		argsStore:     make(map[int][]string, 10),
		nameStore:     make(map[string]int, 10),
		idToNameStore: make(map[int]string, 10),
	}
}

// * list all the processes
// * return a map of ```zid``` to name
func (z *ZepCore) ListProcs() map[int]string {
	return z.idToNameStore
}

/**
* spawn new process or restart existing process
* extId is used to identify existing process
* extId (default to -1) for new process
 */
func (z *ZepCore) SpawnProcess(binaryPath string, extId int, name string, args ...string) error {
	var id int

	name = strings.TrimSpace(name)
	nameLen := len(name)

	// Check for name conflict if creating a new process
	if extId <= 0 && nameLen > 0 {
		_, exists := z.nameStore[name]
		if exists {
			return fmt.Errorf("a process already exists with that name: %s", name)
		}
	}

	// Re-use the zid's if extId provided
	if extId >= 0 {
		extName, ok := z.idToNameStore[extId]
		if !ok || extName == "" {
			// generate a random name if none exists
			name = namesgenerator.GetRandomName(0)
		} else {
			if nameLen > 0 && extName != name {
				return fmt.Errorf("provided name (%s) does not match the existing name (%s) for id %d", name, extName, extId)
			}
			name = extName
		}
		id = extId
	} else if len(z.freeIdSet) > 0 {
		// Reuse freed IDs
		id = z.freeIdSet[0]
		z.freeIdSet = z.freeIdSet[1:]
	} else {
		// Create new id
		id = len(z.pidStore) + 1
	}

	logFileName := fmt.Sprintf("zep-%d.log", id)

	newLogger, err := logger.NewLogger(logFileName)
	if err != nil {
		return fmt.Errorf("error while creating logger: %v", err)
	}

	file, err := os.OpenFile(newLogger.LogPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error while binding logger: %v", err)
	}

	z.zlog[id] = newLogger
	z.logStore[id] = logFileName

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = file
	cmd.Stderr = file

	// Setsid ensures new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("error while running the process: %v", err)
	}

	z.pidStore[id] = cmd.Process.Pid
	z.binPathStore[id] = binaryPath
	z.argsStore[id] = args

	if len(name) > 0 {
		z.nameStore[name] = id
		z.idToNameStore[id] = name
	}

	newLogger.Info(fmt.Sprintf("Process started with PID: %d, Name: %s", cmd.Process.Pid, name))

	return nil
}

// kill existing process using their assinged name
func (z *ZepCore) KillProcessByName(name string) error {

	name = strings.TrimSpace(name)

	zid, ok := z.nameStore[name]
	if !ok {
		return fmt.Errorf("unable to find a process with name :%s", name)
	}

	return z.KillProcess(zid)
}

// kill existing process using their ```zid```
func (z *ZepCore) KillProcess(id int) error {
	pid, ok := z.pidStore[id]
	if !ok {
		return fmt.Errorf("failed to find the process.")
	}

	p, err := os.FindProcess(int(pid))
	if err != nil {
		return fmt.Errorf("failed to find the process:%v", err)
	}

	err = p.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("failed to kill process:%v", err)
	}

	lgr, ok := z.zlog[id]

	if ok {
		lgr.Warn("process terminated.")
	}
	// remove their data from ZepCore instance
	delete(z.zlog, id)
	delete(z.logStore, id)
	delete(z.pidStore, id)
	delete(z.argsStore, id)
	delete(z.binPathStore, id)

	name, _ := z.idToNameStore[id]
	delete(z.nameStore, name)
	delete(z.idToNameStore, id)

	// mark the id as free to re-use
	z.freeIdSet = append(z.freeIdSet, id)

	return nil
}

// show stats of process using their assinged name
func (z *ZepCore) ShowStatsByName(name string) (map[string]interface{}, error) {

	name = strings.TrimSpace(name)

	zid, ok := z.nameStore[name]
	if !ok {
		return map[string]interface{}{}, fmt.Errorf("unable to find a process with name :%s", name)
	}

	return z.ShowStats(zid)
}

// show stats of process using their assinged ```zid```
func (z *ZepCore) ShowStats(id int) (map[string]interface{}, error) {
	pid, ok := z.pidStore[id]

	if !ok {
		return map[string]interface{}{}, fmt.Errorf("failed to find process with ID:%d", id)
	}

	proc, err := process.NewProcess(int32(pid))

	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("failed to montor process:%v", err)
	}

	if isRunning, err := proc.IsRunning(); err != nil || !isRunning {
		return map[string]interface{}{}, fmt.Errorf("failed to montor process:%v", err)
	}

	memInfo, _ := proc.MemoryInfo()
	cpu, _ := proc.CPUPercent()
	name, _ := proc.Name()

	var info map[string]interface{}
	info = map[string]interface{}{
		"mem":  memInfo,
		"cpu":  cpu,
		"name": name,
	}

	return info, nil
}

/**
* * handle exiting of process using their assinged ```zid```
* * simply re-start the process
* TODO: define max number of tries
 */
func (z *ZepCore) HandleProcessExit(id int) {
	if lgr, ok := z.zlog[id]; ok {
		lgr.Error("process has been terminated.")
		lgr.Info("trying to restart the process...")
	}

	extId := -1

	_, ok := z.pidStore[id]
	if ok {
		extId = id
	}

	binPath := z.binPathStore[id]
	args := z.argsStore[id]
	name := z.idToNameStore[id]

	z.SpawnProcess(binPath, extId, name, args...)
}

/**
* * monitor each process on regular interval
* * default interval is 5 seconds
* TODO: allow configuration of time interval
 */
func (z *ZepCore) MonitorProcesses() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Print("Monitor crashed:", r)
			}
		}()

		for range ticker.C {
			for id, pid := range z.pidStore {
				log := z.zlog[id]
				log.Debug("Monitoring process with PID:", pid)

				proc, err := process.NewProcess(int32(pid))
				if err != nil {
					log.Warn("Process.NewProcess failed:", err)
					z.HandleProcessExit(id)
					continue
				}

				statuses, err := proc.Status()
				if err != nil {
					log.Warn("Failed to get process status:", err)
					z.HandleProcessExit(id)
					continue
				}

				isZombie := false
				for _, st := range statuses {
					if st == "zombie" {
						isZombie = true
						break
					}
				}

				if isZombie {
					log.Warn("Process is a zombie (Z), marking as exited")
					z.HandleProcessExit(id)
					continue
				}

				isRunning, err := proc.IsRunning()
				if err != nil {
					log.Warn("IsRunning check failed:", err)
					z.HandleProcessExit(id)
					continue
				}

				if !isRunning {
					log.Warn("Process is not running anymore")
					z.HandleProcessExit(id)
				}
			}
		}

	}()
}

/*
* * shows logs of process using thier name
* * logging is implemented by simply tailing their log files
* * each process has its own seperate log file
 */
func (z *ZepCore) WatchProcessByName(name string) (<-chan string, error) {
	name = strings.TrimSpace(name)

	if len(name) <= 0 {
		return nil, fmt.Errorf("Unable to find process with name: %s", name)
	}

	zid, ok := z.nameStore[name]
	if !ok {
		return nil, fmt.Errorf("unable to find a process with name :%s", name)
	}

	return z.WatchProcess(zid)
}

// shows log using their ```zid```
func (z *ZepCore) WatchProcess(id int) (<-chan string, error) {
	logFile, ok := z.logStore[id]
	if !ok {
		return nil, fmt.Errorf("failed to find the logfile of process with id:%d", id)
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", logDir, logFile))
	if err != nil {
		return nil, fmt.Errorf("failed to open log file(%s):%w", logFile, err)
	}

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		file.Close()
		return nil, err
	}

	lines := make(chan string)

	go func() {
		defer file.Close()
		defer close(lines)

		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if n > 0 {
				lines <- string(buf[:n])
			}
			if err != nil && err != io.EOF {
				fmt.Printf("read error: %v\n", err)
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return lines, nil
}

/*
* * stop core - stops the daemon
* ! this removes all the existing running process
* ! their data is also removed with their logs
 */
func (z *ZepCore) StopCore() {
	for id, pid := range z.pidStore {
		logFile := z.logStore[id]
		os.Remove(fmt.Sprintf("%s/%s", logDir, logFile))
		p, err := os.FindProcess(pid)
		if err == nil {
			p.Signal(syscall.SIGTERM)
		}
	}
}
