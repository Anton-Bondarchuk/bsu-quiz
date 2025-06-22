package services

import (
	"fmt"
	"bsu-quiz/telegram/internal/domain/models"
)

const (
	RoleUser    int = 1 << iota // 1 (0001) - Regular User
	RoleAdmin                   // 2 (0010) - Administrator
	RoleTeacher                 // 4 (0100) - Teacher
	RoleBlocked                 // 8 (1000) - Blocked User
)

// Role names for display purposes
var RoleNames = map[int]string{
	RoleUser:    "User",
	RoleAdmin:   "Admin",
	RoleTeacher: "Teacher",
	RoleBlocked: "Blocked",
}

type Auth struct {
	user *models.User
}

func New(user *models.User) *Auth {
	return &Auth{
		user: user}
}

// HasRole checks if user has a specific role
func (a *Auth) HasRole(role int) bool {
	return int(a.user.Role)&role == role
}

// HasAnyRole checks if user has any of the specified roles
func (a *Auth) HasAnyRole(roles ...int) bool {
	for _, role := range roles {
		if int(a.user.Role)&role != 0 {
			return true
		}
	}
	return false
}

// HasAllRoles checks if user has all of the specified roles
func (a *Auth) HasAllRoles(roles ...int) bool {
	combinedRole := 0
	for _, role := range roles {
		combinedRole |= role
	}
	return int(a.user.Role)&combinedRole == combinedRole
}

// AddRole adds a role to the user
func (a *Auth) AddRole(role int) {
	a.user.Role |= int64(role)
}

// RemoveRole removes a role from the user
func (a *Auth) RemoveRole(role int) {
	a.user.Role &= ^int64(role)
}

// GetRoleNames returns a list of role names that the user has
func (a *Auth) GetRoleNames() []string {
	var roles []string

	// Check each role bit
	for role, name := range RoleNames {
		if a.HasRole(role) {
			roles = append(roles, name)
		}
	}

	return roles
}

// IsAdmin is a convenience method to check if user is an admin
func (a *Auth) IsAdmin() bool {
	return a.HasRole(RoleAdmin)
}

// IsTeacher is a convenience method to check if user is a teacher
func (a *Auth) IsTeacher() bool {
	return a.HasRole(RoleTeacher)
}

// IsBlocked is a convenience method to check if user is blocked
func (a *Auth) IsBlocked() bool {
	return a.HasRole(RoleBlocked)
}

// CanCreateQuiz checks if the user has permission to create quizzes
func (a *Auth) CanCreateQuiz() bool {
	// Blocked users can't create quizzes
	if a.IsBlocked() {
		return false
	}

	// Admins and teachers can create quizzes
	return a.HasAnyRole(RoleAdmin, RoleTeacher)
}

// CanManageUsers checks if the user has permission to manage other users
func (a *Auth) CanManageUsers() bool {
	return a.IsAdmin()
}

// String returns a string representation of the user's roles
func (a *Auth) String() string {
	return fmt.Sprintf("User[%d, %s, Roles: %v]", a.user.ID, a.user.Login, a.GetRoleNames())
}
