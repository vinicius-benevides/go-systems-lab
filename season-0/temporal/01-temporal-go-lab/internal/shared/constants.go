package shared

const (
	// TaskQueue is the logical queue polled by the Worker.
	// The Starter and Worker must use exactly the same value.
	TaskQueue = "greeting-task-queue"

	GreetingWorkflowName        = "greeting.v1"
	ComposeGreetingActivityName = "greeting.compose.v1"
	GreetingStatusQuery         = "greeting.status"

	DefaultTemporalAddress   = "localhost:7233"
	DefaultTemporalNamespace = "default"
)
