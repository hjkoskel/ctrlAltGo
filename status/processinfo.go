package status

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*






W  Paging (only before Linux 2.6.0)

I  Idle (Linux 4.14 onward)
*/

type ProcessState int

const (
	ProcessStateRunning     ProcessState = iota //R  Running
	ProcessStateSleeping                        //S  Sleeping in an interruptible wait
	ProcessStateStopped                         //T  Stopped (on a signal) or (before Linux 2.6.33) trace stopped
	ProcessStateTracingStop                     //t  Tracing stop (Linux 2.6.33 onward)
	ProcessStateZombie                          //Z  Zombie
	ProcessStateDiskSleep                       //D  Waiting in uninterruptible disk sleep
	ProcessStateDead                            //X  Dead (from Linux 2.6.0 onward)
	ProcessStateIdle                            //I Idle
	ProcessStateUnknown
)

func (a ProcessState) String() string {
	result, haz := map[ProcessState]string{
		ProcessStateRunning:     "running",
		ProcessStateSleeping:    "sleeping",
		ProcessStateStopped:     "stopped",
		ProcessStateTracingStop: "tracingStop",
		ProcessStateZombie:      "zombie",
		ProcessStateDiskSleep:   "diskSleep",
		ProcessStateDead:        "dead",
		ProcessStateIdle:        "idle",
		ProcessStateUnknown:     "unknown",
	}[a]
	if !haz {
		return "unknown"
	}
	return result
}

func ParseProcessState(s string) ProcessState {
	result, haz := map[string]ProcessState{
		"R": ProcessStateRunning,
		"S": ProcessStateSleeping,
		"T": ProcessStateStopped,
		"t": ProcessStateTracingStop,
		"Z": ProcessStateZombie,
		"X": ProcessStateDead,
		"I": ProcessStateIdle,
	}[strings.TrimSpace(s)]
	if !haz {
		fmt.Printf("UNKNOW %s\n", s)
		return ProcessStateUnknown
	}
	return result
}

type ProcessInfo struct {
	PID     int
	Command string
	Status  ProcessState
	VmSize  uint64  // Memory usage in kB
	CPU     float64 // CPU usage percentage
}
type ProcessCPUStats struct {
	UTime uint64 // User time in clock ticks
	STime uint64 // System time in clock ticks
}

type ProcessInfos []ProcessInfo

func (a ProcessInfos) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-8s %-20s %-10s %-10s %s\n", "PID", "COMMAND", "MEMORY (kB)", "CPU (%)", "STATE"))
	for _, p := range a {
		sb.WriteString(fmt.Sprintf("%-8d %-20s %-10d %-10.2f %s\n", p.PID, p.Command, p.VmSize, p.CPU, p.Status))
	}
	return sb.String()
}

func ReadProcessInfos(procDirName string) (ProcessInfos, error) { //parameter for unit testing
	// Open the /proc directory
	files, errRead := os.ReadDir(procDirName)
	if errRead != nil {
		return nil, fmt.Errorf("error reading %s dir: %s\n", procDirName, errRead)
	}

	// First pass: Collect initial CPU stats
	initialStats := make(map[int]ProcessCPUStats)
	for _, file := range files {
		if file.IsDir() {
			pid, err := strconv.Atoi(file.Name())
			if err != nil {
				continue // Skip non-PID directories
			}

			stats, err := getProcessCPUStats(pid)
			if err != nil {
				continue // Skip if we can't read the stats
			}

			initialStats[pid] = stats
		}
	}

	// Wait for a short interval to calculate CPU usage
	time.Sleep(1 * time.Second) //TODO separate function and option to get averaging on arbitatry duration?

	var processes []ProcessInfo

	// Iterate over all entries in /proc
	for _, file := range files {
		// Check if the entry is a directory and its name is a number (PID)
		if file.IsDir() {
			pid, err := strconv.Atoi(file.Name())
			if err != nil {
				continue // Skip non-PID directories
			}

			// Read process information
			process, err := getProcessInfo(pid)
			if err != nil {
				continue // Skip if we can't read the process info
			}

			// Get updated CPU stats
			updatedStats, err := getProcessCPUStats(pid)
			if err != nil {
				continue // Skip if we can't read the updated stats
			}

			// Calculate CPU usage percentage
			initial := initialStats[pid]
			totalTime := (updatedStats.UTime + updatedStats.STime) - (initial.UTime + initial.STime)
			process.CPU = float64(totalTime) / float64(100) // Convert to percentage

			processes = append(processes, process)
		}
	}
	return processes, nil
}

// getProcessCPUStats reads CPU stats from /proc/[pid]/stat
func getProcessCPUStats(pid int) (ProcessCPUStats, error) {
	procPath := filepath.Join("/proc", strconv.Itoa(pid), "stat")
	data, err := ioutil.ReadFile(procPath)
	if err != nil {
		return ProcessCPUStats{}, err
	}

	// Split the stat file into fields
	fields := strings.Fields(string(data))
	if len(fields) < 15 {
		return ProcessCPUStats{}, fmt.Errorf("invalid stat file format")
	}

	// Extract utime (field 14) and stime (field 15)
	utime, err := strconv.ParseUint(fields[13], 10, 64)
	if err != nil {
		return ProcessCPUStats{}, err
	}

	stime, err := strconv.ParseUint(fields[14], 10, 64)
	if err != nil {
		return ProcessCPUStats{}, err
	}

	return ProcessCPUStats{
		UTime: utime,
		STime: stime,
	}, nil
}

// getProcessInfo reads process information from /proc/[pid]/
func getProcessInfo(pid int) (ProcessInfo, error) {
	procPath := filepath.Join("/proc", strconv.Itoa(pid))

	// Read the command name from /proc/[pid]/comm
	comm, err := os.ReadFile(filepath.Join(procPath, "comm"))
	if err != nil {
		return ProcessInfo{}, err
	}
	command := strings.TrimSpace(string(comm))

	// Read the status from /proc/[pid]/status
	status, err := os.ReadFile(filepath.Join(procPath, "status"))
	if err != nil {
		return ProcessInfo{}, err
	}

	// Parse memory usage (VmSize) from the status file
	var vmSize uint64
	lines := strings.Split(string(status), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VmSize:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				vmSize, _ = strconv.ParseUint(fields[1], 10, 64)
			}
			break
		}
	}

	// Read the process status (e.g., running, sleeping) from /proc/[pid]/status
	var processStatus string
	for _, line := range lines {
		if strings.HasPrefix(line, "State:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				processStatus = fields[1]
			}
			break
		}
	}

	return ProcessInfo{
		PID:     pid,
		Command: command,
		Status:  ParseProcessState(processStatus),
		VmSize:  vmSize,
	}, nil
}
