package admin

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	DefaultPageLimit = 25
	MaxPageLimit     = 100
	maxSearchLength  = 200

	UserFilterOrphaned = "orphaned"
	UserFilterAdmin    = "admin"
	ClubFilterInactive = "inactive"

	OrderAsc  = "asc"
	OrderDesc = "desc"
)

// ValidationError is returned for invalid list query parameters (HTTP 400).
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string { return e.Message }

// ListParams are the raw query parameters of a list endpoint.
type ListParams struct {
	Search string
	Filter string
	Sort   string
	Order  string
	Page   string
	Limit  string
}

// UserListQuery is a normalized query for ListUsers.
type UserListQuery struct {
	Search string
	Filter string // "", UserFilterOrphaned, UserFilterAdmin
	Sort   string // email, created_at, last_login_at
	Order  string // asc, desc
	Page   int
	Limit  int
}

// ClubListQuery is a normalized query for ListClubs.
type ClubListQuery struct {
	Search string
	Filter string // "", ClubFilterInactive
	Sort   string // name, created_at, last_activity_at
	Order  string // asc, desc
	Page   int
	Limit  int
}

// ParseUserListQuery validates and normalizes the user list parameters.
// Defaults: sort=created_at, order=desc, page=1, limit=25.
func ParseUserListQuery(p ListParams) (UserListQuery, error) {
	var q UserListQuery
	var err error
	if q.Search, err = parseSearch(p.Search); err != nil {
		return q, err
	}
	if q.Filter, err = parseEnum("filter", p.Filter, "", UserFilterOrphaned, UserFilterAdmin); err != nil {
		return q, err
	}
	if q.Sort, err = parseEnum("sort", p.Sort, "created_at", "email", "created_at", "last_login_at"); err != nil {
		return q, err
	}
	defaultOrder := OrderDesc
	if q.Sort == "email" {
		defaultOrder = OrderAsc
	}
	if q.Order, err = parseEnum("order", p.Order, defaultOrder, OrderAsc, OrderDesc); err != nil {
		return q, err
	}
	q.Page, q.Limit, err = parsePaging(p.Page, p.Limit)
	return q, err
}

// ParseClubListQuery validates and normalizes the club list parameters.
// Defaults: sort=name, order=asc, page=1, limit=25.
func ParseClubListQuery(p ListParams) (ClubListQuery, error) {
	var q ClubListQuery
	var err error
	if q.Search, err = parseSearch(p.Search); err != nil {
		return q, err
	}
	if q.Filter, err = parseEnum("filter", p.Filter, "", ClubFilterInactive); err != nil {
		return q, err
	}
	if q.Sort, err = parseEnum("sort", p.Sort, "name", "name", "created_at", "last_activity_at"); err != nil {
		return q, err
	}
	defaultOrder := OrderDesc
	if q.Sort == "name" {
		defaultOrder = OrderAsc
	}
	if q.Order, err = parseEnum("order", p.Order, defaultOrder, OrderAsc, OrderDesc); err != nil {
		return q, err
	}
	q.Page, q.Limit, err = parsePaging(p.Page, p.Limit)
	return q, err
}

func parseSearch(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) > maxSearchLength {
		return "", ValidationError{Message: fmt.Sprintf("Suchbegriff darf höchstens %d Zeichen haben", maxSearchLength)}
	}
	return s, nil
}

func parseEnum(name, value, fallback string, allowed ...string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback, nil
	}
	for _, a := range allowed {
		if value == a {
			return value, nil
		}
	}
	return "", ValidationError{Message: fmt.Sprintf("Ungültiger Wert für %s", name)}
}

func parsePaging(pageStr, limitStr string) (int, int, error) {
	page, limit := 1, DefaultPageLimit
	if s := strings.TrimSpace(pageStr); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 {
			return 0, 0, ValidationError{Message: "Ungültiger Wert für page"}
		}
		page = v
	}
	if s := strings.TrimSpace(limitStr); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 || v > MaxPageLimit {
			return 0, 0, ValidationError{Message: fmt.Sprintf("limit muss zwischen 1 und %d liegen", MaxPageLimit)}
		}
		limit = v
	}
	// Keeps the OFFSET within int32.
	if page > 1_000_000 {
		return 0, 0, ValidationError{Message: "Ungültiger Wert für page"}
	}
	return page, limit, nil
}
