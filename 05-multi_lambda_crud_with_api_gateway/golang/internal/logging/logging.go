package logging

import (
	"log/slog"
	"os"
)

// Setup makes every slog call in the program write JSON instead of plain text.
//
// Lambda forwards anything written to stdout into CloudWatch, so stdout is all
// we need - there is no AWS SDK call involved. Writing JSON rather than text is
// what lets CloudWatch split each line into fields, so Logs Insights can filter
// on them:
//
//	{"level":"ERROR","msg":"failed to check email","err":"..."}

func Setup() {

	// slog.NewJSONHandler(...) means logs should be formatted as JSON.
	// os.Stdout means the JSON should be written to standard output—the normal terminal output stream.
	// nil means no custom handler options are being provided.
	writeJSONToStdout := slog.NewJSONHandler(os.Stdout, nil)

	// Creates a logger that uses the JSON handler.
	logger := slog.New(writeJSONToStdout)

	// Makes that logger the program’s default logger
	slog.SetDefault(logger)
}
