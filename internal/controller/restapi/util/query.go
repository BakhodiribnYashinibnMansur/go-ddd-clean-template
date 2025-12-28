package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// URL Query Argument Helpers

func GetStringQuery(ctx *gin.Context, queryName string) (string, error) {
	param := ctx.Query(queryName)
	if param == "" {
		return "", fmt.Errorf(ErrParamIsEmpty, queryName)
	}
	return param, nil
}

func GetNullStringQuery(ctx *gin.Context, queryName string) string {
	param := ctx.Query(queryName)
	return strings.TrimSpace(param)
}

func GetStringArrayQuery(ctx *gin.Context, queryName string) []string {
	str := ctx.Query(queryName)
	if str == "" {
		return []string{}
	}
	return strings.Split(str, ",")
}

func GetArrayStringQuery(ctx *gin.Context, queryName string) ([]string, error) {
	param := ctx.Query(queryName)
	if param == "" {
		return []string{}, fmt.Errorf(ErrParamIsEmpty, queryName)
	}
	return strings.Split(param, ","), nil
}

func GetNullArrayStringQuery(ctx *gin.Context, queryName string) ([]string, error) {
	queryData := ctx.Query(queryName)
	if queryData != "" {
		return strings.Split(queryData, ","), nil
	}
	return []string{}, nil
}

func GetInt64Query(ctx *gin.Context, queryName string) (int64, error) {
	param := ctx.Query(queryName)
	if param == "" {
		return 0, nil
	}
	paramInt, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(ErrParsingQuery, err.Error())
	}
	return paramInt, nil
}

func GetNullInt64Query(ctx *gin.Context, queryName string) (int64, error) {
	queryData := ctx.Query(queryName)
	if queryData == "" {
		return 0, nil
	}
	queryInt, err := strconv.ParseInt(queryData, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(ErrParsingQuery, err.Error())
	}
	return queryInt, nil
}

func GetNullIntQuery(ctx *gin.Context, queryName string) (int, error) {
	queryData := ctx.Query(queryName)
	if queryData == "" {
		return 0, nil
	}
	queryInt, err := strconv.Atoi(queryData)
	if err != nil {
		return 0, fmt.Errorf(ErrParamIsInvalid, queryData)
	}
	return queryInt, nil
}

func GetFloat64Query(ctx *gin.Context, queryName string) (float64, error) {
	param := ctx.Query(queryName)
	if param == "" {
		return 0, nil
	}
	paramNum, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return 0, fmt.Errorf(ErrParsingQuery, err.Error())
	}
	return paramNum, nil
}

func GetBooleanQuery(ctx *gin.Context, queryName string) (bool, error) {
	param := ctx.Query(queryName)
	if param == "" {
		return false, fmt.Errorf(ErrParamIsInvalid, queryName)
	}
	return strconv.ParseBool(param)
}

func GetNullBooleanQuery(ctx *gin.Context, queryName string) (bool, error) {
	param := ctx.Query(queryName)
	if param != "" {
		return strconv.ParseBool(param)
	}
	return true, nil
}

func GetNullBooleanStringQuery(ctx *gin.Context, queryName string) (string, error) {
	param := ctx.Query(queryName)
	if param != "" {
		_, err := strconv.ParseBool(param)
		if err != nil {
			return "", fmt.Errorf(ErrParamIsInvalid, param)
		}
		return param, nil
	}
	return "", nil
}

func GetUUIDQuery(ctx *gin.Context, queryName string) (uuid.UUID, error) {
	param := ctx.Query(queryName)
	if param == "" {
		return uuid.Nil, fmt.Errorf(ErrParamIsInvalid, queryName)
	}
	paramUUID, err := uuid.Parse(param)
	if err != nil {
		logger.GetLogger().Error(err)
		return uuid.Nil, fmt.Errorf(ErrParamIsInvalid, param)
	}
	return paramUUID, nil
}

func GetNullUUIDQuery(ctx *gin.Context, queryName string) (uuid.UUID, error) {
	queryData := ctx.Query(queryName)
	if queryData != "" {
		queryUUID, err := uuid.Parse(queryData)
		if err != nil {
			return uuid.Nil, fmt.Errorf(QueryInvalid, queryData)
		}
		return queryUUID, nil
	}
	return uuid.Nil, nil
}

