import { csrfTokenFromCookie } from "@/lib/csrf";

/** Antwort von `POST /api/invitations/{token}/accept`. */
export type AcceptedInvitation = { player_id: string; club_id: string };

/**
 * Nimmt eine Einladung für den angemeldeten Benutzer an. Wirft einen Error mit
 * einer anzeigbaren deutschen Meldung, wenn das nicht klappt.
 */
export async function acceptInvitation(token: string): Promise<AcceptedInvitation> {
  const csrfToken = csrfTokenFromCookie();
  if (!csrfToken) {
    throw new Error("CSRF-Token fehlt. Bitte Seite neu laden.");
  }

  const response = await fetch(`/api/invitations/${token}/accept`, {
    method: "POST",
    headers: { "X-CSRF-Token": csrfToken },
    credentials: "include",
  });
  const responseData = await response.json().catch(() => ({}));

  if (!response.ok) {
    if (response.status === 404) {
      throw new Error("Einladung nicht gefunden");
    }
    if (response.status === 410) {
      throw new Error("Einladung ist abgelaufen oder wurde bereits akzeptiert");
    }
    throw new Error(responseData.message || "Fehler beim Akzeptieren der Einladung");
  }
  return responseData as AcceptedInvitation;
}
