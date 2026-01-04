package util

import (
	"errors"

	"gct/consts"
	"gct/pkg/logger"
	"go.uber.org/zap"
)

func LogError(l logger.Log, err error, msg string) {
	l.Errorw(msg,
		zap.Error(err),
		zap.String("type", consts.TypeHandlerError),
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
	ErrCSRFMissing        = errors.New("CSRF token is missing")
	ErrCSRFInvalid        = errors.New("CSRF token is invalid or mismatched")
)

const (
	ParamInvalid = "parameter %s is invalid"
	QueryInvalid = "query parameter %s is invalid"
)
