// proci unit tests
package proci

import (
	"testing"
)

func TestGetMemoryStatus(t *testing.T) {
	mStat, err := GetMemoryStatus()
	if err != nil {
		t.Errorf("GetMemoryStatus returned error: %s", err)
	}
	t.Log("MemoryLoad:", mStat.MemoryLoad, "%")
	if mStat.MemoryLoad == 0 {
		t.Errorf("Memory load cannot be 0")
	}
	t.Log("TotalPhys:", mStat.TotalPhys, "B (", mStat.TotalPhys/1024/1024, "MB )")
	if mStat.TotalPhys == 0 {
		t.Errorf("Total memory cannot be 0")
	}
	t.Log("AvailPhys:", mStat.AvailPhys, "B (", mStat.AvailPhys/1024/1024, "MB )")
	if mStat.AvailPhys == 0 {
		t.Errorf("Available memory cannot be 0")
	}
}

func TestGetProcessPids(t *testing.T) {
	pids := GetProcessPids()
	t.Log("Number of processes:", len(pids))
	if len(pids) == 0 {
		t.Errorf("Number of processes cannot be 0")
	}
}

func TestGetProcessMemoryUsage(t *testing.T) {
	pids := GetProcessPids()
	if len(pids) < 10 {
		t.Errorf("Number of pids very low. Number of pids: %d", len(pids))
	}
	pid := pids[10] // Pick a random process
	t.Log("Get memory for process with pid:", pid)
	memoryUsage, err := GetProcessMemoryUsage(pid)
	if err != nil {
		t.Errorf("GetProcessMemoryUsage returned error: %s", err)
	}
	t.Log("Process with pid", pid, "memory usage:", memoryUsage, "B (", memoryUsage/1024/1024, "MB )")
	if memoryUsage == 0 {
		t.Errorf("Process memory usage cannot be 0 bytes")
	}
}

func TestGetProcessPath(t *testing.T) {
	pids := GetProcessPids()
	if len(pids) < 10 {
		t.Errorf("Number of pids very low. Number of pids: %d", len(pids))
	}
	pid := pids[10] // Pick a random process
	t.Log("Get path for process with pid:", pid)
	path, err := GetProcessPath(pid)
	if err != nil {
		t.Errorf("GetProcessPath returned error: %s", err)
	}
	t.Log("Process with pid", pid, "path:", path)
	if path == "" {
		t.Errorf("Process path cannot be empty string.")
	}
}

// This test requires that you are running as administrator. Also,
// since we are picking a random process we might pick a process
// that we cannot access even if administrative access rights are
// used.
func TestGetProcessCommandLine(t *testing.T) {
	pids := GetProcessPids()
	if len(pids) < 10 {
		t.Errorf("Number of pids very low. Number of pids: %d", len(pids))
	}
	pid := pids[10] // Pick a random process
	t.Log("Get command line for process with pid:", pid)
	commandLine, err := GetProcessCommandLine(pid)
	if err != nil {
		t.Errorf("GetProcessCommandLine returned error: %s", err)
	}
	t.Log("Process with pid", pid, "command line:", commandLine)
	if commandLine == "" {
		t.Errorf("Process command line cannot be empty string.")
	}
}

func TestGetAllProcessInfo(t *testing.T) {
	doLog := true
	pids := GetProcessPids()
	for i := 0; i < len(pids); i++ {
		pid := pids[i]
		if doLog {
			t.Log("PID", pid)
		}
		if pid == 0 {
			// This is the idle process. No operations can be performed
			// on it.
			continue
		}
		path, patherr := GetProcessPath(pid)
		if patherr != nil {
			t.Fatalf("GetProcessPath for PID %d returned error: %s", pid, patherr)
		}
		if doLog {
			t.Log("  Path:", path)
		}
		commandLine, cmderr := GetProcessCommandLine(pid)
		if cmderr != nil {
			// Not an error. Expected for some processes.
			if doLog {
				t.Log("  Unable to read command line for PID", pid, " error: ", cmderr)
			}
		} else {
			if doLog {
				t.Log("  Command line:", commandLine)
			}
		}
		memoryUsage, memerr := GetProcessMemoryUsage(pid)
		if memerr != nil {
			t.Fatalf("GetProcessMemoryUsage for PID %d returned error: %s", pid, memerr)
		}
		if doLog {
			t.Log("  Memory usage:", memoryUsage, "B (", memoryUsage/1024/1024, "MB )")
		}
	}
}

func TestInvalidPids(t *testing.T) {
	_, err := GetProcessMemoryUsage(123456)
	if err == nil {
		t.Fatal("Expected error when providing invalid PID in GetProcessMemoryUsage")
	}
	_, err = GetProcessPath(123456)
	if err == nil {
		t.Fatal("Expected error when providing invalid PID in GetProcessPath")
	}
	_, err = GetProcessCommandLine(123456)
	if err == nil {
		t.Fatal("Expected error when providing invalid PID in GetProcessCommandLine")
	}
}

func TestInterface(t *testing.T) {
	var prociInterface Interface
	prociInterface = Proci{}

	doLog := true
	pids := prociInterface.GetProcessPids()
	for i := 0; i < len(pids); i++ {
		pid := pids[i]
		if doLog {
			t.Log("PID", pid)
		}
		if pid == 0 {
			// This is the idle process. No operations can be performed
			// on it.
			continue
		}
		path, patherr := prociInterface.GetProcessPath(pid)
		if patherr != nil {
			t.Fatalf("GetProcessPath for PID %d returned error: %s", pid, patherr)
		}
		if doLog {
			t.Log("  Path:", path)
		}
		commandLine, cmderr := prociInterface.GetProcessCommandLine(pid)
		if cmderr != nil {
			// Not an error. Expected for some processes.
			if doLog {
				t.Log("  Unable to read command line for PID", pid, " error: ", cmderr)
			}
		} else {
			if doLog {
				t.Log("  Command line:", commandLine)
			}
		}
		memoryUsage, memerr := prociInterface.GetProcessMemoryUsage(pid)
		if memerr != nil {
			t.Fatalf("GetProcessMemoryUsage for PID %d returned error: %s", pid, memerr)
		}
		if doLog {
			t.Log("  Memory usage:", memoryUsage, "B (", memoryUsage/1024/1024, "MB )")
		}
	}
	
	mStat, err := prociInterface.GetMemoryStatus()
	if err != nil {
		t.Errorf("GetMemoryStatus returned error: %s", err)
	}
	t.Log("MemoryLoad:", mStat.MemoryLoad, "%")
	if mStat.MemoryLoad == 0 {
		t.Errorf("Memory load cannot be 0")
	}
	t.Log("TotalPhys:", mStat.TotalPhys, "B (", mStat.TotalPhys/1024/1024, "MB )")
	if mStat.TotalPhys == 0 {
		t.Errorf("Total memory cannot be 0")
	}
	t.Log("AvailPhys:", mStat.AvailPhys, "B (", mStat.AvailPhys/1024/1024, "MB )")
	if mStat.AvailPhys == 0 {
		t.Errorf("Available memory cannot be 0")
	}
}
