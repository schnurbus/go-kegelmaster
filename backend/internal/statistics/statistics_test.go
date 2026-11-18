package statistics

import (
	"reflect"
	"testing"
	"time"
)

func d(s string) time.Time {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

// fixture: 4 published game days, 3 players (me = "p1", male; "p2", male; "p3", female).
// p1 attends gd1, gd2, gd4 (misses gd3).
func fixture() Input {
	from := d("2026-01-01")
	return Input{
		PlayerID:           "p1",
		PlayerName:         "Ich",
		PlayerGender:       "male",
		PlayerStartBalance: 1000,
		From:               &from,
		To:                 d("2026-12-31"),
		GameDays: []GameDayRef{
			{ID: "gd2", Date: d("2026-02-01")},
			{ID: "gd1", Date: d("2026-01-10")},
			{ID: "gd3", Date: d("2026-03-01")},
			{ID: "gd4", Date: d("2026-04-01")},
		},
		Participants: []Participation{
			{ParticipantID: "a1", GameDayID: "gd1", PlayerID: "p1", Gender: "male"},
			{ParticipantID: "a2", GameDayID: "gd1", PlayerID: "p2", Gender: "male"},
			{ParticipantID: "a3", GameDayID: "gd1", PlayerID: "p3", Gender: "female"},
			{ParticipantID: "b1", GameDayID: "gd2", PlayerID: "p1", Gender: "male"},
			{ParticipantID: "b2", GameDayID: "gd2", PlayerID: "p2", Gender: "male"},
			{ParticipantID: "c2", GameDayID: "gd3", PlayerID: "p2", Gender: "male"},
			{ParticipantID: "d1", GameDayID: "gd4", PlayerID: "p1", Gender: "male"},
			{ParticipantID: "d3", GameDayID: "gd4", PlayerID: "p3", Gender: "female"},
			// participation on an unknown game day (e.g. draft) must be ignored
			{ParticipantID: "x1", GameDayID: "draft", PlayerID: "p1", Gender: "male"},
		},
		Fees: []Fee{
			{ParticipantID: "a1", PenaltyTypeID: "pt-pudel", PenaltyTypeName: "Pudel", Price: 50, Count: 3, QuantityScale: 1},
			{ParticipantID: "a1", PenaltyTypeID: "pt-bier", PenaltyTypeName: "Bier", Price: 200, Count: 150, QuantityScale: 100},
			{ParticipantID: "a2", PenaltyTypeID: "pt-pudel", PenaltyTypeName: "Pudel", Price: 50, Count: 1, QuantityScale: 1},
			{ParticipantID: "b1", PenaltyTypeID: "pt-pudel", PenaltyTypeName: "Pudel", Price: 50, Count: 2, QuantityScale: 1},
			{ParticipantID: "b2", PenaltyTypeID: "pt-pudel", PenaltyTypeName: "Pudel", Price: 50, Count: 10, QuantityScale: 1},
			{ParticipantID: "c2", PenaltyTypeID: "pt-kranz", PenaltyTypeName: "Kranz", Price: 100, Count: 1, QuantityScale: 1},
			{ParticipantID: "x1", PenaltyTypeID: "pt-pudel", PenaltyTypeName: "Pudel", Price: 50, Count: 99, QuantityScale: 1},
		},
		Competitions: []CompetitionInfo{
			{ID: "c-pokal", Name: "Pokal", ScoringType: "both", IsGenderSpecific: false},
			{ID: "c-koenig", Name: "König", ScoringType: "winner", IsGenderSpecific: true},
			{ID: "c-leer", Name: "Leer", ScoringType: "winner"},
		},
		CompetitionValues: []CompetitionValue{
			// Pokal gd1: p1 30, p2 20, p3 40 -> p1 neither
			{ParticipantID: "a1", CompetitionID: "c-pokal", Value: 30},
			{ParticipantID: "a2", CompetitionID: "c-pokal", Value: 20},
			{ParticipantID: "a3", CompetitionID: "c-pokal", Value: 40},
			// Pokal gd2: p1 10, p2 50 -> p1 loss
			{ParticipantID: "b1", CompetitionID: "c-pokal", Value: 10},
			{ParticipantID: "b2", CompetitionID: "c-pokal", Value: 50},
			// Pokal gd4: p1 60, p3 5 -> p1 win
			{ParticipantID: "d1", CompetitionID: "c-pokal", Value: 60},
			{ParticipantID: "d3", CompetitionID: "c-pokal", Value: 5},
			// König (gender specific) gd1: p1 7, p2 5, p3 9 -> among men p1 wins
			{ParticipantID: "a1", CompetitionID: "c-koenig", Value: 7},
			{ParticipantID: "a2", CompetitionID: "c-koenig", Value: 5},
			{ParticipantID: "a3", CompetitionID: "c-koenig", Value: 9},
		},
		BalanceDeltaBefore: -200,
		Transactions: []BalanceTransaction{
			{Date: d("2026-01-10"), Type: "base_fee", Amount: -500},
			{Date: d("2026-01-10"), Type: "fee", Amount: -450},
			{Date: d("2026-01-15"), Type: "deposit", Amount: 2000},
			{Date: d("2026-02-01"), Type: "fee", Amount: -100},
		},
	}
}

func TestCompute_Attendance(t *testing.T) {
	s := Compute(fixture())

	a := s.Attendance
	eq(t, 4, a.ClubGameDays)
	eq(t, 3, a.Attended)
	eq(t, 0.75, a.Rate)
	eq(t, 2, a.LongestStreak)
	eq(t, 1, a.CurrentStreak)
	requireTrue(t, a.FirstAttended != nil)
	requireTrue(t, a.LastAttended != nil)
	eq(t, "2026-01-10", *a.FirstAttended)
	eq(t, "2026-04-01", *a.LastAttended)

	requireTrue(t, s.Period.From != nil)
	eq(t, "2026-01-01", *s.Period.From)
	eq(t, "2026-12-31", s.Period.To)
	eq(t, PlayerRef{ID: "p1", Name: "Ich"}, s.Player)
}

func TestCompute_Penalties(t *testing.T) {
	s := Compute(fixture())
	p := s.Penalties

	// p1: gd1 = 150 + 300 = 450, gd2 = 100, gd4 = 0
	eq(t, 550, p.TotalAmount)
	eq(t, 183, p.AveragePerGameDay)
	// club: 450+50+0 + 100+500 + 100 + 0+0 = 1200 over 8 participations
	eq(t, 150, p.ClubAveragePerGameDay)
	eq(t, 1, p.PenaltyFreeGameDays)
	requireTrue(t, p.MostExpensiveGameDay != nil)
	eq(t, "gd1", p.MostExpensiveGameDay.GameDayID)
	eq(t, 450, p.MostExpensiveGameDay.Amount)
	requireTrue(t, p.CheapestGameDay != nil)
	eq(t, "gd4", p.CheapestGameDay.GameDayID)

	// averages: p1 183.3, p2 650/3 = 216.7, p3 0 -> p1 rank 2 of 3
	eq(t, 2, p.Rank)
	eq(t, 3, p.RankOf)

	requireLen(t, len(p.ByType), 2)
	pudel := p.ByType[1]
	eq(t, "Pudel", pudel.Name)
	eq(t, 5.0, pudel.Quantity)
	eq(t, 250, pudel.Amount)
	eq(t, 2, pudel.GameDays)
	eq(t, 1.67, pudel.AveragePerGameDay)
	eq(t, 2.0, pudel.ClubAveragePerGameDay) // 16 / 8

	bier := p.ByType[0]
	eq(t, "Bier", bier.Name)
	eq(t, 1.5, bier.Quantity)
	eq(t, 300, bier.Amount)
}

func TestCompute_Competitions(t *testing.T) {
	s := Compute(fixture())
	requireLen(t, len(s.Competitions), 2)

	pokal := s.Competitions[0]
	eq(t, "Pokal", pokal.Name)
	eq(t, 3, pokal.Participations)
	eq(t, 100, pokal.Total)
	eq(t, 33.33, pokal.Average)
	requireTrue(t, pokal.Best != nil)
	requireTrue(t, pokal.Worst != nil)
	eq(t, 60, *pokal.Best)
	eq(t, 10, *pokal.Worst)
	eq(t, 1, pokal.Wins)
	eq(t, 1, pokal.Losses)
	// averages: p1 33.3, p2 35, p3 22.5 -> rank 2 of 3
	eq(t, 2, pokal.Rank)
	eq(t, 3, pokal.RankOf)
	eq(t, 30.71, pokal.ClubAverage) // 215 / 7
	requireLen(t, len(pokal.History), 3)
	eq(t, "2026-01-10", pokal.History[0].Date)
	eq(t, 30.0, pokal.History[0].ClubAverage)
	eq(t, true, pokal.History[1].IsLoss)
	eq(t, true, pokal.History[2].IsWin)

	koenig := s.Competitions[1]
	eq(t, 1, koenig.Wins)
	eq(t, 0, koenig.Losses)
	eq(t, 1, koenig.Rank)
	eq(t, 2, koenig.RankOf)
	eq(t, 6.0, koenig.ClubAverage)
}

func TestCompute_Finance(t *testing.T) {
	s := Compute(fixture())
	f := s.Finance

	eq(t, 800, f.StartBalance)
	eq(t, 1750, f.EndBalance)
	eq(t, 2000, f.Deposits)
	eq(t, 550, f.Fees)
	eq(t, 500, f.BaseFees)
	eq(t, []BalancePoint{
		{Date: "2026-01-01", Balance: 800},
		{Date: "2026-01-10", Balance: -150},
		{Date: "2026-01-15", Balance: 1850},
		{Date: "2026-02-01", Balance: 1750},
		{Date: "2026-12-31", Balance: 1750},
	}, f.BalanceHistory)
}

func TestCompute_GameDays(t *testing.T) {
	s := Compute(fixture())
	requireLen(t, len(s.GameDays), 4)
	eq(t, "gd1", s.GameDays[0].GameDayID)
	eq(t, true, s.GameDays[0].Attended)
	requireTrue(t, s.GameDays[0].PenaltyAmount != nil)
	eq(t, 450, *s.GameDays[0].PenaltyAmount)
	eq(t, 167, s.GameDays[0].ClubAverageAmount) // 500 / 3
	eq(t, 3, s.GameDays[0].Participants)

	eq(t, false, s.GameDays[2].Attended)
	eq(t, true, s.GameDays[2].PenaltyAmount == nil)
	eq(t, 100, s.GameDays[2].ClubAverageAmount)
}

func TestCompute_Empty(t *testing.T) {
	s := Compute(Input{PlayerID: "p1", To: d("2026-06-30"), PlayerStartBalance: 300})

	eq(t, true, s.Period.From == nil)
	eq(t, 0, s.Attendance.ClubGameDays)
	eq(t, 0.0, s.Attendance.Rate)
	eq(t, 0, s.Penalties.Rank)
	eq(t, true, s.Penalties.MostExpensiveGameDay == nil)
	eq(t, true, s.Penalties.ByType != nil)
	eq(t, true, s.Competitions != nil)
	eq(t, true, s.GameDays != nil)
	eq(t, 300, s.Finance.StartBalance)
	eq(t, 300, s.Finance.EndBalance)
	eq(t, 0, len(s.Finance.BalanceHistory))
}

func TestCompute_NoDecisionOnTieOrSingleParticipant(t *testing.T) {
	in := Input{
		PlayerID: "p1",
		To:       d("2026-06-30"),
		GameDays: []GameDayRef{{ID: "g1", Date: d("2026-01-01")}, {ID: "g2", Date: d("2026-02-01")}},
		Participants: []Participation{
			{ParticipantID: "a", GameDayID: "g1", PlayerID: "p1"},
			{ParticipantID: "b", GameDayID: "g1", PlayerID: "p2"},
			{ParticipantID: "c", GameDayID: "g2", PlayerID: "p1"},
		},
		Competitions: []CompetitionInfo{{ID: "c", Name: "C", ScoringType: "both"}},
		CompetitionValues: []CompetitionValue{
			{ParticipantID: "a", CompetitionID: "c", Value: 5},
			{ParticipantID: "b", CompetitionID: "c", Value: 5},
			{ParticipantID: "c", CompetitionID: "c", Value: 9},
		},
	}
	s := Compute(in)
	requireLen(t, len(s.Competitions), 1)
	eq(t, 0, s.Competitions[0].Wins)
	eq(t, 0, s.Competitions[0].Losses)
	eq(t, 2, s.Competitions[0].Participations)
}

func TestFee_AmountAndQuantity(t *testing.T) {
	eq(t, 0, Fee{Price: 100, Count: 0}.Amount())
	eq(t, 300, Fee{Price: 100, Count: 3}.Amount())
	eq(t, 125, Fee{Price: 50, Count: 250, QuantityScale: 100}.Amount())
	eq(t, 2.5, Fee{Count: 250, QuantityScale: 100}.Quantity())
}

func eq(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want %#v, got %#v", want, got)
	}
}

func requireTrue(t *testing.T, ok bool) {
	t.Helper()
	if !ok {
		t.Fatalf("required condition not met")
	}
}

func requireLen(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("want len %d, got %d", want, got)
	}
}
