package shared

const (
	// TaskQueue is the logical queue polled by the Worker.
	// The Starter and Worker must use exactly the same value.
	TaskQueue = "greeting-task-queue"
)
