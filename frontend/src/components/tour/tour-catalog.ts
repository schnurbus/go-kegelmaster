import type { Side } from "driver.js";

/** Was der Rundgang über den Benutzer und die Ansicht wissen muss. */
export type TourContextInfo = {
  clubName: string;
  isOwner: boolean;
  canSeeRoles: boolean;
  chatEnabled: boolean;
  /** Die Ersteinrichtung-Karte ist auf dem Dashboard zu sehen. */
  showSetupCard: boolean;
};

export type TourStep = {
  id: string;
  /** Wert des data-tour-Attributs am Ziel; ohne Ziel erscheint die Sprechblase mittig. */
  target?: string;
  /** Ziel liegt in der Sidebar (mobil muss sie dafür geöffnet werden). */
  inSidebar?: boolean;
  side?: Side;
  title: string;
  description: string | ((ctx: TourContextInfo) => string);
  /** Sichtbarkeitsbedingung (Rechte, Besitzer, Einstellungen). */
  visible?: (ctx: TourContextInfo) => boolean;
};

/**
 * Katalog aller Schritte des Rundgangs, in der Reihenfolge des vollen Rundgangs.
 * Teilabläufe (nur ein Hinweis) wählen einzelne Schritte per Kennung aus.
 */
export const TOUR_STEPS: TourStep[] = [
  {
    id: "welcome",
    title: "Willkommen im Kegelmaster",
    description:
      "Dieser kurze Rundgang zeigt Ihnen, wo Sie was finden. Mit „Weiter“ geht es zum nächsten Punkt, mit „Überspringen“ oder Esc beenden Sie ihn jederzeit.",
  },
  {
    id: "club-switcher",
    target: "club-switcher",
    inSidebar: true,
    side: "right",
    title: "Ihr Club",
    description: (ctx) =>
      `Hier sehen Sie den aktiven Club „${ctx.clubName}“ mit dem Kassenstand. Sind Sie in mehreren Clubs, wechseln Sie hier zwischen ihnen oder legen einen neuen Club an.`,
  },
  {
    id: "own-balance",
    target: "own-balance",
    side: "bottom",
    title: "Ihr Saldo",
    description:
      "Hier steht Ihr eigener Saldo, sobald Sie im Club einen Spieler haben. Ein negativer Betrag bedeutet, dass Sie der Kasse noch Geld schulden.",
  },
  {
    id: "dashboard-cards",
    target: "dashboard-cards",
    side: "bottom",
    title: "Auf einen Blick",
    description:
      "Die Karten und Diagramme des Dashboards zeigen den letzten Spieltag, Gewinner und Verlierer der Wettbewerbe, die Strafen und die letzten Transaktionen.",
  },
  {
    id: "setup-card",
    target: "setup-card",
    side: "bottom",
    title: "Ersteinrichtung",
    description:
      "Diese Karte führt Sie durch die Einrichtung Ihres Clubs: Strafentypen, Rollen, Spieler und der erste Spieltag. Ein Klick auf einen Schritt öffnet die passende Seite. Sind alle Pflichtschritte erledigt, verschwindet die Karte von selbst.",
    visible: (ctx) => ctx.isOwner && ctx.showSetupCard,
  },
  {
    id: "nav-statistics",
    target: "nav-statistics",
    inSidebar: true,
    side: "right",
    title: "Statistiken",
    description: "Ihre Strafen und Wettbewerbe im Zeitverlauf und im Vergleich mit dem Club.",
  },
  {
    id: "nav-gamedays",
    target: "nav-gamedays",
    inSidebar: true,
    side: "right",
    title: "Spieltage",
    description:
      "Alle Spielabende mit Strafen und Wettbewerben. Erst ein veröffentlichter Spieltag bucht Grundgebühren und Strafen.",
  },
  {
    id: "nav-transactions",
    target: "nav-transactions",
    inSidebar: true,
    side: "right",
    title: "Transaktionen",
    description: "Einzahlungen, Strafen, Trinkgeld und Ausgaben der Clubkasse.",
  },
  {
    id: "nav-roles",
    target: "nav-roles",
    inSidebar: true,
    side: "right",
    title: "Rollen",
    description:
      "Rollen bündeln Rechte. Jeder Spieler hat genau eine Rolle; sie legt fest, was er sehen und ändern darf und ob er Grundgebühr zahlt.",
    visible: (ctx) => ctx.canSeeRoles,
  },
  {
    id: "nav-help",
    target: "nav-help",
    inSidebar: true,
    side: "right",
    title: "Hilfe",
    description:
      "Die Hilfeseite erklärt alle Bereiche. Dort können Sie diesen Rundgang auch erneut starten.",
  },
  {
    id: "help-chat",
    target: "help-chat",
    side: "left",
    title: "Hilfe-Chat",
    description: "Fragen zur Bedienung können Sie hier dem Assistenten stellen.",
    visible: (ctx) => ctx.chatEnabled,
  },
  {
    id: "nav-user",
    target: "nav-user",
    inSidebar: true,
    side: "right",
    title: "Ihr Account",
    description:
      "Im Benutzermenü finden Sie Ihren Account. Dort können Sie eine E-Mail bei abgerechneten Spieltagen einschalten – sie ist anfangs aus. Hier starten Sie auch den Rundgang erneut.",
  },
];

/**
 * full: voller Rundgang; setup-hint: nur die Sprechblase zur Ersteinrichtung-Karte;
 * club-switcher-hint: nur der Hinweis auf den Club-Switcher nach einer Einladung.
 */
export type TourFlow = "full" | "setup-hint" | "club-switcher-hint";

const CLUB_SWITCHER_HINT: TourStep = {
  id: "club-switcher-hint",
  target: "club-switcher",
  inSidebar: true,
  side: "right",
  title: "Neuer Club",
  description: (ctx) =>
    `Sie sind jetzt auch im Club „${ctx.clubName}“. Über den Club-Switcher wechseln Sie zwischen Ihren Clubs.`,
};

/** Die Schritte eines Ablaufs, gefiltert nach dem, was der Benutzer sehen darf. */
export function stepsForFlow(flow: TourFlow, ctx: TourContextInfo): TourStep[] {
  const visible = (s: TourStep) => !s.visible || s.visible(ctx);
  switch (flow) {
    case "full":
      return TOUR_STEPS.filter(visible);
    case "setup-hint":
      return TOUR_STEPS.filter((s) => s.id === "setup-card" && visible(s));
    case "club-switcher-hint":
      return [CLUB_SWITCHER_HINT];
  }
}
