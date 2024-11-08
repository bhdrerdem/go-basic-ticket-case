package errors

type RestErrorInterface interface {
	Error() string
}

type RestError struct {
	Message string
	Status  int
}

func (e RestError) Error() string {
	return e.Message
}

func NewRestError(message string, status int) RestError {
	return RestError{
		Message: message,
		Status:  status,
	}
}
