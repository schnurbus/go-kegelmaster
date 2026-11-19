// Re-authentication for the admin area (KEG-9). Admin endpoints answer
// 401 {"error":"reauth_required"} when the last real login of an instance
// admin is older than ADMIN_REAUTH_MINUTES. Admin requests report that via
// notifyAdminReauthRequired(); AdminReauthDialog listens, asks for the
// password or a passkey and then lets the admin pages load again.

import { getPasskeyAssertion } from "@/lib/passkey";

export const REAUTH_REQUIRED = "reauth_required";

const EVENT_NAME = "kegelmaster:admin-reauth-required";

/** True for the 401 answer that asks an instance admin to confirm the login again. */
export function isReauthRequired(status: number, body: unknown): boolean {
  return (
    status === 401 &&
    typeof body === "object" &&
    body !== null &&
    (body as { error?: unknown }).error === REAUTH_REQUIRED
  );
}

export function notifyAdminReauthRequired(): void {
  window.dispatchEvent(new Event(EVENT_NAME));
}

/** Subscribes to reauth requests; returns the unsubscribe function. */
export function onAdminReauthRequired(callback: () => void): () => void {
  window.addEventListener(EVENT_NAME, callback);
  return () => window.removeEventListener(EVENT_NAME, callback);
}

async function csrfToken(): Promise<string> {
  const fromCookie = document.cookie
    .split("; ")
    .find((row) => row.startsWith("csrf_token="))
    ?.split("=")[1];
  if (fromCookie) return fromCookie;
  const resp = await fetch("/api/auth/csrf-token", { credentials: "include" });
  if (!resp.ok) throw new Error("CSRF-Token konnte nicht geladen werden");
  const data: { csrf_token: string } = await resp.json();
  return data.csrf_token;
}

async function postJSON(url: string, body?: unknown): Promise<Response> {
  const headers: Record<string, string> = { "X-CSRF-Token": await csrfToken() };
  if (body !== undefined) headers["Content-Type"] = "application/json";
  return fetch(url, {
    method: "POST",
    credentials: "include",
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

async function ensureOk(resp: Response, fallback: string): Promise<void> {
  if (resp.ok) return;
  const data = await resp.json().catch(() => ({}));
  throw new Error((data as { message?: string }).message || fallback);
}

/** Confirms the current session with the account password. */
export async function reauthWithPassword(password: string): Promise<void> {
  const resp = await postJSON("/api/auth/reauth", { password });
  await ensureOk(resp, "Bestätigung fehlgeschlagen");
}

/**
 * Confirms the current session with a passkey of the logged-in account.
 * Throws PasskeyCancelledError when the browser dialog was dismissed.
 */
export async function reauthWithPasskey(): Promise<void> {
  const begin = await postJSON("/api/auth/passkeys/login/begin");
  await ensureOk(begin, "Passkey-Bestätigung konnte nicht gestartet werden");
  const credential = await getPasskeyAssertion(await begin.json());
  const resp = await postJSON("/api/auth/passkeys/reauth", { credential });
  await ensureOk(resp, "Bestätigung mit Passkey fehlgeschlagen");
}
