package workflows

import (
	"fmt"
	"strings"
	"time"

	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/model"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/shared"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultDelaySeconds = 2
	maximumDelaySeconds = 60
	maximumNameLength   = 120
)

// GreetingWorkflow orchestrates the durable business process.
//
// Workflow code must be deterministic because Temporal can replay its Event
// History to rebuild the Workflow state after a Worker restart or deployment.
func GreetingWorkflow(ctx workflow.Context, input model.GreetingInput) (model.GreetingResult, error) {
	logger := workflow.GetLogger(ctx)

	input, err := normalizeInput(input)
	if err != nil {
		return model.GreetingResult{}, err
	}

	status := model.GreetingStatus{Phase: "running"}
	if err := workflow.SetQueryHandler(ctx, shared.GreetingStatusQuery, func() (model.GreetingStatus, error) {
		return status, nil
	}); err != nil {
		return model.GreetingResult{}, fmt.Errorf("register status query: %w", err)
	}

	activityTimeout := time.Duration(input.DelaySeconds+10) * time.Second

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout:    activityTimeout,
		ScheduleToCloseTimeout: activityTimeout*3 + 5*time.Second,
		HeartbeatTimeout:       5 * time.Second,
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
	future := workflow.ExecuteActivity(ctx, shared.ComposeGreetingActivityName, input)
	if err := future.Get(ctx, &result); err != nil {
		status = model.GreetingStatus{Phase: "failed", Failure: err.Error()}
		return model.GreetingResult{}, fmt.Errorf("compose greeting Activity failed: %w", err)
	}
	status = model.GreetingStatus{Phase: "completed", Result: &result}

	logger.Info(
		"GreetingWorkflow completed",
		"message", result.Message,
		"activityAttempt", result.Attempt,
	)

	return result, nil
}

func normalizeInput(input model.GreetingInput) (model.GreetingInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return model.GreetingInput{}, fmt.Errorf("name is required")
	}
	if len([]rune(input.Name)) > maximumNameLength {
		return model.GreetingInput{}, fmt.Errorf("name cannot exceed %d characters", maximumNameLength)
	}
	if input.DelaySeconds < 0 {
		return model.GreetingInput{}, fmt.Errorf("delay seconds cannot be negative")
	}
	if input.DelaySeconds == 0 {
		input.DelaySeconds = defaultDelaySeconds
	}
	if input.DelaySeconds > maximumDelaySeconds {
		return model.GreetingInput{}, fmt.Errorf("delay seconds cannot exceed %d", maximumDelaySeconds)
	}

	switch strings.ToLower(strings.TrimSpace(input.Language)) {
	case "", "pt", "pt-br", "portuguese", "português":
		input.Language = "pt-BR"
	case "en", "english":
		input.Language = "en"
	default:
		return model.GreetingInput{}, fmt.Errorf("unsupported language %q: use pt-BR or en", input.Language)
	}
	return input, nil
}
