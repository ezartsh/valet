package valet

// ExistsRule defines a database existence check
type ExistsRule struct {
	Table   string
	Column  string
	Where   []WhereClause
	Message string
}

// UniqueRule defines a database uniqueness check
type UniqueRule struct {
	Table   string
	Column  string
	Ignore  any // Value to ignore (for updates)
	Where   []WhereClause
	Message string
}

// DBCheck represents a pending database check
type DBCheck struct {
	Field    string
	Value    any
	Rule     ExistsRule
	IsUnique bool
	Ignore   any
	Message  MessageArg
}

// DBCheckCollector interface for validators that can collect DB checks
type DBCheckCollector interface {
	GetDBChecks(fieldPath string, value any) []DBCheck
}
