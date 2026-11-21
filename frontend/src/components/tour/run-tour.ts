import { driver, type DriveStep } from "driver.js";
import "driver.js/dist/driver.css";

import type { TourContextInfo, TourStep } from "./tour-catalog";

export type RunTourOptions = {
  steps: TourStep[];
  ctx: TourContextInfo;
  /** Mobil liegt die Sidebar im Offcanvas und muss für Menü-Schritte geöffnet werden. */
  isMobile: boolean;
  setSidebarOpen: (open: boolean) => void;
  /** Wird genau einmal aufgerufen: beendet, übersprungen oder mit Esc geschlossen. */
  onEnd: () => void;
};

const SIDEBAR_ANIMATION_MS = 350;

function findTarget(target: string | undefined): Element | null {
  if (!target) return null;
  return document.querySelector(`[data-tour="${target}"]`);
}

/** Wartet, bis das Ziel im DOM ist (z. B. nach dem Öffnen der Sidebar), höchstens timeoutMs. */
function waitForTarget(target: string | undefined, timeoutMs: number): Promise<void> {
  return new Promise((resolve) => {
    const started = performance.now();
    const tick = () => {
      if (!target || findTarget(target) || performance.now() - started > timeoutMs) {
        // A short pause lets the sheet finish sliding in before driver.js measures.
        window.setTimeout(resolve, target && findTarget(target) ? SIDEBAR_ANIMATION_MS : 0);
        return;
      }
      window.requestAnimationFrame(tick);
    };
    tick();
  });
}

function wait(ms: number) {
  return new Promise<void>((resolve) => window.setTimeout(resolve, ms));
}

/** Steuert einen laufenden Rundgang von außen. */
export type TourHandle = {
  /** Beendet den Rundgang ohne onEnd (z. B. wenn das Dashboard verlassen wird). */
  cancel: () => void;
};

/**
 * Startet einen Rundgang mit driver.js: Sprechblasen mit Weiter, Zurück und
 * Überspringen, bedienbar per Tastatur (Pfeiltasten, Esc = Überspringen).
 * Fehlt ein Ziel, erscheint die Sprechblase mittig ohne Pfeil.
 */
export function runTour({ steps, ctx, isMobile, setSidebarOpen, onEnd }: RunTourOptions): TourHandle {
  if (steps.length === 0) {
    onEnd();
    return { cancel: () => {} };
  }

  let sidebarOpenedByTour = false;
  // Mobile: open the offcanvas sidebar for menu steps and close it again otherwise.
  const prepare = async (index: number) => {
    if (!isMobile) return;
    const step = steps[index];
    if (step?.inSidebar) {
      setSidebarOpen(true);
      sidebarOpenedByTour = true;
      await waitForTarget(step.target, 1000);
    } else {
      // Also closes a sidebar the user opened (e.g. restart from the user menu).
      setSidebarOpen(false);
      if (sidebarOpenedByTour || index === 0) await wait(SIDEBAR_ANIMATION_MS);
      sidebarOpenedByTour = false;
    }
  };

  const driveSteps: DriveStep[] = steps.map((step) => ({
    // A missing target yields a centered popover without arrow.
    element: (() => findTarget(step.target) ?? undefined) as () => Element,
    popover: {
      title: step.title,
      description: typeof step.description === "function" ? step.description(ctx) : step.description,
      side: step.side,
      align: "start",
    },
  }));

  let ended = false;
  // driver.js calls onDestroyed only when it still knows the active step, so
  // every way out (Fertig, Überspringen, Esc, click outside) ends up here.
  const finish = () => {
    if (sidebarOpenedByTour) setSidebarOpen(false);
    sidebarOpenedByTour = false;
    if (!ended) {
      ended = true;
      onEnd();
    }
  };
  const tour = driver({
    steps: driveSteps,
    animate: true,
    // Esc and a click outside end the tour like Überspringen. (driver.js'
    // "nextStep" would bypass onNextClick and with it the mobile sidebar.)
    allowClose: true,
    allowKeyboardControl: true,
    showProgress: steps.length > 1,
    progressText: "{{current}} von {{total}}",
    nextBtnText: "Weiter",
    prevBtnText: "Zurück",
    doneBtnText: "Fertig",
    // Every bubble has Weiter, Zurück and Überspringen (Zurück is disabled on the first one).
    showButtons: ["next", "previous"],
    popoverClass: "kegelmaster-tour",
    // Highlighted elements are not clickable: the tour only explains.
    disableActiveInteraction: true,
    stagePadding: 6,
    stageRadius: 8,
    onPopoverRender: (popover, { driver: d }) => {
      popover.wrapper.setAttribute("role", "dialog");
      popover.wrapper.setAttribute("aria-modal", "false");
      popover.wrapper.setAttribute("aria-labelledby", "driver-popover-title");
      popover.wrapper.setAttribute("aria-describedby", "driver-popover-description");
      const skip = document.createElement("button");
      skip.type = "button";
      skip.className = "kegelmaster-tour-skip";
      skip.textContent = "Überspringen";
      skip.addEventListener("click", () => {
        finish();
        d.destroy();
      });
      popover.footer.insertBefore(skip, popover.footer.firstChild);
      popover.nextButton.focus();
    },
    onNextClick: (_el, _step, { driver: d }) => {
      const index = d.getActiveIndex() ?? 0;
      if (d.isLastStep()) {
        finish();
        d.destroy();
        return;
      }
      void prepare(index + 1).then(() => d.moveNext());
    },
    onPrevClick: (_el, _step, { driver: d }) => {
      const index = d.getActiveIndex() ?? 0;
      if (index === 0) return;
      void prepare(index - 1).then(() => d.movePrevious());
    },
    onDestroyed: finish,
  });

  let cancelled = false;
  void prepare(0).then(() => {
    if (!cancelled) tour.drive(0);
  });
  return {
    cancel: () => {
      cancelled = true;
      ended = true;
      if (tour.isActive()) tour.destroy();
    },
  };
}
