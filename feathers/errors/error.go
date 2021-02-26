package fErr

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

func NewBadRequest(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "BadRequest",
		Message:   message,
		Code:      400,
		ClassName: "bad-request",
		Data:      data,
	}
}

func NewNotAuthenticated(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "NotAuthenticated",
		Message:   message,
		Code:      401,
		ClassName: "not-authenticated",
		Data:      data,
	}
}

func NewPaymentError(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "PaymentError",
		Message:   message,
		Code:      402,
		ClassName: "payment-error",
		Data:      data,
	}
}

func NewForbidden(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "Forbidden",
		Message:   message,
		Code:      403,
		ClassName: "Forbidden",
		Data:      data,
	}
}

func NewNotFound(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "NotFound",
		Message:   message,
		Code:      404,
		ClassName: "not-found",
		Data:      data,
	}
}

func NewMethodNotAllowed(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "MethodNotAllowed",
		Message:   message,
		Code:      405,
		ClassName: "method-not-allowed ",
		Data:      data,
	}
}

func NewNotAcceptable(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "NotAcceptable",
		Message:   message,
		Code:      406,
		ClassName: "not-acceptable ",
		Data:      data,
	}
}

func NewTimeout(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "Timeout",
		Message:   message,
		Code:      407,
		ClassName: "timeout",
		Data:      data,
	}
}

func NewConflict(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "Conflict",
		Message:   message,
		Code:      409,
		ClassName: "conflict",
		Data:      data,
	}
}

func NewGone(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "Gone",
		Message:   message,
		Code:      410,
		ClassName: "gone",
		Data:      data,
	}
}

func NewLengthRequired(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "LengthRequired",
		Message:   message,
		Code:      411,
		ClassName: "length-required",
		Data:      data,
	}
}

func NewUnprocessable(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "Unprocessable",
		Message:   message,
		Code:      422,
		ClassName: "unprocessable",
		Data:      data,
	}
}

func NewTooManyRequests(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "TooManyRequests",
		Message:   message,
		Code:      429,
		ClassName: "too-many-requests",
		Data:      data,
	}
}

func NewGeneralError(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "GeneralError",
		Message:   message,
		Code:      500,
		ClassName: "general-error",
		Data:      data,
	}
}

func NewNotImplemented(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "NotImplemented",
		Message:   message,
		Code:      501,
		ClassName: "not-implemented",
		Data:      data,
	}
}

func NewBadGateway(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "BadGateway",
		Message:   message,
		Code:      502,
		ClassName: "bad-gateway",
		Data:      data,
	}
}

func NewUnavailable(message string, data interface{}) FeathersError {
	return FeathersError{
		Name:      "Unavailable",
		Message:   message,
		Code:      503,
		ClassName: "unavailable",
		Data:      data,
	}
}

func Convert(err error) FeathersError {
	return NewGeneralError(err.Error(), nil)
}
