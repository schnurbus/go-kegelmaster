package admin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"text/tabwriter"
	"time"
)

// AdminStore is the part of Repository the CLI and the env bootstrap need.
type AdminStore interface {
	SetInstanceAdmin(ctx context.Context, email string, isAdmin bool) error
	ListInstanceAdmins(ctx context.Context) ([]InstanceAdmin, error)
}

// Audit actions of the CLI (it has no fiber.Ctx, so it logs directly).
const (
	ActionRoleGranted = "admin_role_granted"
	ActionRoleRevoked = "admin_role_revoked"
)

// CLIUsage describes the "admin" subcommand of the kegelmaster binary.
const CLIUsage = `Verwendung: kegelmaster admin [--database-url URL] <befehl>

Befehle:
  grant <email>    Instanz-Admin-Rolle vergeben
  revoke <email>   Instanz-Admin-Rolle entziehen
  list             Alle Instanz-Admins auflisten

Die Datenbank wird über --database-url oder DATABASE_URL angegeben.
Die Rolle kann bewusst nicht über die Weboberfläche vergeben werden.
`

// RunCLI executes "admin grant|revoke|list" and returns the process exit code.
func RunCLI(ctx context.Context, store AdminStore, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, CLIUsage)
		return 2
	}
	switch cmd := args[0]; cmd {
	case "grant", "revoke":
		if len(args) != 2 || !strings.Contains(args[1], "@") {
			fmt.Fprintf(stderr, "Verwendung: kegelmaster admin %s <email>\n", cmd)
			return 2
		}
		email := strings.ToLower(strings.TrimSpace(args[1]))
		grant := cmd == "grant"
		if err := store.SetInstanceAdmin(ctx, email, grant); err != nil {
			if errors.Is(err, ErrUserNotFound) {
				fmt.Fprintf(stderr, "Kein Konto mit der E-Mail %s gefunden.\n", email)
				return 1
			}
			fmt.Fprintf(stderr, "Fehler: %v\n", err)
			return 1
		}
		// Role changes are audited like admin web accesses (slog, admin=true).
		action := ActionRoleRevoked
		if grant {
			action = ActionRoleGranted
		}
		slog.Info("audit", "event", "audit", "action", action, "user_id", "cli", "admin", true, "target_email", email)
		if grant {
			fmt.Fprintf(stdout, "%s ist jetzt Instanz-Admin.\n", email)
		} else {
			fmt.Fprintf(stdout, "%s ist kein Instanz-Admin mehr.\n", email)
		}
		return 0
	case "list":
		if len(args) != 1 {
			fmt.Fprint(stderr, "Verwendung: kegelmaster admin list\n")
			return 2
		}
		admins, err := store.ListInstanceAdmins(ctx)
		if err != nil {
			fmt.Fprintf(stderr, "Fehler: %v\n", err)
			return 1
		}
		if len(admins) == 0 {
			fmt.Fprintln(stdout, "Keine Instanz-Admins.")
			return 0
		}
		tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "E-MAIL\tREGISTRIERT\tLETZTER LOGIN")
		for _, a := range admins {
			lastLogin := "-"
			if a.LastLoginAt != nil {
				lastLogin = a.LastLoginAt.UTC().Format(time.RFC3339)
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\n", a.Email, a.CreatedAt.UTC().Format(time.RFC3339), lastLogin)
		}
		_ = tw.Flush()
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(stdout, CLIUsage)
		return 0
	default:
		fmt.Fprintf(stderr, "Unbekannter Befehl %q\n\n%s", cmd, CLIUsage)
		return 2
	}
}

// BootstrapFromEnv grants the instance admin role to the existing accounts
// listed in INSTANCE_ADMIN_EMAILS. It never revokes anything, so env and
// database cannot contradict each other. Unknown e-mails are only logged; the
// account may register later and is picked up on the next start.
func BootstrapFromEnv(ctx context.Context, store AdminStore, emails []string) {
	for _, email := range emails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" {
			continue
		}
		err := store.SetInstanceAdmin(ctx, email, true)
		switch {
		case errors.Is(err, ErrUserNotFound):
			slog.Warn("INSTANCE_ADMIN_EMAILS: no account with this e-mail", "email", email)
		case err != nil:
			slog.Error("INSTANCE_ADMIN_EMAILS: grant instance admin", "email", email, "error", err)
		default:
			slog.Info("INSTANCE_ADMIN_EMAILS: instance admin granted", "email", email, "admin", true)
		}
	}
}
