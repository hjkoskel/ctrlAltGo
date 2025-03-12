/*
Check minor and major IDs and name of partitions
*/
package initializing

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

// Partition represents a single partition entry from /proc/partitions
type ProcPartition struct {
	Major  int    // Major device number
	Minor  int    // Minor device number
	Blocks int    // Number of 1K blocks
	Name   string // Device name
}

type ProcPartitions []ProcPartition

// ParseProcPartitions reads and parses the /proc/partitions file
func ParseProcPartitions() (ProcPartitions, error) {
	file, err := os.Open("/proc/partitions")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/partitions: %v", err)
	}
	defer file.Close()

	var partitions []ProcPartition
	var major, minor, blocks int
	var name string

	// Skip the header lines
	scanner := bufio.NewScanner(file)
	for i := 0; i < 2; i++ {
		scanner.Scan()
	}

	// Parse each line
	for scanner.Scan() {
		line := scanner.Text()
		_, err := fmt.Sscanf(line, "%d %d %d %s", &major, &minor, &blocks, &name)
		if err != nil {
			continue // Skip malformed lines
		}

		partitions = append(partitions, ProcPartition{
			Major:  major,
			Minor:  minor,
			Blocks: blocks,
			Name:   name,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/partitions: %v", err)
	}
	return partitions, nil
}

// GetBootPartitionName   prioritylist:{"mmcblk0p1","sda1"}
func GetBootPartitionName(prioritylist []string, timeoutTime time.Duration) (string, error) {
	tStart := time.Now()
	for time.Since(tStart) < timeoutTime { //USB enumeration takes time
		blockDevices, errBlockDevices := GetBlockDevices()
		if errBlockDevices != nil {
			return "", errBlockDevices
		}

		if len(blockDevices) == 0 {
			fmt.Printf("no block devices yet!")
			time.Sleep(time.Millisecond * 250)
			continue // re-try
		}
		fmt.Printf("DEBUG: haz blck devices\n%s\n\n", blockDevices)

		if blockDevices.IsQemu() {
			return "", nil //TODO get drive by some other priority list?  hda?
		}

		//prioritylist := []string{PARTITIONNAME_PRIMARY, PARTITIONNAME_SECONDARY}
		for _, name := range prioritylist {
			if blockDevices.HazPartition(name) {
				return name, nil
			}
		}
		time.Sleep(time.Second)
	}
	return "", fmt.Errorf("get boot partition timeout after %s", time.Since(tStart))
}
