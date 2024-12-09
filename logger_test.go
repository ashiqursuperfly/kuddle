package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorUniqueness(t *testing.T) {
	tests := []struct {
		name        string
		pod1        string
		pod2        string
		shouldMatch bool
	}{
		{"pods having similar names", "web-1", "web-2", false},
		{"pods having different names", "web-1", "db-1", false},
		{"pods having same name", "web-1", "web-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color1 := generateANSIEscapeColorCode(tt.pod1)
			color2 := generateANSIEscapeColorCode(tt.pod2)

			isEqual := color1 == color2

			if isEqual != tt.shouldMatch {
				t.Errorf("Error mismatch: got %v, expected %v", isEqual, tt.shouldMatch)
			}
		})
	}
}

func TestWriteKubectlLogs(t *testing.T) {
	tests := []struct {
		name string
		pod  string
		line string
	}{
		{"simple log line", "web-1", "Simple log line"},
		{"json log line", "web-1", "{\"timestamp\":\"09/Dec/2024:14:21:57 +0000\",\"clientip\":\"10.10.10.1\",\"request\":\"GET /_status HTTP/1.1\",\"status\": \"200\",\"bytes\":\"62\",\"response_time\":\"0.001\",\"referrer\":\"\",\"agent\":\"kube-probe/1.30\"}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()

			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}

			stdout := os.Stdout

			defer func() { os.Stdout = stdout }() // return stdout after the test
			os.Stdout = w

			expectedColor := generateANSIEscapeColorCode(tt.pod)
			expectedOutput := fmt.Sprintf("%s%s ]- %s%s\n", expectedColor, tt.pod, tt.line, "\033[0m")

			WriteKubectlLogs(tt.pod, tt.line)

			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close writer: %v", err)
			}

			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				t.Fatalf("Failed to read from pipe: %v", err)
			}

			assert.Equal(t, expectedOutput, buf.String())
		})
	}

}
