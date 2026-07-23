package main

import (
	"log"

	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/activities"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/shared"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/workflows"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	config := shared.LoadTemporalConfig()
	temporalClient, err := client.Dial(client.Options{
		HostPort:  config.HostPort,
		Namespace: config.Namespace,
	})
	if err != nil {
		log.Fatalf("unable to connect to Temporal Service: %v", err)
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, config.TaskQueue, worker.Options{})
	w.RegisterWorkflowWithOptions(workflows.GreetingWorkflow, workflow.RegisterOptions{Name: shared.GreetingWorkflowName})
	w.RegisterActivityWithOptions(activities.ComposeGreeting, activity.RegisterOptions{Name: shared.ComposeGreetingActivityName})

	log.Printf("Worker started: namespace=%q task-queue=%q", config.Namespace, config.TaskQueue)
	log.Println("Press Ctrl+C to stop the Worker gracefully")

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("unable to run Worker: %v", err)
	}
}
