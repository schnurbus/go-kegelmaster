package handlers

import (
	"time"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/competition"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
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
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Balance            int       `json:"balance"`
	StartBalance       int       `json:"start_balance"`
	BaseFee            int       `json:"base_fee"`
	AutoTipEnabled     bool      `json:"auto_tip_enabled"`
	CouplesModeEnabled bool      `json:"couples_mode_enabled"`
	UserID             string    `json:"user_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func ClubResponseFromEntity(c club.Club) ClubResponse {
	return ClubResponse{
		ID:                 c.ID,
		Name:               c.Name,
		Balance:            c.Balance,
		StartBalance:       c.StartBalance,
		BaseFee:            c.BaseFee,
		AutoTipEnabled:     c.AutoTipEnabled,
		CouplesModeEnabled: c.CouplesModeEnabled,
		UserID:             c.UserID,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
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
	Gender       *string   `json:"gender,omitempty"`
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
		Gender:       p.Gender,
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
	ID                    string    `json:"id"`
	ClubID                string    `json:"club_id"`
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	Price                 int       `json:"price"`
	DisplayOrder          int       `json:"display_order"`
	AllowsDecimalQuantity bool      `json:"allows_decimal_quantity"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	ReplacedByID          *string   `json:"replaced_by_id,omitempty"`
}

func PenaltyTypeResponseFromEntity(p penaltytype.PenaltyType) PenaltyTypeResponse {
	return PenaltyTypeResponse{
		ID:                    p.ID,
		ClubID:                p.ClubID,
		Name:                  p.Name,
		Description:           p.Description,
		Price:                 p.Price,
		DisplayOrder:          p.DisplayOrder,
		AllowsDecimalQuantity: p.AllowsDecimalQuantity,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
		ReplacedByID:          p.ReplacedByID,
	}
}

func PenaltyTypesResponseFromEntities(penaltyTypes []penaltytype.PenaltyType) []PenaltyTypeResponse {
	result := make([]PenaltyTypeResponse, len(penaltyTypes))
	for i, p := range penaltyTypes {
		result[i] = PenaltyTypeResponseFromEntity(p)
	}
	return result
}