func GetDateQuery(ctx *gin.Context, queryName string) (string, error) {
	queryDate := GetNullStringQuery(ctx, queryName)
	if queryDate != "" {
		parseDate, err := time.Parse(ParseDate, queryDate)
		if err != nil {
			return "", err
		}
		return parseDate.Format(FormatDate), nil
	}
	return "", nil
}

func GetNullDateQuery(ctx *gin.Context, queryName string) (time.Time, error) {
	queryData := ctx.Query(queryName)
	if queryData != "" {
		return time.Parse(ParseDate, queryData)
	}
	return time.Time{}, nil
}

// Search, Sort and Filters

func GetSearchParamsQuery(ctx *gin.Context, queryName string, extraFields map[string]string) (map[string]string, error) {
	searchParams := make(map[string]string)
	searchQuery := strings.TrimSpace(ctx.Query(queryName))
	for key, value := range extraFields {
		searchParams[key] = value
	}
	if searchQuery == "" {
		return searchParams, nil
	}
	searchFields := strings.Split(searchQuery, ";")
	for _, field := range searchFields {
		parts := strings.Split(field, ":")
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf(ErrParamIsInvalid, searchQuery)
		}
		searchParams[parts[0]] = parts[1]
	}
	return searchParams, nil
}

func GetSortParamsQuery(ctx *gin.Context, queryName string) (map[string]string, error) {
	sortParams := make(map[string]string)
	sortQuery := strings.TrimSpace(ctx.Query(queryName))
	if sortQuery == "" {
		return sortParams, nil
	}
	sortFields := strings.Split(sortQuery, ",")
	for _, field := range sortFields {
		parts := strings.Split(field, ":")
		if len(parts) != 2 || (strings.ToLower(parts[1]) != OrderAsc && strings.ToLower(parts[1]) != OrderDesc) || parts[0] == "" {
			return nil, fmt.Errorf(ErrParamIsInvalid, sortQuery)
		}
		sortParams[parts[0]] = parts[1]
	}
	return sortParams, nil
}

func GetFieldsParamsQuery(ctx *gin.Context, queryName string, extraParams map[string]string) map[string]string {
	fieldsParams := make(map[string]string)
	fieldsQuery := strings.TrimSpace(ctx.Query(queryName))
	for _, v := range extraParams {
		if fieldsQuery != "" {
			fieldsQuery += ","
		}
		fieldsQuery += v
	}
	if fieldsQuery == "" {
		return fieldsParams
	}
	fields := strings.Split(fieldsQuery, ",")
	for _, field := range fields {
		fieldsParams[field] = field
	}
	return fieldsParams
}

func GetDateOrderQuery(ctx *gin.Context, queryName string) (string, error) {
	dateOrder := strings.ToLower(ctx.Query(queryName))
	if dateOrder == "" {
		return "", errors.New("date-order query is empty")
	}
	if dateOrder != OrderAsc && dateOrder != OrderDesc {
		return "", errors.New("invalid date-order value. it is not same with asc or desc")
	}
	return dateOrder, nil
}

// Pagination Helpers

func GetPageQuery(ctx *gin.Context) (int64, error) {
	offsetStr := ctx.DefaultQuery("page", "1")
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(ErrParsingQuery, err.Error())
	}
	if offset < 0 {
		return 0, fmt.Errorf(ErrDataUnsignedInt, offset)
	}
	return offset, nil
}

func GetPageSizeQuery(ctx *gin.Context) (int64, error) {
	limitStr := ctx.DefaultQuery("pageSize", "10")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(ErrParsingQuery, err.Error())
	}
	if limit < 0 {
		return 0, fmt.Errorf(ErrDataUnsignedInt, limit)
	}
	return limit, nil
}

func GetPagination(ctx *gin.Context) (domain.Pagination, error) {
	limit, err := GetInt64Query(ctx, "limit")
	if err != nil {
		return domain.Pagination{}, err
	}
	offset, err := GetInt64Query(ctx, "offset")
	if err != nil {
		return domain.Pagination{}, err
	}
	if limit == 0 {
		limit = 20
	}
	return domain.Pagination{Limit: limit, Offset: offset}, nil
}

func ListPagination(ctx *gin.Context) (domain.Pagination, error) {
	return GetPagination(ctx)
}
