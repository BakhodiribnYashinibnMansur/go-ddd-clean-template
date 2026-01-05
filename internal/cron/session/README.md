# Internal - Cron - Session

## Overview
The `session` package.

## Symbols
### Exported Types
- `Activity`
- `CronConfig`
- `CronJobs`
- `PgxPool`

### Exported Functions
- `GetExpireOldSessionsConfig`
- `GetSyncSessionActivityConfig`
- `NewSessionCronJobs`
- `TestCronJobs_ExpireOldSessions`
- `TestCronJobs_ExpireOldSessions_Error`
- `TestCronJobs_SyncSessionActivityToPostgres`
- `TestCronJobs_SyncSessionActivityToPostgres_NoData`
- `TestCronJobs_SyncSessionActivityToPostgres_RedisError`



## Usage
```go
import "gct/internal/cron/session"
```
