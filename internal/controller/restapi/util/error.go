package util

import (
	"errors"

	"gct/pkg/logger"
)

func LogError(l logger.Log, err error, msg string) {
	l.Errorw(msg,
		"error", err,
		"type", "handler_error",
	)
}

var (
	ErrParamIsEmpty       = errors.New("parameter is empty")
	ErrParsingQuery       = errors.New("error while parsing query parameter")
	ErrUnmarshalData      = errors.New("error while unmarshalling data")
	ErrParamIsInvalid     = errors.New("parameter is invalid")
	ErrUserIdNotFound     = errors.New("user id not found in context")
	ErrParsingUUID        = errors.New("error while parsing UUID")
	ErrRoleNotFound       = errors.New("user role not found in context")
	ErrDataUnsignedInt    = errors.New("value is not a positive integer")
	ErrSessionIDNotFound  = errors.New("sessionID not found")
	ErrInvalidSessionID   = errors.New("sessionID is not a string or UUID")
	ErrDateOrderEmpty     = errors.New("date-order query is empty")
	ErrDateOrderInvalid   = errors.New("invalid date-order value. it is not same with asc or desc")
	ErrApiKeyTypeNotFound = errors.New("API key type not found")
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
