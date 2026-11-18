// Package statistics computes per-player statistics (attendance, penalties,
// competitions, finances) for a selectable period. Only published game days
// (is_draft = false) with at least one attended participant are taken into
// account; finances are based on the ledger and include every game day.
package statistics

import (
	"math"
	"sort"
	"time"
)

// DateLayout is the date format used in the statistics API.
const DateLayout = "2006-01-02"

// ==================== INPUT ====================

// Input is the raw data needed to compute statistics for one player.
// The repository loads it; Compute is a pure function over it.
type Input struct {
	PlayerID           string
	PlayerName         string
	PlayerGender       string // "male", "female" or ""
	PlayerStartBalance int    // cents
	From               *time.Time
	To                 time.Time

	GameDays           []GameDayRef         // published game days of the club in the period (days without attendees are skipped)
	Participants       []Participation      // all participations of all players in the period
	Fees               []Fee                // all fees of all participants in the period
	CompetitionValues  []CompetitionValue   // all competition values in the period
	Competitions       []CompetitionInfo    // competitions of the club, in display order
	BalanceDeltaBefore int                  // sum of balance-relevant transactions before From
	Transactions       []BalanceTransaction // player's balance-relevant transactions in the period, ordered by date
}

type GameDayRef struct {
	ID   string
	Date time.Time
}

// Participation is one player on one game day. Attended is false for a player
// who was only booked for penalties (e.g. a forgotten item) without being there:
// their fees count, but not their attendance or competition values.
type Participation struct {
	ParticipantID string
	GameDayID     string
	PlayerID      string
	Attended      bool
	Gender        string
}

type Fee struct {
	ParticipantID   string
	PenaltyTypeID   string
	PenaltyTypeName string
	Price           int // cents per unit (snapshot)
	Count           int
	QuantityScale   int // 1 = integer, 100 = two decimal places
}

// Amount returns the fee amount in cents (positive). Mirrors transaction.FeeAmount.
func (f Fee) Amount() int {
	if f.Count <= 0 {
		return 0
	}
	return f.Price * f.Count / f.scale()
}

// Quantity returns the displayed quantity (count / scale).
func (f Fee) Quantity() float64 {
	return float64(f.Count) / float64(f.scale())
}

func (f Fee) scale() int {
	if f.QuantityScale < 1 {
		return 1
	}
	return f.QuantityScale
}

type CompetitionValue struct {
	ParticipantID string
	CompetitionID string
	Value         int
}

type CompetitionInfo struct {
	ID               string
	Name             string
	ScoringType      string // winner, loser, both
	IsGenderSpecific bool
}

type BalanceTransaction struct {
	Date   time.Time
	Type   string // base_fee, fee, deposit
	Amount int    // cents, signed
}

// ==================== OUTPUT ====================

type PlayerStatistics struct {
	Player       PlayerRef               `json:"player"`
	Period       Period                  `json:"period"`
	Attendance   Attendance              `json:"attendance"`
	Penalties    Penalties               `json:"penalties"`
	Competitions []CompetitionStatistics `json:"competitions"`
	Finance      Finance                 `json:"finance"`
	GameDays     []GameDayEntry          `json:"game_days"`
}

type PlayerRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Period struct {
	From *string `json:"from"`
	To   string  `json:"to"`
}

type Attendance struct {
	ClubGameDays  int     `json:"club_game_days"`
	Attended      int     `json:"attended"`
	Rate          float64 `json:"rate"` // 0..1
	CurrentStreak int     `json:"current_streak"`
	LongestStreak int     `json:"longest_streak"`
	FirstAttended *string `json:"first_attended"`
	LastAttended  *string `json:"last_attended"`
}

type Penalties struct {
	TotalAmount           int                  `json:"total_amount"`              // cents, including penalties booked while absent
	AbsentAmount          int                  `json:"absent_amount"`             // cents, part of TotalAmount booked while absent
	AveragePerGameDay     int                  `json:"average_per_game_day"`      // cents, TotalAmount per attended game day
	ClubAveragePerGameDay int                  `json:"club_average_per_game_day"` // cents, all club penalties per attended participation
	PenaltyFreeGameDays   int                  `json:"penalty_free_game_days"`
	MostExpensiveGameDay  *GameDayAmount       `json:"most_expensive_game_day"`
	CheapestGameDay       *GameDayAmount       `json:"cheapest_game_day"`
	Rank                  int                  `json:"rank"`    // 1 = highest average penalty amount per game day; 0 = not ranked
	RankOf                int                  `json:"rank_of"` // number of ranked players
	ByType                []PenaltyTypeSummary `json:"by_type"`
}

