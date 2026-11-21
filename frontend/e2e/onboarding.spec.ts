import { test, expect, type Page } from "@playwright/test";

// Rundgang und Ersteinrichtung (#91). Die Tests registrieren frische Konten;
// der Einladungstest braucht die offene Einladung aus dem Dev-Seed
// (make dev-fresh), die nur einmal angenommen werden kann.

const SEED_INVITE_TOKEN = "e2e-einladung-kc-alle-neune";
const SEED_INVITE_EMAIL = "eingeladen@kegelmaster.local";
const SEED_CLUB = "KC Alle Neune";

const tour = (page: Page) => page.locator(".driver-popover");
const tourTitle = (page: Page) => page.locator(".driver-popover-title");

async function register(page: Page, email: string, query = "") {
  await page.goto(`/register${query}`);
  await page.getByLabel("E-Mail").fill(email);
  await page.getByLabel("Passwort", { exact: true }).fill("securepass123");
  await page.getByRole("button", { name: "Account erstellen" }).click();
  await expect(page).toHaveURL(/\/app/, { timeout: 10000 });
}

async function createClubFromWelcome(page: Page, name: string) {
  await expect(page.getByText("Willkommen beim Kegelmaster")).toBeVisible();
  await page.getByRole("button", { name: "Club anlegen" }).click();
  await page.getByPlaceholder("Club Name").fill(name);
  await page.getByRole("button", { name: "Erstellen", exact: true }).click();
}

/** Klickt sich durch den Rundgang und liefert die Titel der Sprechblasen. */
async function walkTour(page: Page): Promise<string[]> {
  const titles: string[] = [];
  await expect(tour(page)).toBeVisible({ timeout: 10000 });
  for (let i = 0; i < 20; i++) {
    const title = (await tourTitle(page).innerText()).trim();
    if (titles[titles.length - 1] !== title) titles.push(title);
    const next = tour(page).locator(".driver-popover-next-btn");
    const label = (await next.innerText()).trim();
    await next.click();
    if (label === "Fertig") break;
    await expect(tourTitle(page)).not.toHaveText(title);
  }
  await expect(tour(page)).toHaveCount(0);
  return titles;
}

async function tourFlags(page: Page) {
  return page.evaluate(async () => {
    const me = await fetch("/api/auth/me", { credentials: "include" }).then((r) => r.json());
    return { tour: me.tour_seen_at as string | null, hint: me.setup_hint_seen_at as string | null };
  });
}

