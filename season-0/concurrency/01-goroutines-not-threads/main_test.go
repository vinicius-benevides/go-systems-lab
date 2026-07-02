package main

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestRunWritesBothMeasurements(t *testing.T) {
	var output bytes.Buffer

	if err := run([]string{"-tasks", "3", "-latency", "0"}, &output); err != nil {
		t.Fatalf("run returned an error: %v", err)
	}

	records, err := csv.NewReader(&output).ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d CSV records, want 3", len(records))
	}
	if records[1][2] != modeSequential || records[2][2] != modeGoroutines {
		t.Fatalf("unexpected modes: %q and %q", records[1][2], records[2][2])
	}
	if records[1][0] != "3" || records[2][0] != "3" {
		t.Fatalf("unexpected task counts: %q and %q", records[1][0], records[2][0])
	}
	if _, err := strconv.ParseFloat(records[2][3], 64); err != nil {
		t.Fatalf("elapsed_ms is not numeric: %q", records[2][3])
	}
}

func TestRunRejectsInvalidConfiguration(t *testing.T) {
	testCases := []struct {
		name string
		args []string
		want string
	}{
		{name: "zero tasks", args: []string{"-tasks", "0"}, want: "tasks must be greater than zero"},
		{name: "negative latency", args: []string{"-latency", "-1ms"}, want: "latency must not be negative"},
		{name: "positional argument", args: []string{"extra"}, want: "unexpected positional arguments"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := run(testCase.args, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("run error = %v, want error containing %q", err, testCase.want)
			}
		})
	}
}

func TestExecutionModels(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sequential := runSequential(3, 10*time.Millisecond)
		concurrent := runConcurrent(3, 10*time.Millisecond)

		if sequential.Elapsed != 30*time.Millisecond {
			t.Errorf("sequential elapsed = %v, want 30ms", sequential.Elapsed)
		}
		if concurrent.Elapsed != 10*time.Millisecond {
			t.Errorf("concurrent elapsed = %v, want 10ms", concurrent.Elapsed)
		}
	})
}

func BenchmarkSequential(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		runSequential(20, 2*time.Millisecond)
	}
}

func BenchmarkConcurrent(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		runConcurrent(20, 2*time.Millisecond)
	}
}
