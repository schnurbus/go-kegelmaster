package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"
)

// GameDaySettledPlayer is one of the recipient's players in a club. Amounts are cents.
type GameDaySettledPlayer struct {
	Name string
	// GameDayAmount is the sum of all bookings of the game day (negative = charges).
	GameDayAmount int
	// Balance is the player's balance after the game day was settled.
	Balance int
}

// GameDaySettledEmail holds the content of the "Spieltag abgerechnet" e-mail.
type GameDaySettledEmail struct {
	ClubName string
	Date     time.Time
	Players  []GameDaySettledPlayer
	AppURL   string
}

var gameDaySettledTemplate = template.Must(template.New("gameday_settled").Funcs(template.FuncMap{
	"euro": FormatEuro,
}).Parse(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>{{.Subject}}</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
	<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
		<h1 style="color: #2563eb;">Spieltag abgerechnet</h1>
		<p>Hallo,</p>
		<p>der Spieltag vom <strong>{{.Date}}</strong> im Club <strong>{{.ClubName}}</strong> wurde abgerechnet.</p>
		{{range .Players}}
		<table style="border-collapse: collapse; margin: 16px 0; min-width: 280px;">
			<tr><th colspan="2" style="text-align: left; padding: 4px 0;">{{.Name}}</th></tr>
			<tr><td style="padding: 2px 16px 2px 0;">Buchungen am Spieltag</td><td style="text-align: right;">{{euro .GameDayAmount}}</td></tr>
			<tr><td style="padding: 2px 16px 2px 0;">Aktueller Saldo</td><td style="text-align: right;"><strong>{{euro .Balance}}</strong></td></tr>
		</table>
		{{end}}
		{{if .AppURL}}
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.AppURL}}" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Zum Kegelmaster</a>
		</p>
		{{end}}
		<hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
		<p style="color: #666; font-size: 12px;">Sie erhalten diese E-Mail, weil Sie die Benachrichtigung über abgerechnete Spieltage in Ihrem Account aktiviert haben. Sie können sie dort unter „Mein Account“ jederzeit abschalten.</p>
	</div>
</body>
</html>
`))

// GameDaySettledSubject returns the subject line of the "Spieltag abgerechnet" e-mail.
func GameDaySettledSubject(mail GameDaySettledEmail) string {
	return fmt.Sprintf("Spieltag vom %s abgerechnet – %s", mail.Date.UTC().Format("02.01.2006"), mail.ClubName)
}

// RenderGameDaySettled renders the HTML body of the "Spieltag abgerechnet" e-mail.
// Club and player names are escaped.
func RenderGameDaySettled(mail GameDaySettledEmail) (string, error) {
	var buf bytes.Buffer
	err := gameDaySettledTemplate.Execute(&buf, struct {
		GameDaySettledEmail
		Subject string
		Date    string
	}{
		GameDaySettledEmail: mail,
		Subject:             GameDaySettledSubject(mail),
		Date:                mail.Date.UTC().Format("02.01.2006"),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SendGameDaySettledEmail informs a user that a game day of their club was settled.
func (s *Service) SendGameDaySettledEmail(to string, mail GameDaySettledEmail) error {
	if !s.Enabled() {
		return fmt.Errorf("email service not configured: RESEND_API_KEY or RESEND_FROM_EMAIL not set")
	}
	html, err := RenderGameDaySettled(mail)
	if err != nil {
		return fmt.Errorf("render game day settled email: %w", err)
	}
	result, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: GameDaySettledSubject(mail),
		Html:    html,
	})
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}
	slog.Info("game day settled email sent", "email", to, "resend_id", result.Id)
	return nil
}

// Enabled reports whether e-mails can be sent (API key and sender configured).
func (s *Service) Enabled() bool {
	return s != nil && s.client != nil && s.fromEmail != ""
}

// FormatEuro formats cents as a German currency string, e.g. -350 → "-3,50 €".
func FormatEuro(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	s := fmt.Sprintf("%d", cents/100)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteRune('.')
		}
		b.WriteRune(r)
	}
	return fmt.Sprintf("%s%s,%02d €", sign, b.String(), cents%100)
}
