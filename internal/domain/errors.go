package domain

import "errors"

type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

func AsDomainError(err error) (*DomainError, bool) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}

var (
	ErrProductNotFound     = &DomainError{Code: "PRODUCT_NOT_FOUND", Message: "producto no encontrado"}
	ErrInsufficientStock   = &DomainError{Code: "INSUFFICIENT_STOCK", Message: "stock insuficiente"}
	ErrOrderNotFound       = &DomainError{Code: "ORDER_NOT_FOUND", Message: "orden no encontrada"}
	ErrOrderNotOwned       = &DomainError{Code: "ORDER_NOT_OWNED", Message: "la orden no pertenece al usuario"}
	ErrOrderNotCancellable = &DomainError{Code: "ORDER_NOT_CANCELLABLE", Message: "solo se pueden cancelar ordenes en estado PENDING"}
	ErrInvalidCredentials  = &DomainError{Code: "INVALID_CREDENTIALS", Message: "credenciales invalidas"}
	ErrEmailAlreadyExists  = &DomainError{Code: "EMAIL_ALREADY_EXISTS", Message: "el email ya esta registrado"}
	ErrUnauthorized        = &DomainError{Code: "UNAUTHORIZED", Message: "no autorizado"}
	ErrInvalidInput        = &DomainError{Code: "INVALID_INPUT", Message: "datos de entrada invalidos"}
)

func invalidInput(message string) *DomainError {
	return &DomainError{
		Code:    ErrInvalidInput.Code,
		Message: message,
	}
}
