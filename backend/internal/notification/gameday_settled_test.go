package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/schnurbus/go-kegelmaster/backend/internal/email"
)

type fakeStore struct {
	claimed    map[string]bool
	claimErr   error
	recipients []Recipient
	listCalls  int
}

func (s *fakeStore) ClaimGameDaySettled(_ context.Context, gameDayID string) (bool, error) {
	if s.claimErr != nil {
		return false, s.claimErr
	}
	if s.claimed == nil {
		s.claimed = map[string]bool{}
	}
	if s.claimed[gameDayID] {
		return false, nil
	}
	s.claimed[gameDayID] = true
	return true, nil
}

func (s *fakeStore) ListGameDaySettledRecipients(context.Context, string, string) ([]Recipient, error) {
	s.listCalls++
	return s.recipients, nil
}

func (s *fakeStore) ClubName(context.Context, string) (string, error) {
	return "KC Alle Neune", nil
}

type sentMail struct {
	to   string
	mail email.GameDaySettledEmail
}

type fakeMailer struct {
	enabled bool
	failFor string
	sent    []sentMail
}

func (m *fakeMailer) Enabled() bool { return m.enabled }

func (m *fakeMailer) SendGameDaySettledEmail(to string, mail email.GameDaySettledEmail) error {
	if to == m.failFor {
		return errors.New("boom")
	}
	m.sent = append(m.sent, sentMail{to: to, mail: mail})
	return nil
}

var day = time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

func TestNotifyGameDaySettled_SendsOneMailPerUser(t *testing.T) {
	store := &fakeStore{recipients: []Recipient{
		{Email: "anna@example.com", PlayerName: "Anna", PlayerBalance: -1250, GameDayAmount: -750},
		{Email: "bernd@example.com", PlayerName: "Bernd", PlayerBalance: 300, GameDayAmount: -500},
		{Email: "Anna@example.com", PlayerName: "Anna (2)", PlayerBalance: 0, GameDayAmount: 0},
	}}
	mailer := &fakeMailer{enabled: true}
	n := NewNotifier(store, mailer, "https://kegel.example.org/")

	if sent := n.NotifyGameDaySettled(context.Background(), "club", "day-1", day); sent != 2 {
		t.Fatalf("sent = %d, want 2", sent)
	}
	if len(mailer.sent) != 2 {
		t.Fatalf("mails = %+v", mailer.sent)
	}
	anna := mailer.sent[0]
	if anna.to != "anna@example.com" || len(anna.mail.Players) != 2 {
		t.Fatalf("anna mail = %+v", anna)
	}
	if anna.mail.ClubName != "KC Alle Neune" || !anna.mail.Date.Equal(day) || anna.mail.AppURL != "https://kegel.example.org/app" {
		t.Fatalf("anna mail content = %+v", anna.mail)
	}
	if p := anna.mail.Players[0]; p.Name != "Anna" || p.Balance != -1250 || p.GameDayAmount != -750 {
		t.Fatalf("anna player = %+v", p)
	}
	if mailer.sent[1].to != "bernd@example.com" {
		t.Fatalf("second mail = %+v", mailer.sent[1])
	}
}

func TestNotifyGameDaySettled_OnlyOncePerGameDay(t *testing.T) {
	store := &fakeStore{recipients: []Recipient{{Email: "anna@example.com", PlayerName: "Anna"}}}
	mailer := &fakeMailer{enabled: true}
	n := NewNotifier(store, mailer, "")

	n.NotifyGameDaySettled(context.Background(), "club", "day-1", day)
	// Unpublished and published again: no second e-mail.
	if sent := n.NotifyGameDaySettled(context.Background(), "club", "day-1", day); sent != 0 {
		t.Fatalf("second run sent %d mails", sent)
	}
	if len(mailer.sent) != 1 || store.listCalls != 1 {
		t.Fatalf("mails = %d, list calls = %d", len(mailer.sent), store.listCalls)
	}
}

