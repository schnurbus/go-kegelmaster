package permission

import (
	"context"
	"errors"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
)

// Checker provides methods to check permissions.
type Checker struct {
	clubRepo   *club.Repository
	roleRepo   *role.Repository
	playerRepo *player.Repository
}

// NewChecker creates a new permission checker.
func NewChecker(clubRepo *club.Repository, roleRepo *role.Repository, playerRepo *player.Repository) *Checker {
	return &Checker{
		clubRepo:   clubRepo,
		roleRepo:   roleRepo,
		playerRepo: playerRepo,
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
// Other users get permissions from their player's role.
// Users without a player for the club have no permissions.
func (c *Checker) HasPermission(ctx context.Context, userID, clubID string, entityType role.EntityType, permissionType role.PermissionType) (bool, error) {
	// Club owners always have all permissions
	isOwner, err := c.IsClubOwner(ctx, userID, clubID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	// Check if user has a player for this club
	playerEntity, err := c.playerRepo.GetByUserIDAndClubID(ctx, userID, clubID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			// User has no player for this club, so no permissions
			return false, nil
		}
		return false, err
	}

	// If player has no role, no permissions
	if playerEntity.RoleID == nil {
		return false, nil
	}

	// Check if the role has the requested permission
	return c.roleRepo.HasPermission(ctx, *playerEntity.RoleID, entityType, permissionType)
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

