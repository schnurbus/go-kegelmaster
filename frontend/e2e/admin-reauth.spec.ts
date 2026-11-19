import { createHmac } from "node:crypto";
import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test, expect, type Page } from "@playwright/test";

// Same database and JWT secret as the E2E backend in playwright.local.config.ts
// (which leaves JWT_SECRET at its development default).
const databaseURL =
  process.env.E2E_DATABASE_URL ??
  "postgres://kegelmaster:e2e_test_password@localhost:5433/e2e_testdb?sslmode=disable";
const jwtSecret = process.env.E2E_JWT_SECRET ?? "dev-secret-change-me";
const backendDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../backend");
const password = "securepass123";

function grantInstanceAdmin(email: string) {
  execFileSync("go", ["run", "./cmd/api", "admin", "--database-url", databaseURL, "grant", email], {
    cwd: backendDir,
    stdio: "pipe",
  });
}

async function register(page: Page, email: string) {
  await page.goto("/register");
  await page.getByLabel("E-Mail").fill(email);
  await page.getByLabel("Passwort").fill(password);
  await page.getByRole("button", { name: "Account erstellen" }).click();
  await expect(page).toHaveURL(/\/app/, { timeout: 10000 });
}

function base64url(value: string | Buffer): string {
  return Buffer.from(value).toString("base64url");
}

/** Signs a session JWT like the backend, with an auth_time in the past (stale session). */
function staleSessionToken(userID: string, ageSeconds: number): string {
  const now = Math.floor(Date.now() / 1000);
  const header = base64url(JSON.stringify({ alg: "HS256", typ: "JWT" }));
  const payload = base64url(
    JSON.stringify({ uid: userID, sub: userID, iat: now - 5, exp: now + 3600, auth_time: now - ageSeconds }),
  );
  const signature = createHmac("sha256", jwtSecret).update(`${header}.${payload}`).digest("base64url");
  return `${header}.${payload}.${signature}`;
}

test.describe("Admin area re-authentication", () => {
  test("stale admin session asks for the password and then shows the admin area", async ({ page, context, baseURL }) => {
    test.setTimeout(90000);
    const adminEmail = `e2e-reauth-${Date.now()}@example.com`;
    await register(page, adminEmail);
    grantInstanceAdmin(adminEmail);

    const me = await page.request.get("/api/auth/me");
    expect(me.ok()).toBeTruthy();
    const { id: userID } = (await me.json()) as { id: string };

    // Replace the fresh session by one whose last login was an hour ago.
    await context.addCookies([
      { name: "auth_token", value: staleSessionToken(userID, 3600), url: baseURL ?? "http://localhost:5174", httpOnly: true },
    ]);
    const stale = await page.request.get("/api/admin/stats");
    expect(stale.status()).toBe(401);
    expect(((await stale.json()) as { error: string }).error).toBe("reauth_required");

    await page.goto("/admin");
    const dialog = page.getByRole("dialog", { name: "Anmeldung bestätigen" });
    await expect(dialog).toBeVisible();

    // Wrong password: stays in the dialog with an error.
    await dialog.getByLabel("Passwort", { exact: true }).fill("falsches-passwort");
    await dialog.getByRole("button", { name: "Bestätigen", exact: true }).click();
    await expect(dialog.getByRole("alert")).toContainText("Passwort ist falsch");

    await dialog.getByLabel("Passwort", { exact: true }).fill(password);
    await dialog.getByRole("button", { name: "Bestätigen", exact: true }).click();
    await expect(dialog).toBeHidden();
    await expect(page.getByText("Auffälligkeiten", { exact: true })).toBeVisible();
    expect((await page.request.get("/api/admin/stats")).ok()).toBeTruthy();
  });

  test("stale admin session can be confirmed with a passkey", async ({ page, context, baseURL }) => {
    test.setTimeout(90000);
    // Chromium's virtual authenticator stands in for Touch ID / Windows Hello.
    const cdp = await context.newCDPSession(page);
    await cdp.send("WebAuthn.enable");
    await cdp.send("WebAuthn.addVirtualAuthenticator", {
      options: {
        protocol: "ctap2",
        transport: "internal",
        hasResidentKey: true,
        hasUserVerification: true,
        isUserVerified: true,
        automaticPresenceSimulation: true,
      },
    });

    const adminEmail = `e2e-reauth-passkey-${Date.now()}@example.com`;
    await register(page, adminEmail);
    grantInstanceAdmin(adminEmail);

    await page.goto("/app/account");
    await page.getByRole("button", { name: "Passkey hinzufügen" }).click();
    const addDialog = page.getByRole("dialog");
    await addDialog.getByLabel("Name").fill("Testgerät");
    await addDialog.getByLabel("Aktuelles Passwort").fill(password);
    await addDialog.getByRole("button", { name: "Weiter" }).click();
    await expect(page.getByText("Passkey hinzugefügt")).toBeVisible({ timeout: 10000 });

    const { id: userID } = (await (await page.request.get("/api/auth/me")).json()) as { id: string };
    await context.addCookies([
      { name: "auth_token", value: staleSessionToken(userID, 3600), url: baseURL ?? "http://localhost:5174", httpOnly: true },
    ]);

    await page.goto("/admin");
    const dialog = page.getByRole("dialog", { name: "Anmeldung bestätigen" });
    await dialog.getByRole("button", { name: "Mit Passkey bestätigen" }).click();
    await expect(dialog).toBeHidden({ timeout: 10000 });
    await expect(page.getByText("Auffälligkeiten", { exact: true })).toBeVisible();
  });
});
