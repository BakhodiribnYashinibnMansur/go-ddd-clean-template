package util

import (
	"errors"

	"github.com/evrone/go-clean-template/pkg/logger"
)

func LogError(l logger.Log, err error, msg string) {
	l.Errorw(msg,
		"error", err,
		"type", "handler_error",
	)
}

var (
	ErrParamIsEmpty    = "parameter %s is empty"
	ErrParsingQuery    = "error while parsing query parameter: %s"
	ErrUnmarshalData   = "error while unmarshalling data: %s"
	ErrParamIsInvalid  = "parameter %s is invalid"
	ErrUserIdNotFound  = errors.New("user id not found in context")
	ErrParsingUUID     = "error while parsing UUID: %s"
	ErrRoleNotFound    = errors.New("user role not found in context")
	ErrDataUnsignedInt = "value %d is not a positive integer"
)

const (
	ParamInvalid     = "parameter %s is invalid"
	QueryInvalid     = "query parameter %s is invalid"
	ParseDate        = "2006-01-02"
	FormatDate       = "2006-01-02"
	OrderAsc         = "asc"
	OrderDesc        = "desc"
	AcceptedLanguage = "Accept-Language"
	Language         = "Language"
	ApiKeyTypeHeader = "X-Api-Key-Type"
	AppVersionHeader = "appVersion"
)
