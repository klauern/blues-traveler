package hooks

import "testing"

func BenchmarkParseCursorResponse(b *testing.B) {
	output := `{"permission":"deny","userMessage":"User message","agentMessage":"Agent message","continue":false}`

	for i := 0; i < b.N; i++ {
		_, _ = parseCursorResponse(output)
	}
}

func BenchmarkParseCursorResponseMinimal(b *testing.B) {
	output := `{"permission":"allow"}`

	for i := 0; i < b.N; i++ {
		_, _ = parseCursorResponse(output)
	}
}

func BenchmarkParseCursorResponseNonJSON(b *testing.B) {
	output := "This is plain text output"

	for i := 0; i < b.N; i++ {
		_, _ = parseCursorResponse(output)
	}
}
