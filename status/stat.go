/*
parsing /proc/stat
https://docs.kernel.org/filesystems/proc.html
*/
package status

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hjkoskel/timegopher"
)

type CPUStats struct {
	User      int64
	Nice      int64
	System    int64
	Idle      int64
	IOWait    int64
	IRQ       int64
	SoftIRQ   int64
	Steal     int64
	Guest     int64
	GuestNice int64
}

func (p *CPUStats) TotalCpuTime() int64 {
	return p.User + p.Nice + p.System + p.Idle + p.IOWait + p.IRQ + p.SoftIRQ + p.Steal + p.System
}

func (p *CPUStats) CpuPercent() float64 {
	return 100 * (1 - float64(p.Idle)/float64(p.TotalCpuTime()))
}

type ProcStat struct { //TODO later time series from array of ProcStats
	CPU             CPUStats //cpu with no number
	CPUs            []CPUStats
	Interrupts      int64 //Interrupts since boot
	InterruptsByInt []int64
	Ctxt            int64 //context switches
	BootEpoch       int64 //TODO time package?
	Processes       int64
	ProcsRunning    int64
	ProcsBlocked    int64

	SoftIRQTotal int64
	SoftIRQs     []int64

	//Store stat time for creating time series for plot etc..
	Timestamp time.Time
	Uptime    int64
}

func GetProcStat() (ProcStat, error) {
	byt, errRead := os.ReadFile("/proc/stat")
	if errRead != nil {
		return ProcStat{}, errRead
	}
	return ParseProcStat(byt)
}

/*
The “intr” line gives counts of interrupts serviced since boot time, for each of the possible system interrupts. The first column is the total of all interrupts serviced including unnumbered architecture specific interrupts; each subsequent column is the total for that particular numbered interrupt. Unnumbered interrupts are not shown, only summed into the total.

The “ctxt” line gives the total number of context switches across all CPUs.

The “btime” line gives the time at which the system booted, in seconds since the Unix epoch.

The “processes” line gives the number of processes and threads created, which includes (but is not limited to) those created by calls to the fork() and clone() system calls.

The “procs_running” line gives the total number of threads that are running or ready to run (i.e., the total number of runnable threads).

The “procs_blocked” line gives the number of processes currently blocked, waiting for I/O to complete.

The “softirq” line gives counts of softirqs serviced since boot time, for each of the possible system softirqs. The first column is the total of all softirqs serviced; each subsequent column is the total for that particular softirq.
*/
func ParseProcStat(procStatContent []byte) (ProcStat, error) {
	lines := strings.Split(string(procStatContent), "\n")

	result := ProcStat{}

	cpumap := make(map[string]CPUStats)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Parse CPU lines
		if strings.HasPrefix(fields[0], "cpu") {
			if len(fields) < 11 {
				continue
			}

			var stats CPUStats
			fmt.Sscanf(
				strings.Join(fields[1:], " "),
				"%d %d %d %d %d %d %d %d %d %d",
				&stats.User, &stats.Nice, &stats.System, &stats.Idle, &stats.IOWait,
				&stats.IRQ, &stats.SoftIRQ, &stats.Steal, &stats.Guest, &stats.GuestNice,
			)
			cpumap[fields[0]] = stats
		}

		firstNumber, _ := strconv.ParseInt(fields[1], 10, 64)
		arrOtherNumbers := make([]int64, len(fields)-2)
		for i, s := range fields {
			if i < 2 {
				continue
			}
			arrOtherNumbers[i-2], _ = strconv.ParseInt(s, 10, 64)
		}

		switch fields[0] {
		case "intr":
			result.Interrupts = firstNumber
			result.InterruptsByInt = arrOtherNumbers
		case "ctxt":
			result.Ctxt = firstNumber
		case "btime":
			result.BootEpoch = firstNumber
		case "processes":
			result.Processes = firstNumber
		case "procs_running":
			result.ProcsRunning = firstNumber
		case "procs_blocked":
			result.ProcsBlocked = firstNumber
		case "softirq":
			result.SoftIRQTotal = firstNumber
			result.SoftIRQs = arrOtherNumbers
		}
	}

	ut, errUp := timegopher.GetDirectUptime()
	result.Uptime = int64(ut)
	result.Timestamp = time.Now()

	result.CPUs = make([]CPUStats, len(cpumap)-1)

	for cpuIndex := range result.CPUs {
		name := fmt.Sprintf("cpu%v", cpuIndex)
		cpu, haz := cpumap[name]
		if !haz {
			return result, fmt.Errorf("missing cpu %s", name)
		}
		result.CPUs[cpuIndex] = cpu
	}

	var hazCpu bool
	result.CPU, hazCpu = cpumap["cpu"]
	if !hazCpu {
		return result, fmt.Errorf("missing key cpu")
	}

	return result, errUp

}
