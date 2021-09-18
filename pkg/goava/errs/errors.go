package errs

import (
	"errors"
	"github.com/lingdor/stackerror"
	"reflect"
)

const (
	ReasonOther = "REASON_OTHER"
)

type BaseError struct {
	error

	Stack   stackerror.StackError
	Cause   error
	Message string
	Details string
	Reason  string
}

func NewBaseErrorFromCause(cause error) *BaseError {
	return &BaseError{
		Stack:   stackerror.New(GetRootCauser(cause).Error()),
		Cause:   cause,
		Message: cause.Error(),
		Reason:  ReasonOther,
	}
}

func GetRootCauser(e error) error {
	for {
		switch ex := e.(type) {
		case *BaseError:
			if ex.Cause == nil {
				return ex
			}
			e = ex.Cause
		default:
			return ex
		}
	}
}

func NewBaseErrorWithReason(message string, reason string) *BaseError {
	r := NewBaseError(message)
	r.Reason = reason
	return r
}

func NewBaseErrorWithReasonDetails(message string, details string, reason string) *BaseError {
	r := NewBaseError(message)
	r.Reason = reason
	r.Details = details
	return r
}

func NewBaseError(message string) *BaseError {
	return &BaseError{
		Stack:   stackerror.New(message),
		Cause:   nil,
		Message: message,
		Reason:  ReasonOther,
	}
}

func NewBaseErrorFromCauseMsgReason(cause error, message string, reason string) *BaseError {
	r := NewBaseErrorFromCauseMsg(cause, message)
	r.Reason = reason
	return r
}

func NewBaseErrorFromCauseMsg(cause error, message string) *BaseError {
	return &BaseError{
		Stack:   stackerror.New(GetRootCauser(cause).Error()),
		Cause:   cause,
		Message: message,
		Reason:  ReasonOther,
	}
}

func (e *BaseError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return ""
}

func (e *BaseError) Unwrap() error {
	return e.Cause
}

func FindBaseError(e error) *BaseError {
	var ese *BaseError
	if errors.As(e, &ese) {
		return ese
	}

	r := reflect.ValueOf(e)
	if r.IsValid() {
		f := reflect.Indirect(r)
		if f.IsValid() {
			f = f.FieldByName("BaseError")
			if f.IsValid() {
				e, casted := f.Interface().(BaseError)
				if casted {
					return &e
				}
			}
		}
	}
	return nil
}