type GameDayAmount struct {
	GameDayID string `json:"game_day_id"`
	Date      string `json:"date"`
	Amount    int    `json:"amount"`
}

type PenaltyTypeSummary struct {
	PenaltyTypeID         string  `json:"penalty_type_id"`
	Name                  string  `json:"name"`
	Quantity              float64 `json:"quantity"`
	Amount                int     `json:"amount"`    // cents
	GameDays              int     `json:"game_days"` // game days with this penalty (attended or not)
	AveragePerGameDay     float64 `json:"average_per_game_day"`
	ClubAveragePerGameDay float64 `json:"club_average_per_game_day"`
}

type CompetitionStatistics struct {
	CompetitionID    string               `json:"competition_id"`
	Name             string               `json:"name"`
	ScoringType      string               `json:"scoring_type"`
	IsGenderSpecific bool                 `json:"is_gender_specific"`
	Participations   int                  `json:"participations"`
	Total            int                  `json:"total"`
	Average          float64              `json:"average"`
	ClubAverage      float64              `json:"club_average"`
	Best             *int                 `json:"best"`
	Worst            *int                 `json:"worst"`
	Wins             int                  `json:"wins"`
	Losses           int                  `json:"losses"`
	Rank             int                  `json:"rank"` // 1 = highest average; 0 = not ranked
	RankOf           int                  `json:"rank_of"`
	History          []CompetitionHistory `json:"history"`
}

type CompetitionHistory struct {
	GameDayID   string  `json:"game_day_id"`
	Date        string  `json:"date"`
	Value       int     `json:"value"`
	ClubAverage float64 `json:"club_average"`
	IsWin       bool    `json:"is_win"`
	IsLoss      bool    `json:"is_loss"`
}

type Finance struct {
	StartBalance   int            `json:"start_balance"`
	EndBalance     int            `json:"end_balance"`
	Deposits       int            `json:"deposits"`  // cents, positive
	Fees           int            `json:"fees"`      // cents, positive (charged penalties)
	BaseFees       int            `json:"base_fees"` // cents, positive
	BalanceHistory []BalancePoint `json:"balance_history"`
}

type BalancePoint struct {
	Date    string `json:"date"`
	Balance int    `json:"balance"`
}

type GameDayEntry struct {
	GameDayID         string `json:"game_day_id"`
	Date              string `json:"date"`
	Attended          bool   `json:"attended"`
	Participants      int    `json:"participants"`   // attended participants
	PenaltyAmount     *int   `json:"penalty_amount"` // nil if not booked on that day; set for absent days with penalties
	ClubAverageAmount int    `json:"club_average_amount"`
}

// ==================== COMPUTE ====================

