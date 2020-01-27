package smtpdetails

type AlreadyExistsError struct {
	Message string
}

func (e *AlreadyExistsError) Error() string {
	return e.Message
}

func IsAlreadyExistsError(err error) bool {
	_, ok := err.(*AlreadyExistsError)
	return ok
}

type NotExistError struct {
	Message string
}

func (e *NotExistError) Error() string {
	return e.Message
}

func IsNotExistError(err error) bool {
	_, ok := err.(*NotExistError)
	return ok
}
