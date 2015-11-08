package ectx

// A ContextErr is an error with some context information
type ContextErr struct {
	context string
	err     error
}

func (c ContextErr) Error() string {
	return c.context + ": " + c.err.Error()
}

// Err creates an error that wraps err with some context information.
func Err(context string, err error) error {
	if err == nil {
		return nil
	}
	return ContextErr{context: context, err: err}
}
