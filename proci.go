// Package proci includes functionality to get information about running
// processes.
package proci

// MemoryStatus reflects the total physical memory utilization.
type MemoryStatus struct {
	MemoryLoad uint32 // Current memory load in percent 0-100
	TotalPhys  uint64 // Total physical memory in bytes
	AvailPhys  uint64 // Available memory in bytes
}

// GetMemoryStatus gets the physical memory utilization.
func GetMemoryStatus() (MemoryStatus, error) {
	return getMemoryStatus()
}

// GetProcessPids lists all the process identities (PIDS) running in the system.
//
// Returns a slice with PIDS and with the length corresponding to number of PIDS.
//
// Note that PID 0 is reserved for the idle process in Windows which is special
// in that you cannot read it with the other functions in this package.
func GetProcessPids() []uint32 {
	return getProcessPids()
}

// GetProcessMemoryUsage gets the number of bytes used by the specific process.
func GetProcessMemoryUsage(pid uint32) (uint64, error) {
	return getProcessMemoryUsage(pid)
}

// GetProcessPath gets the path of the process (which also includes the
// process name).
func GetProcessPath(pid uint32) (string, error) {
	return getProcessPath(pid)
}

// GetProcessCommandLine reads the process command line. This function
// requires that you are running using as administrator. If you are
// running as a normal user you will get the error "Access Denied".
//
// Also note that some system processes (usually the ones with lowest PIDs)
// will give "Access Denied" even if you are running as administrator.
func GetProcessCommandLine(pid uint32) (string, error) {
	return getProcessCommandLine(pid)
}