func TestNotifyGameDaySettled_SkipsWhenMailNotConfigured(t *testing.T) {
	store := &fakeStore{recipients: []Recipient{{Email: "anna@example.com", PlayerName: "Anna"}}}
	n := NewNotifier(store, &fakeMailer{enabled: false}, "")

	if sent := n.NotifyGameDaySettled(context.Background(), "club", "day-1", day); sent != 0 {
		t.Fatalf("sent = %d", sent)
	}
	if store.claimed["day-1"] {
		t.Fatal("game day must not be marked as notified without e-mail")
	}
}

func TestNotifyGameDaySettled_ContinuesAfterSendError(t *testing.T) {
	store := &fakeStore{recipients: []Recipient{
		{Email: "anna@example.com", PlayerName: "Anna"},
		{Email: "bernd@example.com", PlayerName: "Bernd"},
	}}
	mailer := &fakeMailer{enabled: true, failFor: "anna@example.com"}
	n := NewNotifier(store, mailer, "")

	if sent := n.NotifyGameDaySettled(context.Background(), "club", "day-1", day); sent != 1 {
		t.Fatalf("sent = %d, want 1", sent)
	}
	if len(mailer.sent) != 1 || mailer.sent[0].to != "bernd@example.com" {
		t.Fatalf("mails = %+v", mailer.sent)
	}
}

func TestNotifyGameDaySettled_ClaimErrorSendsNothing(t *testing.T) {
	store := &fakeStore{claimErr: errors.New("db down"), recipients: []Recipient{{Email: "anna@example.com"}}}
	mailer := &fakeMailer{enabled: true}
	n := NewNotifier(store, mailer, "")

	if sent := n.NotifyGameDaySettled(context.Background(), "club", "day-1", day); sent != 0 || len(mailer.sent) != 0 {
		t.Fatalf("sent = %d, mails = %+v", sent, mailer.sent)
	}
}

func TestGameDaySettled_RunsInBackground(t *testing.T) {
	store := &fakeStore{recipients: []Recipient{{Email: "anna@example.com", PlayerName: "Anna"}}}
	mailer := &fakeMailer{enabled: true}
	n := NewNotifier(store, mailer, "")

	n.GameDaySettled("club", "day-1", day)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if !n.Wait(ctx) {
		t.Fatal("background notification did not finish")
	}
	if len(mailer.sent) != 1 {
		t.Fatalf("mails = %+v", mailer.sent)
	}
}

func TestNilNotifierIsNoop(t *testing.T) {
	var n *Notifier
	n.GameDaySettled("club", "day-1", day)
}

func TestRepository_ClaimGameDaySettled(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer conn.Close()
	repo := NewRepository(conn)

	mock.ExpectExec(`INSERT INTO game_day_notifications .* ON CONFLICT \(game_day_id\) DO NOTHING`).
		WithArgs(sqlmock.AnyArg(), "day-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO game_day_notifications`).
		WithArgs(sqlmock.AnyArg(), "day-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	first, err := repo.ClaimGameDaySettled(context.Background(), "day-1")
	if err != nil || !first {
		t.Fatalf("first claim = %v, %v", first, err)
	}
	second, err := repo.ClaimGameDaySettled(context.Background(), "day-1")
	if err != nil || second {
		t.Fatalf("second claim = %v, %v", second, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_ListGameDaySettledRecipients(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer conn.Close()
	repo := NewRepository(conn)

	mock.ExpectQuery(`ListGameDaySettledRecipients`).
		WithArgs("day-1", "club-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "email", "player_id", "player_name", "player_balance", "game_day_amount"}).
			AddRow("u1", "anna@example.com", "p1", "Anna", -1250, -750))

	got, err := repo.ListGameDaySettledRecipients(context.Background(), "club-1", "day-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	want := Recipient{Email: "anna@example.com", PlayerName: "Anna", PlayerBalance: -1250, GameDayAmount: -750}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("recipients = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
