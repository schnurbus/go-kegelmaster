/** CSRF-Token aus dem Cookie (Double-Submit), für schreibende Anfragen. */
export function csrfTokenFromCookie(): string {
  return (
    document.cookie
      .split("; ")
      .find((row) => row.startsWith("csrf_token="))
      ?.split("=")[1] ?? ""
  );
}
