import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { useLocation } from "react-router-dom";

type PageTitleContextValue = {
  title: string | null;
  setTitle: (title: string | null) => void;
};

const PageTitleContext = createContext<PageTitleContextValue | undefined>(
  undefined
);

export function PageTitleProvider({
  children,
  routeTitle,
}: {
  children: ReactNode;
  routeTitle: string;
}) {
  const [overrideTitle, setOverrideTitle] = useState<string | null>(null);
  const location = useLocation();

  // Beim Routenwechsel Override zurücksetzen, damit die neue Seite ihren Titel setzen kann
  useEffect(() => {
    setOverrideTitle(null);
  }, [location.pathname]);

  const setTitle = useCallback((title: string | null) => {
    setOverrideTitle(title);
  }, []);

  const title = overrideTitle ?? routeTitle;

  return (
    <PageTitleContext.Provider value={{ title, setTitle }}>
      {children}
    </PageTitleContext.Provider>
  );
}

export function usePageTitle() {
  const ctx = useContext(PageTitleContext);
  if (!ctx) {
    throw new Error("usePageTitle must be used within PageTitleProvider");
  }
  return ctx;
}
