import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider, createBrowserRouter } from "react-router-dom";
import App from "./App.tsx";
import LandingPage from "./routes/LandingPage.tsx";
import DashboardPage from "./routes/DashboardPage.tsx";
import LoginPage from "./routes/LoginPage.tsx";
import RegisterPage from "./routes/RegisterPage.tsx";
import PlayersPage from "./routes/PlayersPage.tsx";
import PlayerDetailPage from "./routes/PlayerDetailPage.tsx";
import RolesPage from "./routes/RolesPage.tsx";
import RoleDetailPage from "./routes/RoleDetailPage.tsx";
import PenaltyTypesPage from "./routes/PenaltyTypesPage.tsx";
import GameDaysPage from "./routes/GameDaysPage.tsx";
import GameDayDetailPage from "./routes/GameDayDetailPage.tsx";
import TransactionsPage from "./routes/TransactionsPage.tsx";
import RequireAuth from "./components/RequireAuth.tsx";
import { AuthProvider } from "./context/AuthContext.tsx";
import { ClubProvider } from "./context/ClubContext.tsx";
import { ThemeProvider } from "./components/theme-provider.tsx";
import "./index.css";

const router = createBrowserRouter([
  {
    element: <App />,
    children: [
      { index: true, element: <LandingPage /> },
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      {
        element: <RequireAuth />,
        children: [
          { path: "/app", element: <DashboardPage /> },
          { path: "/app/players", element: <PlayersPage /> },
          { path: "/app/players/:id", element: <PlayerDetailPage /> },
          { path: "/app/roles", element: <RolesPage /> },
          { path: "/app/roles/:id", element: <RoleDetailPage /> },
          { path: "/app/penalty-types", element: <PenaltyTypesPage /> },
          { path: "/app/gamedays", element: <GameDaysPage /> },
          { path: "/app/gamedays/:id", element: <GameDayDetailPage /> },
          { path: "/app/transactions", element: <TransactionsPage /> },
        ],
      },
    ],
  },
]);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ThemeProvider defaultTheme="dark" storageKey="kegelmaster-theme">
      <AuthProvider>
        <ClubProvider>
          <RouterProvider router={router} />
        </ClubProvider>
      </AuthProvider>
    </ThemeProvider>
  </StrictMode>,
);
