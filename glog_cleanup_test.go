package glog

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestCleanup(t *testing.T) {
	logFile, logFileName, err := create("INFO", time.Now())
	if err != nil {
		t.Fatalf("Could not create logfile: %v", err)
	}
	fmt.Fprintf(logFile, "some bytes\n")
	logFile.Close()

	// Get file size
	fi, err := os.Stat(logFileName)
	if err != nil {
		t.Fatalf("Could not stat logfile: %v", err)
	}

	flag.Set("log_total_bytes", strconv.FormatInt(fi.Size(), 10))

	cleanup()

	// Verify that the file was not cleaned up yet.
	if _, err := os.Stat(logFileName); err != nil {
		t.Fatalf("got %v, want %v", err, nil)
	}

	// Write just two more bytes to a new logfile. Increase time by one second
	// to make sure we create a new logfile.
	newLogFile, _, err := create("INFO", time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatalf("Could not create logfile: %v", err)
	}
	fmt.Fprintf(newLogFile, "x\n")
	newLogFile.Close()

	cleanup()

	// Verify that the old file is now gone.
	if _, err := os.Stat(logFileName); !os.IsNotExist(err) {
		t.Fatalf("%q: got %v, want %v", logFileName, err, os.ErrNotExist)
	}
}

// When running on docker, hostnames are auto-generated. Ensure that sorting
// the logfiles to see which one is the oldest ignores the hostname, otherwise
// we might delete the logfile into which we are writing.
func TestCleanupMultipleHostnames(t *testing.T) {
	oldLogPath := filepath.Join(os.TempDir(), "glog.test.db7860cc55a8."+userName+".log.ERROR.20160219-170516.1")
	newLogPath := filepath.Join(os.TempDir(), "glog.test.5b867334831d."+userName+".log.INFO.20160602-074008.1")
	if err := ioutil.WriteFile(oldLogPath, []byte("123456"), 0644); err != nil {
		t.Fatalf("Could not create logfile: %v", err)
	}
	if err := ioutil.WriteFile(newLogPath, []byte("1234"), 0644); err != nil {
		t.Fatalf("Could not create logfile: %v", err)
	}

	flag.Set("log_total_bytes", "10")

	cleanup()

	// Verify that the file was not cleaned up yet.
	if _, err := os.Stat(newLogPath); err != nil {
		t.Fatalf("got %v, want %v", err, nil)
	}
}
