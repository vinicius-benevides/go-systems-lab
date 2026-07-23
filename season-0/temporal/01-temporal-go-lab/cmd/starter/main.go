package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/model"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/shared"
	"go.temporal.io/sdk/client"
)

func main() {
	name := flag.String("name", "Vinícius", "name used in the greeting")
	language := flag.String("language", "pt-BR", "greeting language: pt-BR or en")
	delaySeconds := flag.Int("delay", 2, "simulated Activity duration in seconds")
	waitTimeout := flag.Duration("wait-timeout", 0, "maximum time to wait for the result; 0 waits indefinitely")
	simulateFailure := flag.Bool("simulate-failure", false, "fail the first Activity attempt to demonstrate retry")
	workflowID := flag.String("workflow-id", "", "custom Workflow ID. generated automatically when empty")
	flag.Parse()

	if strings.TrimSpace(*workflowID) == "" {
		*workflowID = fmt.Sprintf("greeting-%d", time.Now().UnixNano())
	}

	config := shared.LoadTemporalConfig()
	temporalClient, err := client.Dial(client.Options{
		HostPort:  config.HostPort,
		Namespace: config.Namespace,
	})
	if err != nil {
		log.Fatalf("unable to connect to Temporal Service: %v", err)
	}
	defer temporalClient.Close()

	input := model.GreetingInput{
		Name:            *name,
		Language:        *language,
		DelaySeconds:    *delaySeconds,
		SimulateFailure: *simulateFailure,
	}

	options := client.StartWorkflowOptions{
		ID:                       *workflowID,
		TaskQueue:                config.TaskQueue,
		WorkflowExecutionTimeout: 5 * time.Minute,
	}

	run, err := temporalClient.ExecuteWorkflow(
		context.Background(),
		options,
		shared.GreetingWorkflowName,
		input,
	)
	if err != nil {
		log.Fatalf("unable to start Workflow: %v", err)
	}

	log.Printf("Workflow started: WorkflowID=%s RunID=%s", run.GetID(), run.GetRunID())

	waitContext := context.Background()
	if *waitTimeout > 0 {
		var cancel context.CancelFunc
		waitContext, cancel = context.WithTimeout(waitContext, *waitTimeout)
		defer cancel()
	}

	var result model.GreetingResult
	if err := run.Get(waitContext, &result); err != nil {
		if waitContext.Err() == context.DeadlineExceeded {
			log.Printf("Workflow is still running; it was not cancelled. Query %q with ID %s for its status.", shared.GreetingStatusQuery, run.GetID())
			return
		}
		log.Fatalf("Workflow execution failed: %v", err)
	}

	fmt.Println("Workflow completed")
	fmt.Printf("Message:      %s\n", result.Message)
	fmt.Printf("Generated at: %s\n", result.GeneratedAt)
	fmt.Printf("Activity try: %d\n", result.Attempt)
	fmt.Printf("Workflow ID:  %s\n", run.GetID())
	fmt.Printf("Run ID:       %s\n", run.GetRunID())
}
