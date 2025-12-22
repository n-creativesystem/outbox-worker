package logging

type LoggingHandle string

func (h LoggingHandle) String() string {
	return string(h)
}

const (
	JsonHandler    = LoggingHandle("json")
	TextHandler    = LoggingHandle("text")
	SentryHandler  = LoggingHandle("sentry")
	RollbarHandler = LoggingHandle("rollbar")
	DatadogHandler = LoggingHandle("datadog")
)

var (
	LoggingHandlers = []LoggingHandle{JsonHandler, TextHandler, SentryHandler, RollbarHandler, DatadogHandler}
)

func LoggingHandlersToInf() []any {
	results := make([]any, len(LoggingHandlers))
	for i := range LoggingHandlers {
		results[i] = LoggingHandlers[i]
	}
	return results
}
