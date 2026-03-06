package main

import (
	"os"
	"testing"
)

func TestMainHelp(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"edge-tts", "--list-voices"}
	err := run()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	os.Args = []string{"edge-tts", "-t", "Test output"}
	err = run()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	tempFile, _ := os.CreateTemp("", "test")
	tempFile.WriteString("file content")
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	os.Args = []string{"edge-tts", "-f", tempFile.Name(), "--write-media", "-", "--write-subtitles", "-"}
	err = run()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
}
