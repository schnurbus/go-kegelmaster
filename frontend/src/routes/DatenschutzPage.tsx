import { useEffect, useState } from "react";

function getLegalUrl(path: string): string {
  const base = import.meta.env.VITE_BACKEND_URL;
  if (base) {
    return `${String(base).replace(/\/$/, "")}/api/legal/${path}`;
  }
  return `/api/legal/${path}`;
}

export default function DatenschutzPage() {
  const [html, setHtml] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    fetch(getLegalUrl("datenschutz"))
      .then((res) => (res.ok ? res.text() : Promise.reject(new Error(res.statusText))))
      .then(setHtml)
      .catch(() => setError(true))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="page mx-auto w-full max-w-3xl px-4 py-8">
      <h1 className="mb-6 text-2xl font-semibold">Datenschutzerklärung</h1>
      {loading && <p className="text-muted-foreground">Laden …</p>}
      {error && (
        <p className="text-destructive">Datenschutzerklärung konnte nicht geladen werden.</p>
      )}
      {!loading && !error && (
        <div
          className="prose prose-sm dark:prose-invert max-w-none"
          dangerouslySetInnerHTML={{ __html: html }}
        />
      )}
    </div>
  );
}
