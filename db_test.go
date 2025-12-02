package valet

import (
	"testing"
)

func TestExistsRule(t *testing.T) {
	rule := ExistsRule{
		Table:  "users",
		Column: "id",
		Where: []WhereClause{
			{Column: "status", Operator: "=", Value: "active"},
		},
	}

	if rule.Table != "users" {
		t.Errorf("Table = %s, want users", rule.Table)
	}
	if rule.Column != "id" {
		t.Errorf("Column = %s, want id", rule.Column)
	}
	if len(rule.Where) != 1 {
		t.Errorf("Where length = %d, want 1", len(rule.Where))
	}
}

func TestUniqueRule(t *testing.T) {
	rule := UniqueRule{
		Table:  "users",
		Column: "email",
		Ignore: 123,
		Where: []WhereClause{
			{Column: "deleted", Operator: "=", Value: false},
		},
	}

	if rule.Table != "users" {
		t.Errorf("Table = %s, want users", rule.Table)
	}
	if rule.Column != "email" {
		t.Errorf("Column = %s, want email", rule.Column)
	}
	if rule.Ignore != 123 {
		t.Errorf("Ignore = %v, want 123", rule.Ignore)
	}
}

func TestDBCheck(t *testing.T) {
	check := DBCheck{
		Field: "user_id",
		Value: 1,
		Rule: ExistsRule{
			Table:  "users",
			Column: "id",
		},
		IsUnique: false,
		Message:  "User not found",
	}

	if check.Field != "user_id" {
		t.Errorf("Field = %s, want user_id", check.Field)
	}
	if check.Value != 1 {
		t.Errorf("Value = %v, want 1", check.Value)
	}
	if check.IsUnique {
		t.Error("IsUnique should be false")
	}
	if check.Message != "User not found" {
		t.Errorf("Message = %s, want 'User not found'", check.Message)
	}
}

func TestDBCheck_Unique(t *testing.T) {
	check := DBCheck{
		Field: "email",
		Value: "test@example.com",
		Rule: ExistsRule{
			Table:  "users",
			Column: "email",
		},
		IsUnique: true,
		Ignore:   123,
		Message:  "Email already taken",
	}

	if !check.IsUnique {
		t.Error("IsUnique should be true")
	}
	if check.Ignore != 123 {
		t.Errorf("Ignore = %v, want 123", check.Ignore)
	}
}
