package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const modeSequential = "sequential"
const modeGoroutines = "goroutines"

type result struct {
	Tasks              int
	Latency            time.Duration
	Mode               string
	Elapsed            time.Duration
	GoroutinesObserved int
}

func simulateIO(latency time.Duration) {
	time.Sleep(latency)
}

func runSequential(tasks int, latency time.Duration) result {
	start := time.Now()
	observed := runtime.NumGoroutine()

	for range tasks {
		simulateIO(latency)
	}

	return result{
		Tasks:              tasks,
		Latency:            latency,
		Mode:               modeSequential,
		Elapsed:            time.Since(start),
		GoroutinesObserved: observed,
	}
}

func runConcurrent(tasks int, latency time.Duration) result {
	var ready sync.WaitGroup
	var finished sync.WaitGroup
	startTasks := make(chan struct{})

	ready.Add(tasks)

	start := time.Now()
	for range tasks {
		finished.Go(func() {
			ready.Done()
			<-startTasks
			simulateIO(latency)
		})
	}

	ready.Wait()
	observed := runtime.NumGoroutine()
	close(startTasks)
	finished.Wait()

	return result{
		Tasks:              tasks,
		Latency:            latency,
		Mode:               modeGoroutines,
		Elapsed:            time.Since(start),
		GoroutinesObserved: observed,
	}
}

func validate(tasks int, latency time.Duration) error {
	if tasks < 1 {
		return errors.New("tasks must be greater than zero")
	}
	if latency < 0 {
		return errors.New("latency must not be negative")
	}
	return nil
}

func writeCSV(output io.Writer, results ...result) error {
	writer := csv.NewWriter(output)
	if err := writer.Write([]string{"tasks", "latency_ms", "mode", "elapsed_ms", "goroutines_observed"}); err != nil {
		return err
	}

	for _, measurement := range results {
		record := []string{
			strconv.Itoa(measurement.Tasks),
			strconv.FormatFloat(float64(measurement.Latency)/float64(time.Millisecond), 'f', 3, 64),
			measurement.Mode,
			strconv.FormatFloat(float64(measurement.Elapsed)/float64(time.Millisecond), 'f', 3, 64),
			strconv.Itoa(measurement.GoroutinesObserved),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func run(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("goroutines-not-threads", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	tasks := flags.Int("tasks", 20, "number of simulated I/O tasks")
	latency := flags.Duration("latency", 100*time.Millisecond, "simulated latency per task")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if err := validate(*tasks, *latency); err != nil {
		return err
	}

	return writeCSV(
		output,
		runSequential(*tasks, *latency),
		runConcurrent(*tasks, *latency),
	)
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
