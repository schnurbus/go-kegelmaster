package mcpserver

// Minimal mirrors of the REST API's JSON responses (see internal/handlers/responses.go).

type apiClub struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Balance            int    `json:"balance"`
	BaseFee            int    `json:"base_fee"`
	AutoTipEnabled     bool   `json:"auto_tip_enabled"`
	CouplesModeEnabled bool   `json:"couples_mode_enabled"`
}

type apiPermissions struct {
	IsOwner     bool `json:"is_owner"`
	Permissions []struct {
		EntityType     string `json:"entity_type"`
		PermissionType string `json:"permission_type"`
	} `json:"permissions"`
}

type apiPlayer struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Balance     int     `json:"balance"`
	RoleID      *string `json:"role_id"`
	Gender      *string `json:"gender"`
	Inactive    bool    `json:"inactive"`
	PartnerID   *string `json:"partner_id"`
	PairBalance *int    `json:"pair_balance"`
}

type apiRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PaysBaseFee bool   `json:"pays_base_fee"`
}

type apiPenaltyType struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	Price                 int     `json:"price"`
	AllowsDecimalQuantity bool    `json:"allows_decimal_quantity"`
	ReplacedByID          *string `json:"replaced_by_id"`
}

type apiCompetition struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ScoringType      string `json:"scoring_type"`
	IsGenderSpecific bool   `json:"is_gender_specific"`
}

type apiGameDay struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Notes   string `json:"notes"`
	IsDraft bool   `json:"is_draft"`
}

type apiGameDaySummary struct {
	apiGameDay
	ParticipantCount int `json:"participant_count"`
	PenaltyFeeTotal  int `json:"penalty_fee_total"`
}

type apiFee struct {
	PenaltyTypeID    string  `json:"penalty_type_id"`
	PenaltyTypeName  string  `json:"penalty_type_name"`
	PenaltyTypePrice int     `json:"penalty_type_price"`
	Count            int     `json:"count"`
	QuantityScale    int     `json:"quantity_scale"`
	Quantity         float64 `json:"quantity"`
}

type apiGameDayDetail struct {
	GameDay      apiGameDay `json:"game_day"`
	Participants []struct {
		Participant struct {
			PlayerID   string `json:"player_id"`
			PlayerName string `json:"player_name"`
		} `json:"participant"`
		Fees              []apiFee `json:"fees"`
		CompetitionValues []struct {
			CompetitionID string `json:"competition_id"`
			Value         int    `json:"value"`
		} `json:"competition_values"`
	} `json:"participants"`
}

type apiTxSummary struct {
	BaseFeeTotal    int   `json:"base_fee_total"`
	PenaltyFeeTotal int   `json:"penalty_fee_total"`
	BaseFeeCount    int64 `json:"base_fee_count"`
	PenaltyFeeCount int64 `json:"penalty_fee_count"`
	Total           int   `json:"total"`
}

type apiTransaction struct {
	ID                 string  `json:"id"`
	PlayerID           *string `json:"player_id"`
	PlayerName         string  `json:"player_name"`
	TransactionType    string  `json:"transaction_type"`
	Amount             int     `json:"amount"`
	Description        string  `json:"description"`
	GameDayID          *string `json:"game_day_id"`
	PlayerBalanceAfter *int    `json:"player_balance_after"`
	ClubBalanceAfter   int     `json:"club_balance_after"`
	TransactionDate    string  `json:"transaction_date"`
}

type apiTxPage struct {
	Data       []apiTransaction `json:"data"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	Total      int64            `json:"total"`
	TotalPages int              `json:"total_pages"`
}

type apiPenaltyHistoryDay struct {
	GameDayID string `json:"game_day_id"`
	Date      string `json:"date"`
	Penalties []struct {
		PenaltyTypeName string  `json:"penalty_type_name"`
		Quantity        float64 `json:"quantity"`
	} `json:"penalties"`
}

type apiCompetitionHistoryDay struct {
	GameDayID string `json:"game_day_id"`
	Date      string `json:"date"`
	Values    []struct {
		CompetitionName string `json:"competition_name"`
		Value           int    `json:"value"`
	} `json:"values"`
}
