import { useEffect, useState } from "react";
import { Link, NavLink, useLocation } from "react-router-dom";
import { toast } from "sonner";
import { ArrowUpCircleIcon, LayoutDashboardIcon, LogOutIcon, MenuIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { ThemeToggle } from "@/components/theme-toggle";
import { useAuth } from "@/context/AuthContext";
import { useRegistrationMode } from "@/hooks/use-public-config";
import { cn } from "@/lib/utils";

const sections = [
  { href: "/#funktionen", label: "Funktionen" },
  { href: "/#so-gehts", label: "So geht's" },
  { href: "/#automatisierung", label: "KI & API" },
];

export function Brand({ className }: { className?: string }) {
  return (
    <Link to="/" className={cn("flex items-center gap-2 font-semibold tracking-tight", className)}>
      <span className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
        <ArrowUpCircleIcon className="size-5" aria-hidden />
      </span>
      <span className="text-lg">Kegelmaster</span>
    </Link>
  );
}

/** Kopfzeile der öffentlichen Seiten (Startseite, Login, Registrierung, Rechtliches). */
export function PublicSiteHeader() {
  const { status, user, logout } = useAuth();
  const isAuthenticated = status === "authenticated";
  const location = useLocation();
  // Ohne offene Registrierung führt "Jetzt starten" ins Leere: Link ausblenden.
  const canRegister = useRegistrationMode() === "open";
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const handleLogout = () => {
    logout().catch(() => toast.error("Abmelden fehlgeschlagen"));
  };

  const navLinkClass =
    "rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground";

  return (
    <header
      className={cn(
        "sticky top-0 z-40 w-full border-b transition-colors duration-200",
        scrolled
          ? "border-border bg-background/80 backdrop-blur-lg supports-[backdrop-filter]:bg-background/60"
          : "border-transparent bg-transparent",
      )}
    >
      <div className="mx-auto flex h-16 max-w-6xl items-center gap-6 px-4 sm:px-6">
        <Brand />

        <nav className="hidden items-center gap-1 md:flex" aria-label="Hauptnavigation">
          {sections.map((s) => (
            <Link key={s.href} to={s.href} className={navLinkClass}>
              {s.label}
            </Link>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <ThemeToggle />
          <div className="hidden items-center gap-2 md:flex">
            {isAuthenticated ? (
              <>
                <span className="max-w-48 truncate text-sm text-muted-foreground" title={user?.email}>
                  {user?.email}
                </span>
                <Button variant="ghost" size="sm" onClick={handleLogout}>
                  Abmelden
                </Button>
                <Button asChild size="sm">
                  <Link to="/app">Zum Dashboard</Link>
                </Button>
              </>
            ) : (
              <>
                <Button asChild variant="ghost" size="sm">
                  <NavLink to="/login">Anmelden</NavLink>
                </Button>
                {canRegister && (
                  <Button asChild size="sm">
                    <NavLink to="/register">Jetzt starten</NavLink>
                  </Button>
                )}
              </>
            )}
          </div>

          <Sheet key={location.pathname}>
            <SheetTrigger asChild>
              <Button variant="ghost" size="icon" className="md:hidden" aria-label="Menü öffnen">
                <MenuIcon className="size-5" />
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="w-72">
              <SheetHeader>
                <SheetTitle className="text-left">Menü</SheetTitle>
              </SheetHeader>
              <nav className="flex flex-col gap-1 px-4" aria-label="Mobile Navigation">
                {sections.map((s) => (
                  <SheetClose asChild key={s.href}>
                    <Link to={s.href} className="rounded-md px-3 py-2.5 text-base font-medium hover:bg-accent">
                      {s.label}
                    </Link>
                  </SheetClose>
                ))}
              </nav>
              <div className="mt-auto flex flex-col gap-2 border-t p-4">
                {isAuthenticated ? (
                  <>
                    <p className="truncate px-1 pb-1 text-sm text-muted-foreground">{user?.email}</p>
                    <Button asChild>
                      <Link to="/app">
                        <LayoutDashboardIcon /> Zum Dashboard
                      </Link>
                    </Button>
                    <Button variant="outline" onClick={handleLogout}>
                      <LogOutIcon /> Abmelden
                    </Button>
                  </>
                ) : (
                  <>
                    {canRegister && (
                      <Button asChild>
                        <Link to="/register">Jetzt starten</Link>
                      </Button>
                    )}
                    <Button asChild variant="outline">
                      <Link to="/login">Anmelden</Link>
                    </Button>
                  </>
                )}
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </header>
  );
}