// Compute derives all statistics from the given input.
func Compute(in Input) PlayerStatistics {
	out := PlayerStatistics{
		Player:       PlayerRef{ID: in.PlayerID, Name: in.PlayerName},
		Period:       Period{To: in.To.Format(DateLayout)},
		Competitions: []CompetitionStatistics{},
		GameDays:     []GameDayEntry{},
	}
	if in.From != nil {
		f := in.From.Format(DateLayout)
		out.Period.From = &f
	}

	// A game day without any attended participant did not take place; it may
	// exist only to charge the base fee. It is left out entirely, including
	// penalties booked on it, so it does not show up as an empty evening.
	held := make(map[string]bool)
	for _, p := range in.Participants {
		if p.Attended {
			held[p.GameDayID] = true
		}
	}
	gameDays := make([]GameDayRef, 0, len(in.GameDays))
	for _, gd := range in.GameDays {
		if held[gd.ID] {
			gameDays = append(gameDays, gd)
		}
	}
	sort.SliceStable(gameDays, func(i, j int) bool { return gameDays[i].Date.Before(gameDays[j].Date) })
	gameDayByID := make(map[string]GameDayRef, len(gameDays))
	for _, gd := range gameDays {
		gameDayByID[gd.ID] = gd
	}

	// participants holds every booking (attended or not) so fees can be
	// attributed; attendance-based figures only use attended ones.
	participants := make(map[string]Participation, len(in.Participants))
	participantsPerDay := make(map[string]int)    // attended participants per game day
	myParticipantByDay := make(map[string]string) // game day id -> participant id (attended only)
	myAbsentByDay := make(map[string]string)      // game day id -> participant id (absent only)
	for _, p := range in.Participants {
		if _, ok := gameDayByID[p.GameDayID]; !ok {
			continue
		}
		participants[p.ParticipantID] = p
		if p.Attended {
			participantsPerDay[p.GameDayID]++
		}
		if p.PlayerID == in.PlayerID {
			if p.Attended {
				myParticipantByDay[p.GameDayID] = p.ParticipantID
			} else {
				myAbsentByDay[p.GameDayID] = p.ParticipantID
			}
		}
	}

	// Penalty amounts per participant (all players).
	amountByParticipant := make(map[string]int)
	for _, f := range in.Fees {
		if _, ok := participants[f.ParticipantID]; !ok {
			continue
		}
		amountByParticipant[f.ParticipantID] += f.Amount()
	}

	out.Attendance = computeAttendance(gameDays, myParticipantByDay)
	out.Penalties = computePenalties(in, gameDays, participants, myParticipantByDay, myAbsentByDay, amountByParticipant)
	out.Competitions = computeCompetitions(in, gameDays, gameDayByID, participants, myParticipantByDay)
	out.Finance = computeFinance(in)

	dayTotals := make(map[string]int)
	for pid, p := range participants {
		dayTotals[p.GameDayID] += amountByParticipant[pid]
	}
	for _, gd := range gameDays {
		entry := GameDayEntry{
			GameDayID:    gd.ID,
			Date:         gd.Date.Format(DateLayout),
			Participants: participantsPerDay[gd.ID],
		}
		if entry.Participants > 0 {
			entry.ClubAverageAmount = roundDiv(dayTotals[gd.ID], entry.Participants)
		}
		if pid, ok := myParticipantByDay[gd.ID]; ok {
			entry.Attended = true
			amount := amountByParticipant[pid]
			entry.PenaltyAmount = &amount
		} else if pid, ok := myAbsentByDay[gd.ID]; ok {
			if amount := amountByParticipant[pid]; amount > 0 {
				entry.PenaltyAmount = &amount
			}
		}
		out.GameDays = append(out.GameDays, entry)
	}

	return out
}

func computeAttendance(gameDays []GameDayRef, mine map[string]string) Attendance {
	a := Attendance{ClubGameDays: len(gameDays)}
	streak := 0
	for _, gd := range gameDays {
		if _, ok := mine[gd.ID]; ok {
			a.Attended++
			streak++
			if streak > a.LongestStreak {
				a.LongestStreak = streak
			}
			d := gd.Date.Format(DateLayout)
			if a.FirstAttended == nil {
				a.FirstAttended = &d
			}
			a.LastAttended = &d
		} else {
			streak = 0
		}
	}
	a.CurrentStreak = streak
	if a.ClubGameDays > 0 {
		a.Rate = round2(float64(a.Attended) / float64(a.ClubGameDays))
	}
	return a
}

