package logging

func ToInterface[T any](values []T) []any {
	inf := make([]any, len(values))
	for idx, v := range values {
		inf[idx] = v
	}
	return inf
}
