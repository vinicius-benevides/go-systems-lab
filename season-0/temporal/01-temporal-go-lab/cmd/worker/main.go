package main

import (
	"log"

	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/activities"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/shared"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("unable to connect to Temporal Service: %v", err)
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, shared.TaskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.GreetingWorkflow)
	w.RegisterActivity(activities.ComposeGreeting)

	log.Printf("Worker started and polling Task Queue %q", shared.TaskQueue)
	log.Println("Press Ctrl+C to stop the Worker gracefully")

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("unable to run Worker: %v", err)
	}
}
