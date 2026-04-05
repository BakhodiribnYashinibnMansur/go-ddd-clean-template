package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/platform/domain/consts"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCtxWithValue(key string, value any) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if value != nil {
		c.Set(key, value)
	}
	return c
}

func TestGetUserID(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		ctx     *gin.Context
		want    uuid.UUID
		wantErr bool
	}{
		{
			name: "uuid_type",
			ctx:  newCtxWithValue(consts.CtxUserID, validUUID),
			want: validUUID,
		},
		{
			name: "string_type",
			ctx:  newCtxWithValue(consts.CtxUserID, validUUID.String()),
			want: validUUID,
		},
		{
			name:    "missing",
			ctx:     newCtxWithValue(consts.CtxUserID, nil),
			wantErr: true,
		},
		{
			name:    "invalid_string",
			ctx:     newCtxWithValue(consts.CtxUserID, "not-a-uuid"),
			wantErr: true,
		},
		{
			name:    "wrong_type",
			ctx:     newCtxWithValue(consts.CtxUserID, 12345),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserID(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetCtxSessionID(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		ctx     *gin.Context
		want    uuid.UUID
		wantErr bool
	}{
		{
			name: "string_type",
			ctx:  newCtxWithValue(consts.CtxSessionID, validUUID.String()),
			want: validUUID,
		},
		{
			name: "uuid_type",
			ctx:  newCtxWithValue(consts.CtxSessionID, validUUID),
			want: validUUID,
		},
		{
			name:    "missing",
			ctx:     newCtxWithValue(consts.CtxSessionID, nil),
			wantErr: true,
		},
		{
			name:    "invalid_string",
			ctx:     newCtxWithValue(consts.CtxSessionID, "bad"),
			wantErr: true,
		},
		{
			name:    "wrong_type",
			ctx:     newCtxWithValue(consts.CtxSessionID, 999),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCtxSessionID(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetUserRole(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name    string
		ctx     *gin.Context
		want    uuid.UUID
		wantErr bool
	}{
		{
			name: "uuid_type",
			ctx:  newCtxWithValue(consts.CtxRoleID, validUUID),
			want: validUUID,
		},
		{
			name: "string_type",
			ctx:  newCtxWithValue(consts.CtxRoleID, validUUID.String()),
			want: validUUID,
		},
		{
			name:    "missing",
			ctx:     newCtxWithValue(consts.CtxRoleID, nil),
			wantErr: true,
		},
		{
			name:    "wrong_type",
			ctx:     newCtxWithValue(consts.CtxRoleID, 42),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserRole(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
