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
import CompetitionsPage from "./routes/CompetitionsPage.tsx";
import GameDaysPage from "./routes/GameDaysPage.tsx";
import GameDayDetailPage from "./routes/GameDayDetailPage.tsx";
import TransactionsPage from "./routes/TransactionsPage.tsx";
import ClubEditPage from "./routes/ClubEditPage.tsx";
import HelpPage from "./routes/HelpPage.tsx";
import InvitationPage from "./routes/InvitationPage.tsx";
import ForgotPasswordPage from "./routes/ForgotPasswordPage.tsx";
import ResetPasswordPage from "./routes/ResetPasswordPage.tsx";
import ImpressumPage from "./routes/ImpressumPage.tsx";
import DatenschutzPage from "./routes/DatenschutzPage.tsx";
import RequireAuth from "./components/RequireAuth.tsx";
import { DashboardLayout } from "./components/DashboardLayout.tsx";
import { AuthProvider } from "./context/AuthContext.tsx";
import { ClubProvider } from "./context/ClubContext.tsx";
import { ThemeProvider } from "./components/theme-provider.tsx";
import { Toaster } from "./components/ui/sonner.tsx";
import "./index.css";

const router = createBrowserRouter([
  {
    element: <App />,
    children: [
      { index: true, element: <LandingPage /> },
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      { path: "/forgot-password", element: <ForgotPasswordPage /> },
      { path: "/reset-password", element: <ResetPasswordPage /> },
      { path: "/impressum", element: <ImpressumPage /> },
      { path: "/datenschutz", element: <DatenschutzPage /> },
      { path: "/invite/:token", element: <InvitationPage /> },
      {
        element: <RequireAuth />,
        children: [
          {
            path: "/app",
            element: <DashboardLayout />,
            children: [
              { index: true, element: <DashboardPage />, handle: { title: "Dashboard" } },
              { path: "players", element: <PlayersPage />, handle: { title: "Spieler" } },
              { path: "players/:id", element: <PlayerDetailPage />, handle: { title: "Spieler" } },
              { path: "roles", element: <RolesPage />, handle: { title: "Rollen" } },
              { path: "roles/:id", element: <RoleDetailPage />, handle: { title: "Rollen Details" } },
              { path: "penalty-types", element: <PenaltyTypesPage />, handle: { title: "Strafentypen" } },
              { path: "competitions", element: <CompetitionsPage />, handle: { title: "Wettbewerbe" } },
              { path: "gamedays", element: <GameDaysPage />, handle: { title: "Spieltage" } },
              { path: "gamedays/:id", element: <GameDayDetailPage />, handle: { title: "Spieltag" } },
              { path: "transactions", element: <TransactionsPage />, handle: { title: "Transaktionen" } },
              { path: "club/:clubId", element: <ClubEditPage />, handle: { title: "Club bearbeiten" } },
              { path: "help", element: <HelpPage />, handle: { title: "Hilfe" } },
            ],
          },
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
          <Toaster richColors position="top-center" />
        </ClubProvider>
      </AuthProvider>
    </ThemeProvider>
  </StrictMode>,
);