// Competition responses
type CompetitionResponse struct {
	ID               string    `json:"id"`
	ClubID           string    `json:"club_id"`
	Name             string    `json:"name"`
	ScoringType      string    `json:"scoring_type"`
	IsGenderSpecific bool      `json:"is_gender_specific"`
	DisplayOrder     int       `json:"display_order"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func CompetitionResponseFromEntity(c competition.Competition) CompetitionResponse {
	return CompetitionResponse{
		ID:               c.ID,
		ClubID:           c.ClubID,
		Name:             c.Name,
		ScoringType:      c.ScoringType,
		IsGenderSpecific: c.IsGenderSpecific,
		DisplayOrder:     c.DisplayOrder,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}

func CompetitionsResponseFromEntities(competitions []competition.Competition) []CompetitionResponse {
	result := make([]CompetitionResponse, len(competitions))
	for i, c := range competitions {
		result[i] = CompetitionResponseFromEntity(c)
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

type GameDaySummaryResponse struct {
	ID               string `json:"id"`
	ClubID           string `json:"club_id"`
	Date             string `json:"date"`
	Notes            string `json:"notes"`
	ParticipantCount int    `json:"participant_count"`
	PenaltyFeeTotal  int    `json:"penalty_fee_total"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

func GameDaySummaryResponseFromEntity(gds gameday.GameDaySummary) GameDaySummaryResponse {
	return GameDaySummaryResponse{
		ID:               gds.ID,
		ClubID:           gds.ClubID,
		Date:             gds.Date.Format("2006-01-02"),
		Notes:            gds.Notes,
		ParticipantCount: gds.ParticipantCount,
		PenaltyFeeTotal:  gds.PenaltyFeeTotal,
		CreatedAt:        gds.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        gds.UpdatedAt.Format(time.RFC3339),
	}
}

func GameDaySummariesResponseFromEntities(summaries []gameday.GameDaySummary) []GameDaySummaryResponse {
	responses := make([]GameDaySummaryResponse, len(summaries))
	for i, s := range summaries {
		responses[i] = GameDaySummaryResponseFromEntity(s)
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
	ID                     string  `json:"id"`
	GameDayParticipantID   string  `json:"game_day_participant_id"`
	PenaltyTypeID          string  `json:"penalty_type_id"`
	PenaltyTypeName        string  `json:"penalty_type_name"`
	PenaltyTypeDescription string  `json:"penalty_type_description"`
	PenaltyTypePrice       int     `json:"penalty_type_price"`
	Count                  int     `json:"count"`          // stored value (scaled when quantity_scale > 1)
	QuantityScale          int     `json:"quantity_scale"` // 1 or 100
	Quantity               float64 `json:"quantity"`       // display: count / quantity_scale
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

func FeeResponseFromEntity(f gameday.GameDayFee) FeeResponse {
	quantity := float64(f.Count)
	if f.QuantityScale > 0 {
		quantity = float64(f.Count) / float64(f.QuantityScale)
	}
	return FeeResponse{
		ID:                     f.ID,
		GameDayParticipantID:   f.GameDayParticipantID,
		PenaltyTypeID:          f.PenaltyTypeID,
		PenaltyTypeName:        f.PenaltyTypeName,
		PenaltyTypeDescription: f.PenaltyTypeDescription,
		PenaltyTypePrice:       f.PenaltyTypePrice,
		Count:                  f.Count,
		QuantityScale:          f.QuantityScale,
		Quantity:               quantity,
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

type CompetitionValueResponse struct {
	ID                   string `json:"id"`
	GameDayParticipantID string `json:"game_day_participant_id"`
	CompetitionID        string `json:"competition_id"`
	Value                int    `json:"value"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

func CompetitionValueResponseFromEntity(cv gameday.GameDayCompetitionValue) CompetitionValueResponse {
	return CompetitionValueResponse{
		ID:                   cv.ID,
		GameDayParticipantID: cv.GameDayParticipantID,
		CompetitionID:        cv.CompetitionID,
		Value:                cv.Value,
		CreatedAt:            cv.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            cv.UpdatedAt.Format(time.RFC3339),
	}
}

func CompetitionValuesResponseFromEntities(cvs []gameday.GameDayCompetitionValue) []CompetitionValueResponse {
	responses := make([]CompetitionValueResponse, len(cvs))
	for i, cv := range cvs {
		responses[i] = CompetitionValueResponseFromEntity(cv)
	}
	return responses
}

type ParticipantWithFeesResponse struct {
	Participant       ParticipantResponse        `json:"participant"`
	Fees              []FeeResponse              `json:"fees"`
	CompetitionValues []CompetitionValueResponse `json:"competition_values"`
}

type GameDayDetailResponse struct {
	GameDay      GameDayResponse               `json:"game_day"`
	Participants []ParticipantWithFeesResponse `json:"participants"`
}

func GameDayDetailResponseFromEntity(detail gameday.GameDayDetail) GameDayDetailResponse {
	participants := make([]ParticipantWithFeesResponse, len(detail.Participants))
	for i, pwf := range detail.Participants {
		participants[i] = ParticipantWithFeesResponse{
			Participant:       ParticipantResponseFromEntity(pwf.Participant),
			Fees:              FeesResponseFromEntities(pwf.Fees),
			CompetitionValues: CompetitionValuesResponseFromEntities(pwf.CompetitionValues),
		}
	}

	return GameDayDetailResponse{
		GameDay:      GameDayResponseFromEntity(detail.GameDay),
		Participants: participants,
	}
}

// Transaction responses
type TransactionResponse struct {
	ID                  string  `json:"id"`
	ClubID              string  `json:"club_id"`
	PlayerID            *string `json:"player_id"`
	PlayerName          string  `json:"player_name,omitempty"`
	TransactionType     string  `json:"transaction_type"`
	Amount              int     `json:"amount"`
	Description         string  `json:"description"`
	GameDayFeeID        *string `json:"game_day_fee_id,omitempty"`
	GameDayID           *string `json:"game_day_id,omitempty"`
	PlayerBalanceBefore *int    `json:"player_balance_before,omitempty"`
	PlayerBalanceAfter  *int    `json:"player_balance_after,omitempty"`
	ClubBalanceBefore   int     `json:"club_balance_before"`
	ClubBalanceAfter    int     `json:"club_balance_after"`
	TransactionDate     string  `json:"transaction_date"` // effective date (game day date for fee/base_fee)
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

func TransactionResponseFromEntity(t transaction.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:                  t.ID,
		ClubID:              t.ClubID,
		PlayerID:            t.PlayerID,
		PlayerName:          t.PlayerName,
		TransactionType:     string(t.TransactionType),
		Amount:              t.Amount,
		Description:         t.Description,
		GameDayFeeID:        t.GameDayFeeID,
		GameDayID:           t.GameDayID,
		PlayerBalanceBefore: t.PlayerBalanceBefore,
		PlayerBalanceAfter:  t.PlayerBalanceAfter,
		ClubBalanceBefore:   t.ClubBalanceBefore,
		ClubBalanceAfter:    t.ClubBalanceAfter,
		TransactionDate:     t.TransactionDate.Format("2006-01-02"),
		CreatedAt:           t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           t.UpdatedAt.Format(time.RFC3339),
	}
}

func TransactionsResponseFromEntities(transactions []transaction.Transaction) []TransactionResponse {
	responses := make([]TransactionResponse, len(transactions))
	for i, t := range transactions {
		responses[i] = TransactionResponseFromEntity(t)
	}
	return responses
}

type PaginatedTransactionsResponse struct {
	Data       []TransactionResponse `json:"data"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"total_pages"`
}

func PaginatedTransactionsResponseFromEntity(p *transaction.PaginatedTransactions) PaginatedTransactionsResponse {
	return PaginatedTransactionsResponse{
		Data:       TransactionsResponseFromEntities(p.Data),
		Page:       p.Page,
		Limit:      p.Limit,
		Total:      p.Total,
		TotalPages: p.TotalPages,
	}
}

type GameDayTransactionSummaryResponse struct {
	BaseFeeTotal    int   `json:"base_fee_total"`
	PenaltyFeeTotal int   `json:"penalty_fee_total"`
	BaseFeeCount    int64 `json:"base_fee_count"`
	PenaltyFeeCount int64 `json:"penalty_fee_count"`
	Total           int   `json:"total"`
}

func GameDayTransactionSummaryResponseFromEntity(s *transaction.GameDayTransactionSummary) GameDayTransactionSummaryResponse {
	return GameDayTransactionSummaryResponse{
		BaseFeeTotal:    s.BaseFeeTotal,
		PenaltyFeeTotal: s.FeeTotal,
		BaseFeeCount:    s.BaseFeeCount,
		PenaltyFeeCount: s.FeeCount,
		Total:           s.BaseFeeTotal + s.FeeTotal,
	}
}
