// Package proci implements the proci interface for Windows
package proci

import (
	"fmt"
	"syscall"
	"unsafe"
)

//////////////////////////////////////////////////////////////////////////////
// External types

//////////////////////////////////////////////////////////////////////////////
// Internal variables

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")
	ntDll    = syscall.NewLazyDLL("Ntdll.dll")
	advapi32 = syscall.NewLazyDLL("Advapi32.dll")

	globalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	getCurrentProcess    = kernel32.NewProc("GetCurrentProcess")
	openProcess          = kernel32.NewProc("OpenProcess")
	closeHandle          = kernel32.NewProc("CloseHandle")
	getLastError         = kernel32.NewProc("GetLastError")
	readProcessMemory    = kernel32.NewProc("ReadProcessMemory")

	enumProcesses           = psapi.NewProc("EnumProcesses")
	getProcessMemoryInfo    = psapi.NewProc("GetProcessMemoryInfo")
	getProcessImageFileName = psapi.NewProc("GetProcessImageFileNameW")

	ntQueryInformationProcess = ntDll.NewProc("NtQueryInformationProcess")

	openProcessToken      = advapi32.NewProc("OpenProcessToken")
	lookupPrivilegeValue  = advapi32.NewProc("LookupPrivilegeValueW")
	adjustTokenPrivileges = advapi32.NewProc("AdjustTokenPrivileges")
)

//////////////////////////////////////////////////////////////////////////////
// Windows Data Types (internal)
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa383751(v=vs.85).aspx

type winByte uint8
type winUShort uint16
type winLong int32
type winULong uint32
type winDWord uint32
type winDWordLong uint64
type winSizeT uint64
type winPVoid uint64
type winPointer uint64 // Generic for all kinds of pointers

type winUnicodeString struct {
	Length        winUShort
	MaximumLength winUShort
	Buffer        winPointer
}

//////////////////////////////////////////////////////////////////////////////
// Initialization - set priviledges to allow memory write of other processes

type winLUID struct {
	LowPart  winDWord
	HighPart winLong
}

// LUID_AND_ATTRIBUTES
type winLUIDAndAttributes struct {
	Luid       winLUID
	Attributes winDWord
}

// TOKEN_PRIVILEGES
type winTokenPriviledges struct {
	PrivilegeCount winDWord
	Privileges     [1]winLUIDAndAttributes
}

// Give current process priviledges to read in other processes memory
// i.e. call the readProcessMemory function. All errors lead to panic.
func init() {
	handle, _, err := getCurrentProcess.Call()
	if handle == 0 {
		panic(fmt.Sprintf("Unable to open current process. Reason: %s\n", err))
	}

	var token uintptr
	ret2, _, err2 := openProcessToken.Call(
		handle,
		uintptr(0x0028),
		uintptr(unsafe.Pointer(&token)))
	if ret2 == 0 {
		panic(fmt.Sprintf("Unable to open token. Reason: %s\n", err2))
	}

	tokenPriviledges := winTokenPriviledges{PrivilegeCount: 1}
	lpName := syscall.StringToUTF16("SeDebugPrivilege")
	ret3, _, err3 := lookupPrivilegeValue.Call(
		0,
		uintptr(unsafe.Pointer(&lpName[0])),
		uintptr(unsafe.Pointer(&tokenPriviledges.Privileges[0].Luid)))
	if ret3 == 0 {
		panic(fmt.Sprintf("Unable to lookup priviledges. Reason: %s\n", err3))
	}

	tokenPriviledges.Privileges[0].Attributes = 0x00000002 // SE_PRIVILEGE_ENABLED

	ret4, _, err4 := adjustTokenPrivileges.Call(
		token,
		0,
		uintptr(unsafe.Pointer(&tokenPriviledges)),
		uintptr(unsafe.Sizeof(tokenPriviledges)),
		0,
		0)
	if ret4 == 0 {
		panic(fmt.Sprintf("Unable to adjust token priviledges. Reason: %s\n", err4))
	}
}

//////////////////////////////////////////////////////////////////////////////
// Get physical memory status

type winMemoryStatusEx struct {
	DwLength                winDWord
	DwMemoryLoad            winDWord
	UllTotalPhys            winDWordLong
	UllAvailPhys            winDWordLong
	UllTotalPageFile        winDWordLong
	UllAvailPageFile        winDWordLong
	UllTotalVirtual         winDWordLong
	UllAvailVirtual         winDWordLong
	UllAvailExtendedVirtual winDWordLong
}

