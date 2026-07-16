package workflows

import (
	"fmt"
	"strings"
	"time"

	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/activities"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/model"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreetingWorkflow orchestrates the durable business process.
//
// Workflow code must be deterministic because Temporal can replay its Event
// History to rebuild the Workflow state after a Worker restart or deployment.
func GreetingWorkflow(ctx workflow.Context, input model.GreetingInput) (model.GreetingResult, error) {
	logger := workflow.GetLogger(ctx)

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return model.GreetingResult{}, fmt.Errorf("name is required")
	}
	if input.DelaySeconds < 0 {
		return model.GreetingResult{}, fmt.Errorf("delay seconds cannot be negative")
	}

	if input.Language == "" {
		input.Language = "pt-BR"
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger.Info(
		"GreetingWorkflow started",
		"name", input.Name,
		"language", input.Language,
		"simulateFailure", input.SimulateFailure,
	)

	var result model.GreetingResult
	future := workflow.ExecuteActivity(ctx, activities.ComposeGreeting, input)
	if err := future.Get(ctx, &result); err != nil {
		return model.GreetingResult{}, fmt.Errorf("compose greeting Activity failed: %w", err)
	}

	logger.Info(
		"GreetingWorkflow completed",
		"message", result.Message,
		"activityAttempt", result.Attempt,
	)

	return result, nil
}
