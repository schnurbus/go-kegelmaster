import { defineConfig, devices } from "@playwright/test";

const baseURL = process.env.BASE_URL ?? "http://localhost:5174";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: "html",
  webServer: [
    {
      command: 'cd ../backend && CORS_ALLOW_ORIGINS=http://localhost:5174 BACKEND_PORT=8081 DATABASE_URL="postgres://kegelmaster:e2e_test_password@localhost:5433/e2e_testdb?sslmode=disable" go run cmd/api/main.go',
      url: 'http://localhost:8081/healthz', // Warte bis Backend bereit
      reuseExistingServer: false, // Wichtig: Für E2E immer einen frischen Server
      stdout: 'pipe', // Optional: Backend-Logs in Playwright sehen
    },
    {
      command: 'cd ../frontend && VITE_BACKEND_URL=http://localhost:8081 VITE_PORT=5174 npm run dev -- --port 5174', // Startet Vite auf 5173
      url: 'http://localhost:5174',
      reuseExistingServer: false,
    },
  ],
  use: {
    baseURL,
    trace: "on-first-retry",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
