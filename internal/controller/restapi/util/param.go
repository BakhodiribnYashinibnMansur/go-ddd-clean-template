package util

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// URI Parameter Helpers

func GetStringParam(ctx *gin.Context, paramName string) (string, error) {
	param := ctx.Param(paramName)
	if param == "" {
		return "", fmt.Errorf(ErrParamIsEmpty, paramName)
	}
	return param, nil
}

func GetNullStringParam(ctx *gin.Context, paramName string) (string, error) {
	return ctx.Param(paramName), nil
}

func GetUUIDParam(ctx *gin.Context, paramName string) (uuid.UUID, error) {
	queryData := ctx.Param(paramName)
	if queryData == "" {
		return uuid.Nil, fmt.Errorf(ErrParamIsEmpty, paramName)
	}
	queryUUID, err := uuid.Parse(queryData)
	if err != nil {
		return uuid.Nil, fmt.Errorf(ErrParamIsInvalid, queryData)
	}
	return queryUUID, nil
}

func GetNullUUIDParam(ctx *gin.Context, paramName string) (uuid.UUID, error) {
	queryData := ctx.Param(paramName)
	if queryData != "" {
		queryUUID, err := uuid.Parse(queryData)
		if err != nil {
			return uuid.Nil, fmt.Errorf(ErrParsingUUID, err.Error())
		}
		return queryUUID, nil
	}
	return uuid.Nil, nil
}

func GetInt64Param(ctx *gin.Context, paramName string) (int64, error) {
	queryData := ctx.Param(paramName)
	if queryData == "" {
		return 0, fmt.Errorf(ErrParamIsEmpty, paramName)
	}
	queryInt, err := strconv.ParseInt(queryData, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(ErrParsingQuery, err.Error())
	}
	return queryInt, nil
}

func GetNullInt64Param(ctx *gin.Context, paramName string) (int64, error) {
	queryData := ctx.Param(paramName)
	if queryData != "" {
		queryInt, err := strconv.ParseInt(queryData, 10, 64)
		if err != nil {
			return 0, fmt.Errorf(ErrParsingQuery, err.Error())
		}
		return queryInt, nil
	}
	return 0, nil
}

func GetNullIntParam(ctx *gin.Context, paramName string) (int, error) {
	queryData := ctx.Param(paramName)
	if queryData != "" {
		queryInt, err := strconv.Atoi(queryData)
		if err != nil {
			return 0, fmt.Errorf(ErrParamIsInvalid, queryData)
		}
		return queryInt, nil
	}
	return 0, nil
}

func GetNullFloat64Param(ctx *gin.Context, paramName string) (float64, error) {
	queryData := ctx.Param(paramName)
	if queryData != "" {
		queryFloat, err := strconv.ParseFloat(queryData, 64)
		if err != nil {
			return 0, fmt.Errorf(ErrParsingQuery, err.Error())
		}
		return queryFloat, nil
	}
	return 0, nil
}