func computePenalties(
	in Input,
	gameDays []GameDayRef,
	participants map[string]Participation,
	mine map[string]string,
	myAbsent map[string]string,
	amountByParticipant map[string]int,
) Penalties {
	p := Penalties{ByType: []PenaltyTypeSummary{}}

	// Own totals per game day.
	for _, gd := range gameDays {
		pid, ok := mine[gd.ID]
		if !ok {
			continue
		}
		amount := amountByParticipant[pid]
		p.TotalAmount += amount
		if amount == 0 {
			p.PenaltyFreeGameDays++
		}
		ga := &GameDayAmount{GameDayID: gd.ID, Date: gd.Date.Format(DateLayout), Amount: amount}
		if p.MostExpensiveGameDay == nil || amount > p.MostExpensiveGameDay.Amount {
			p.MostExpensiveGameDay = ga
		}
		if p.CheapestGameDay == nil || amount < p.CheapestGameDay.Amount {
			p.CheapestGameDay = ga
		}
	}
	// Penalties booked while absent count towards the totals, but the day
	// itself does not (no penalty-free day, no most expensive/cheapest day).
	for _, pid := range myAbsent {
		p.AbsentAmount += amountByParticipant[pid]
	}
	p.TotalAmount += p.AbsentAmount
	attended := len(mine)
	if attended > 0 {
		p.AveragePerGameDay = roundDiv(p.TotalAmount, attended)
	}

	// Club average per attended participation and ranking by average per
	// attended game day. Penalties of absent players are included in the sums.
	totalParticipations := 0
	clubTotal := 0
	perPlayerTotal := make(map[string]int)
	perPlayerDays := make(map[string]int)
	for pid, part := range participants {
		clubTotal += amountByParticipant[pid]
		perPlayerTotal[part.PlayerID] += amountByParticipant[pid]
		if part.Attended {
			totalParticipations++
			perPlayerDays[part.PlayerID]++
		}
	}
	if totalParticipations > 0 {
		p.ClubAveragePerGameDay = roundDiv(clubTotal, totalParticipations)
	}
	if attended > 0 {
		averages := make(map[string]float64, len(perPlayerDays))
		for playerID, days := range perPlayerDays {
			averages[playerID] = float64(perPlayerTotal[playerID]) / float64(days)
		}
		p.Rank, p.RankOf = rankDesc(averages, in.PlayerID)
	}

	// Per penalty type.
	type typeAgg struct {
		summary PenaltyTypeSummary
		days    map[string]struct{}
		clubQty float64
		hasOwn  bool
	}
	types := make(map[string]*typeAgg)
	myParticipants := make(map[string]struct{}, len(mine)+len(myAbsent))
	for _, pid := range mine {
		myParticipants[pid] = struct{}{}
	}
	for _, pid := range myAbsent {
		myParticipants[pid] = struct{}{}
	}
	for _, f := range in.Fees {
		part, ok := participants[f.ParticipantID]
		if !ok {
			continue
		}
		agg, ok := types[f.PenaltyTypeID]
		if !ok {
			agg = &typeAgg{
				summary: PenaltyTypeSummary{PenaltyTypeID: f.PenaltyTypeID, Name: f.PenaltyTypeName},
				days:    make(map[string]struct{}),
			}
			types[f.PenaltyTypeID] = agg
		}
		agg.clubQty += f.Quantity()
		if _, own := myParticipants[f.ParticipantID]; own {
			agg.hasOwn = true
			agg.summary.Name = f.PenaltyTypeName
			agg.summary.Quantity += f.Quantity()
			agg.summary.Amount += f.Amount()
			agg.days[part.GameDayID] = struct{}{}
		}
	}
	for _, agg := range types {
		if !agg.hasOwn {
			continue
		}
		s := agg.summary
		s.Quantity = round2(s.Quantity)
		s.GameDays = len(agg.days)
		if attended > 0 {
			s.AveragePerGameDay = round2(agg.summary.Quantity / float64(attended))
		}
		if totalParticipations > 0 {
			s.ClubAveragePerGameDay = round2(agg.clubQty / float64(totalParticipations))
		}
		p.ByType = append(p.ByType, s)
	}
	sort.Slice(p.ByType, func(i, j int) bool {
		if p.ByType[i].Amount != p.ByType[j].Amount {
			return p.ByType[i].Amount > p.ByType[j].Amount
		}
		return p.ByType[i].Name < p.ByType[j].Name
	})
	return p
}

