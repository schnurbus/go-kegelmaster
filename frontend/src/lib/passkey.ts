// Thin wrapper around the WebAuthn browser API for passkey sign-in.
// The server sends options as JSON ({ publicKey: ... }) and expects the
// result of PublicKeyCredential.toJSON() back.

export type Passkey = {
  id: string;
  name: string;
  transports: string[];
  synced: boolean;
  last_used_at: string | null;
  created_at: string;
};

/** Thrown when the user dismissed the browser dialog or it was aborted. */
export class PasskeyCancelledError extends Error {
  constructor() {
    super("Vorgang abgebrochen");
    this.name = "PasskeyCancelledError";
  }
}

/** True when the browser supports passkeys including the JSON helpers we rely on. */
export function passkeysSupported(): boolean {
  return (
    typeof window !== "undefined" &&
    typeof window.PublicKeyCredential !== "undefined" &&
    typeof PublicKeyCredential.parseRequestOptionsFromJSON === "function" &&
    typeof PublicKeyCredential.parseCreationOptionsFromJSON === "function"
  );
}

/** True when passkeys can be offered in the autofill of the e-mail field. */
export async function conditionalMediationAvailable(): Promise<boolean> {
  if (!passkeysSupported() || typeof PublicKeyCredential.isConditionalMediationAvailable !== "function") {
    return false;
  }
  try {
    return await PublicKeyCredential.isConditionalMediationAvailable();
  } catch {
    return false;
  }
}

function translateError(error: unknown, ceremony: "create" | "get"): Error {
  if (error instanceof DOMException) {
    switch (error.name) {
      case "NotAllowedError":
      case "AbortError":
        return new PasskeyCancelledError();
      case "InvalidStateError":
        if (ceremony === "create") {
          return new Error("Für dieses Gerät ist bereits ein Passkey registriert.");
        }
        break;
      case "SecurityError":
        return new Error("Passkeys sind auf dieser Adresse nicht erlaubt.");
    }
  }
  return error instanceof Error ? error : new Error("Passkey-Vorgang fehlgeschlagen");
}

export async function getPasskeyAssertion(
  options: { publicKey: PublicKeyCredentialRequestOptionsJSON },
  { conditional = false, signal }: { conditional?: boolean; signal?: AbortSignal } = {},
): Promise<PublicKeyCredentialJSON> {
  try {
    const credential = await navigator.credentials.get({
      publicKey: PublicKeyCredential.parseRequestOptionsFromJSON(options.publicKey),
      mediation: conditional ? "conditional" : "optional",
      signal,
    });
    if (!(credential instanceof PublicKeyCredential)) {
      throw new PasskeyCancelledError();
    }
    return credential.toJSON();
  } catch (error) {
    throw translateError(error, "get");
  }
}

export async function createPasskey(options: {
  publicKey: PublicKeyCredentialCreationOptionsJSON;
}): Promise<PublicKeyCredentialJSON> {
  try {
    const credential = await navigator.credentials.create({
      publicKey: PublicKeyCredential.parseCreationOptionsFromJSON(options.publicKey),
    });
    if (!(credential instanceof PublicKeyCredential)) {
      throw new PasskeyCancelledError();
    }
    return credential.toJSON();
  } catch (error) {
    throw translateError(error, "create");
  }
}

/** A readable default name for a new passkey, derived from the platform. */
export function suggestPasskeyName(): string {
  const ua = navigator.userAgent;
  if (/iPhone/.test(ua)) return "iPhone";
  if (/iPad/.test(ua)) return "iPad";
  if (/Android/.test(ua)) return "Android";
  if (/Mac OS X/.test(ua)) return "Mac";
  if (/Windows/.test(ua)) return "Windows";
  if (/Linux/.test(ua)) return "Linux";
  return "Passkey";
}
