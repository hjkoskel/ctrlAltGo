/*
Check minor and major IDs and name of partitions
*/
package initializing

import (
	"bufio"
	"fmt"
	"os"
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