test.describe("Neuer Besitzer", () => {
  test("Willkommensseite → Club → Rundgang → Ersteinrichtung bis die Karte verschwindet", async ({ page }) => {
    await register(page, `e2e-owner-${Date.now()}@example.com`);
    await createClubFromWelcome(page, "KC E2E");

    const titles = await walkTour(page);
    expect(titles[0]).toBe("Willkommen im Kegelmaster");
    expect(titles).toContain("Ersteinrichtung");
    expect(titles).toContain("Rollen");
    expect(titles[titles.length - 1]).toBe("Ihr Benutzerkonto");
    await expect.poll(async () => (await tourFlags(page)).tour).not.toBeNull();

    const card = page.locator('[data-tour="setup-card"]');
    await expect(card).toBeVisible();
    await expect(card.getByText("1 von 5 Pflichtschritten erledigt")).toBeVisible();

    // Strafentyp: Klick führt zur Seite mit offenem Dialog.
    await card.getByRole("button", { name: /Strafentypen anlegen/ }).click();
    await expect(page.getByRole("dialog", { name: "Neuen Strafentyp erstellen" })).toBeVisible();
    await page.getByLabel("Name *").fill("Pudel");
    await page.getByLabel("Preis (€) *").fill("0.20");
    await page.getByRole("button", { name: "Erstellen", exact: true }).click();
    await expect(page.getByRole("dialog")).toHaveCount(0);

    // Rollen per Vorschlag.
    await page.goto("/app");
    await card.getByRole("button", { name: "Vorschlag übernehmen" }).click();
    await expect(card.getByText("3 von 5 Pflichtschritten erledigt")).toBeVisible();

    // Spieler.
    await card.getByRole("button", { name: /Spieler anlegen/ }).click();
    const playerDialog = page.getByRole("dialog", { name: "Neuen Player erstellen" });
    await expect(playerDialog).toBeVisible();
    await playerDialog.getByLabel("Name").fill("Hans");
    await playerDialog.getByRole("button", { name: "Erstellen" }).click();
    await expect(page.getByRole("dialog")).toHaveCount(0);

    // Erster veröffentlichter Spieltag.
    await page.goto("/app");
    await card.getByRole("button", { name: /Ersten Spieltag veröffentlichen/ }).click();
    await expect(page).toHaveURL(/\/app\/gamedays\/new/);
    await page.getByLabel("Datum").fill(new Date().toISOString().slice(0, 10));
    await page.getByLabel("Vorläufig").uncheck();
    await page.getByRole("button", { name: "Speichern" }).click();
    await expect(page.getByText("Spieltag erstellt")).toBeVisible();

    await page.goto("/app");
    await expect(page.locator('[data-sidebar="sidebar"]').first()).toBeVisible();
    await expect(page.locator('[data-tour="own-balance"]')).toBeVisible();
    await expect(card).toHaveCount(0);
  });

  test("„Das bin ich“ legt den eigenen Spieler an", async ({ page }) => {
    await register(page, `e2e-self-${Date.now()}@example.com`);
    await createClubFromWelcome(page, "KC Selbst");
    await expect(tour(page)).toBeVisible({ timeout: 10000 });
    await tour(page).getByRole("button", { name: "Überspringen" }).click();

    const card = page.locator('[data-tour="setup-card"]');
    await card.getByRole("button", { name: "Vorschlag übernehmen" }).click();
    await card.getByRole("button", { name: /Eigener Spieler/ }).click();
    const dialog = page.getByRole("dialog", { name: "Neuen Player erstellen" });
    await expect(dialog.getByLabel("Das bin ich")).toBeChecked();
    await dialog.getByLabel("Name").fill("Ich");
    await dialog.getByRole("button", { name: "Erstellen" }).click();
    await expect(page.getByRole("dialog")).toHaveCount(0);

    // Danach wird „Das bin ich“ nicht mehr angeboten.
    await page.getByRole("button", { name: "Player hinzufügen" }).click();
    await expect(page.getByRole("dialog").getByLabel("Das bin ich")).toHaveCount(0);
  });

  test("Ausblenden versteckt die Ersteinrichtung", async ({ page }) => {
    await register(page, `e2e-dismiss-${Date.now()}@example.com`);
    await createClubFromWelcome(page, "KC Ausblenden");
    await expect(tour(page)).toBeVisible({ timeout: 10000 });
    await page.keyboard.press("Escape");
    await expect(tour(page)).toHaveCount(0);

    const card = page.locator('[data-tour="setup-card"]');
    await card.getByRole("button", { name: "Ausblenden" }).click();
    await expect(card).toHaveCount(0);
    await page.reload();
    await expect(page.locator('[data-tour="dashboard-cards"]')).toBeVisible();
    await expect(card).toHaveCount(0);
  });
});

