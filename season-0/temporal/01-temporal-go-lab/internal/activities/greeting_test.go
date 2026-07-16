package activities

import "testing"

func TestBuildMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		language string
		want     string
	}{
		{
			name:     "English language code",
			language: "en",
			want:     "Hello, Ada! Your first Temporal Workflow completed successfully.",
		},
		{
			name:     "English language name with surrounding spaces",
			language: " english ",
			want:     "Hello, Ada! Your first Temporal Workflow completed successfully.",
		},
		{
			name:     "Portuguese is the default",
			language: "pt-BR",
			want:     "Olá, Ada! Seu primeiro Workflow com Temporal foi concluído com sucesso.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := buildMessage("Ada", tt.language); got != tt.want {
				t.Fatalf("buildMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}
