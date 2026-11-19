import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test, expect, type Page } from "@playwright/test";

// Same database as the E2E backend in playwright.local.config.ts.
const databaseURL =
  process.env.E2E_DATABASE_URL ??
  "postgres://kegelmaster:e2e_test_password@localhost:5433/e2e_testdb?sslmode=disable";
const backendDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../backend");

/** Grants the instance admin role the way operators do it: via the CLI of the main binary. */
function grantInstanceAdmin(email: string) {
  execFileSync("go", ["run", "./cmd/api", "admin", "--database-url", databaseURL, "grant", email], {
    cwd: backendDir,
    stdio: "pipe",
  });
}

async function register(page: Page, email: string) {
  await page.goto("/register");
  await page.getByLabel("E-Mail").fill(email);
  await page.getByLabel("Passwort").fill("securepass123");
  await page.getByRole("button", { name: "Account erstellen" }).click();
  await expect(page).toHaveURL(/\/app/, { timeout: 10000 });
}

async function createClub(page: Page, name: string) {
  const status = await page.evaluate(async (clubName) => {
    const csrf =
      document.cookie
        .split("; ")
        .find((row) => row.startsWith("csrf_token="))
        ?.split("=")[1] ?? "";
    const resp = await fetch("/api/clubs", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", "X-CSRF-Token": csrf },
      body: JSON.stringify({ name: clubName }),
    });
    return resp.status;
  }, name);
  expect(status).toBe(201);
}

test.describe("Admin area", () => {
  test("instance admin sees users and clubs", async ({ browser, page }) => {
    test.setTimeout(90000);
    const stamp = Date.now();
    const ownerEmail = `e2e-admin-owner-${stamp}@example.com`;
    const adminEmail = `e2e-admin-${stamp}@example.com`;
    const clubName = `E2E Adminklub ${stamp}`;

    // A regular user owns a club.
    const ownerContext = await browser.newContext();
    const ownerPage = await ownerContext.newPage();
    await register(ownerPage, ownerEmail);
    await createClub(ownerPage, clubName);
    await ownerContext.close();

    await register(page, adminEmail);
    grantInstanceAdmin(adminEmail);
    await page.reload();

    await page.getByRole("link", { name: "Administration" }).click();
    await expect(page).toHaveURL(/\/admin$/);
    await expect(page.getByText("Auffälligkeiten", { exact: true })).toBeVisible();
    await expect(page.getByText("Offene Einladungen", { exact: true })).toBeVisible();

    // Users: search for the club owner and open the detail page.
    await page.getByRole("navigation", { name: "Administration" }).getByRole("link", { name: "Benutzer" }).click();
    await expect(page).toHaveURL(/\/admin\/users/);
    await page.getByLabel("Benutzer suchen").fill(ownerEmail);
    await expect(page).toHaveURL(/search=/);
    const table = page.getByRole("table");
    await expect(table.getByRole("link", { name: ownerEmail })).toBeVisible();
    await expect(table.getByRole("link", { name: adminEmail })).toHaveCount(0);
    await table.getByRole("link", { name: ownerEmail }).click();
    await expect(page).toHaveURL(/\/admin\/users\/[0-9a-f-]+$/);
    await expect(page.getByText("Eigene Klubs (1)")).toBeVisible();

    // Cross link user -> club.
    await page.getByRole("link", { name: clubName }).click();
    await expect(page).toHaveURL(/\/admin\/clubs\/[0-9a-f-]+$/);
    await expect(page.getByRole("link", { name: ownerEmail }).first()).toBeVisible();

    // Club list contains the club, with a link to its owner.
    await page.getByRole("navigation", { name: "Administration" }).getByRole("link", { name: "Klubs" }).click();
    await expect(page).toHaveURL(/\/admin\/clubs$/);
    await page.getByLabel("Klubs suchen").fill(clubName);
    await expect(page.getByRole("table").getByRole("link", { name: clubName })).toBeVisible();
    await expect(page.getByRole("table").getByRole("link", { name: ownerEmail })).toBeVisible();

    // The admin filter shows the admin with the Admin badge.
    await page.goto(`/admin/users?filter=admin&search=${encodeURIComponent(adminEmail)}`);
    const adminRow = page.getByRole("row").filter({ hasText: adminEmail });
    await expect(adminRow.getByText("Admin", { exact: true })).toBeVisible();

    // CSV export of the memberships.
    const csv = await page.request.get("/api/admin/export/memberships.csv");
    expect(csv.ok()).toBeTruthy();
    expect(await csv.text()).toContain(clubName);
  });

  test("regular user has no menu entry and gets an error page on /admin", async ({ page }) => {
    await register(page, `e2e-noadmin-${Date.now()}@example.com`);

    await expect(page.getByRole("link", { name: "Hilfe" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Administration" })).toHaveCount(0);

    await page.goto("/admin");
    await expect(page.getByRole("heading", { name: "Seite nicht gefunden" })).toBeVisible();
    await page.goto("/admin/users");
    await expect(page.getByRole("heading", { name: "Seite nicht gefunden" })).toBeVisible();

    const resp = await page.request.get("/api/admin/users");
    expect(resp.status()).toBe(404);
  });
});
