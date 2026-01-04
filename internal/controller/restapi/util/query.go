package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// URL Query Argument Helpers

func GetStringQuery(ctx *gin.Context, queryName string) (string, error) {
	param := ctx.Query(queryName)
	if param == "" {
		return "", fmt.Errorf("%w: %s", ErrParamIsEmpty, queryName)
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
		return []string{}, fmt.Errorf("%w: %s", ErrParamIsEmpty, queryName)
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
		return 0, fmt.Errorf("%w: %s", ErrParsingQuery, err.Error())
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
		return 0, fmt.Errorf("%w: %s", ErrParsingQuery, err.Error())
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
		return 0, fmt.Errorf("%w: %s", ErrParamIsInvalid, queryData)
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
		return 0, fmt.Errorf("%w: %s", ErrParsingQuery, err.Error())
	}
	return paramNum, nil
}

func GetBooleanQuery(ctx *gin.Context, queryName string) (bool, error) {
	param := ctx.Query(queryName)
	if param == "" {
		return false, fmt.Errorf("%w: %s", ErrParamIsInvalid, queryName)
	}
	result, err := strconv.ParseBool(param)
	if err != nil {
		return false, fmt.Errorf("failed to parse boolean parameter %s: %w", queryName, err)
	}
	return result, nil
}

func GetNullBooleanQuery(ctx *gin.Context, queryName string) (bool, error) {
	param := ctx.Query(queryName)
	if param != "" {
		result, err := strconv.ParseBool(param)
		if err != nil {
			return false, fmt.Errorf("failed to parse boolean parameter %s: %w", queryName, err)
		}
		return result, nil
	}
	return true, nil
}

func GetNullBooleanStringQuery(ctx *gin.Context, queryName string) (string, error) {
	param := ctx.Query(queryName)
	if param != "" {
		_, err := strconv.ParseBool(param)
		if err != nil {
			return "", fmt.Errorf("%w: %s", ErrParamIsInvalid, param)
		}
		return param, nil
	}
	return "", nil
}

func GetUUIDQuery(ctx *gin.Context, queryName string) (uuid.UUID, error) {
	param := ctx.Query(queryName)
	if param == "" {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrParamIsInvalid, queryName)
	}
	paramUUID, err := uuid.Parse(param)
	if err != nil {
		logger.GetLogger().Errorw("GetUUIDQuery - Parse", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%w: %s", ErrParamIsInvalid, param)
	}
	return paramUUID, nil
}

func GetNullUUIDQuery(ctx *gin.Context, queryName string) (uuid.UUID, error) {
	queryData := ctx.Query(queryName)
	if queryData != "" {
		queryUUID, err := uuid.Parse(queryData)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%w: %s", ErrParamIsInvalid, queryData)
		}
		return queryUUID, nil
	}
	return uuid.Nil, nil
}

func GetDateQuery(ctx *gin.Context, queryName string) (string, error) {
	queryDate := GetNullStringQuery(ctx, queryName)
	if queryDate != "" {
		parseDate, err := time.Parse(consts.FormatDate, queryDate)
		if err != nil {
			return "", fmt.Errorf("failed to parse date parameter %s: %w", queryName, err)
		}
		return parseDate.Format(consts.FormatDate), nil
	}
	return "", nil
}

func GetNullDateQuery(ctx *gin.Context, queryName string) (time.Time, error) {
	queryData := ctx.Query(queryName)
	if queryData != "" {
		result, err := time.Parse(consts.FormatDate, queryData)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse date parameter %s: %w", queryName, err)
		}
		return result, nil
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
			return nil, fmt.Errorf("%w: %s", ErrParamIsInvalid, searchQuery)
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
		if len(parts) != 2 || (strings.ToLower(parts[1]) != consts.OrderAsc && strings.ToLower(parts[1]) != consts.OrderDesc) || parts[0] == "" {
			return nil, fmt.Errorf("%w: %s", ErrParamIsInvalid, sortQuery)
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
		return "", ErrDateOrderEmpty
	}
	if dateOrder != consts.OrderAsc && dateOrder != consts.OrderDesc {
		return "", ErrDateOrderInvalid
	}
	return dateOrder, nil
}

// Pagination Helpers

func GetPageQuery(ctx *gin.Context) (int64, error) {
	offsetStr := ctx.DefaultQuery(consts.QueryPage, "1")
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrParsingQuery, err.Error())
	}
	if offset < 0 {
		return 0, fmt.Errorf("%w: %d", ErrDataUnsignedInt, offset)
	}
	return offset, nil
}

func GetPageSizeQuery(ctx *gin.Context) (int64, error) {
	limitStr := ctx.DefaultQuery(consts.QueryPageSize, "10")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrParsingQuery, err.Error())
	}
	if limit < 0 {
		return 0, fmt.Errorf("%w: %d", ErrDataUnsignedInt, limit)
	}
	return limit, nil
}

func GetPagination(ctx *gin.Context) (domain.Pagination, error) {
	limit, err := GetInt64Query(ctx, consts.QueryLimit)
	if err != nil {
		return domain.Pagination{}, err
	}
	offset, err := GetInt64Query(ctx, consts.QueryOffset)
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
