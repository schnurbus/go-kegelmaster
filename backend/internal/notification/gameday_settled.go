// Package notification sends e-mails users opted in to, such as the notice
// that a game day of their club was settled.
package notification

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
	"github.com/schnurbus/go-kegelmaster/backend/internal/email"
)

// sendTimeout bounds one notification run (database reads and all e-mails).
const sendTimeout = 2 * time.Minute

// Recipient is a player linked to a user who enabled game day e-mails.
type Recipient struct {
	Email         string
	PlayerName    string
	PlayerBalance int
	GameDayAmount int
}

// Store is the data access the notifier needs.
type Store interface {
	// ClaimGameDaySettled records that the game day was announced. It returns
	// false when it was announced before, e.g. after it was re-published.
	ClaimGameDaySettled(ctx context.Context, gameDayID string) (bool, error)
	ListGameDaySettledRecipients(ctx context.Context, clubID, gameDayID string) ([]Recipient, error)
	ClubName(ctx context.Context, clubID string) (string, error)
}

// Mailer sends the e-mail; *email.Service implements it.
type Mailer interface {
	Enabled() bool
	SendGameDaySettledEmail(to string, mail email.GameDaySettledEmail) error
}

// Notifier announces settled game days by e-mail.
type Notifier struct {
	store  Store
	mailer Mailer
	appURL string
	wg     sync.WaitGroup
}

// NewNotifier creates a notifier. appURL is the link in the e-mail (PUBLIC_BASE_URL).
func NewNotifier(store Store, mailer Mailer, baseURL string) *Notifier {
	appURL := ""
	if base := strings.TrimRight(strings.TrimSpace(baseURL), "/"); base != "" {
		appURL = base + "/app"
	}
	return &Notifier{store: store, mailer: mailer, appURL: appURL}
}

// GameDaySettled sends the e-mails in the background. Call it only after the
// game day was published and its bookings were committed; errors are logged and
// never reach the caller.
func (n *Notifier) GameDaySettled(clubID, gameDayID string, date time.Time) {
	if n == nil {
		return
	}
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				slog.Error("game day settled notification panicked", "gameday_id", gameDayID, "panic", r)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
		defer cancel()
		n.NotifyGameDaySettled(ctx, clubID, gameDayID, date)
	}()
}

// Wait blocks until all background notifications have finished or ctx ends.
// It reports whether everything finished.
func (n *Notifier) Wait(ctx context.Context) bool {
	done := make(chan struct{})
	go func() {
		n.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

// NotifyGameDaySettled sends one e-mail per opted-in user synchronously and
// returns the number of e-mails sent.
func (n *Notifier) NotifyGameDaySettled(ctx context.Context, clubID, gameDayID string, date time.Time) int {
	if n.mailer == nil || !n.mailer.Enabled() {
		slog.Info("e-mail not configured, skipping game day settled notification", "club_id", clubID, "gameday_id", gameDayID)
		return 0
	}

	claimed, err := n.store.ClaimGameDaySettled(ctx, gameDayID)
	if err != nil {
		slog.Error("claim game day settled notification", "error", err, "gameday_id", gameDayID)
		return 0
	}
	if !claimed {
		slog.Info("game day settled notification already sent", "gameday_id", gameDayID)
		return 0
	}

	recipients, err := n.store.ListGameDaySettledRecipients(ctx, clubID, gameDayID)
	if err != nil {
		slog.Error("list game day settled recipients", "error", err, "gameday_id", gameDayID)
		return 0
	}
	if len(recipients) == 0 {
		return 0
	}
	clubName, err := n.store.ClubName(ctx, clubID)
	if err != nil {
		slog.Error("load club for game day settled notification", "error", err, "club_id", clubID)
		return 0
	}

	sent := 0
	for _, mail := range groupByEmail(recipients) {
		if ctx.Err() != nil {
			slog.Error("game day settled notification timed out", "gameday_id", gameDayID, "sent", sent)
			break
		}
		mail.content.ClubName = clubName
		mail.content.Date = date
		mail.content.AppURL = n.appURL
		if err := n.mailer.SendGameDaySettledEmail(mail.to, mail.content); err != nil {
			slog.Error("send game day settled email", "error", err, "email", mail.to, "gameday_id", gameDayID)
			continue
		}
		sent++
	}
	slog.Info("game day settled notifications sent", "gameday_id", gameDayID, "sent", sent)
	return sent
}

type groupedMail struct {
	to      string
	content email.GameDaySettledEmail
}

// groupByEmail sends a user with several players in the club a single e-mail.
// The order of first appearance is kept.
func groupByEmail(recipients []Recipient) []groupedMail {
	index := make(map[string]int, len(recipients))
	mails := make([]groupedMail, 0, len(recipients))
	for _, r := range recipients {
		key := strings.ToLower(strings.TrimSpace(r.Email))
		if key == "" {
			continue
		}
		i, ok := index[key]
		if !ok {
			i = len(mails)
			index[key] = i
			mails = append(mails, groupedMail{to: r.Email})
		}
		mails[i].content.Players = append(mails[i].content.Players, email.GameDaySettledPlayer{
			Name:          r.PlayerName,
			GameDayAmount: r.GameDayAmount,
			Balance:       r.PlayerBalance,
		})
	}
	return mails
}

// Repository implements Store with the sqlc queries.
type Repository struct {
	queries *db.Queries
}

// NewRepository creates the database-backed Store.
func NewRepository(conn *sql.DB) *Repository {
	return &Repository{queries: db.New(conn)}
}

func (r *Repository) ClaimGameDaySettled(ctx context.Context, gameDayID string) (bool, error) {
	rows, err := r.queries.ClaimGameDaySettledNotification(ctx, db.ClaimGameDaySettledNotificationParams{
		ID:        uuid.NewString(),
		GameDayID: gameDayID,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *Repository) ListGameDaySettledRecipients(ctx context.Context, clubID, gameDayID string) ([]Recipient, error) {
	id := gameDayID
	rows, err := r.queries.ListGameDaySettledRecipients(ctx, db.ListGameDaySettledRecipientsParams{
		GameDayID: &id,
		ClubID:    clubID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]Recipient, 0, len(rows))
	for _, row := range rows {
		out = append(out, Recipient{
			Email:         row.Email,
			PlayerName:    row.PlayerName,
			PlayerBalance: int(row.PlayerBalance),
			GameDayAmount: int(row.GameDayAmount),
		})
	}
	return out, nil
}

func (r *Repository) ClubName(ctx context.Context, clubID string) (string, error) {
	club, err := r.queries.GetClubByID(ctx, clubID)
	if err != nil {
		return "", err
	}
	return club.Name, nil
}