func computeCompetitions(
	in Input,
	gameDays []GameDayRef,
	gameDayByID map[string]GameDayRef,
	participants map[string]Participation,
	mine map[string]string,
) []CompetitionStatistics {
	type entry struct {
		playerID string
		gender   string
		value    int
	}
	// competition id -> game day id -> entries
	values := make(map[string]map[string][]entry)
	for _, cv := range in.CompetitionValues {
		part, ok := participants[cv.ParticipantID]
		if !ok || !part.Attended {
			continue
		}
		byDay, ok := values[cv.CompetitionID]
		if !ok {
			byDay = make(map[string][]entry)
			values[cv.CompetitionID] = byDay
		}
		byDay[part.GameDayID] = append(byDay[part.GameDayID], entry{playerID: part.PlayerID, gender: part.Gender, value: cv.Value})
	}

	result := make([]CompetitionStatistics, 0, len(in.Competitions))
	for _, comp := range in.Competitions {
		cs := CompetitionStatistics{
			CompetitionID:    comp.ID,
			Name:             comp.Name,
			ScoringType:      comp.ScoringType,
			IsGenderSpecific: comp.IsGenderSpecific,
			History:          []CompetitionHistory{},
		}
		byDay := values[comp.ID]
		if len(byDay) == 0 {
			continue
		}
		inGroup := func(e entry) bool {
			return !comp.IsGenderSpecific || e.gender == in.PlayerGender
		}
		countsWins := comp.ScoringType == "winner" || comp.ScoringType == "both"
		countsLosses := comp.ScoringType == "loser" || comp.ScoringType == "both"

		clubSum, clubCount := 0, 0
		playerSum := make(map[string]int)
		playerCount := make(map[string]int)
		for _, gd := range gameDays {
			entries := byDay[gd.ID]
			group := make([]entry, 0, len(entries))
			for _, e := range entries {
				if inGroup(e) {
					group = append(group, e)
				}
			}
			if len(group) == 0 {
				continue
			}
			minV, maxV, daySum := group[0].value, group[0].value, 0
			var own *entry
			for i, e := range group {
				daySum += e.value
				clubSum += e.value
				clubCount++
				playerSum[e.playerID] += e.value
				playerCount[e.playerID]++
				minV = min(minV, e.value)
				maxV = max(maxV, e.value)
				if e.playerID == in.PlayerID {
					own = &group[i]
				}
			}
			if own == nil {
				continue
			}
			if _, attended := mine[gd.ID]; !attended {
				continue
			}
			h := CompetitionHistory{
				GameDayID:   gd.ID,
				Date:        gameDayByID[gd.ID].Date.Format(DateLayout),
				Value:       own.value,
				ClubAverage: round2(float64(daySum) / float64(len(group))),
			}
			decisive := len(group) >= 2 && minV != maxV
			if decisive && countsWins && own.value == maxV {
				h.IsWin = true
				cs.Wins++
			}
			if decisive && countsLosses && own.value == minV {
				h.IsLoss = true
				cs.Losses++
			}
			cs.Participations++
			cs.Total += own.value
			if cs.Best == nil || own.value > *cs.Best {
				v := own.value
				cs.Best = &v
			}
			if cs.Worst == nil || own.value < *cs.Worst {
				v := own.value
				cs.Worst = &v
			}
			cs.History = append(cs.History, h)
		}
		if clubCount > 0 {
			cs.ClubAverage = round2(float64(clubSum) / float64(clubCount))
		}
		if cs.Participations > 0 {
			cs.Average = round2(float64(cs.Total) / float64(cs.Participations))
			averages := make(map[string]float64, len(playerCount))
			for playerID, n := range playerCount {
				averages[playerID] = float64(playerSum[playerID]) / float64(n)
			}
			cs.Rank, cs.RankOf = rankDesc(averages, in.PlayerID)
		}
		result = append(result, cs)
	}
	return result
}

func computeFinance(in Input) Finance {
	f := Finance{BalanceHistory: []BalancePoint{}}
	f.StartBalance = in.PlayerStartBalance + in.BalanceDeltaBefore
	balance := f.StartBalance
	if in.From != nil {
		f.BalanceHistory = append(f.BalanceHistory, BalancePoint{Date: in.From.Format(DateLayout), Balance: balance})
	}
	for _, t := range in.Transactions {
		switch t.Type {
		case "deposit":
			f.Deposits += t.Amount
		case "fee":
			f.Fees -= t.Amount
		case "base_fee":
			f.BaseFees -= t.Amount
		}
		balance += t.Amount
		d := t.Date.Format(DateLayout)
		if n := len(f.BalanceHistory); n > 0 && f.BalanceHistory[n-1].Date == d {
			f.BalanceHistory[n-1].Balance = balance
		} else {
			f.BalanceHistory = append(f.BalanceHistory, BalancePoint{Date: d, Balance: balance})
		}
	}
	f.EndBalance = balance
	if n := len(f.BalanceHistory); n > 0 {
		if to := in.To.Format(DateLayout); f.BalanceHistory[n-1].Date < to {
			f.BalanceHistory = append(f.BalanceHistory, BalancePoint{Date: to, Balance: balance})
		}
	}
	return f
}

// rankDesc returns the 1-based rank of playerID by descending value (ties share
// the better rank) and the number of ranked players. Returns 0, n if not ranked.
func rankDesc(values map[string]float64, playerID string) (int, int) {
	own, ok := values[playerID]
	if !ok {
		return 0, len(values)
	}
	rank := 1
	for id, v := range values {
		if id != playerID && v > own+1e-9 {
			rank++
		}
	}
	return rank, len(values)
}

func roundDiv(a, b int) int {
	if b == 0 {
		return 0
	}
	return int(math.Round(float64(a) / float64(b)))
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
