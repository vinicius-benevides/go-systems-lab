package shared

import "os"

// TemporalConfig contains the connection and routing configuration shared by
// the Starter and Worker. Environment variables keep the sample portable
// between local development, CI and a deployed Temporal cluster.
type TemporalConfig struct {
	HostPort  string
	Namespace string
	TaskQueue string
}

func LoadTemporalConfig() TemporalConfig {
	return TemporalConfig{
		HostPort:  envOrDefault("TEMPORAL_ADDRESS", DefaultTemporalAddress),
		Namespace: envOrDefault("TEMPORAL_NAMESPACE", DefaultTemporalNamespace),
		TaskQueue: envOrDefault("TEMPORAL_TASK_QUEUE", TaskQueue),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
