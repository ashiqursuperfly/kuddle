package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

// generateColor takes a string (e.g., pod or container name) and returns an ANSI escape code for a unique color.
func generateColor(name string) string {
	hasher := fnv.New32()
	_, _ = hasher.Write([]byte(name))
	hash := hasher.Sum32()

	// Generate a unique RGB color from the hash
	red := 30 + int((hash>>16)&0x7F)%200
	green := 30 + int((hash>>8)&0x7F)%200
	blue := 30 + int(hash&0x7F)%200

	// Return an ANSI escape code for the color
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", red, green, blue)
}

var mutex sync.Mutex

func WriteKubectlLogs(color string, args []string, line string) {
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Printf("%s%s\n%s%s\n", color, args, line, "\033[0m")
}
