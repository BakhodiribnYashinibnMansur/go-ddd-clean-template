# HTTP Utility Package (`pkg/httpx`)

## Overview
The `httpx` package provides generic HTTP utilities for Gin-based applications.

## Symbols
### Exported Types
- `MockType`

### Exported Functions
- `ExtractBearerToken`
- `ExtractBasicToken`
- `ParseAuthorizationType`
- `DownloadFile`
- `FileTransfer`
- `GenerateToken`
- `GetAPIKey`
- `GetApiKeyType`
- `GetArrayStringQuery`
- `GetAuthorization`
- `GetBooleanQuery`
- `GetClientDomain`
- `GetCtxSessionID`
- `GetDateOrderQuery`
- `GetDateQuery`
- `GetDeviceID`
- `GetDeviceIDUUID`
- `GetFieldsParamsQuery`
- `GetFloat64Query`
- `GetForwardedProto`
- `GetHeader`
- `GetIPAddress`
- `GetInt64Param`
- `GetInt64Query`
- `GetLanguage`
- `GetMock`
- `GetMocks`
- `GetNullArrayStringQuery`
- `GetNullBooleanQuery`
- `GetNullBooleanStringQuery`
- `GetNullDateQuery`
- `GetNullFloat64Param`
- `GetNullInt64Param`
- `GetNullInt64Query`
- `GetNullIntParam`
- `GetNullIntQuery`
- `GetNullStringParam`
- `GetNullStringQuery`
- `GetNullUUIDParam`
- `GetNullUUIDQuery`
- `GetPageQuery`
- `GetPageSizeQuery`
- `GetPagination`
- `GetRequestID`
- `GetSearchParamsQuery`
- `GetSortParamsQuery`
- `GetStringArrayQuery`
- `GetStringParam`
- `GetStringQuery`
- `GetUUIDParam`
- `GetUUIDQuery`
- `GetUserAgent`
- `GetUserID`
- `GetUserRole`
- `GetVersion`
- `HandleMockDelay`
- `HandleMockEmpty`
- `HandleMockError`
- `IsMockMode`
- `ListPagination`
- `LogError`
- `Mock`
- `MockCreated`
- `MockDelete`
- `MockResponse`
- `MockSuccess`
- `MockUpdate`
- `ResponseHeaderXTotalCountWrite`

## Usage
```go
import "gct/pkg/httpx"
```

