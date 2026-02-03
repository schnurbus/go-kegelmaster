import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { toast } from "sonner";

const initialState = { email: "", password: "" };

function LoginPage() {
  const [form, setForm] = useState(initialState);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const inviteToken = searchParams.get("invite_token");
  const redirectTo = (location.state as { from?: string })?.from ?? "/app";

  const acceptInvitation = async (token: string) => {
    try {
      toast.loading("Einladung wird akzeptiert...", { id: "accept-invite-after-login" });
      
      const csrfToken = document.cookie
        .split("; ")
        .find((row) => row.startsWith("csrf_token="))
        ?.split("=")[1] || "";

      if (!csrfToken) {
        throw new Error("CSRF-Token fehlt. Bitte Seite neu laden.");
      }

      const response = await fetch(`/api/invitations/${token}/accept`, {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrfToken,
        },
        credentials: "include",
      });

      const responseData = await response.json().catch(() => ({}));

      if (!response.ok) {
        const errorMessage = responseData.message || "Fehler beim Akzeptieren der Einladung";
        
        if (response.status === 404) {
          throw new Error("Einladung nicht gefunden");
        }
        if (response.status === 410) {
          throw new Error("Einladung ist abgelaufen oder wurde bereits akzeptiert");
        }
        
        throw new Error(errorMessage);
      }

      const data = await response.json();
      toast.success("Einladung erfolgreich akzeptiert!", { 
        id: "accept-invite-after-login",
        duration: 4000 
      });
      return data.player_id;
    } catch (error) {
      console.error("Error accepting invitation:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Akzeptieren der Einladung",
        { id: "accept-invite-after-login", duration: 6000 }
      );
      throw error;
    }
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);
    try {
      await login(form);
      toast.success("Anmeldung erfolgreich");
      // If there's an invite token, accept the invitation
      if (inviteToken) {
        try {
          const playerId = await acceptInvitation(inviteToken);
          navigate(`/app/players/${playerId}`, { replace: true });
        } catch (inviteError) {
          console.error("Failed to accept invitation:", inviteError);
          toast.warning(
            "Anmeldung erfolgreich, aber Einladung konnte nicht akzeptiert werden. " +
            "Sie können die Einladung später erneut versuchen.",
            { duration: 8000 }
          );
          navigate(redirectTo, { replace: true });
        }
      } else {
        navigate(redirectTo, { replace: true });
      }
    } catch (err) {
      const message = (err as Error).message;
      setError(message);
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <section className="page auth">
      <div className="auth-card">
        <h1>Anmelden</h1>
        <p>Willkommen zurück! Bitte melde dich mit deinem Account an.</p>

        {error && <div className="error-banner">{error}</div>}

        <form className="auth-form" onSubmit={handleSubmit}>
          <label className="form-group">
            <span>E-Mail</span>
            <input
              type="email"
              value={form.email}
              onChange={(event) =>
                setForm((prev) => ({ ...prev, email: event.target.value }))
              }
              required
            />
          </label>

          <label className="form-group">
            <span>Passwort</span>
            <input
              type="password"
              minLength={8}
              value={form.password}
              onChange={(event) =>
                setForm((prev) => ({ ...prev, password: event.target.value }))
              }
              required
            />
          </label>

          <button className="btn primary" disabled={isSubmitting}>
            {isSubmitting ? "Melde an ..." : "Login"}
          </button>
        </form>

        <p className="auth-hint">
          Kein Account? <Link to="/register">Jetzt registrieren</Link>
        </p>
      </div>
    </section>
  );
}

export default LoginPage;

