package httperrors

type FeathersError struct {
	Name      string      `json:"name"`
	Message   string      `json:"message"`
	Code      int         `json:"code"`
	ClassName string      `json:"className"`
	Data      interface{} `json:"data,omitEmpty"`
	Errors    interface{} `json:"errors,omitEmpty"`
}

func (err FeathersError) Error() string {
	return err.Message
}

func retrieveData(data []interface{}) interface{} {
	if len(data) > 1 {
		return data[0]
	}
	return nil
}

func NewBadRequest(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "BadRequest",
		Message:   message,
		Code:      400,
		ClassName: "bad-request",
		Data:      retrieveData(data),
	}
}

func NewNotAuthenticated(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "NotAuthenticated",
		Message:   message,
		Code:      401,
		ClassName: "not-authenticated",
		Data:      retrieveData(data),
	}
}

func NewPaymentError(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "PaymentError",
		Message:   message,
		Code:      402,
		ClassName: "payment-error",
		Data:      retrieveData(data),
	}
}

func NewForbidden(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "Forbidden",
		Message:   message,
		Code:      403,
		ClassName: "Forbidden",
		Data:      retrieveData(data),
	}
}

func NewNotFound(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "NotFound",
		Message:   message,
		Code:      404,
		ClassName: "not-found",
		Data:      retrieveData(data),
	}
}

func NewMethodNotAllowed(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "MethodNotAllowed",
		Message:   message,
		Code:      405,
		ClassName: "method-not-allowed ",
		Data:      retrieveData(data),
	}
}

func NewNotAcceptable(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "NotAcceptable",
		Message:   message,
		Code:      406,
		ClassName: "not-acceptable ",
		Data:      retrieveData(data),
	}
}

func NewTimeout(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "Timeout",
		Message:   message,
		Code:      407,
		ClassName: "timeout",
		Data:      retrieveData(data),
	}
}

func NewConflict(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "Conflict",
		Message:   message,
		Code:      409,
		ClassName: "conflict",
		Data:      retrieveData(data),
	}
}

func NewGone(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "Gone",
		Message:   message,
		Code:      410,
		ClassName: "gone",
		Data:      retrieveData(data),
	}
}

func NewLengthRequired(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "LengthRequired",
		Message:   message,
		Code:      411,
		ClassName: "length-required",
		Data:      retrieveData(data),
	}
}

func NewUnprocessable(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "Unprocessable",
		Message:   message,
		Code:      422,
		ClassName: "unprocessable",
		Data:      retrieveData(data),
	}
}

func NewTooManyRequests(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "TooManyRequests",
		Message:   message,
		Code:      429,
		ClassName: "too-many-requests",
		Data:      retrieveData(data),
	}
}

func NewGeneralError(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "GeneralError",
		Message:   message,
		Code:      500,
		ClassName: "general-error",
		Data:      retrieveData(data),
	}
}

func NewNotImplemented(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "NotImplemented",
		Message:   message,
		Code:      501,
		ClassName: "not-implemented",
		Data:      retrieveData(data),
	}
}

func NewBadGateway(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "BadGateway",
		Message:   message,
		Code:      502,
		ClassName: "bad-gateway",
		Data:      retrieveData(data),
	}
}

func NewUnavailable(message string, data ...interface{}) FeathersError {
	return FeathersError{
		Name:      "Unavailable",
		Message:   message,
		Code:      503,
		ClassName: "unavailable",
		Data:      retrieveData(data),
	}
}

func Convert(err error) FeathersError {
	return NewGeneralError(err.Error())
}
