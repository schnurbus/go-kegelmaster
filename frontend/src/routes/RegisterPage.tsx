import type { FormEvent } from "react";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const initialState = { email: "", password: "" };

function RegisterPage() {
  const [form, setForm] = useState(initialState);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { register } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);
    try {
      await register(form);
      navigate("/app", { replace: true });
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

