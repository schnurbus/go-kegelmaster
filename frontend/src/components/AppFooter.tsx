import { Link } from "react-router-dom";
import { useEffect, useState } from "react";

export function AppFooter({ className = "" }: { className?: string }) {
  const [version, setVersion] = useState<string | null>(null);
  const [buyMeACoffeeUrl, setBuyMeACoffeeUrl] = useState<string | null>(null);
  const year = new Date().getFullYear();

  useEffect(() => {
    fetch("/api/config/public")
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data) {
          if (typeof data.version === "string") setVersion(data.version);
          if (typeof data.buy_me_a_coffee_url === "string" && data.buy_me_a_coffee_url) {
            setBuyMeACoffeeUrl(data.buy_me_a_coffee_url);
          }
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
            <span>{version}</span>
          </>
        )}
        {buyMeACoffeeUrl != null && (
          <>
            <span aria-hidden>·</span>
            <a
              href={buyMeACoffeeUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="underline hover:no-underline"
            >
              Buy Me a Coffee
            </a>
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
