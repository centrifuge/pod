package errors

type GenericError struct {
	S string
}

func (e *GenericError) Error() string {
	return "Generic Error: " + e.S
}
