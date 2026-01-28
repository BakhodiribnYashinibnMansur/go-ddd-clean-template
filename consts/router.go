package consts

// Path Parameters
const (
	ParamUserID     string = "user_id"
	ParamID         string = "id"
	ParamSessionID  string = "session_id"
	ParamRoleID     string = "role_id"
	ParamPermID     string = "perm_id"
	ParamPolicyID   string = "policy_id"
	ParamRelationID string = "relation_id"
)

// Query Parameters
const (
	QueryName     string = "name"
	QueryPath     string = "path"
	QueryMethod   string = "method"
	QueryPhone    string = "phone"
	QueryLimit    string = "limit"
	QueryOffset   string = "offset"
	QueryPage     string = "page"
	QueryPageSize string = "pageSize"

	// Mocking
	QueryMock      string = "mock"
	QueryMockDelay string = "mock_delay"
	QueryMockError string = "mock_error"
	QueryMockEmpty string = "mock_empty"

	// Filtering
	QueryUserID       string = "user_id"
	QueryAction       string = "action"
	QueryResourceType string = "resource_type"
	QueryResourceID   string = "resource_id"
	QuerySuccess      string = "success"
	QueryFromDate     string = "from_date"
	QueryToDate       string = "to_date"
)
