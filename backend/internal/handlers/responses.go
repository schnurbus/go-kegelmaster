package handlers

import (
	"time"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func UserResponseFromEntity(u user.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type ClubResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Balance   int       `json:"balance"`
	BaseFee   int       `json:"base_fee"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ClubResponseFromEntity(c club.Club) ClubResponse {
	return ClubResponse{
		ID:        c.ID,
		Name:      c.Name,
		Balance:   c.Balance,
		BaseFee:   c.BaseFee,
		UserID:    c.UserID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ClubsResponseFromEntities(clubs []club.Club) []ClubResponse {
	result := make([]ClubResponse, len(clubs))
	for i, c := range clubs {
		result[i] = ClubResponseFromEntity(c)
	}
	return result
}

type PermissionResponse struct {
	ID             string    `json:"id"`
	RoleID         string    `json:"role_id"`
	EntityType     string    `json:"entity_type"`
	PermissionType string    `json:"permission_type"`
	CreatedAt      time.Time `json:"created_at"`
}

type RoleResponse struct {
	ID          string               `json:"id"`
	ClubID      string               `json:"club_id"`
	Name        string               `json:"name"`
	PaysBaseFee bool                 `json:"pays_base_fee"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Permissions []PermissionResponse `json:"permissions"`
}

func PermissionResponseFromEntity(p role.RolePermission) PermissionResponse {
	return PermissionResponse{
		ID:             p.ID,
		RoleID:         p.RoleID,
		EntityType:     string(p.EntityType),
		PermissionType: string(p.PermissionType),
		CreatedAt:      p.CreatedAt,
	}
}

func RoleResponseFromEntity(r role.Role, perms []role.RolePermission) RoleResponse {
	permResponses := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		permResponses[i] = PermissionResponseFromEntity(p)
	}

	return RoleResponse{
		ID:          r.ID,
		ClubID:      r.ClubID,
		Name:        r.Name,
		PaysBaseFee: r.PaysBaseFee,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		Permissions: permResponses,
	}
}

type PlayerResponse struct {
	ID           string    `json:"id"`
	ClubID       string    `json:"club_id"`
	UserID       *string   `json:"user_id"`
	RoleID       *string   `json:"role_id"`
	Name         string    `json:"name"`
	Balance      int       `json:"balance"`
	StartBalance int       `json:"start_balance"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func PlayerResponseFromEntity(p player.Player) PlayerResponse {
	return PlayerResponse{
		ID:           p.ID,
		ClubID:       p.ClubID,
		UserID:       p.UserID,
		RoleID:       p.RoleID,
		Name:         p.Name,
		Balance:      p.Balance,
		StartBalance: p.StartBalance,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func PlayersResponseFromEntities(players []player.Player) []PlayerResponse {
	result := make([]PlayerResponse, len(players))
	for i, p := range players {
		result[i] = PlayerResponseFromEntity(p)
	}
	return result
}

type PenaltyTypeResponse struct {
	ID           string    `json:"id"`
	ClubID       string    `json:"club_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        int       `json:"price"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ReplacedByID *string   `json:"replaced_by_id,omitempty"`
}

func PenaltyTypeResponseFromEntity(p penaltytype.PenaltyType) PenaltyTypeResponse {
	return PenaltyTypeResponse{
		ID:           p.ID,
		ClubID:       p.ClubID,
		Name:         p.Name,
		Description:  p.Description,
		Price:        p.Price,
		DisplayOrder: p.DisplayOrder,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		ReplacedByID: p.ReplacedByID,
	}
}

func PenaltyTypesResponseFromEntities(penaltyTypes []penaltytype.PenaltyType) []PenaltyTypeResponse {
	result := make([]PenaltyTypeResponse, len(penaltyTypes))
	for i, p := range penaltyTypes {
		result[i] = PenaltyTypeResponseFromEntity(p)
	}
	return result
}

// GameDay responses
type GameDayResponse struct {
	ID        string `json:"id"`
	ClubID    string `json:"club_id"`
	Date      string `json:"date"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func GameDayResponseFromEntity(gd gameday.GameDay) GameDayResponse {
	return GameDayResponse{
		ID:        gd.ID,
		ClubID:    gd.ClubID,
		Date:      gd.Date.Format("2006-01-02"),
		Notes:     gd.Notes,
		CreatedAt: gd.CreatedAt.Format(time.RFC3339),
		UpdatedAt: gd.UpdatedAt.Format(time.RFC3339),
	}
}

func GameDaysResponseFromEntities(gameDays []gameday.GameDay) []GameDayResponse {
	responses := make([]GameDayResponse, len(gameDays))
	for i, gd := range gameDays {
		responses[i] = GameDayResponseFromEntity(gd)
	}
	return responses
}

type ParticipantResponse struct {
	ID         string `json:"id"`
	GameDayID  string `json:"game_day_id"`
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name,omitempty"`
	CreatedAt  string `json:"created_at"`
}

func ParticipantResponseFromEntity(p gameday.GameDayParticipant) ParticipantResponse {
	return ParticipantResponse{
		ID:         p.ID,
		GameDayID:  p.GameDayID,
		PlayerID:   p.PlayerID,
		PlayerName: p.PlayerName,
		CreatedAt:  p.CreatedAt.Format(time.RFC3339),
	}
}

type FeeResponse struct {
	ID                     string `json:"id"`
	GameDayParticipantID   string `json:"game_day_participant_id"`
	PenaltyTypeID          string `json:"penalty_type_id"`
	PenaltyTypeName        string `json:"penalty_type_name"`
	PenaltyTypeDescription string `json:"penalty_type_description"`
	PenaltyTypePrice       int    `json:"penalty_type_price"`
	Count                  int    `json:"count"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

func FeeResponseFromEntity(f gameday.GameDayFee) FeeResponse {
	return FeeResponse{
		ID:                     f.ID,
		GameDayParticipantID:   f.GameDayParticipantID,
		PenaltyTypeID:          f.PenaltyTypeID,
		PenaltyTypeName:        f.PenaltyTypeName,
		PenaltyTypeDescription: f.PenaltyTypeDescription,
		PenaltyTypePrice:       f.PenaltyTypePrice,
		Count:                  f.Count,
		CreatedAt:              f.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              f.UpdatedAt.Format(time.RFC3339),
	}
}

func FeesResponseFromEntities(fees []gameday.GameDayFee) []FeeResponse {
	responses := make([]FeeResponse, len(fees))
	for i, f := range fees {
		responses[i] = FeeResponseFromEntity(f)
	}
	return responses
}

type ParticipantWithFeesResponse struct {
	Participant ParticipantResponse `json:"participant"`
	Fees        []FeeResponse       `json:"fees"`
}

type GameDayDetailResponse struct {
	GameDay      GameDayResponse               `json:"game_day"`
	Participants []ParticipantWithFeesResponse `json:"participants"`
}

func GameDayDetailResponseFromEntity(detail gameday.GameDayDetail) GameDayDetailResponse {
	participants := make([]ParticipantWithFeesResponse, len(detail.Participants))
	for i, pwf := range detail.Participants {
		participants[i] = ParticipantWithFeesResponse{
			Participant: ParticipantResponseFromEntity(pwf.Participant),
			Fees:        FeesResponseFromEntities(pwf.Fees),
		}
	}

	return GameDayDetailResponse{
		GameDay:      GameDayResponseFromEntity(detail.GameDay),
		Participants: participants,
	}
}
