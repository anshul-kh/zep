package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/anshul-kh/zep-core/internal/logger"
	"github.com/shirou/gopsutil/v3/process"
)

type ZepCore struct {
	zlog         map[int]*logger.Logger
	pidStore     map[int]int
	logStore     map[int]string
	freeIdSet    []int
	binPathStore map[int]string
	argsStore    map[int][]string
}

func NewCore() *ZepCore {
	return &ZepCore{
		pidStore:     make(map[int]int, 10),
		logStore:     make(map[int]string, 10),
		zlog:         make(map[int]*logger.Logger, 10),
		freeIdSet:    []int{},
		binPathStore: make(map[int]string, 10),
		argsStore:    make(map[int][]string, 10),
	}
}

func (z *ZepCore) SpawnNewProcess(binaryPath string, extId int, args ...string) error {
	var id int

	if extId >= 0 {
		id = extId
	} else if len(z.freeIdSet) > 0 {
		id = z.freeIdSet[0]
		z.freeIdSet = z.freeIdSet[1:]
	} else {
		id = int(len(z.pidStore) + 1)
	}

	var logFileName string
	var newLogger *logger.Logger

	if extId > 0 {
		logFileName = z.logStore[extId]
		newLogger = z.zlog[extId]
	} else {
		logFileName = fmt.Sprintf("zep-%d.log", id)
		lgr, err := logger.NewLogger(logFileName)
		if err != nil {
			return fmt.Errorf("error while creating logger:%v", err)
		}
		newLogger = lgr
	}

	file, err := os.OpenFile(newLogger.LogPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)

	if err != nil {
		return fmt.Errorf("error while binding logger:%v", err)
	}

	z.zlog[id] = newLogger
	z.logStore[id] = logFileName

	cmd := exec.Command(binaryPath, args...)

	cmd.Stdout = file
	cmd.Stderr = file

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("error while running the process:%v", err)
	}

	z.pidStore[id] = int(cmd.Process.Pid)
	z.binPathStore[id] = binaryPath
	z.argsStore[id] = args

	newLogger.Info(fmt.Sprintf("Process is started with PID:%s", cmd.Process.Pid))

	return nil
}

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
		lgr.Warn(fmt.Sprint("process terminated."))
	}

	delete(z.zlog, id)
	delete(z.logStore, id)
	delete(z.pidStore, id)
	delete(z.argsStore, id)
	delete(z.binPathStore, id)

	z.freeIdSet = append(z.freeIdSet, id)

	return nil
}

func (z *ZepCore) ShowStats(id int) (map[string]interface{}, error) {
	pid, ok := z.pidStore[id]
	if !ok {
		return nil, fmt.Errorf("failed to find process with ID:%d", id)
	}

	proc, err := process.NewProcess(int32(pid))

	if err != nil {
		return nil, fmt.Errorf("failed to montor process:%v", err)
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

	z.SpawnNewProcess(binPath, extId, args...)
}

func (z *ZepCore) MonitorProcesses() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for range ticker.C {
			for id, pid := range z.pidStore {
				proc, err := process.NewProcess(int32(pid))
				if err != nil {
					z.HandleProcessExit(id)
					continue
				}

				isRunning, err := proc.IsRunning()
				if err != nil || !isRunning {
					z.HandleProcessExit(id)
				}
			}
		}
	}()
}


func (z *ZepCore) WatchProcess(id int) {
	logFile , ok := z.logStore[id]
	if !ok {
		fmt.Printf("failed to find the logfile of process with id:%d",id)
		return
	}

	file , err := os.Open(logFile)
	if err != nil {
		fmt.Printf("failed to open log file")
		return
	}
	defer file.Close()

	file.Seek(0,io.SeekEnd)

	sc := bufio.NewScanner(file)
	for {
		for sc.Scan() {
			fmt.Print(sc.Text())
		}
		
		time.Sleep(time.Second)
	}
}
