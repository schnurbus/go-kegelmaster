import { test, expect, type Page } from "@playwright/test";

async function registerAndOpenAccount(page: Page, email: string, password: string) {
  await page.goto("/register");
  await page.getByLabel("E-Mail").fill(email);
  await page.getByLabel("Passwort").fill(password);
  await page.getByRole("button", { name: "Account erstellen" }).click();
  await expect(page).toHaveURL(/\/app/, { timeout: 10000 });

  await page.getByRole("button", { name: new RegExp(email, "i") }).click();
  await page.getByRole("menuitem", { name: "Account" }).click();
  await expect(page).toHaveURL(/\/app\/account/);
}

test.describe("Account page", () => {
  test("user menu has no placeholder entries", async ({ page }) => {
    const email = `e2e-menu-${Date.now()}@example.com`;
    await page.goto("/register");
    await page.getByLabel("E-Mail").fill(email);
    await page.getByLabel("Passwort").fill("securepass123");
    await page.getByRole("button", { name: "Account erstellen" }).click();
    await expect(page).toHaveURL(/\/app/, { timeout: 10000 });

    await page.getByRole("button", { name: new RegExp(email, "i") }).click();
    await expect(page.getByRole("menuitem", { name: "Account" })).toBeVisible();
    await expect(page.getByRole("menuitem", { name: "Billing" })).toHaveCount(0);
    await expect(page.getByRole("menuitem", { name: "Notifications" })).toHaveCount(0);
  });

  test("shows account details", async ({ page }) => {
    const email = `e2e-account-${Date.now()}@example.com`;
    await registerAndOpenAccount(page, email, "securepass123");

    const main = page.getByRole("main");
    await expect(main.getByText(email)).toBeVisible();
    await expect(main.getByText("Noch nie")).toBeVisible();
  });

  test("change password, then log in with the new one", async ({ page }) => {
    const email = `e2e-chpw-${Date.now()}@example.com`;
    const oldPassword = "oldpass12345";
    const newPassword = "newpass67890";
    await registerAndOpenAccount(page, email, oldPassword);

    await page.getByLabel("Aktuelles Passwort").fill(oldPassword);
    await page.getByLabel("Neues Passwort", { exact: true }).fill(newPassword);
    await page.getByLabel("Neues Passwort bestätigen").fill(newPassword);
    await page.getByRole("button", { name: "Passwort ändern" }).click();

    await expect(page.getByText("Passwort geändert")).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole("main").getByText("Noch nie")).toHaveCount(0);

    // The session survives the change: a reload stays signed in.
    await page.reload();
    await expect(page).toHaveURL(/\/app\/account/);

    // Logging out still works with the rotated CSRF token.
    await page.getByRole("button", { name: new RegExp(email, "i") }).click();
    await page.getByRole("menuitem", { name: "Log out" }).click();
    await expect(page).toHaveURL(/\/login/);

    await page.getByLabel("E-Mail").fill(email);
    await page.getByLabel("Passwort").fill(oldPassword);
    await page.getByRole("button", { name: "Login" }).click();
    await expect(page.getByRole("alert")).toBeVisible({ timeout: 10000 });

    await page.getByLabel("Passwort").fill(newPassword);
    await page.getByRole("button", { name: "Login" }).click();
    await expect(page).toHaveURL(/\/app/, { timeout: 10000 });
  });

  test("wrong current password shows an error", async ({ page }) => {
    const email = `e2e-chpw-wrong-${Date.now()}@example.com`;
    await registerAndOpenAccount(page, email, "rightpass123");

    await page.getByLabel("Aktuelles Passwort").fill("wrongpass123");
    await page.getByLabel("Neues Passwort", { exact: true }).fill("newpass67890");
    await page.getByLabel("Neues Passwort bestätigen").fill("newpass67890");
    await page.getByRole("button", { name: "Passwort ändern" }).click();

    await expect(page.getByRole("alert")).toContainText("Aktuelles Passwort ist falsch", { timeout: 10000 });
  });

  test("mismatched confirmation is rejected client-side", async ({ page }) => {
    const email = `e2e-chpw-mismatch-${Date.now()}@example.com`;
    await registerAndOpenAccount(page, email, "rightpass123");

    await page.getByLabel("Aktuelles Passwort").fill("rightpass123");
    await page.getByLabel("Neues Passwort", { exact: true }).fill("newpass67890");
    await page.getByLabel("Neues Passwort bestätigen").fill("otherpass678");
    await page.getByRole("button", { name: "Passwort ändern" }).click();

    await expect(page.getByRole("alert")).toContainText("stimmen nicht überein");
  });
});
