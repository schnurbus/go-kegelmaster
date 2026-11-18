import { test, expect, type Page } from "@playwright/test";

// Chromium's virtual authenticator (via CDP) stands in for Touch ID / Windows Hello.
async function addVirtualAuthenticator(page: Page) {
  const client = await page.context().newCDPSession(page);
  await client.send("WebAuthn.enable");
  await client.send("WebAuthn.addVirtualAuthenticator", {
    options: {
      protocol: "ctap2",
      transport: "internal",
      hasResidentKey: true,
      hasUserVerification: true,
      isUserVerified: true,
      automaticPresenceSimulation: true,
    },
  });
}

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

async function addPasskey(page: Page, name: string, password: string) {
  await page.getByRole("button", { name: "Passkey hinzufügen" }).click();
  const dialog = page.getByRole("dialog");
  await dialog.getByLabel("Name").fill(name);
  await dialog.getByLabel("Aktuelles Passwort").fill(password);
  await dialog.getByRole("button", { name: "Weiter" }).click();
}

// Leaves the app for /login. Note: the virtual authenticator answers the
// login page's autofill request (conditional mediation) right away.
async function logout(page: Page, email: string) {
  await page.getByRole("button", { name: new RegExp(email, "i") }).click();
  await page.getByRole("menuitem", { name: "Log out" }).click();
}

// Simulates a browser without passkey autofill, so only the button is used.
async function disableAutofill(page: Page) {
  await page.addInitScript(() => {
    PublicKeyCredential.isConditionalMediationAvailable = async () => false;
  });
}

test.describe("Passkeys", () => {
  test("add a passkey, then sign in with the button", async ({ page }) => {
    await disableAutofill(page);
    await addVirtualAuthenticator(page);
    const email = `e2e-passkey-${Date.now()}@example.com`;
    const password = "securepass123";
    await registerAndOpenAccount(page, email, password);

    await addPasskey(page, "Testgerät", password);
    await expect(page.getByText("Passkey hinzugefügt")).toBeVisible({ timeout: 10000 });
    const list = page.getByRole("list", { name: "Passkeys" });
    await expect(list.getByText("Testgerät")).toBeVisible();
    await expect(list.getByText("Noch nie")).toBeVisible();

    await logout(page, email);
    await expect(page).toHaveURL(/\/login/);
    await page.getByRole("button", { name: "Mit Passkey anmelden" }).click();
    await expect(page).toHaveURL(/\/app/, { timeout: 10000 });

    // The session is real and the passkey records its use.
    await page.goto("/app/account");
    await expect(page.getByRole("main").getByText(email)).toBeVisible();
    await expect(page.getByRole("list", { name: "Passkeys" }).getByText("Noch nie")).toHaveCount(0);
  });

  test("passkey autofill on the login page signs in", async ({ page }) => {
    await addVirtualAuthenticator(page);
    const email = `e2e-passkey-autofill-${Date.now()}@example.com`;
    const password = "securepass123";
    await registerAndOpenAccount(page, email, password);
    await addPasskey(page, "Testgerät", password);
    await expect(page.getByText("Passkey hinzugefügt")).toBeVisible({ timeout: 10000 });

    // Logging out lands on /login, whose autofill request the authenticator answers.
    await logout(page, email);
    await page.waitForURL(/\/login/);
    // Back to where the user logged out (/app/account), now signed in again.
    await expect(page).toHaveURL(/\/app\//, { timeout: 10000 });
    await page.goto("/app/account");
    await expect(page.getByRole("main").getByText(email)).toBeVisible();
  });

  test("adding a passkey requires the current password", async ({ page }) => {
    await addVirtualAuthenticator(page);
    const email = `e2e-passkey-pw-${Date.now()}@example.com`;
    await registerAndOpenAccount(page, email, "securepass123");

    await addPasskey(page, "Testgerät", "wrongpass123");
    await expect(page.getByRole("dialog").getByRole("alert")).toHaveText("Aktuelles Passwort ist falsch");
    await expect(page.getByText("Noch keine Passkeys eingerichtet.")).toBeVisible();
  });

  test("a removed passkey can no longer sign in", async ({ page }) => {
    await addVirtualAuthenticator(page);
    const email = `e2e-passkey-del-${Date.now()}@example.com`;
    const password = "securepass123";
    await registerAndOpenAccount(page, email, password);

    await addPasskey(page, "Altes Handy", password);
    await expect(page.getByText("Passkey hinzugefügt")).toBeVisible({ timeout: 10000 });

    await page.getByRole("button", { name: "Passkey Altes Handy umbenennen" }).click();
    await page.getByRole("dialog").getByLabel("Name").fill("Diensthandy");
    await page.getByRole("dialog").getByRole("button", { name: "Speichern" }).click();
    await expect(page.getByRole("list", { name: "Passkeys" }).getByText("Diensthandy")).toBeVisible();

    await page.getByRole("button", { name: "Passkey Diensthandy entfernen" }).click();
    await page.getByRole("alertdialog").getByRole("button", { name: "Entfernen" }).click();
    await expect(page.getByText("Noch keine Passkeys eingerichtet.")).toBeVisible();

    // The authenticator still holds the key, but the server no longer accepts it
    // (neither via autofill nor via the button).
    await logout(page, email);
    await expect(page).toHaveURL(/\/login/);
    const alert = page.getByRole("alert");
    await expect(alert).toContainText("Anmeldung mit Passkey fehlgeschlagen", { timeout: 10000 });
    await page.getByRole("button", { name: "Mit Passkey anmelden" }).click();
    await expect(alert).toContainText("Anmeldung mit Passkey fehlgeschlagen", { timeout: 10000 });
    await expect(page).toHaveURL(/\/login/);
  });
});
