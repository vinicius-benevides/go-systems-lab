package activities

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/model"
	"go.temporal.io/sdk/activity"
)

const defaultDelaySeconds = 2

// ComposeGreeting simulates an operation that interacts with the outside world.
// In a real system, this Activity could call an API, query a database or publish
// a message. Activities may use time.Now, native timers, network clients and I/O.
func ComposeGreeting(ctx context.Context, input model.GreetingInput) (model.GreetingResult, error) {
	logger := activity.GetLogger(ctx)
	info := activity.GetInfo(ctx)

	logger.Info(
		"ComposeGreeting Activity started",
		"name", input.Name,
		"language", input.Language,
		"attempt", info.Attempt,
	)

	// This optional failure lets us observe Temporal's automatic retry behavior.
	if input.SimulateFailure && info.Attempt == 1 {
		logger.Warn("Simulating a transient failure on the first attempt")
		return model.GreetingResult{}, fmt.Errorf("simulated transient failure")
	}

	delay := input.DelaySeconds
	if delay <= 0 {
		delay = defaultDelaySeconds
	}

	timer := time.NewTimer(time.Duration(delay) * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		// The simulated external work has completed.
	case <-ctx.Done():
		return model.GreetingResult{}, ctx.Err()
	}

	message := buildMessage(input.Name, input.Language)
	result := model.GreetingResult{
		Message:     message,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Attempt:     info.Attempt,
	}

	logger.Info(
		"ComposeGreeting Activity completed",
		"message", result.Message,
		"attempt", result.Attempt,
	)

	return result, nil
}

func buildMessage(name, language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "en", "english":
		return fmt.Sprintf("Hello, %s! Your first Temporal Workflow completed successfully.", name)
	default:
		return fmt.Sprintf("Olá, %s! Seu primeiro Workflow com Temporal foi concluído com sucesso.", name)
	}
}