test.describe("Rundgang", () => {
  test("Überspringen merkt den Rundgang, nach Neuladen kommt er nicht wieder", async ({ page }) => {
    await register(page, `e2e-skip-${Date.now()}@example.com`);
    await createClubFromWelcome(page, "KC Skip");
    await expect(tour(page)).toBeVisible({ timeout: 10000 });
    await tour(page).getByRole("button", { name: "Überspringen" }).click();
    await expect(tour(page)).toHaveCount(0);
    await expect.poll(async () => (await tourFlags(page)).tour).not.toBeNull();

    await page.reload();
    await expect(page.locator('[data-tour="dashboard-cards"]')).toBeVisible();
    await page.waitForTimeout(1500);
    await expect(tour(page)).toHaveCount(0);
  });

  test("Neustart über das Benutzermenü trotz gesetztem Merker", async ({ page }) => {
    const email = `e2e-restart-${Date.now()}@example.com`;
    await register(page, email);
    await createClubFromWelcome(page, "KC Neustart");
    await expect(tour(page)).toBeVisible({ timeout: 10000 });
    await page.keyboard.press("Escape");
    await expect(tour(page)).toHaveCount(0);

    await page.goto("/app/help");
    await page.getByRole("button", { name: new RegExp(email, "i") }).click();
    await page.getByRole("menuitem", { name: "Rundgang erneut starten" }).click();
    await expect(page).toHaveURL(/\/app$/);
    await expect(tourTitle(page)).toHaveText("Willkommen im Kegelmaster", { timeout: 10000 });
  });

  test("Bereits bekannter Rundgang: neuer Club zeigt nur die Sprechblase zur Ersteinrichtung", async ({ page }) => {
    await register(page, `e2e-hint-${Date.now()}@example.com`);
    // Rundgang wurde schon anderswo gesehen (z. B. nach einer Einladung).
    await page.evaluate(async () => {
      const csrf = document.cookie.split("; ").find((r) => r.startsWith("csrf_token="))?.split("=")[1] ?? "";
      await fetch("/api/auth/me", {
        method: "PATCH",
        credentials: "include",
        headers: { "Content-Type": "application/json", "X-CSRF-Token": csrf },
        body: JSON.stringify({ tour_seen: true }),
      });
    });
    await page.reload();
    await createClubFromWelcome(page, "KC Hinweis");

    const titles = await walkTour(page);
    expect(titles).toEqual(["Ersteinrichtung"]);
    await expect.poll(async () => (await tourFlags(page)).hint).not.toBeNull();
  });
});

test.describe("Mobil", () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test("Rundgang öffnet die Sidebar für Menü-Schritte", async ({ page }) => {
    await register(page, `e2e-mobile-${Date.now()}@example.com`);
    await page.getByRole("button", { name: "Club anlegen" }).click();
    await page.getByPlaceholder("Club Name").fill("KC Mobil");
    await page.getByRole("button", { name: "Erstellen", exact: true }).click();

    await expect(tourTitle(page)).toHaveText("Willkommen im Kegelmaster", { timeout: 10000 });
    await tour(page).locator(".driver-popover-next-btn").click();
    await expect(tourTitle(page)).toHaveText("Ihr Club");
    // The offcanvas sidebar is open and the club switcher is highlighted.
    await expect(page.locator('[data-mobile="true"][data-sidebar="sidebar"]')).toBeVisible();
    await expect(page.locator('[data-tour="club-switcher"].driver-active-element')).toBeVisible();

    const titles = await walkTour(page);
    expect(titles[titles.length - 1]).toBe("Ihr Benutzerkonto");
    await expect(page.locator('[data-mobile="true"][data-sidebar="sidebar"]')).toHaveCount(0);
  });
});

test.describe("Eingeladener Benutzer", () => {
  test("landet im richtigen Club und bekommt den Rundgang ohne Rollen", async ({ page, request }) => {
    const invitation = await request.get(`/api/invitations/${SEED_INVITE_TOKEN}`);
    test.skip(!invitation.ok(), "Seed-Einladung schon angenommen – Dev-Datenbank neu aufsetzen (make dev-fresh)");

    await page.goto(`/register?invite_token=${SEED_INVITE_TOKEN}`);
    await page.getByLabel("E-Mail").fill(SEED_INVITE_EMAIL);
    await page.getByLabel("Passwort", { exact: true }).fill("securepass123");
    await page.getByRole("button", { name: "Account erstellen" }).click();
    await expect(page).toHaveURL(/\/app$/, { timeout: 10000 });

    const titles = await walkTour(page);
    expect(titles[0]).toBe("Willkommen im Kegelmaster");
    expect(titles).not.toContain("Rollen");
    expect(titles).not.toContain("Ersteinrichtung");
    expect(titles).toContain("Ihr Benutzerkonto");
    await expect(page.locator('[data-tour="club-switcher"]')).toContainText(SEED_CLUB);

    await page.reload();
    await expect(page.locator('[data-tour="dashboard-cards"]')).toBeVisible();
    await page.waitForTimeout(1500);
    await expect(tour(page)).toHaveCount(0);
  });
});
