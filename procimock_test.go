// proci unit tests
package proci

import (
	"testing"
)

func TestMock(t *testing.T) {
	pm := GenerateMock(10)
	nbrProcesses := len(pm.GetProcessPids())
	if nbrProcesses != 10 {
		t.Fatalf("Expected 10 mocked processes but it was %d", nbrProcesses)
	}
	
	// GetProcessMemoryUsage
	mem, errMem := pm.GetProcessMemoryUsage(8)
	if errMem != nil {
		t.Fatal("Expected no error for GetProcessMemoryUsage")
	}
	if mem != 1008 {
		t.Fatal("Expected 1008 bytes for PID 8")
	}
	pm.Processes[8].DoFailMemoryUsage = true
	_, errMem2 := pm.GetProcessMemoryUsage(8)
	if errMem2 == nil {
		t.Fatal("Expected error for GetProcessMemoryUsage")
	}
	_, errMem3 := pm.GetProcessMemoryUsage(1234)
	if errMem3 == nil {
		t.Fatal("Expected error for GetProcessMemoryUsage for invalid PID")
	}
	
	// GetProcessPath
	path, errPath := pm.GetProcessPath(3)
	if errPath != nil {
		t.Fatal("Expected no error for GetProcessPath")
	}
	if path != "path_3" {
		t.Fatal("Expected path path_3 PID 3")
	}
	pm.Processes[3].DoFailPath = true
	_, errPath2 := pm.GetProcessPath(3)
	if errPath2 == nil {
		t.Fatal("Expected error for GetProcessPath")
	}
	_, errPath3 := pm.GetProcessPath(1234)
	if errPath3 == nil {
		t.Fatal("Expected error for GetProcessPath for invalid PID")
	}
	
	// GetProcessCommandLine
	cmd, errCmd := pm.GetProcessCommandLine(9)
	if errCmd != nil {
		t.Fatal("Expected no error for GetProcessCommandLine")
	}
	if cmd != "command_line_9" {
		t.Fatal("Expected path path_9 PID 9")
	}
	pm.Processes[9].DoFailCommandLine = true
	_, errCmd2 := pm.GetProcessCommandLine(9)
	if errCmd2 == nil {
		t.Fatal("Expected error for GetProcessCommandLine")
	}
	_, errCmd3 := pm.GetProcessCommandLine(1234)
	if errCmd3 == nil {
		t.Fatal("Expected error for GetProcessCommandLine for invalid PID")
	}
	
	// GetMemoryStatus
	memStatus, errMemStat := pm.GetMemoryStatus()
	if errMemStat != nil {
		t.Fatal("Expected no error for GetMemoryStatus")
	}
	if (memStatus.MemoryLoad != 50 && 
	    memStatus.TotalPhys != 4 * 1024*1024*1024 &&
			memStatus.AvailPhys != 2 * 1024*1024*1024) {
			t.Fatal("Invalid values in GetMemoryStatus")
	}
	pm.DoFailMemStatus = true
	_, errMemStat2 := pm.GetMemoryStatus()
	if errMemStat2 == nil {
		t.Fatal("Expected error for GetMemoryStatus")
	}
	
	

	
}

