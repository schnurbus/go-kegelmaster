import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { toast } from "sonner";

const initialState = { email: "", password: "" };

function RegisterPage() {
  const [form, setForm] = useState(initialState);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { register } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const inviteToken = searchParams.get("invite_token");

  const acceptInvitation = async (token: string) => {
    try {
      toast.loading("Einladung wird akzeptiert...", { id: "accept-invite-after-register" });
      
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
        id: "accept-invite-after-register",
        duration: 4000 
      });
      return data.player_id;
    } catch (error) {
      console.error("Error accepting invitation:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Akzeptieren der Einladung",
        { id: "accept-invite-after-register", duration: 6000 }
      );
      throw error;
    }
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);
    try {
      await register(form);
      
      // If there's an invite token, accept the invitation
      if (inviteToken) {
        try {
          const playerId = await acceptInvitation(inviteToken);
          navigate(`/app/players/${playerId}`, { replace: true });
        } catch (inviteError) {
          console.error("Failed to accept invitation:", inviteError);
          toast.warning(
            "Registrierung erfolgreich, aber Einladung konnte nicht akzeptiert werden. " +
            "Sie können die Einladung später erneut versuchen.",
            { duration: 8000 }
          );
          navigate("/app", { replace: true });
        }
      } else {
        navigate("/app", { replace: true });
      }
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <section className="page auth">
      <div className="auth-card">
        <h1>Registrieren</h1>
        <p>Erstelle einen Account, um Kegelabende zu verwalten.</p>

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
            {isSubmitting ? "Registriere ..." : "Account erstellen"}
          </button>
        </form>

        <p className="auth-hint">
          Bereits registriert? <Link to="/login">Zum Login</Link>
        </p>
      </div>
    </section>
  );
}

export default RegisterPage;

