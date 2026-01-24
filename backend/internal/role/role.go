package role

import (
	"time"
)

// EntityType represents the type of entity a permission applies to.
type EntityType string

const (
	EntityTypeRoles       EntityType = "roles"
	EntityTypePlayers     EntityType = "players"
	EntityTypeGameNights  EntityType = "game_nights"
	EntityTypePenalties   EntityType = "penalties"
	EntityTypePenaltyTypes EntityType = "penalty_types"
)

// ValidEntityTypes returns a list of all valid entity types.
func ValidEntityTypes() []EntityType {
	return []EntityType{
		EntityTypeRoles,
		EntityTypePlayers,
		EntityTypeGameNights,
		EntityTypePenalties,
		EntityTypePenaltyTypes,
	}
}

// IsValid checks if the entity type is valid.
func (e EntityType) IsValid() bool {
	for _, valid := range ValidEntityTypes() {
		if e == valid {
			return true
		}
	}
	return false
}

// PermissionType represents the type of permission.
type PermissionType string

const (
	PermissionTypeList   PermissionType = "list"
	PermissionTypeView   PermissionType = "view"
	PermissionTypeCreate PermissionType = "create"
	PermissionTypeUpdate PermissionType = "update"
	PermissionTypeDelete PermissionType = "delete"
)

// ValidPermissionTypes returns a list of all valid permission types.
func ValidPermissionTypes() []PermissionType {
	return []PermissionType{
		PermissionTypeList,
		PermissionTypeView,
		PermissionTypeCreate,
		PermissionTypeUpdate,
		PermissionTypeDelete,
	}
}

// IsValid checks if the permission type is valid.
func (p PermissionType) IsValid() bool {
	for _, valid := range ValidPermissionTypes() {
		if p == valid {
			return true
		}
	}
	return false
}

// Role represents a role entity in the system.
type Role struct {
	ID          string
	ClubID      string
	Name        string
	PaysBaseFee bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RolePermission represents a permission assigned to a role.
type RolePermission struct {
	ID             string
	RoleID         string
	EntityType     EntityType
	PermissionType PermissionType
	CreatedAt      time.Time
}

// RoleWithPermissions represents a role with its associated permissions.
type RoleWithPermissions struct {
	Role        Role
	Permissions []RolePermission
}

