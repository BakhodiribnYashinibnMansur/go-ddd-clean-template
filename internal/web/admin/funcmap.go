package admin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// Shared Template Functions
var templateFuncs = template.FuncMap{
	"currUserEmail": func(u *domain.User) string {
		if u != nil {
			if u.Username != nil && *u.Username != "" {
				return *u.Username
			}
			if u.Email != nil && *u.Email != "" {
				return *u.Email
			}
			if u.Phone != nil && *u.Phone != "" {
				return *u.Phone
			}
		}
		return "Admin"
	},
	"derefString": func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	},
	"derefBool": func(b *bool) bool {
		if b == nil {
			return false
		}
		return *b
	},
	"derefDeviceType": func(d *domain.SessionDeviceType) string {
		if d == nil {
			return ""
		}
		return string(*d)
	},
	"add": func(a, b any) int64 {
		return toInt64(a) + toInt64(b)
	},
	"sub": func(a, b any) int64 {
		return toInt64(a) - toInt64(b)
	},
	"formatUUID": func(id any) string {
		if id == nil {
			return ""
		}
		switch v := id.(type) {
		case uuid.UUID:
			return v.String()
		case string:
			return v
		case []byte:
			if len(v) == 16 {
				u, err := uuid.FromBytes(v)
				if err == nil {
					return u.String()
				}
			}
			return string(v)
		default:
			return fmt.Sprintf("%v", v)
		}
	},
	"formatTime": func(t any) string {
		if t == nil {
			return ""
		}
		switch v := t.(type) {
		case time.Time:
			return v.Format("02 Jan 2006, 15:04")
		case *time.Time:
			if v != nil {
				return v.Format("02 Jan 2006, 15:04")
			}
		}
		return ""
	},
	"seq": func(start, end int) []int {
		var res []int
		for i := start; i <= end; i++ {
			res = append(res, i)
		}
		return res
	},
	"totalPages": func(total, limit int64) int64 {
		if limit == 0 {
			return 1
		}
		return int64(math.Ceil(float64(total) / float64(limit)))
	},
	"currPage": func(offset, limit int64) int64 {
		if limit == 0 {
			return 1
		}
		return (offset / limit) + 1
	},
	"toJSON": func(v interface{}) template.JS {
		b, err := json.Marshal(v)
		if err != nil {
			return template.JS("{}")
		}
		return template.JS(b)
	},
	"paginationLink": func(currURL url.Values, page int64, limit int64) string {
		currURL.Set("page", strconv.FormatInt(page, 10))
		currURL.Set("limit", strconv.FormatInt(limit, 10))
		return "?" + currURL.Encode()
	},
	"default": func(d string, v any) any {
		if v == nil {
			return d
		}
		if s, ok := v.(*string); ok {
			if s == nil || *s == "" {
				return d
			}
			return *s
		}
		if s, ok := v.(string); ok {
			if s == "" {
				return d
			}
			return s
		}
		return v
	},
	"gt": func(a, b any) bool {
		return toInt64(a) > toInt64(b)
	},
	"lt": func(a, b any) bool {
		return toInt64(a) < toInt64(b)
	},
	"ge": func(a, b any) bool {
		return toInt64(a) >= toInt64(b)
	},
	"le": func(a, b any) bool {
		return toInt64(a) <= toInt64(b)
	},
	"eq": func(a, b any) bool {
		return toInt64(a) == toInt64(b)
	},
	"ne": func(a, b any) bool {
		return toInt64(a) != toInt64(b)
	},
	"contains": func(s any, substr string) bool {
		var str string
		switch v := s.(type) {
		case string:
			str = v
		case *string:
			if v != nil {
				str = *v
			} else {
				return false
			}
		default:
			return false
		}
		return strings.Contains(str, substr)
	},
}

func toInt64(v any) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	default:
		return 0
	}
}
