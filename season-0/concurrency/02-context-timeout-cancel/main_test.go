package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestRunWritesAllScenarios(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var output bytes.Buffer

		if err := run(nil, &output); err != nil {
			t.Fatalf("run returned an error: %v", err)
		}

		records, err := csv.NewReader(&output).ReadAll()
		if err != nil {
			t.Fatalf("read CSV: %v", err)
		}
		if len(records) != len(allScenarios)+1 {
			t.Fatalf("got %d CSV records, want %d", len(records), len(allScenarios)+1)
		}
		if got := records[0]; strings.Join(got, ",") != "scenario,total_ms,completed_steps,canceled_steps,error" {
			t.Fatalf("unexpected header: %v", got)
		}

		for i, scenario := range allScenarios {
			if records[i+1][0] != scenario {
				t.Fatalf("record %d scenario = %q, want %q", i+1, records[i+1][0], scenario)
			}
			if _, err := strconv.ParseFloat(records[i+1][1], 64); err != nil {
				t.Fatalf("total_ms is not numeric: %q", records[i+1][1])
			}
		}
	})
}

func TestRunWritesSingleScenario(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var output bytes.Buffer

		if err := run([]string{"-scenario", scenarioTimeout}, &output); err != nil {
			t.Fatalf("run returned an error: %v", err)
		}

		records, err := csv.NewReader(&output).ReadAll()
		if err != nil {
			t.Fatalf("read CSV: %v", err)
		}
		if len(records) != 2 {
			t.Fatalf("got %d CSV records, want 2", len(records))
		}
		if records[1][0] != scenarioTimeout {
			t.Fatalf("scenario = %q, want %q", records[1][0], scenarioTimeout)
		}
	})
}

func TestRunRejectsInvalidConfiguration(t *testing.T) {
	testCases := []struct {
		name string
		args []string
		want string
	}{
		{name: "unknown scenario", args: []string{"-scenario", "missing"}, want: "scenario must be one of"},
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

func TestScenarioResults(t *testing.T) {
	testCases := []struct {
		scenario       string
		total          time.Duration
		completedSteps int
		canceledSteps  int
		err            error
	}{
		{scenario: scenarioSuccess, total: 360 * time.Millisecond, completedSteps: 3, canceledSteps: 0},
		{scenario: scenarioTimeout, total: 150 * time.Millisecond, completedSteps: 1, canceledSteps: 1, err: context.DeadlineExceeded},
		{scenario: scenarioManualCancel, total: 90 * time.Millisecond, completedSteps: 0, canceledSteps: 1, err: context.Canceled},
		{scenario: scenarioParentCancel, total: 90 * time.Millisecond, completedSteps: 0, canceledSteps: 1, err: context.Canceled},
		{scenario: scenarioIgnoreContext, total: 360 * time.Millisecond, completedSteps: 3, canceledSteps: 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				got, err := runScenario(testCase.scenario)
				if err != nil {
					t.Fatalf("runScenario returned an error: %v", err)
				}

				if got.Total != testCase.total {
					t.Errorf("total = %v, want %v", got.Total, testCase.total)
				}
				if got.CompletedSteps != testCase.completedSteps {
					t.Errorf("completed steps = %d, want %d", got.CompletedSteps, testCase.completedSteps)
				}
				if got.CanceledSteps != testCase.canceledSteps {
					t.Errorf("canceled steps = %d, want %d", got.CanceledSteps, testCase.canceledSteps)
				}
				if got.Err != testCase.err {
					t.Errorf("error = %v, want %v", got.Err, testCase.err)
				}
			})
		})
	}
}
