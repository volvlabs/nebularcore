package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Role int

const (
	Nil Role = iota
	Admin
	User
)

var RoleToString = map[Role]string{
	Nil:   "Nil",
	Admin: "Admin",
	User:  "User",
}

var StringToRole = map[string]Role{
	"Nil":   Nil,
	"Admin": Admin,
	"User":  User,
}

// MarshalJSON converts the Role to its corresponding string representation for JSON serialization.
func (r Role) MarshalJSON() ([]byte, error) {
	str, ok := RoleToString[r]
	if !ok {
		return nil, fmt.Errorf("invalid role: %d", r)
	}
	return json.Marshal(str)
}

// UnmarshalJSON converts a JSON string to the corresponding Role value.
func (r *Role) UnmarshalJSON(data []byte) error {
	var roleStr string
	if err := json.Unmarshal(data, &roleStr); err != nil {
		return err
	}
	role, ok := StringToRole[roleStr]
	if !ok {
		return errors.New("invalid role")
	}
	*r = role
	return nil
}

func RoleFromString(roleStr string) (string, error) {
	// Convert the string to an integer
	roleInt, err := strconv.Atoi(roleStr)
	if err != nil {
		return "", fmt.Errorf("invalid role string: %v", err)
	}

	// Convert the integer to the Role type
	role := Role(roleInt)

	// Find the corresponding string for the Role
	roleName, ok := RoleToString[role]
	if !ok {
		return "", fmt.Errorf("role not found for value: %d", roleInt)
	}

	return roleName, nil
}
