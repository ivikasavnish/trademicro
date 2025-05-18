package services

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/vikasavnish/trademicro/internal/models"
)

// ProcessService defines the interface for process management
type ProcessService interface {
	StartProcess(req models.StartTradeProcessRequest) (int, error)
	StopProcess(pid int) error
	ListProcesses() ([]string, error)
}

// processService implements the ProcessService interface
type processService struct {
}

// NewProcessService creates a new process service
func NewProcessService() ProcessService {
	return &processService{}
}

// StartProcess starts a new process
func (s *processService) StartProcess(req models.StartTradeProcessRequest) (int, error) {
	cmdArgs := append([]string{"start", req.Script}, req.Args...)
	cmd := exec.Command("python3", append([]string{"trade_manager.py"}, cmdArgs...)...)
	cmd.Dir = "." // run in project root
	err := cmd.Start()
	if err != nil {
		return 0, err
	}

	go cmd.Wait() // Don't block
	return cmd.Process.Pid, nil
}

// StopProcess stops a running process
func (s *processService) StopProcess(pid int) error {
	cmd := exec.Command("python3", "trade_manager.py", "stop", "--pid", strconv.Itoa(pid))
	cmd.Dir = "."
	return cmd.Run()
}

// ListProcesses returns a list of running processes
func (s *processService) ListProcesses() ([]string, error) {
	cmd := exec.Command("python3", "trade_manager.py", "list")
	cmd.Dir = "."
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// The output is plain text, convert to array of lines
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}
