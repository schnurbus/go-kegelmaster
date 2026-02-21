import { Link } from "react-router-dom";
import { useEffect, useState } from "react";

function getHealthzUrl(): string {
  const base = import.meta.env.VITE_BACKEND_URL;
  if (base) {
    return `${String(base).replace(/\/$/, "")}/healthz`;
  }
  return "/healthz";
}

export function AppFooter({ className = "" }: { className?: string }) {
  const [version, setVersion] = useState<string | null>(null);
  const year = new Date().getFullYear();

  useEffect(() => {
    const url = getHealthzUrl();
    fetch(url)
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data && typeof data.version === "string") {
          setVersion(data.version);
        }
      })
      .catch(() => setVersion(null));
  }, []);

  return (
    <footer className={className || "app-footer"}>
      <small className="flex flex-wrap items-center justify-center gap-x-3 gap-y-1 text-muted-foreground">
        <span>© {year} Schnurbus</span>
        {version != null && (
          <>
            <span aria-hidden>·</span>
            <span>v{version}</span>
          </>
        )}
        <span aria-hidden>·</span>
        <Link to="/impressum" className="underline hover:no-underline">
          Impressum
        </Link>
        <span aria-hidden>·</span>
        <Link to="/datenschutz" className="underline hover:no-underline">
          Datenschutz
        </Link>
      </small>
    </footer>
  );
}
