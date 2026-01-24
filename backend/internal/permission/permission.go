package permission

import (
	"context"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
)

// Checker provides methods to check permissions.
type Checker struct {
	clubRepo *club.Repository
	roleRepo *role.Repository
}

// NewChecker creates a new permission checker.
func NewChecker(clubRepo *club.Repository, roleRepo *role.Repository) *Checker {
	return &Checker{
		clubRepo: clubRepo,
		roleRepo: roleRepo,
	}
}

// IsClubOwner checks if a user is the owner of a club.
func (c *Checker) IsClubOwner(ctx context.Context, userID, clubID string) (bool, error) {
	clubEntity, err := c.clubRepo.GetByID(ctx, clubID)
	if err != nil {
		return false, err
	}
	return clubEntity.UserID == userID, nil
}

// HasPermission checks if a user has a specific permission for a club.
// Club owners always have all permissions.
func (c *Checker) HasPermission(ctx context.Context, userID, clubID string, entityType role.EntityType, permissionType role.PermissionType) (bool, error) {
	// Club owners always have all permissions
	isOwner, err := c.IsClubOwner(ctx, userID, clubID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	// TODO: Check permissions via player-role assignment when Player entity is implemented
	// For now, only owners have permissions
	return false, nil
}

// HasAnyPermission checks if a user has any of the specified permissions for a club.
func (c *Checker) HasAnyPermission(ctx context.Context, userID, clubID string, entityType role.EntityType, permissionTypes []role.PermissionType) (bool, error) {
	for _, permType := range permissionTypes {
		has, err := c.HasPermission(ctx, userID, clubID, entityType, permType)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}

