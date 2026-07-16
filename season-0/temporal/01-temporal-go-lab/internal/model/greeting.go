package model

// GreetingInput contains the data required by the Workflow.
// Temporal serializes this value when the Workflow is started.
type GreetingInput struct {
	Name            string `json:"name"`
	Language        string `json:"language"`
	DelaySeconds    int    `json:"delaySeconds"`
	SimulateFailure bool   `json:"simulateFailure"`
}

// GreetingResult is returned by the Activity, propagated by the Workflow,
// and decoded by the Starter process.
type GreetingResult struct {
	Message     string `json:"message"`
	GeneratedAt string `json:"generatedAt"`
	Attempt     int32  `json:"attempt"`
}
