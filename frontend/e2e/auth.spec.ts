import { test, expect } from "@playwright/test";

test.describe("Auth E2E", () => {
  test("positive: register then user is logged in", async ({ page }) => {
    const email = `e2e-reg-${Date.now()}@example.com`;
    const password = "securepass123";

    await page.goto("/register");
    await page.getByLabel("E-Mail").fill(email);
    await page.getByLabel("Passwort").fill(password);
    await page.getByRole("button", { name: "Account erstellen" }).click();

    await expect(page).toHaveURL(/\/app/, { timeout: 10000 });
    // Dashboard: user email in sidebar indicates logged in
    await expect(page.getByText(email)).toBeVisible({ timeout: 5000 });
  });

  test("positive: logout then login with same credentials", async ({ page }) => {
    const email = `e2e-login-${Date.now()}@example.com`;
    const password = "pass12345";

    await page.goto("/register");
    await page.getByLabel("E-Mail").fill(email);
    await page.getByLabel("Passwort").fill(password);
    await page.getByRole("button", { name: "Account erstellen" }).click();
    await expect(page).toHaveURL(/\/app/);

    // Open sidebar user menu and click Log out
    await page.getByRole("button", { name: new RegExp(email, "i") }).click();
    await page.getByRole("menuitem", { name: "Log out" }).click();
    await expect(page).toHaveURL(/\/login/);

    await page.goto("/login");
    await page.getByLabel("E-Mail").fill(email);
    await page.getByLabel("Passwort").fill(password);
    await page.getByRole("button", { name: "Login" }).click();
    await expect(page).toHaveURL(/\/app/);
  });

  test("negative: register with invalid email shows error", async ({ page }) => {
    await page.goto("/register");
    await page.getByLabel("E-Mail").fill("invalid");
    await page.getByLabel("Passwort").fill("12345678");
    // Bypass HTML5 validation and submit in one tick so React cannot re-render in between
    const [response] = await Promise.all([
      page.waitForResponse(
        (resp) =>
          resp.url().includes("/api/auth/register") &&
          resp.status() >= 400 &&
          resp.status() < 500,
        { timeout: 10000 }
      ),
      page.evaluate(() => {
        const emailInput = document.querySelector("#register-email") as HTMLInputElement;
        if (emailInput) {
          emailInput.type = "text";
          emailInput.removeAttribute("required");
        }
        document.querySelector<HTMLButtonElement>('button[type="submit"]')?.click();
      }),
    ]);
    await expect(response.status()).toBe(400);
    await expect(page.getByRole("alert")).toBeVisible({ timeout: 10000 });
    await expect(page).not.toHaveURL(/\/app/);
  });

  test("negative: register with short password shows error", async ({ page }) => {
    await page.goto("/register");
    await page.getByLabel("E-Mail").fill("a@b.com");
    // Bypass HTML5 minLength so form submits and backend returns 400
    await page.locator("#register-password").evaluate((el) => {
      (el as HTMLInputElement).removeAttribute("minLength");
      (el as HTMLInputElement).removeAttribute("minlength");
    });
    await page.getByLabel("Passwort").fill("123");
    await page.getByRole("button", { name: "Account erstellen" }).click();

    await expect(page.getByRole("alert")).toBeVisible({ timeout: 10000 });
    await expect(page).not.toHaveURL(/\/app/);
  });

  test("negative: login with wrong password shows error", async ({ page }) => {
    const email = `e2e-wrong-${Date.now()}@example.com`;
    await page.goto("/register");
    await page.getByLabel("E-Mail").fill(email);
    await page.getByLabel("Passwort").fill("correct123");
    await page.getByRole("button", { name: "Account erstellen" }).click();
    await expect(page).toHaveURL(/\/app/);

    await page.getByRole("button", { name: new RegExp(email, "i") }).click();
    await page.getByRole("menuitem", { name: "Log out" }).click();
    await page.goto("/login");
    await page.getByLabel("E-Mail").fill(email);
    await page.getByLabel("Passwort").fill("wrong");
    // Bypass HTML5 minLength so form submits and backend returns 401
    await page.locator("#login-password").evaluate((el) => {
      (el as HTMLInputElement).removeAttribute("minLength");
      (el as HTMLInputElement).removeAttribute("minlength");
    });
    // Wait for 401 response then assert alert
    await Promise.all([
      page.waitForResponse(
        (resp) => resp.url().includes("/api/auth/login") && resp.status() === 401,
        { timeout: 10000 }
      ),
      page.getByRole("button", { name: "Login" }).click(),
    ]);
    await expect(page.getByRole("alert")).toBeVisible({ timeout: 5000 });
  });

  test("negative: login with unknown email shows error", async ({ page }) => {
    await page.goto("/login");
    await page.getByLabel("E-Mail").fill("unknown@example.com");
    await page.getByLabel("Passwort").fill("anypass123");
    await page.getByRole("button", { name: "Login" }).click();

    await expect(page.getByRole("alert")).toBeVisible({ timeout: 10000 });
  });
});
