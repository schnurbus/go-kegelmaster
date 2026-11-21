import * as React from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { CheckCircle2Icon, ChevronRightIcon, CircleIcon, SparklesIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { ClubSetup, SetupStep, SetupStepId } from "@/hooks/use-club-setup";
import { csrfTokenFromCookie } from "@/lib/csrf";
import { cn } from "@/lib/utils";

type StepText = { title: string; hint?: string; to: string };

/** Texte und Zielseiten der Schritte; ?neu=… öffnet dort den Dialog zum Anlegen. */
const STEP_TEXT: Record<SetupStepId, StepText> = {
  club: { title: "Club anlegen", to: "/app" },
  penalty_types: {
    title: "Strafentypen anlegen",
    hint: "Zum Beispiel Pudel oder Kranz, jeweils mit Preis.",
    to: "/app/penalty-types?neu=1",
  },
  roles: {
    title: "Rollen anlegen",
    hint: "Spieler brauchen eine Rolle. Sie legt Rechte und die Grundgebühr fest.",
    to: "/app/roles?neu=1",
  },
  players: { title: "Spieler anlegen", hint: "Alle, die mitkegeln.", to: "/app/players?neu=1" },
  first_game_day: {
    title: "Ersten Spieltag veröffentlichen",
    hint: "Erst ein veröffentlichter Spieltag bucht Grundgebühren und Strafen.",
    to: "/app/gamedays/new",
  },
  own_player: {
    title: "Eigener Spieler",
    hint: "Spielen Sie selbst mit? Dann legen Sie sich als Spieler an, um Saldo, Statistiken und Benachrichtigungen zu bekommen.",
    to: "/app/players?neu=ich",
  },
  invitations: {
    title: "Spieler einladen",
    hint: "Eingeladene sehen ihren Saldo und ihre Statistiken selbst.",
    to: "/app/players?einladen=1",
  },
  competitions: {
    title: "Wettbewerbe anlegen",
    hint: "Zum Beispiel Pokal oder Pumpe.",
    to: "/app/competitions?neu=1",
  },
};

type SetupCardProps = {
  clubId: string;
  setup: ClubSetup;
  onChanged: () => void;
};

/** Ersteinrichtung-Karte auf dem Dashboard (nur für den Besitzer). */
export function SetupCard({ clubId, setup, onChanged }: SetupCardProps) {
  const navigate = useNavigate();
  const [busy, setBusy] = React.useState<"dismiss" | "suggestion" | null>(null);

  const required = setup.steps.filter((s) => s.required);
  const optional = setup.steps.filter((s) => !s.required);
  const doneCount = required.filter((s) => s.done).length;

  const post = async (url: string, kind: "dismiss" | "suggestion", success: string) => {
    setBusy(kind);
    try {
      const resp = await fetch(url, {
        method: "POST",
        credentials: "include",
        headers: { "X-CSRF-Token": csrfTokenFromCookie() },
      });
      if (!resp.ok) {
        const data = await resp.json().catch(() => ({}));
        throw new Error(data.message || "Das hat nicht geklappt");
      }
      toast.success(success);
      onChanged();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Das hat nicht geklappt");
    } finally {
      setBusy(null);
    }
  };

  const renderStep = (step: SetupStep) => {
    const text = STEP_TEXT[step.id];
    const clickable = step.id !== "club";
    return (
      <li key={step.id} className="flex flex-col gap-1">
        <button
          type="button"
          disabled={!clickable}
          onClick={() => navigate(text.to)}
          className={cn(
            "flex flex-1 items-start gap-3 rounded-md p-2 text-left transition-colors",
            clickable && "hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
          )}
          aria-label={`${text.title}: ${step.done ? "erledigt" : "offen"}`}
        >
          {step.done ? (
            <CheckCircle2Icon className="mt-0.5 size-5 shrink-0 text-primary" aria-hidden />
          ) : (
            <CircleIcon className="mt-0.5 size-5 shrink-0 text-muted-foreground" aria-hidden />
          )}
          <span className="flex-1">
            <span className={cn("block font-medium", step.done && "text-muted-foreground line-through")}>
              {text.title}
            </span>
            {text.hint && !step.done && <span className="block text-sm text-muted-foreground">{text.hint}</span>}
          </span>
          {clickable && <ChevronRightIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden />}
        </button>
        {step.id === "roles" && !step.done && (
          <Button
            variant="outline"
            size="sm"
            className="ml-10 self-start"
            disabled={busy !== null}
            onClick={() =>
              post(`/api/clubs/${clubId}/roles/suggestion`, "suggestion", "Rollen Präsident, Kassenwart, Mitglied und Gast angelegt")
            }
          >
            <SparklesIcon />
            Vorschlag übernehmen
          </Button>
        )}
      </li>
    );
  };

  return (
    <Card className="mx-4 lg:mx-6" data-tour="setup-card" aria-labelledby="setup-card-title">
      <CardHeader>
        <CardTitle id="setup-card-title">Ersteinrichtung</CardTitle>
        <CardDescription>
          {doneCount} von {required.length} Pflichtschritten erledigt. Sind alle erledigt, verschwindet diese Karte.
        </CardDescription>
        <CardAction>
          <Button
            variant="ghost"
            size="sm"
            disabled={busy !== null}
            onClick={() => post(`/api/clubs/${clubId}/setup/dismiss`, "dismiss", "Ersteinrichtung ausgeblendet")}
          >
            Ausblenden
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-6 md:grid-cols-2">
        <div>
          <h3 className="mb-2 text-sm font-semibold">Pflichtschritte</h3>
          <ol className="flex flex-col gap-1">{required.map(renderStep)}</ol>
        </div>
        <div>
          <h3 className="mb-2 text-sm font-semibold">Optional</h3>
          <ul className="flex flex-col gap-1">{optional.map(renderStep)}</ul>
        </div>
      </CardContent>
    </Card>
  );
}
