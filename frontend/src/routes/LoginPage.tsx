import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const initialState = { email: "", password: "" };

function LoginPage() {
  const [form, setForm] = useState(initialState);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const redirectTo = (location.state as { from?: string })?.from ?? "/app";

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);
    try {
      await login(form);
      navigate(redirectTo, { replace: true });
    } catch (err) {
      setError((err as Error).message);
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

