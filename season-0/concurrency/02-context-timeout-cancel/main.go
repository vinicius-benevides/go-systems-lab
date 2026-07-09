package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	scenarioAll           = "all"
	scenarioSuccess       = "success"
	scenarioTimeout       = "timeout"
	scenarioManualCancel  = "manual-cancel"
	scenarioParentCancel  = "parent-cancel"
	scenarioIgnoreContext = "ignore-context"
)

const (
	checkFraudDuration       = 120 * time.Millisecond
	reserveInventoryDuration = 180 * time.Millisecond
	persistPaymentDuration   = 60 * time.Millisecond
)

var allScenarios = []string{
	scenarioSuccess,
	scenarioTimeout,
	scenarioManualCancel,
	scenarioParentCancel,
	scenarioIgnoreContext,
}

type metrics struct {
	completedSteps int
	canceledSteps  int
}

type result struct {
	Scenario       string
	Total          time.Duration
	CompletedSteps int
	CanceledSteps  int
	Err            error
}

func (m *metrics) complete() {
	m.completedSteps++
}

func (m *metrics) cancel() {
	m.canceledSteps++
}

func processOrder(ctx context.Context, measurements *metrics) error {
	if err := checkFraud(ctx, measurements); err != nil {
		return err
	}
	if err := reserveInventory(ctx, measurements); err != nil {
		return err
	}
	if err := persistPayment(ctx, measurements); err != nil {
		return err
	}
	return nil
}

func processOrderIgnoringContext(ctx context.Context, measurements *metrics) error {
	if err := checkFraudIgnoringContext(ctx, measurements); err != nil {
		return err
	}
	if err := reserveInventoryIgnoringContext(ctx, measurements); err != nil {
		return err
	}
	if err := persistPaymentIgnoringContext(ctx, measurements); err != nil {
		return err
	}
	return nil
}

func checkFraud(ctx context.Context, measurements *metrics) error {
	return doWork(ctx, checkFraudDuration, measurements)
}

func reserveInventory(ctx context.Context, measurements *metrics) error {
	return doWork(ctx, reserveInventoryDuration, measurements)
}

func persistPayment(ctx context.Context, measurements *metrics) error {
	return doWork(ctx, persistPaymentDuration, measurements)
}

func checkFraudIgnoringContext(ctx context.Context, measurements *metrics) error {
	return doWorkIgnoringContext(ctx, checkFraudDuration, measurements)
}

func reserveInventoryIgnoringContext(ctx context.Context, measurements *metrics) error {
	return doWorkIgnoringContext(ctx, reserveInventoryDuration, measurements)
}

func persistPaymentIgnoringContext(ctx context.Context, measurements *metrics) error {
	return doWorkIgnoringContext(ctx, persistPaymentDuration, measurements)
}

func doWork(ctx context.Context, duration time.Duration, measurements *metrics) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		measurements.complete()
		return nil
	case <-ctx.Done():
		measurements.cancel()
		return ctx.Err()
	}
}

func doWorkIgnoringContext(ctx context.Context, duration time.Duration, measurements *metrics) error {
	_ = ctx
	time.Sleep(duration)
	measurements.complete()
	return nil
}

func runScenario(scenario string) (result, error) {
	start := time.Now()
	measurements := &metrics{}
	err := executeScenario(scenario, measurements)

	return result{
		Scenario:       scenario,
		Total:          time.Since(start),
		CompletedSteps: measurements.completedSteps,
		CanceledSteps:  measurements.canceledSteps,
		Err:            err,
	}, nil
}

func executeScenario(scenario string, measurements *metrics) error {
	switch scenario {
	case scenarioSuccess:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return processOrder(ctx, measurements)
	case scenarioTimeout:
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()
		return processOrder(ctx, measurements)
	case scenarioManualCancel:
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			time.Sleep(90 * time.Millisecond)
			cancel()
		}()

		return processOrder(ctx, measurements)
	case scenarioParentCancel:
		parentCtx, cancelParent := context.WithCancel(context.Background())
		defer cancelParent()

		childCtx, cancelChild := context.WithCancel(parentCtx)
		defer cancelChild()

		go func() {
			time.Sleep(90 * time.Millisecond)
			cancelParent()
		}()

		return processOrder(childCtx, measurements)
	case scenarioIgnoreContext:
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			time.Sleep(90 * time.Millisecond)
			cancel()
		}()

		return processOrderIgnoringContext(ctx, measurements)
	default:
		return fmt.Errorf("unknown scenario %q", scenario)
	}
}

func validateScenario(scenario string) error {
	if scenario == scenarioAll {
		return nil
	}
	for _, available := range allScenarios {
		if scenario == available {
			return nil
		}
	}
	return fmt.Errorf("scenario must be one of: %v", append([]string{scenarioAll}, allScenarios...))
}

func writeCSV(output io.Writer, results ...result) error {
	writer := csv.NewWriter(output)
	if err := writer.Write([]string{"scenario", "total_ms", "completed_steps", "canceled_steps", "error"}); err != nil {
		return err
	}

	for _, measurement := range results {
		record := []string{
			measurement.Scenario,
			strconv.FormatFloat(float64(measurement.Total)/float64(time.Millisecond), 'f', 3, 64),
			strconv.Itoa(measurement.CompletedSteps),
			strconv.Itoa(measurement.CanceledSteps),
			errorText(measurement.Err),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func errorText(err error) string {
	if err == nil {
		return "nil"
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled.Error()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded.Error()
	}
	return err.Error()
}

func run(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("context-timeout-cancel", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	scenario := flags.String("scenario", scenarioAll, "scenario to run: all, success, timeout, manual-cancel, parent-cancel, ignore-context")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if err := validateScenario(*scenario); err != nil {
		return err
	}

	if *scenario != scenarioAll {
		measurement, err := runScenario(*scenario)
		if err != nil {
			return err
		}
		return writeCSV(output, measurement)
	}

	results := make([]result, 0, len(allScenarios))
	for _, name := range allScenarios {
		measurement, err := runScenario(name)
		if err != nil {
			return err
		}
		results = append(results, measurement)
	}
	return writeCSV(output, results...)
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
