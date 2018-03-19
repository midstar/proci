package proci

import (
	"fmt"
)

type ProcessMock struct{
	Pid                uint32
	Path               string
	CommandLine        string
	MemoryUsage        uint64
	
	DoFailPath         bool   // If true, fail GetProcessPath
	DoFailCommandLine  bool   // If true, fail GetProcessCommandLine
	DoFailMemoryUsage  bool   // If true, fail GetProcessMemoryUsage
}

// ProciMock is a mock implementation for the proci Interface. It is intended
// for mocking of proci during unit testing.
type ProciMock struct{
	MemStatus        *MemoryStatus
	DoFailMemStatus  bool   // If true, fail GetMemoryStatus
	
	Processes        map[uint32]*ProcessMock
}

// GenerateMock generate a mock with mock processes. It will start from
// PID 0 up to numberOfProcesses - 1. 
func GenerateMock(numberOfProcesses int) *ProciMock {
	memoryStatus := MemoryStatus{MemoryLoad: 50, TotalPhys: 4 * 1024*1024*1024, AvailPhys: 2 * 1024*1024*1024}
	processes := make(map[uint32]*ProcessMock)
	for i := 0; i < numberOfProcesses; i++ {
		pid := uint32(i)
		process := ProcessMock{
			Pid : pid,
			Path : fmt.Sprintf("path_%d", i),
			CommandLine : fmt.Sprintf("command_line_%d", i),
			MemoryUsage : uint64(1000 + i),
			DoFailPath : false,
			DoFailCommandLine : false,
			DoFailMemoryUsage : false}
		processes[pid] = &process
	}
	return &ProciMock{
		MemStatus : &memoryStatus,
		DoFailMemStatus : false,
		Processes : processes}
}

func (s ProciMock) GetMemoryStatus() (*MemoryStatus, error) {
	if s.DoFailMemStatus {
		return nil, fmt.Errorf("GetMemoryStatus Mock intentional failure")
	}
	return s.MemStatus, nil
}

func (s ProciMock) GetProcessPids() []uint32 {
	pids := make([]uint32, 0, len(s.Processes))
  for pid, _ := range s.Processes { 
    pids = append(pids, pid)
  }
	return pids
}

func (s ProciMock) GetProcessMemoryUsage(pid uint32) (uint64, error) {
	process, hasPid := s.Processes[pid]
	if !hasPid {
		return 0, fmt.Errorf("PID %d does not exist", pid)
	}
	if process.DoFailMemoryUsage {
		return 0, fmt.Errorf("GetProcessMemoryUsage Mock intentional failure")
	}
	return process.MemoryUsage, nil
}

func (s ProciMock) GetProcessPath(pid uint32) (string, error) {
	process, hasPid := s.Processes[pid]
	if !hasPid {
		return "", fmt.Errorf("PID %d does not exist", pid)
	}
	if process.DoFailPath {
		return "", fmt.Errorf("GetProcessPath Mock intentional failure")
	}
	return process.Path, nil
}

func (s ProciMock) GetProcessCommandLine(pid uint32) (string, error) {
	process, hasPid := s.Processes[pid]
	if !hasPid {
		return "", fmt.Errorf("PID %d does not exist", pid)
	}
	if process.DoFailCommandLine {
		return "", fmt.Errorf("GetProcessCommandLine Mock intentional failure")
	}
	return process.CommandLine, nil
}

