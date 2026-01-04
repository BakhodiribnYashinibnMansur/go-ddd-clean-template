package cache_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/persistent/redis"
	"gct/internal/usecase/cache"
	"gct/pkg/logger"

	"github.com/go-redis/redismock/v9"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPublicCache(_ *testing.T) (*cache.Cache, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	log := logger.New("debug")
	repo := redis.New(db, log)
	c := cache.NewCache(repo, log)
	return c, mock
}

type TestData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestCache_CreatePublicCache(t *testing.T) {
	t.Parallel()
	c, mock := setupPublicCache(t)

	tableName := consts.TableUsers
	lang := "en"
	pagination := &domain.Pagination{Limit: 10, Offset: 0}
	duration := time.Minute
	data := TestData{Name: "John", Age: 30}
	dataBytes, _ := json.Marshal(data)

	// Expected cache key: "users_en_0_10"
	cacheKey := "users_en_0_10"

	mock.ExpectSet(cacheKey, dataBytes, duration).SetVal("OK")

	err := c.CreatePublicCache(data, tableName, lang, pagination, duration)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_GetPublicCache(t *testing.T) {
	t.Parallel()
	c, mock := setupPublicCache(t)

	tableName := consts.TableUsers
	lang := "en"
	pagination := &domain.Pagination{Limit: 10, Offset: 0}
	data := TestData{Name: "John", Age: 30}
	dataBytes, _ := json.Marshal(data)

	cacheKey := "users_en_0_10"

	mock.ExpectGet(cacheKey).SetVal(string(dataBytes))

	var out TestData
	err := c.GetPublicCache(tableName, lang, pagination, &out)
	require.NoError(t, err)
	assert.Equal(t, data, out)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_GetPublicCache_Miss(t *testing.T) {
	t.Parallel()
	c, mock := setupPublicCache(t)

	tableName := consts.TableUsers
	lang := "en"
	pagination := &domain.Pagination{Limit: 10, Offset: 0}
	cacheKey := "users_en_0_10"

	mock.ExpectGet(cacheKey).SetErr(redisClient.Nil)

	var out TestData
	err := c.GetPublicCache(tableName, lang, pagination, &out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis get")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_DeletePublicCache(t *testing.T) {
	t.Parallel()
	c, mock := setupPublicCache(t)

	tableName := consts.TableUsers
	lang := "en"
	pagination := &domain.Pagination{Limit: 10, Offset: 0}
	cacheKey := "users_en_0_10"

	mock.ExpectDel(cacheKey).SetVal(1)

	err := c.DeletePublicCache(tableName, lang, pagination)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_DeletePublicCaches(t *testing.T) {
	t.Parallel()
	c, mock := setupPublicCache(t)

	tableName := consts.TableUsers
	keys := []string{"users_en_0_10", "users_ru_0_10"}

	// Expect Scan
	mock.ExpectScan(0, tableName+"*", 100).SetVal(keys, 0)
	// Expect Del for each key
	mock.ExpectDel(keys[0]).SetVal(1)
	mock.ExpectDel(keys[1]).SetVal(1)

	err := c.DeletePublicCaches(tableName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_DeletePublicCaches_ScanError(t *testing.T) {
	t.Parallel()
	c, mock := setupPublicCache(t)

	tableName := consts.TableUsers

	mock.ExpectScan(0, tableName+"*", 100).SetErr(errors.New("scan error"))

	err := c.DeletePublicCaches(tableName)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis scan")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_CreatePublicCache_NoLangOrPagination(t *testing.T) {
	t.Parallel()
	c, mock := setupPublicCache(t)

	tableName := consts.TableUsers
	duration := time.Minute
	data := TestData{Name: "John", Age: 30}
	dataBytes, _ := json.Marshal(data)

	// Case 1: No Lang, No Pagination
	// Expected key: "users"
	mock.ExpectSet("users", dataBytes, duration).SetVal("OK")
	err := c.CreatePublicCache(data, tableName, "", nil, duration)
	require.NoError(t, err)

	// Case 2: No Lang, With Pagination
	// Expected key: "users_0_10"
	pagination := &domain.Pagination{Limit: 10, Offset: 0}
	mock.ExpectSet("users_0_10", dataBytes, duration).SetVal("OK")
	err = c.CreatePublicCache(data, tableName, "", pagination, duration)
	require.NoError(t, err)

	// Case 3: With Lang, No Pagination
	// Expected key: "users_en"
	mock.ExpectSet("users_en", dataBytes, duration).SetVal("OK")
	err = c.CreatePublicCache(data, tableName, "en", nil, duration)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
