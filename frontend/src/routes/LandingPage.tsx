import { useEffect, type ComponentType, type ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import {
  ArrowRightIcon,
  BotIcon,
  Building2Icon,
  CalendarCheckIcon,
  CheckIcon,
  KeyRoundIcon,
  MailPlusIcon,
  ShieldCheckIcon,
  SmartphoneIcon,
  TrophyIcon,
  WalletIcon,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { useAuth } from "@/context/AuthContext";
import { useRegistrationMode } from "@/hooks/use-public-config";
import { cn } from "@/lib/utils";

type IconType = ComponentType<{ className?: string }>;

/** Pixelmaße der Dateien in /public/screenshots, damit der Platz vor dem Laden reserviert ist. */
const screenshotSizes: Record<string, [number, number]> = {
  dashboard: [2160, 1350],
  gamedays: [2160, 1350],
  transactions: [2160, 1350],
  penalties: [1644, 1005],
};

/** Zeigt je nach Theme den hellen oder dunklen Screenshot aus /public/screenshots. */
function Screenshot({
  name,
  alt,
  className,
  priority,
}: {
  name: string;
  alt: string;
  className?: string;
  priority?: boolean;
}) {
  const [width, height] = screenshotSizes[name] ?? [];
  const common = {
    alt,
    width,
    height,
    loading: priority ? ("eager" as const) : ("lazy" as const),
    decoding: "async" as const,
  };
  return (
    <>
      <img
        src={`/screenshots/${name}-light.jpg`}
        {...common}
        className={cn("block h-auto w-full dark:hidden", className)}
      />
      <img
        src={`/screenshots/${name}-dark.jpg`}
        {...common}
        className={cn("hidden h-auto w-full dark:block", className)}
      />
    </>
  );
}

/** Rahmen im Stil eines Browserfensters für Screenshots. */
function BrowserFrame({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div
      className={cn(
        "overflow-hidden rounded-xl border bg-card shadow-2xl shadow-black/10 ring-1 ring-black/5 dark:shadow-black/50 dark:ring-white/5",
        className,
      )}
    >
      <div className="flex items-center gap-1.5 border-b bg-muted/60 px-4 py-2.5">
        <span className="size-2.5 rounded-full bg-foreground/15" />
        <span className="size-2.5 rounded-full bg-foreground/15" />
        <span className="size-2.5 rounded-full bg-foreground/15" />
        <span className="ml-3 hidden h-5 flex-1 max-w-xs rounded-md bg-background/70 sm:block" />
      </div>
      {children}
    </div>
  );
}

function SectionHeading({ eyebrow, title, children }: { eyebrow: string; title: string; children?: ReactNode }) {
  return (
    <div className="mx-auto max-w-2xl text-center">
      <p className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">{eyebrow}</p>
      <h2 className="mt-3 text-3xl font-bold tracking-tight text-balance sm:text-4xl">{title}</h2>
      {children && <p className="mt-4 text-lg text-muted-foreground text-pretty">{children}</p>}
    </div>
  );
}

function Spotlight({
  icon: Icon,
  eyebrow,
  title,
  text,
  bullets,
  screenshot,
  alt,
  reverse,
  framed = true,
}: {
  icon: IconType;
  eyebrow: string;
  title: string;
  text: string;
  bullets: string[];
  screenshot: string;
  alt: string;
  reverse?: boolean;
  framed?: boolean;
}) {
  const image = framed ? (
    <BrowserFrame>
      <Screenshot name={screenshot} alt={alt} />
    </BrowserFrame>
  ) : (
    <div className="overflow-hidden rounded-xl border shadow-2xl shadow-black/10 ring-1 ring-black/5 dark:shadow-black/50 dark:ring-white/5">
      <Screenshot name={screenshot} alt={alt} />
    </div>
  );

  return (
    <div className="grid items-center gap-10 lg:grid-cols-12 lg:gap-16">
      <div className={cn("lg:col-span-5", reverse && "lg:order-last")}>
        <div className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
          <span className="flex size-8 items-center justify-center rounded-lg border bg-card text-foreground">
            <Icon className="size-4" aria-hidden />
          </span>
          {eyebrow}
        </div>
        <h3 className="mt-4 text-2xl font-bold tracking-tight text-balance sm:text-3xl">{title}</h3>
        <p className="mt-4 text-muted-foreground text-pretty">{text}</p>
        <ul className="mt-6 space-y-3">
          {bullets.map((b) => (
            <li key={b} className="flex gap-3 text-sm">
              <CheckIcon className="mt-0.5 size-4 shrink-0 text-emerald-600 dark:text-emerald-400" aria-hidden />
              <span>{b}</span>
            </li>
          ))}
        </ul>
      </div>
      <div className="lg:col-span-7">{image}</div>
    </div>
  );
}

const moreFeatures: { icon: IconType; title: string; text: string }[] = [
  {
    icon: Building2Icon,
    title: "Mehrere Klubs",
    text: "Kegelclub, Stammtisch, Betriebsgruppe – alles in einem Account, mit einem Klick gewechselt.",
  },
  {
    icon: ShieldCheckIcon,
    title: "Rollen & Rechte",
    text: "Kassenwart darf buchen, Mitglieder schauen nur rein. Rechte pro Rolle fein einstellbar.",
  },
  {
    icon: MailPlusIcon,
    title: "Mitglieder einladen",
    text: "Spieler per Einladungslink mit einem Account verknüpfen – jeder sieht sein eigenes Guthaben.",
  },
  {
    icon: TrophyIcon,
    title: "Wettbewerbe",
    text: "Pudelkönig, Pokal oder Pumpe: Eigene Wettbewerbe mit Sieger oder Verlierer pro Abend.",
  },
  {
    icon: BotIcon,
    title: "Hilfe-Assistent",
    text: "Fragen zur Bedienung? Der eingebaute Chat-Assistent erklärt die App direkt im Dashboard.",
  },
  {
    icon: SmartphoneIcon,
    title: "Am Handy dabei",
    text: "Läuft im Browser auf Laptop, Tablet und Smartphone – hell oder dunkel, ganz ohne Installation.",
  },
];

const steps = [
  {
    title: "Account anlegen",
    text: "In einer Minute registriert – nur E-Mail und Passwort.",
  },
  {
    title: "Club einrichten",
    text: "Spieler, Strafenkatalog und Wettbewerbe anlegen. Einmal, danach nie wieder.",
  },
  {
    title: "Kegeln & erfassen",
    text: "Am Abend Strafen eintragen, speichern – Kasse und Guthaben stimmen automatisch.",
  },
];

function LandingPage() {
  const { status } = useAuth();
  const isAuthed = status === "authenticated";
  // Ohne offene Registrierung verweisen die Aufrufe zum Handeln auf den Login.
  const canRegister = useRegistrationMode() === "open";
  const { hash } = useLocation();

  // Anker-Links aus dem Header (/#funktionen) auch nach Client-Navigation anspringen
  useEffect(() => {
    if (!hash) return;
    document.getElementById(hash.slice(1))?.scrollIntoView({ behavior: "smooth" });
  }, [hash]);

  const primaryCta = isAuthed ? (
    <Button asChild size="lg" className="h-11 rounded-full px-6">
      <Link to="/app">
        Zum Dashboard <ArrowRightIcon />
      </Link>
    </Button>
  ) : (
    <Button asChild size="lg" className="h-11 rounded-full px-6">
      <Link to={canRegister ? "/register" : "/login"}>
        {canRegister ? "Jetzt starten" : "Anmelden"} <ArrowRightIcon />
      </Link>
    </Button>
  );

  return (
    <div className="-mt-16 overflow-x-clip">
      {/* Hero */}
      <section className="relative isolate pt-32 sm:pt-40">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[46rem] bg-[radial-gradient(60%_50%_at_50%_0%,oklch(0.75_0.12_250/0.28),transparent_70%)] dark:bg-[radial-gradient(60%_50%_at_50%_0%,oklch(0.5_0.2_264/0.35),transparent_70%)]"
        />
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 -z-10 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:48px_48px] [mask-image:radial-gradient(ellipse_70%_50%_at_50%_0%,black,transparent_75%)] opacity-60"
        />

        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <div className="mx-auto max-w-3xl text-center animate-in fade-in slide-in-from-bottom-3 duration-700">
            <a
              href="#automatisierung"
              className="inline-flex items-center gap-2 rounded-full border bg-background/70 px-3 py-1 text-sm text-muted-foreground backdrop-blur transition-colors hover:text-foreground"
            >
              <span className="rounded-full bg-primary px-2 py-0.5 text-xs font-semibold text-primary-foreground">
                Neu
              </span>
              API-Tokens & MCP für KI-Agenten
              <ArrowRightIcon className="size-3.5" aria-hidden />
            </a>
            <h1 className="mt-6 text-4xl font-bold tracking-tight text-balance sm:text-6xl">
              Die Kegelkasse,{" "}
              <span className="bg-gradient-to-r from-foreground to-foreground/55 bg-clip-text text-transparent">
                endlich ohne Zettelwirtschaft.
              </span>
            </h1>
            <p className="mx-auto mt-6 max-w-2xl text-lg text-muted-foreground text-pretty">
              Kegelmaster erfasst Spielabende, Strafen und Wettbewerbe und führt automatisch die Kasse –
              mit Guthaben für jedes Mitglied und einem Dashboard für den ganzen Club.
            </p>
            <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
              {primaryCta}
              {!isAuthed && (
                <Button asChild size="lg" variant="outline" className="h-11 rounded-full bg-background/60 px-6">
                  <Link to="/login">Anmelden</Link>
                </Button>
              )}
            </div>
            <p className="mt-5 text-sm text-muted-foreground">
              Läuft im Browser · Für beliebig viele Clubs · Rollen für Vorstand & Mitglieder
            </p>
          </div>

          <div className="relative mx-auto mt-16 max-w-5xl animate-in fade-in slide-in-from-bottom-6 duration-1000 sm:mt-20">
            <div
              aria-hidden
              className="absolute -inset-x-8 -top-8 bottom-0 -z-10 rounded-[2rem] bg-gradient-to-b from-primary/10 to-transparent blur-2xl"
            />
            <BrowserFrame>
              <Screenshot
                priority
                name="dashboard"
                alt="Kegelmaster-Dashboard mit Guthaben, letztem Spieltag und Strafen-Diagramm"
              />
            </BrowserFrame>
            <div
              aria-hidden
              className="pointer-events-none absolute inset-x-0 bottom-0 h-1/3 bg-gradient-to-t from-background to-transparent"
            />
          </div>
        </div>
      </section>

      {/* Spotlights */}
      <section id="funktionen" className="scroll-mt-20 py-20 sm:py-28">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <SectionHeading eyebrow="Funktionen" title="Alles, was der Kassenwart braucht.">
            Vom Strafenkatalog bis zum Kassenstand – Kegelmaster nimmt euch die Rechnerei nach dem Kegeln ab.
          </SectionHeading>

          <div className="mt-16 space-y-24 sm:mt-20 sm:space-y-32">
            <Spotlight
              icon={CalendarCheckIcon}
              eyebrow="Spieltage"
              title="Den Abend in wenigen Minuten erfassen."
              text="Teilnehmer auswählen, Pudel, Kränze und Biere eintragen – die Summen pro Spieler rechnet Kegelmaster sofort aus."
              bullets={[
                "Eigener Strafenkatalog mit Beträgen pro Strafe",
                "Wettbewerbswerte im selben Formular erfassen",
                "Als vorläufig speichern und später veröffentlichen",
              ]}
              screenshot="penalties"
              alt="Erfassung von Strafen pro Spieler an einem Spieltag"
              framed={false}
            />
            <Spotlight
              reverse
              icon={WalletIcon}
              eyebrow="Kasse"
              title="Jeder Cent nachvollziehbar."
              text="Grundgebühr und Strafen werden beim Speichern automatisch gebucht. Einzahlungen und Ausgaben wie die Bahnmiete trägst du mit zwei Klicks nach."
              bullets={[
                "Guthaben jedes Spielers immer aktuell",
                "Filter nach Spieler, Typ und Beschreibung",
                "Clubkasse und Einzelkonten getrennt im Blick",
              ]}
              screenshot="transactions"
              alt="Transaktionsliste mit Grundgebühren, Strafen und Ausgaben"
            />
            <Spotlight
              icon={TrophyIcon}
              eyebrow="Überblick"
              title="Die ganze Saison auf einen Blick."
              text="Alle Spielabende mit Teilnehmern und Strafgebühren, dazu ein Dashboard mit Sieger, Verlierer und Strafen-Verlauf der letzten zwölf Monate."
              bullets={[
                "Chronik aller Spieltage",
                "Strafen-Statistik als Diagramm",
                "Eigene Wettbewerbsergebnisse für jedes Mitglied",
              ]}
              screenshot="gamedays"
              alt="Liste aller Spieltage mit Teilnehmern und Strafgebühren"
            />
          </div>
        </div>
      </section>

      {/* Weitere Features */}
      <section className="border-y bg-muted/30 py-20 sm:py-28">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <SectionHeading eyebrow="Und noch mehr" title="Gemacht für echte Clubs." />
          <div className="mt-14 grid gap-px overflow-hidden rounded-2xl border bg-border sm:grid-cols-2 lg:grid-cols-3">
            {moreFeatures.map(({ icon: Icon, title, text }) => (
              <div key={title} className="bg-background p-6 sm:p-8">
                <Icon className="size-5 text-foreground" aria-hidden />
                <h3 className="mt-4 font-semibold">{title}</h3>
                <p className="mt-2 text-sm text-muted-foreground text-pretty">{text}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* So geht's */}
      <section id="so-gehts" className="scroll-mt-20 py-20 sm:py-28">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <SectionHeading eyebrow="So geht's" title="In drei Schritten startklar." />
          <ol className="mt-14 grid gap-6 md:grid-cols-3">
            {steps.map((s, i) => (
              <li key={s.title} className="relative rounded-2xl border bg-card p-6">
                <span className="flex size-9 items-center justify-center rounded-full bg-primary text-sm font-bold text-primary-foreground">
                  {i + 1}
                </span>
                <h3 className="mt-5 font-semibold">{s.title}</h3>
                <p className="mt-2 text-sm text-muted-foreground">{s.text}</p>
              </li>
            ))}
          </ol>
        </div>
      </section>

      {/* Automatisierung */}
      <section id="automatisierung" className="scroll-mt-20 pb-20 sm:pb-28">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <div className="grid items-center gap-10 overflow-hidden rounded-3xl border bg-card p-6 sm:p-10 lg:grid-cols-2 lg:gap-16 lg:p-14">
            <div>
              <div className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
                <span className="flex size-8 items-center justify-center rounded-lg border bg-background text-foreground">
                  <KeyRoundIcon className="size-4" aria-hidden />
                </span>
                KI & API
              </div>
              <h2 className="mt-4 text-2xl font-bold tracking-tight text-balance sm:text-3xl">
                Lass deinen KI-Assistenten die Kasse führen.
              </h2>
              <p className="mt-4 text-muted-foreground text-pretty">
                Mit persönlichen API-Tokens und dem eingebauten MCP-Server können Skripte und KI-Agenten wie
                Claude Spieltage anlegen, Strafen eintragen oder Kontostände abfragen – mit genau deinen
                Rechten.
              </p>
              <ul className="mt-6 space-y-3 text-sm">
                {[
                  "REST-API mit OpenAPI-Dokumentation",
                  "MCP-Server unter /mcp für KI-Agenten",
                  "Tokens jederzeit im Dashboard widerrufbar",
                ].map((b) => (
                  <li key={b} className="flex gap-3">
                    <CheckIcon className="mt-0.5 size-4 shrink-0 text-emerald-600 dark:text-emerald-400" aria-hidden />
                    {b}
                  </li>
                ))}
              </ul>
            </div>
            <div className="min-w-0 rounded-xl border bg-zinc-950 p-5 font-mono text-[13px] leading-relaxed text-zinc-300 shadow-xl dark:bg-black/60">
              <div className="mb-4 flex items-center gap-1.5">
                <span className="size-2.5 rounded-full bg-zinc-700" />
                <span className="size-2.5 rounded-full bg-zinc-700" />
                <span className="size-2.5 rounded-full bg-zinc-700" />
              </div>
              <pre className="overflow-x-auto whitespace-pre">
                <code>
                  <span className="text-zinc-500"># Kontostand aller Spieler abfragen</span>
                  {"\n"}
                  <span className="text-emerald-400">curl</span> https://…/api/clubs/<span className="text-sky-300">$CLUB</span>/players \{"\n"}
                  {"  "}-H <span className="text-amber-300">"Authorization: Bearer kgm_…"</span>
                  {"\n\n"}
                  <span className="text-zinc-500"># oder einfach fragen:</span>
                  {"\n"}
                  <span className="text-sky-300">&gt;</span> Trag für heute bei Klaus 3 Pudel ein.
                  {"\n"}
                  <span className="text-emerald-400">✓</span> Spieltag 18.11. aktualisiert – Klaus: 0,60 €
                </code>
              </pre>
            </div>
          </div>
        </div>
      </section>

      {/* Abschluss-CTA */}
      <section className="pb-24 sm:pb-32">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <div className="relative isolate overflow-hidden rounded-3xl bg-primary px-6 py-14 text-center text-primary-foreground sm:px-12 sm:py-20">
            <div
              aria-hidden
              className="absolute inset-0 -z-10 bg-[radial-gradient(50%_80%_at_50%_0%,oklch(0.6_0.2_264/0.45),transparent)]"
            />
            <h2 className="text-3xl font-bold tracking-tight text-balance sm:text-4xl">
              Bereit für den nächsten Kegelabend?
            </h2>
            <p className="mx-auto mt-4 max-w-xl text-primary-foreground/75 text-pretty">
              Richte deinen Club heute ein – beim nächsten Treffen ist die Abrechnung in Minuten erledigt.
            </p>
            <div className="mt-8 flex flex-wrap justify-center gap-3">
              <Button asChild size="lg" variant="secondary" className="h-11 rounded-full px-6">
                <Link to={isAuthed ? "/app" : canRegister ? "/register" : "/login"}>
                  {isAuthed ? "Zum Dashboard" : canRegister ? "Account erstellen" : "Anmelden"} <ArrowRightIcon />
                </Link>
              </Button>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}

export default LandingPage;
