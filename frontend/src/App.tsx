import { NavLink, Outlet, useLocation } from "react-router-dom";
import "./App.css";
import { useAuth } from "./context/AuthContext";

function App() {
  const { status, user, logout } = useAuth();
  const isAuthenticated = status === "authenticated";
  const location = useLocation();
  const isDashboard = location.pathname.startsWith("/app");

  // Dashboard hat seine eigene Sidebar, daher keinen Header anzeigen
  if (isDashboard) {
    return <Outlet />;
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <NavLink to="/" className="brand">
          go-kegelmaster
        </NavLink>
        <nav className="app-nav">
          <NavLink to="/" end>
            Landing
          </NavLink>
          {isAuthenticated ? (
            <>
              <NavLink to="/app">Dashboard</NavLink>
              <span className="user-chip">{user?.email}</span>
              <button
                className="link-btn"
                onClick={() => logout().catch(() => undefined)}
              >
                Logout
              </button>
            </>
          ) : (
            <>
              <NavLink to="/login">Login</NavLink>
              <NavLink to="/register">Registrieren</NavLink>
            </>
          )}
        </nav>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
      <footer className="app-footer">
        <small>© {new Date().getFullYear()} Schnurbus</small>
      </footer>
    </div>
  );
}

export default App;