// getMemoryStatus implements GetMemoryStatus.
func getMemoryStatus() (MemoryStatus, error) {
	mStatEx := new(winMemoryStatusEx)
	mStatEx.DwLength = winDWord(unsafe.Sizeof(*mStatEx))

	ret, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(mStatEx)))
	if ret == 0 {
		return MemoryStatus{}, fmt.Errorf("unable to get physical memory info. Reason: %s", err)
	}

	return MemoryStatus{
		MemoryLoad: uint32(mStatEx.DwMemoryLoad),
		TotalPhys:  uint64(mStatEx.UllTotalPhys),
		AvailPhys:  uint64(mStatEx.UllAvailPhys)}, nil
}

//////////////////////////////////////////////////////////////////////////////
// List processes

// Maximum number of processes
const maxProcesses = 2000

// getProcessPids implements GetProcessPids
func getProcessPids() []uint32 {
	var bytesReturned uint32
	pids := make([]uint32, maxProcesses, maxProcesses)
	ret, _, err := enumProcesses.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(len(pids)*4),
		uintptr(unsafe.Pointer(&bytesReturned)))
	if ret == 0 {
		panic(fmt.Sprintf("Unable to list processes. Reason: %s", err))
	}
	nbrOfPids := int(bytesReturned / 4)
	if nbrOfPids >= len(pids) {
		panic(fmt.Sprintf("More running processes than configured max capacity of %d\n", maxProcesses))
	}
	pids = pids[:nbrOfPids] // Shrik to fit number of PIDS
	return pids
}

//////////////////////////////////////////////////////////////////////////////
// Get process memory utilization

// PROCESS_MEMORY_COUNTERS_EX
type winProcessMemoryCountersEx struct {
	cb                         winDWord
	PageFaultCount             winDWord
	PeakWorkingSetSize         winSizeT
	WorkingSetSize             winSizeT
	QuotaPeakPagedPoolUsage    winSizeT
	QuotaPagedPoolUsage        winSizeT
	QuotaPeakNonPagedPoolUsage winSizeT
	QuotaNonPagedPoolUsage     winSizeT
	PagefileUsage              winSizeT
	PeakPagefileUsage          winSizeT
	PrivateUsage               winSizeT
}

// getProcessMemoryUsage implements GetProcessMemoryUsage.
func getProcessMemoryUsage(pid uint32) (uint64, error) {
	handle, err := openProc(pid, opBasic)
	if err != nil {
		return 0, err
	}
	defer closeProc(handle)

	procMemCntrEx := new(winProcessMemoryCountersEx)
	cb := uintptr(unsafe.Sizeof(*procMemCntrEx))

	ret, _, err2 := getProcessMemoryInfo.Call(handle, uintptr(unsafe.Pointer(procMemCntrEx)), cb)
	if ret == 0 {
		return 0, fmt.Errorf("unable to get physical memory info. Reason: %s", err2)
	}
	return uint64(procMemCntrEx.PrivateUsage), nil
}

//////////////////////////////////////////////////////////////////////////////
// Get process path (which also includes the process name).

// getProcessPath implements GetProcessPath.
func getProcessPath(pid uint32) (string, error) {
	handle, err := openProc(pid, opBasic)
	if err != nil {
		return "", err
	}
	defer closeProc(handle)

	var lpImageFileName = make([]uint16, syscall.MAX_PATH+1)
	var nSize = uintptr(len(lpImageFileName))

	ret, _, _ := getProcessImageFileName.Call(handle, uintptr(unsafe.Pointer(&lpImageFileName[0])), nSize)
	if ret == 0 {
		// Here we don't know if there was no path (for example the Kernel process
		// with PID 4 don't have it) or if there was an error. We assume that the
		// path is empty.
		return "", nil
	}
	return syscall.UTF16ToString(lpImageFileName), nil
}

//////////////////////////////////////////////////////////////////////////////
// Get process command line arguments

type winProcessBasicInformation struct {
	Reserved1       winPVoid
	PebBaseAddress  winPointer
	Reserved2       [2]winPVoid
	UniqueProcessID winPointer
	Reserved3       winPVoid
}

