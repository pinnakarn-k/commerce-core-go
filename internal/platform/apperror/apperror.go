package apperror

type ErrorKind string

const (
	KindInternal     ErrorKind = "internal"
	KindBadRequest   ErrorKind = "bad_request"
	KindNotFound     ErrorKind = "not_found"
	KindConflict     ErrorKind = "conflict"
	KindUnauthorized ErrorKind = "unauthorized"
	KindForbidden    ErrorKind = "forbidden"
)

type AppError struct {
	Kind    ErrorKind
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(
	kind ErrorKind,
	code string,
	message string,
	err error,
) *AppError {
	return &AppError{
		Kind:    kind,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func Internal(err error) *AppError {
	return New(
		KindInternal,
		"INTERNAL_SERVER_ERROR",
		"internal server error",
		err,
	)
}

func BadRequest(
	code string,
	message string,
	err error,
) *AppError {
	return New(
		KindBadRequest,
		code,
		message,
		err,
	)
}

func NotFound(
	code string,
	message string,
	err error,
) *AppError {
	return New(
		KindNotFound,
		code,
		message,
		err,
	)
}

func Conflict(
	code string,
	message string,
	err error,
) *AppError {
	return New(
		KindConflict,
		code,
		message,
		err,
	)
}

func Unauthorized(
	code string,
	message string,
	err error,
) *AppError {
	return New(
		KindUnauthorized,
		code,
		message,
		err,
	)
}

func Forbidden(
	code string,
	message string,
	err error,
) *AppError {
	return New(
		KindForbidden,
		code,
		message,
		err,
	)
}
