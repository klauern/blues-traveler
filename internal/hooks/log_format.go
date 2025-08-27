package hooks

const (
	LoggingFormatJSONL  = "jsonl"
	LoggingFormatPretty = "pretty"
)

// IsValidLoggingFormat returns true if the provided format is supported.
func IsValidLoggingFormat(f string) bool {
	return f == LoggingFormatJSONL || f == LoggingFormatPretty
}
