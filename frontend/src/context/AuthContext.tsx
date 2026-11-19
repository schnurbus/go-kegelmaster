import type { ReactNode } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

import { getPasskeyAssertion } from "@/lib/passkey";

type AuthUser = {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
  password_changed_at: string | null;
  notify_gameday_settled: boolean;
  /** Instance admin (operator): may open the read-only admin area (/api/admin). */
  is_instance_admin: boolean;
};

type AuthStatus = "loading" | "authenticated" | "unauthenticated";

type AuthContextValue = {
  user: AuthUser | null;
  status: AuthStatus;
  csrfToken: string | null;
  login: (payload: AuthCredentials) => Promise<AuthUser>;
  register: (payload: AuthCredentials) => Promise<AuthUser>;
  logout: () => Promise<void>;
  refreshCsrf: () => Promise<string>;
  changePassword: (payload: ChangePasswordPayload) => Promise<AuthUser>;
  updateSettings: (payload: UserSettingsPayload) => Promise<AuthUser>;
  loginWithPasskey: (payload: PasskeyLoginOptions) => Promise<AuthUser>;
};

type PasskeyLoginOptions = {
  /** A function is evaluated after the user picked a passkey (autofill may wait a long time). */
  remember_me?: boolean | (() => boolean);
  /** Offer passkeys in the autofill of the e-mail field instead of a dialog. */
  conditional?: boolean;
  signal?: AbortSignal;
};

type ChangePasswordPayload = {
  current_password: string;
  new_password: string;
};

type UserSettingsPayload = {
  notify_gameday_settled?: boolean;
};

type AuthCredentials = {
  email: string;
  password: string;
  remember_me?: boolean;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [status, setStatus] = useState<AuthStatus>("loading");
  const [csrfToken, setCsrfToken] = useState<string | null>(null);

  const fetchCsrf = useCallback(async () => {
    const resp = await fetch("/api/auth/csrf-token", {
      credentials: "include",
    });
    if (!resp.ok) {
      throw new Error("CSRF Token konnte nicht geladen werden");
    }
    const data: { csrf_token: string } = await resp.json();
    setCsrfToken(data.csrf_token);
    return data.csrf_token;
  }, []);

  const ensureCsrf = useCallback(async () => {
    if (csrfToken) {
      return csrfToken;
    }
    return fetchCsrf();
  }, [csrfToken, fetchCsrf]);

  const handleAuthSuccess = useCallback(
    async (resp: Response) => {
      if (!resp.ok) {
        throw new Error(await extractError(resp));
      }
      const data: AuthUser = await resp.json();
      setUser(data);
      setStatus("authenticated");
      await fetchCsrf(); // rotate token after login/register
      return data;
    },
    [fetchCsrf],
  );

  const login = useCallback(
    async (payload: AuthCredentials) => {
      setStatus("loading");
      const token = await ensureCsrf();
      const resp = await fetch("/api/auth/login", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": token,
        },
        body: JSON.stringify(payload),
      });
      try {
        return await handleAuthSuccess(resp);
      } catch (error) {
        setStatus("unauthenticated");
        throw error;
      }
    },
    [ensureCsrf, handleAuthSuccess],
  );

  const register = useCallback(
    async (payload: AuthCredentials) => {
      setStatus("loading");
      const token = await ensureCsrf();
      const resp = await fetch("/api/auth/register", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": token,
        },
        body: JSON.stringify(payload),
      });
      try {
        return await handleAuthSuccess(resp);
      } catch (error) {
        setStatus("unauthenticated");
        throw error;
      }
    },
    [ensureCsrf, handleAuthSuccess],
  );

  const logout = useCallback(async () => {
    setStatus("loading");
    const token = await ensureCsrf();
    const resp = await fetch("/api/auth/logout", {
      method: "POST",
      credentials: "include",
      headers: {
        "X-CSRF-Token": token,
      },
    });
    if (!resp.ok) {
      setStatus("authenticated");
      throw new Error(await extractError(resp));
    }
    setUser(null);
    setStatus("unauthenticated");
    await fetchCsrf();
  }, [ensureCsrf, fetchCsrf]);

  const loginWithPasskey = useCallback(
    async ({ remember_me = false, conditional = false, signal }: PasskeyLoginOptions) => {
      // Read the CSRF cookie right before each request: the login page can
      // mount (and start the autofill request) while a logout is still
      // rotating the token, so the token in state may already be stale.
      const currentCsrf = async () => readCsrfCookie() ?? (await fetchCsrf());
      const beginResp = await fetch("/api/auth/passkeys/login/begin", {
        method: "POST",
        credentials: "include",
        headers: { "X-CSRF-Token": await currentCsrf() },
        signal,
      });
      if (!beginResp.ok) {
        throw new Error(await extractError(beginResp));
      }
      // Waits for the user (dialog or autofill); throws PasskeyCancelledError when dismissed.
      const credential = await getPasskeyAssertion(await beginResp.json(), { conditional, signal });

      setStatus("loading");
      const resp = await fetch("/api/auth/passkeys/login/finish", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": await currentCsrf(),
        },
        body: JSON.stringify({
          credential,
          remember_me: typeof remember_me === "function" ? remember_me() : remember_me,
        }),
      });
      try {
        return await handleAuthSuccess(resp);
      } catch (error) {
        setStatus("unauthenticated");
        throw error;
      }
    },
    [fetchCsrf, handleAuthSuccess],
  );

  const changePassword = useCallback(
    async (payload: ChangePasswordPayload) => {
      const token = await ensureCsrf();
      const resp = await fetch("/api/auth/change-password", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": token,
        },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        throw new Error(await extractError(resp));
      }
      const data: AuthUser = await resp.json();
      setUser(data);
      await fetchCsrf(); // the server reissued the session and rotated the CSRF cookie
      return data;
    },
    [ensureCsrf, fetchCsrf],
  );

  const updateSettings = useCallback(
    async (payload: UserSettingsPayload) => {
      const token = await ensureCsrf();
      const resp = await fetch("/api/auth/me", {
        method: "PATCH",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": token,
        },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        throw new Error(await extractError(resp));
      }
      const data: AuthUser = await resp.json();
      setUser(data);
      return data;
    },
    [ensureCsrf],
  );

  useEffect(() => {
    const bootstrap = async () => {
      try {
        await fetchCsrf();
        const resp = await fetch("/api/auth/me", {
          credentials: "include",
        });
        if (!resp.ok) {
          setStatus("unauthenticated");
          setUser(null);
          return;
        }
        const data: AuthUser = await resp.json();
        setUser(data);
        setStatus("authenticated");
      } catch {
        setStatus("unauthenticated");
      }
    };
    bootstrap();
  }, [fetchCsrf]);

  const value = useMemo(
    () => ({
      user,
      status,
      csrfToken,
      login,
      register,
      logout,
      refreshCsrf: fetchCsrf,
      changePassword,
      updateSettings,
      loginWithPasskey,
    }),
    [user, status, csrfToken, login, register, logout, fetchCsrf, changePassword, updateSettings, loginWithPasskey],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}

function readCsrfCookie(): string | null {
  const value = document.cookie
    .split("; ")
    .find((row) => row.startsWith("csrf_token="))
    ?.split("=")[1];
  return value ? value : null;
}

async function extractError(resp: Response) {
  try {
    const data = await resp.json();
    if (typeof data === "string") {
      return data;
    }
    return data?.error ?? data?.message ?? resp.statusText;
  } catch {
    try {
      const text = await resp.text();
      if (text.trim()) return text;
    } catch {
      // ignore
    }
    return resp.statusText;
  }
}

