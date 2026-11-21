/** Parameters of an OAuth authorization request that the consent page passes back to the backend. */
export const AUTHORIZE_PARAMS = [
  "client_id",
  "redirect_uri",
  "response_type",
  "code_challenge",
  "code_challenge_method",
  "state",
  "scope",
  "resource",
] as const;

function readCsrfCookie(): string {
  return (
    document.cookie
      .split("; ")
      .find((row) => row.startsWith("csrf_token="))
      ?.split("=")[1] ?? ""
  );
}

/** Returns the CSRF token, fetching a fresh one when the cookie expired. */
export async function oauthCsrfToken(): Promise<string> {
  const token = readCsrfCookie();
  if (token) return token;
  const resp = await fetch("/api/auth/csrf-token", { credentials: "include" });
  if (!resp.ok) throw new Error("CSRF-Token konnte nicht geladen werden");
  const data: { csrf_token: string } = await resp.json();
  return data.csrf_token;
}

/** Reads the server's error message from a failed response. */
export async function errorMessage(resp: Response, fallback: string): Promise<string> {
  const data = await resp.json().catch(() => ({}));
  return typeof data.message === "string" && data.message ? data.message : fallback;
}
