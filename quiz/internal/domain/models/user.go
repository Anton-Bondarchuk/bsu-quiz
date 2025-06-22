package models

// NOTE: review and compare with telegram/internal/domain/models/user.go

const (
	RoleUser    = 1 // 0001
	RoleAdmin   = 2 // 0010
	RoleTeacher = 4 // 0100
	RoleBlocked = 8 // 1000
)

type User struct {
	ID        int64  `json:"id" db:"id"`
	Login     string `json:"login" db:"login"`
	Password  string `json:"-" db:"password"` // Not in DB schema but needed for auth
	RoleFlags int    `json:"role_flags" db:"role_flags"`
}

// HasRole checks if a user has a specific role
func (u *User) HasRole(role int) bool {
	return (u.RoleFlags & role) == role
}

// IsAdmin checks if a user has admin privileges
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// IsTeacher checks if a user has teacher privileges
func (u *User) IsTeacher() bool {
	return u.HasRole(RoleTeacher)
}

// IsBlocked checks if a user is blocked
func (u *User) IsBlocked() bool {
	return u.HasRole(RoleBlocked)
}