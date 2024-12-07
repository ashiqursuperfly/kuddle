package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

func generateANSIEscapeColorCode(name string) string {
	hasher := fnv.New32()
	_, _ = hasher.Write([]byte(name))
	hash := hasher.Sum32()

	red := 30 + int((hash>>16)&0x7F)%200
	green := 30 + int((hash>>8)&0x7F)%200
	blue := 30 + int(hash&0x7F)%200

	// Return an ANSI escape code for the color
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", red, green, blue)
}

var mutex sync.Mutex

func WriteKubectlLogs(group string, line string) {
	mutex.Lock()
	defer mutex.Unlock()

	color := generateANSIEscapeColorCode(group)

	fmt.Printf("%s%s ]- %s%s\n", color, group, line, "\033[0m")
}
