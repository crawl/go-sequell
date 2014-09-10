package ectx

type ContextErr struct {
	context string
	err     error
}

func (c ContextErr) Error() string {
	return c.context + ": " + c.err.Error()
}

func Err(context string, err error) error {
	if err == nil {
		return nil
	}
	return ContextErr{context: context, err: err}
}
