package workflows

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/activities"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/model"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

func TestGreetingWorkflow(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterActivityWithOptions(activities.ComposeGreeting, activity.RegisterOptions{Name: shared.ComposeGreetingActivityName})
	env.OnActivity(
		shared.ComposeGreetingActivityName,
		mock.Anything,
		mock.MatchedBy(func(input model.GreetingInput) bool {
			return input.Name == "Ada" && input.Language == "pt-BR"
		}),
	).Return(model.GreetingResult{Message: "Olá, Ada!", Attempt: 1}, nil).Once()

	env.ExecuteWorkflow(GreetingWorkflow, model.GreetingInput{Name: "  Ada  "})

	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("GreetingWorkflow() error = %v", err)
	}

	var result model.GreetingResult
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("GetWorkflowResult() error = %v", err)
	}
	if result.Message != "Olá, Ada!" || result.Attempt != 1 {
		t.Fatalf("GreetingWorkflow() result = %#v, want successful greeting", result)
	}

	queryResult, err := env.QueryWorkflow(shared.GreetingStatusQuery)
	if err != nil {
		t.Fatalf("QueryWorkflow() error = %v", err)
	}
	var status model.GreetingStatus
	if err := queryResult.Get(&status); err != nil {
		t.Fatalf("query result Get() error = %v", err)
	}
	if status.Phase != "completed" || status.Result == nil || status.Result.Message != result.Message {
		t.Fatalf("status = %#v, want completed status with result", status)
	}
}

func TestGreetingWorkflowRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input model.GreetingInput
		want  string
	}{
		{
			name:  "missing name",
			input: model.GreetingInput{Name: "  "},
			want:  "name is required",
		},
		{
			name:  "negative delay",
			input: model.GreetingInput{Name: "Ada", DelaySeconds: -1},
			want:  "delay seconds cannot be negative",
		},
		{
			name:  "unsupported language",
			input: model.GreetingInput{Name: "Ada", Language: "es"},
			want:  "unsupported language",
		},
		{
			name:  "delay above the activity contract",
			input: model.GreetingInput{Name: "Ada", DelaySeconds: 61},
			want:  "delay seconds cannot exceed 60",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.ExecuteWorkflow(GreetingWorkflow, tt.input)

			err := env.GetWorkflowError()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("GreetingWorkflow() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}
