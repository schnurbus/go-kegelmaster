import { Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

function LandingPage() {
  const { status } = useAuth();
  const isAuthed = status === "authenticated";

  return (
    <section className="page landing">
      <h1>Kegelabende mühelos verwalten</h1>
      <p>
        Erfasse Spielabende, Strafen und Einzahlungen für mehrere Klubs in
        einer zentralen Oberfläche.
      </p>
      <div className="cta-row">
        {isAuthed ? (
          <Link className="btn primary" to="/app">
            Zum Dashboard
          </Link>
        ) : (
          <>
            <Link className="btn primary" to="/register">
              Account erstellen
            </Link>
            <Link className="btn secondary" to="/login">
              Ich habe bereits einen Account
            </Link>
          </>
        )}
      </div>
    </section>
  );
}

export default LandingPage;

