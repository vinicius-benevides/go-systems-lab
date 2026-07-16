package workflows

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/activities"
	"github.com/vinicius-benevides/go-systems-lab/season-0/temporal/01-temporal-go-lab/internal/model"
	"go.temporal.io/sdk/testsuite"
)

func TestGreetingWorkflow(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.OnActivity(
		activities.ComposeGreeting,
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
