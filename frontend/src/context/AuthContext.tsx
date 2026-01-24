import type { ReactNode } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

type AuthUser = {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
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
};

type AuthCredentials = {
  email: string;
  password: string;
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
    }),
    [user, status, csrfToken, login, register, logout, fetchCsrf],
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

async function extractError(resp: Response) {
  try {
    const data = await resp.json();
    if (typeof data === "string") {
      return data;
    }
    return data?.error ?? data?.message ?? resp.statusText;
  } catch {
    return resp.statusText;
  }
}

