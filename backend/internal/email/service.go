package email

import (
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v3"
)

// Service handles sending emails via Resend.
type Service struct {
	client    *resend.Client
	fromEmail string
}

// NewService creates a new email service.
func NewService(apiKey, fromEmail string) *Service {
	if apiKey == "" {
		slog.Warn("RESEND_API_KEY not set, email service will not work", "from_email", fromEmail)
		return &Service{
			client:    nil,
			fromEmail: fromEmail,
		}
	}

	slog.Info("email service initialized", "from_email", fromEmail, "api_key_set", apiKey != "")
	return &Service{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
	}
}

// SendInvitationEmail sends an invitation email to the specified address.
func (s *Service) SendInvitationEmail(email, playerName, invitationLink string) error {
	slog.Info("sending invitation email", "email", email, "player", playerName, "link", invitationLink)
	
	if s.client == nil {
		err := fmt.Errorf("email service not configured: RESEND_API_KEY not set")
		slog.Error("email service not configured", "error", err)
		return err
	}

	if s.fromEmail == "" {
		err := fmt.Errorf("from email not configured: RESEND_FROM_EMAIL not set")
		slog.Error("from email not configured", "error", err)
		return err
	}

	subject := fmt.Sprintf("Einladung für %s", playerName)
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>%s</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
	<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
		<h1 style="color: #2563eb;">Einladung zum Kegelclub</h1>
		<p>Hallo,</p>
		<p>Sie wurden eingeladen, dem Spieler <strong>%s</strong> beizutreten.</p>
		<p>Klicken Sie auf den folgenden Link, um die Einladung anzunehmen:</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="%s" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Einladung annehmen</a>
		</p>
		<p>Dieser Link ist 7 Tage gültig.</p>
		<p>Falls Sie diese Einladung nicht angefordert haben, können Sie diese E-Mail ignorieren.</p>
		<hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
		<p style="color: #666; font-size: 12px;">Diese E-Mail wurde automatisch generiert. Bitte antworten Sie nicht auf diese E-Mail.</p>
	</div>
</body>
</html>
`, subject, playerName, invitationLink)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{email},
		Subject: subject,
		Html:    htmlBody,
	}

	slog.Debug("sending email via Resend", "from", s.fromEmail, "to", email, "subject", subject)
	result, err := s.client.Emails.Send(params)
	if err != nil {
		slog.Error("failed to send invitation email via Resend", 
			"error", err, 
			"email", email, 
			"player", playerName,
			"from", s.fromEmail)
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}

	slog.Info("invitation email sent successfully",
		"email", email,
		"player", playerName,
		"resend_id", result.Id,
		"from", s.fromEmail)
	return nil
}

// SendPasswordResetEmail sends a password reset email with the given link.
func (s *Service) SendPasswordResetEmail(email, resetLink string) error {
	slog.Info("sending password reset email", "email", email, "link", resetLink)

	if s.client == nil {
		err := fmt.Errorf("email service not configured: RESEND_API_KEY not set")
		slog.Error("email service not configured", "error", err)
		return err
	}

	if s.fromEmail == "" {
		err := fmt.Errorf("from email not configured: RESEND_FROM_EMAIL not set")
		slog.Error("from email not configured", "error", err)
		return err
	}

	subject := "Passwort zurücksetzen – Kegelmaster"
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>%s</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
	<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
		<h1 style="color: #2563eb;">Passwort zurücksetzen</h1>
		<p>Hallo,</p>
		<p>Sie haben angefordert, Ihr Passwort zurückzusetzen. Klicken Sie auf den folgenden Link, um ein neues Passwort zu setzen:</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="%s" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Passwort zurücksetzen</a>
		</p>
		<p>Dieser Link ist eine Stunde gültig.</p>
		<p>Falls Sie diese Anfrage nicht gestellt haben, können Sie diese E-Mail ignorieren.</p>
		<hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
		<p style="color: #666; font-size: 12px;">Diese E-Mail wurde automatisch generiert. Bitte antworten Sie nicht auf diese E-Mail.</p>
	</div>
</body>
</html>
`, subject, resetLink)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{email},
		Subject: subject,
		Html:    htmlBody,
	}

	result, err := s.client.Emails.Send(params)
	if err != nil {
		slog.Error("failed to send password reset email via Resend",
			"error", err, "email", email, "from", s.fromEmail)
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}

	slog.Info("password reset email sent successfully",
		"email", email, "resend_id", result.Id, "from", s.fromEmail)
	return nil
}