type winPEB struct {
	Reserved1              [2]winByte
	BeingDebugged          winByte
	Reserved2              [1]winByte
	Reserved3              [2]winPVoid
	Ldr                    winPointer // PPEB_LDR_DATA
	ProcessParameters      winPointer // PRTL_USER_PROCESS_PARAMETERS
	Reserved4              [104]winByte
	Reserved5              [52]winPVoid
	PostProcessInitRoutine winPointer // PPS_POST_PROCESS_INIT_ROUTINE
	Reserved6              [128]winByte
	Reserved7              [1]winPVoid
	SessionID              winULong
}

type winRTLUserProcessParameters struct {
	Reserved1     [16]winByte
	Reserved2     [10]winPVoid
	ImagePathName winUnicodeString
	CommandLine   winUnicodeString
}

// getProcessCommandLine implements GetProcessCommandLine.
func getProcessCommandLine(pid uint32) (string, error) {
	handle, err := openProc(pid, opReadVM)
	if err != nil {
		return "", err
	}
	defer closeProc(handle)

	/////////////
	procBasicInf := new(winProcessBasicInformation)
	procBasicInfSize := uintptr(unsafe.Sizeof(*procBasicInf))
	var returnLength uint32
	ret, _, err2 := ntQueryInformationProcess.Call(
		handle,
		0,
		uintptr(unsafe.Pointer(procBasicInf)),
		procBasicInfSize,
		uintptr(unsafe.Pointer(&returnLength)))
	if ret != 0 {
		return "", fmt.Errorf("unable to query process information. Reason: %s", err2)
	}

	/////////////
	peb := new(winPEB)

	err3 := readProcMemory(
		handle,
		uintptr(procBasicInf.PebBaseAddress),
		uintptr(unsafe.Pointer(peb)),
		uintptr(unsafe.Sizeof(*peb)))
	if err3 != nil {
		return "", fmt.Errorf("unable to read PEB structure. Reason: %s", err3)
	}

	/////////////
	userProcessParameters := new(winRTLUserProcessParameters)

	err4 := readProcMemory(
		handle,
		uintptr(peb.ProcessParameters),
		uintptr(unsafe.Pointer(userProcessParameters)),
		uintptr(unsafe.Sizeof(*userProcessParameters)))
	if err4 != nil {
		return "", fmt.Errorf("unable to read PEB process memory. Reason: %s", err4)
	}

	/////////////
	cmdUnicode := userProcessParameters.CommandLine
	commandLineBuffer := make([]uint16, cmdUnicode.Length)

	err5 := readProcMemory(
		handle,
		uintptr(cmdUnicode.Buffer),
		uintptr(unsafe.Pointer(&commandLineBuffer[0])),
		uintptr(cmdUnicode.Length))
	if err5 != nil {
		return "", fmt.Errorf("unable to read command line. Reason: %s", err5)
	}

	return syscall.UTF16ToString(commandLineBuffer), nil
}

//////////////////////////////////////////////////////////////////////////////
// Internal functions

const opReadVM = 0x00000410 // PROCESS_QUERY_INFORMATION | PROCESS_VM_READ
const opBasic = 0x00001000  // PROCESS_QUERY_LIMITED_INFORMATION

// Opens a process and returns the process handle
// Note! Close the process with closeProcess
func openProc(pid uint32, accessLevel uint32) (uintptr, error) {
	handle, _, err := openProcess.Call(
		uintptr(accessLevel),
		0,
		uintptr(pid))
	if handle == 0 {
		return 0, fmt.Errorf("unable to open process %d. Reason: %s", pid, err)
	}
	return handle, nil
}

func closeProc(handle uintptr) error {
	ret, _, err := closeHandle.Call(handle)
	if ret == 0 {
		return fmt.Errorf("unable to close process handle %d. Reason: %s", handle, err)
	}
	return nil
}

// Read the memory of a another process. Requires the process being opened with
// opReadVM AND that this (current) process is running as administrator.
func readProcMemory(handle uintptr, address uintptr, bufferPointer uintptr, bufferSize uintptr) error {
	var numberOfBytesRead uintptr
	ret, _, err := readProcessMemory.Call(
		handle,
		address,
		bufferPointer,
		bufferSize,
		uintptr(unsafe.Pointer(&numberOfBytesRead)))
	if ret == 0 {
		return fmt.Errorf("unable to read memory. Reason: %s", err)
	}
	if bufferSize != numberOfBytesRead {
		return fmt.Errorf("unable to read memory all memory. Reason: %s", err)
	}
	return nil
}
