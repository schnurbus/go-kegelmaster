import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useClub } from "@/context/ClubContext";
import { usePermissions } from "@/hooks/use-permissions";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { toast } from "sonner";
import { Loader2Icon, SaveIcon, TrashIcon, ArrowLeftIcon, XIcon, ChevronsUpDownIcon, CheckIcon } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";

const GAMEDAY_HIDE_INACTIVE_KEY_PREFIX = "kegelmaster_gameday_hide_inactive_";

function loadGamedayHideInactive(clubId: string): boolean {
  try {
    const raw = localStorage.getItem(GAMEDAY_HIDE_INACTIVE_KEY_PREFIX + clubId);
    return raw === "true";
  } catch {
    return false;
  }
}

function saveGamedayHideInactive(clubId: string, value: boolean) {
  try {
    localStorage.setItem(GAMEDAY_HIDE_INACTIVE_KEY_PREFIX + clubId, String(value));
  } catch {
    // ignore
  }
}

interface GameDayDetail {
  game_day: {
    id: string;
    club_id: string;
    date: string;
    notes: string;
    is_draft: boolean;
    created_at: string;
    updated_at: string;
  };
  participants: Array<{
    participant: {
      id: string;
      game_day_id: string;
      player_id: string;
      player_name: string;
      created_at: string;
    };
    fees: Array<{
      id: string;
      game_day_participant_id: string;
      penalty_type_id: string;
      penalty_type_name: string;
      penalty_type_description: string;
      penalty_type_price: number;
      count: number;
      quantity_scale: number;
      quantity: number;
      created_at: string;
      updated_at: string;
    }>;
    competition_values: Array<{
      id: string;
      game_day_participant_id: string;
      competition_id: string;
      value: number;
      created_at: string;
      updated_at: string;
    }>;
  }>;
}

interface Player {
  id: string;
  name: string;
  club_id: string;
  balance: number;
  start_balance: number;
  role_id: string | null;
  user_id: string | null;
  inactive: boolean;
  created_at: string;
  updated_at: string;
}

interface PenaltyType {
  id: string;
  club_id: string;
  name: string;
  description: string;
  price: number;
  display_order: number;
  allows_decimal_quantity: boolean;
  created_at: string;
  updated_at: string;
}

interface Competition {
  id: string;
  club_id: string;
  name: string;
  scoring_type: string;
  is_gender_specific: boolean;
  display_order: number;
  created_at: string;
  updated_at: string;
}

interface OrderedParticipant {
  player_id: string;
  player_name: string;
  created_at: string;
}

function GameDayDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { activeClub } = useClub();
  const { canCreate, canUpdate, canDelete } = usePermissions(activeClub?.id ?? null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isSaving, setIsSaving] = React.useState(false);
  
  // Fix: Derive isNew from id parameter instead of state
  const isNew = id === "new";
  const canEditGameday = isNew ? canCreate("game_days") : canUpdate("game_days");

  const [date, setDate] = React.useState("");
  const [notes, setNotes] = React.useState("");
  const [isDraft, setIsDraft] = React.useState(true); // default true for new gamedays
  const [detail, setDetail] = React.useState<GameDayDetail | null>(null);

  // State for participant and fee management
  const [allPlayers, setAllPlayers] = React.useState<Player[]>([]);
  const [allPenaltyTypes, setAllPenaltyTypes] = React.useState<PenaltyType[]>([]);
  const [allCompetitions, setAllCompetitions] = React.useState<Competition[]>([]);
  const [orderedParticipants, setOrderedParticipants] = React.useState<OrderedParticipant[]>([]);
  const [selectedPlayerIds, setSelectedPlayerIds] = React.useState<Set<string>>(new Set());
  const [feeInputs, setFeeInputs] = React.useState<Map<string, Map<string, number>>>(new Map());
  const [competitionValueInputs, setCompetitionValueInputs] = React.useState<Map<string, Map<string, number>>>(new Map());
  const [isLoadingPlayers, setIsLoadingPlayers] = React.useState(false);
  const [isLoadingPenaltyTypes, setIsLoadingPenaltyTypes] = React.useState(false);
  const [isLoadingCompetitions, setIsLoadingCompetitions] = React.useState(false);
  const [isSavingFees, setIsSavingFees] = React.useState(false);
  const [isSavingCompetitionValues, setIsSavingCompetitionValues] = React.useState(false);
  
  // Combobox state
  const [comboboxOpen, setComboboxOpen] = React.useState(false);
  const [selectedPlayerToAdd, setSelectedPlayerToAdd] = React.useState<string>("");

  // Hide inactive players in selection (per-club, persisted)
  const [hideInactive, setHideInactive] = React.useState<boolean>(false);

  React.useEffect(() => {
    if (activeClub) {
      setHideInactive(loadGamedayHideInactive(activeClub.id));
    }
  }, [activeClub?.id]);

  // Transaction summary state
  const [transactionSummary, setTransactionSummary] = React.useState<{
    base_fee_total: number;
    base_fee_count: number;
    penalty_fee_total: number;
    penalty_fee_count: number;
    total: number;
  } | null>(null);
  const [isLoadingTransactions, setIsLoadingTransactions] = React.useState(false);

  const fetchCSRFToken = React.useCallback(async () => {
    const response = await fetch("/api/auth/csrf-token", {
      credentials: "include",
    });
    if (!response.ok) {
      throw new Error("Failed to fetch CSRF token");
    }
    const data = await response.json();
    return data.csrf_token;
  }, []);

  const fetchGameDay = React.useCallback(async () => {
    if (!activeClub || isNew) {
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    try {
      const response = await fetch(`/api/clubs/${activeClub.id}/gamedays/${id}`, {
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Fehler beim Laden des Spieltags");
      }

      const data: GameDayDetail = await response.json();
      setDetail(data);
      setDate(data.game_day.date);
      setNotes(data.game_day.notes);
      setIsDraft(data.game_day.is_draft);
    } catch (error) {
      console.error("Error fetching game day:", error);
      toast.error("Fehler beim Laden des Spieltags");
    } finally {
      setIsLoading(false);
    }
  }, [activeClub, id, isNew]);

  const fetchPlayers = React.useCallback(async () => {
    if (!activeClub) return;

    setIsLoadingPlayers(true);
    try {
      const response = await fetch(`/api/clubs/${activeClub.id}/players`, {
        credentials: "include",
      });

      if (!response.ok) throw new Error("Failed to fetch players");

      const data: Player[] = await response.json();
      setAllPlayers(data);
    } catch (error) {
      console.error("Error fetching players:", error);
      toast.error("Fehler beim Laden der Spieler");
    } finally {
      setIsLoadingPlayers(false);
    }
  }, [activeClub]);

  const fetchPenaltyTypes = React.useCallback(async () => {
    if (!activeClub) return;

    setIsLoadingPenaltyTypes(true);
    try {
      const response = await fetch(`/api/clubs/${activeClub.id}/penalty-types`, {
        credentials: "include",
      });

      if (!response.ok) throw new Error("Failed to fetch penalty types");

      const data: PenaltyType[] = await response.json();
      setAllPenaltyTypes(data.sort((a, b) => a.display_order - b.display_order));
    } catch (error) {
      console.error("Error fetching penalty types:", error);
      toast.error("Fehler beim Laden der Strafarten");
    } finally {
      setIsLoadingPenaltyTypes(false);
    }
  }, [activeClub]);

  const fetchTransactionSummary = React.useCallback(async () => {
    if (!activeClub || isNew || !id) return;

    setIsLoadingTransactions(true);
    try {
      const response = await fetch(
        `/api/clubs/${activeClub.id}/gamedays/${id}/transaction-summary`,
        {
          credentials: "include",
        }
      );

      if (!response.ok) throw new Error("Failed to fetch transaction summary");

      const data = await response.json();
      setTransactionSummary(data);
    } catch (error) {
      console.error("Error fetching transaction summary:", error);
      // Don't show error toast - this is optional data
      setTransactionSummary(null);
    } finally {
      setIsLoadingTransactions(false);
    }
  }, [activeClub, id, isNew]);

  React.useEffect(() => {
    fetchGameDay();
  }, [fetchGameDay]);

  const fetchCompetitions = React.useCallback(async () => {
    if (!activeClub) return;
    setIsLoadingCompetitions(true);
    try {
      const response = await fetch(`/api/clubs/${activeClub.id}/competitions`, {
        credentials: "include",
      });
      if (!response.ok) throw new Error("Failed to fetch competitions");
      const data: Competition[] = await response.json();
      setAllCompetitions(data.sort((a, b) => a.display_order - b.display_order));
    } catch (error) {
      console.error("Error fetching competitions:", error);
      toast.error("Fehler beim Laden der Wettbewerbe");
    } finally {
      setIsLoadingCompetitions(false);
    }
  }, [activeClub]);

  // Fetch players, penalty types and competitions when not creating new
  React.useEffect(() => {
    if (!isNew && activeClub) {
      fetchPlayers();
      fetchPenaltyTypes();
      fetchCompetitions();
      fetchTransactionSummary();
    }
  }, [isNew, activeClub, fetchPlayers, fetchPenaltyTypes, fetchCompetitions]);

  // Initialize state from backend detail (maintains insertion order)
  React.useEffect(() => {
    if (!detail) return;

    // Build ordered participants array from backend response (already ordered by created_at)
    const participants = detail.participants.map(p => ({
      player_id: p.participant.player_id,
      player_name: p.participant.player_name,
      created_at: p.participant.created_at,
    }));
    setOrderedParticipants(participants);

    // Build selectedPlayerIds Set for quick lookup
    const participantIds = new Set(participants.map(p => p.player_id));
    setSelectedPlayerIds(participantIds);

    // Build fee inputs map (use quantity for display when scale > 1)
    const fees = new Map<string, Map<string, number>>();
    const compValues = new Map<string, Map<string, number>>();
    detail.participants.forEach(p => {
      const playerFees = new Map<string, number>();
      p.fees.forEach(f => {
        playerFees.set(f.penalty_type_id, f.quantity ?? f.count);
      });
      fees.set(p.participant.player_id, playerFees);

      const playerCompValues = new Map<string, number>();
      (p.competition_values || []).forEach(cv => {
        playerCompValues.set(cv.competition_id, cv.value);
      });
      compValues.set(p.participant.player_id, playerCompValues);
    });
    setFeeInputs(fees);
    setCompetitionValueInputs(compValues);
  }, [detail]);

  const handleSave = async () => {
    if (!activeClub) return;

    if (!date) {
      toast.error("Bitte geben Sie ein Datum ein");
      return;
    }

    setIsSaving(true);
    try {
      const csrfToken = await fetchCSRFToken();

      const url = isNew
        ? `/api/clubs/${activeClub.id}/gamedays`
        : `/api/clubs/${activeClub.id}/gamedays/${id}`;

      const method = isNew ? "POST" : "PUT";

      const response = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrfToken,
        },
        credentials: "include",
        body: JSON.stringify({
          date,
          notes,
          is_draft: isDraft,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Speichern");
      }

      const savedGameDay = await response.json();
      toast.success(
        isNew ? "Spieltag erstellt" : "Spieltag aktualisiert"
      );

      if (isNew) {
        navigate(`/app/gamedays/${savedGameDay.id}`);
      } else {
        fetchGameDay();
      }
    } catch (error: any) {
      console.error("Error saving game day:", error);
      toast.error(error.message || "Fehler beim Speichern");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!activeClub || isNew || !id) return;

    if (!confirm("Möchten Sie diesen Spieltag wirklich löschen?")) {
      return;
    }

    try {
      const csrfToken = await fetchCSRFToken();

      const response = await fetch(`/api/clubs/${activeClub.id}/gamedays/${id}`, {
        method: "DELETE",
        headers: {
          "X-CSRF-Token": csrfToken,
        },
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Fehler beim Löschen");
      }

      navigate("/app/gamedays", { replace: true });
      toast.success("Spieltag gelöscht");
    } catch (error: any) {
      console.error("Error deleting game day:", error);
      toast.error(error.message || "Fehler beim Löschen");
    }
  };

  const handleAddPlayer = async (playerId: string) => {
    if (!activeClub || !id || isNew || !playerId) return;

    try {
      const csrfToken = await fetchCSRFToken();

      const response = await fetch(
        `/api/clubs/${activeClub.id}/gamedays/${id}/participants`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": csrfToken,
          },
          credentials: "include",
          body: JSON.stringify({ player_id: playerId }),
        }
      );

      if (!response.ok) throw new Error("Failed to add participant");

      toast.success("Teilnehmer hinzugefügt");
      
      // Clear selection
      setSelectedPlayerToAdd("");
      
      // Refresh to get updated order from backend
      await fetchGameDay();
    } catch (error: any) {
      console.error("Error adding participant:", error);
      toast.error(error.message || "Fehler beim Hinzufügen");
    }
  };

  const handleRemovePlayer = async (playerId: string) => {
    if (!activeClub || !id || isNew) return;

    // Check if player has fees
    const playerFees = feeInputs.get(playerId);
    const hasFees = playerFees && Array.from(playerFees.values()).some(count => count > 0);

    if (hasFees && !confirm("Dieser Spieler hat Strafen. Wirklich entfernen?")) {
      return;
    }

    try {
      const csrfToken = await fetchCSRFToken();

      const response = await fetch(
        `/api/clubs/${activeClub.id}/gamedays/${id}/participants/${playerId}`,
        {
          method: "DELETE",
          headers: {
            "X-CSRF-Token": csrfToken,
          },
          credentials: "include",
        }
      );

      if (!response.ok) throw new Error("Failed to remove participant");

      toast.success("Teilnehmer entfernt");
      
      // Refresh to get updated state from backend
      await fetchGameDay();
    } catch (error: any) {
      console.error("Error removing participant:", error);
      toast.error(error.message || "Fehler beim Entfernen");
    }
  };

  const handleFeeChange = (playerId: string, penaltyTypeId: string, count: number) => {
    setFeeInputs(prev => {
      const newMap = new Map(prev);
      const playerFees = new Map(newMap.get(playerId) || []);

      if (count <= 0 || Number.isNaN(count)) {
        playerFees.delete(penaltyTypeId);
      } else {
        playerFees.set(penaltyTypeId, count);
      }

      newMap.set(playerId, playerFees);
      return newMap;
    });
  };

  const handleSaveFees = async () => {
    if (!activeClub || !id || isNew) return;

    setIsSavingFees(true);
    try {
      const csrfToken = await fetchCSRFToken();

      // Save fees for each participant: send ALL penalty types with count (0 = delete fee)
      const savePromises = orderedParticipants.map(async (participant) => {
        const playerId = participant.player_id;
        const playerFees = feeInputs.get(playerId) || new Map();

        const fees = allPenaltyTypes.map((pt) => ({
          penalty_type_id: pt.id,
          count: playerFees.has(pt.id) ? Number(playerFees.get(pt.id)) : 0,
        }));

        const response = await fetch(
          `/api/clubs/${activeClub.id}/gamedays/${id}/participants/${playerId}/fees`,
          {
            method: "PUT",
            headers: {
              "Content-Type": "application/json",
              "X-CSRF-Token": csrfToken,
            },
            credentials: "include",
            body: JSON.stringify({ fees }),
          }
        );

        if (!response.ok) {
          throw new Error(`Fehler beim Speichern für ${participant.player_name}`);
        }
      });

      await Promise.all(savePromises);

      toast.success("Strafen gespeichert");

      // Refresh detail and transaction summary to get updated data
      await fetchGameDay();
      await fetchTransactionSummary();
    } catch (error: any) {
      console.error("Error saving fees:", error);
      toast.error(error.message || "Fehler beim Speichern der Strafen");
    } finally {
      setIsSavingFees(false);
    }
  };

  const handleCompetitionValueChange = (playerId: string, competitionId: string, value: number) => {
    setCompetitionValueInputs(prev => {
      const newMap = new Map(prev);
      const playerValues = new Map(newMap.get(playerId) || []);
      playerValues.set(competitionId, value);
      newMap.set(playerId, playerValues);
      return newMap;
    });
  };

  const handleSaveCompetitionValues = async () => {
    if (!activeClub || !id || isNew) return;
    setIsSavingCompetitionValues(true);
    try {
      const csrfToken = await fetchCSRFToken();
      const savePromises = orderedParticipants.map(async (participant) => {
        const playerId = participant.player_id;
        const playerValues = competitionValueInputs.get(playerId) || new Map();
        const values = Array.from(playerValues.entries()).map(([competition_id, value]) => ({
          competition_id,
          value,
        }));
        const response = await fetch(
          `/api/clubs/${activeClub.id}/gamedays/${id}/participants/${playerId}/competition-values`,
          {
            method: "PUT",
            headers: {
              "Content-Type": "application/json",
              "X-CSRF-Token": csrfToken,
            },
            credentials: "include",
            body: JSON.stringify({ values }),
          }
        );
        if (!response.ok) throw new Error(`Fehler beim Speichern für ${participant.player_name}`);
      });
      await Promise.all(savePromises);
      toast.success("Wettbewerbs-Werte gespeichert");
      await fetchGameDay();
    } catch (error: unknown) {
      console.error("Error saving competition values:", error);
      toast.error(error instanceof Error ? error.message : "Fehler beim Speichern der Wettbewerbs-Werte");
    } finally {
      setIsSavingCompetitionValues(false);
    }
  };

  // Filter available players (not yet added), optionally hide inactive, sort alphabetically
  const availablePlayers = React.useMemo(() => {
    let list = allPlayers.filter((player) => !selectedPlayerIds.has(player.id));
    if (hideInactive) {
      list = list.filter((p) => !p.inactive);
    }
    return [...list].sort((a, b) => a.name.localeCompare(b.name, "de"));
  }, [allPlayers, selectedPlayerIds, hideInactive]);

  if (!activeClub) {
    return (
      <div className="p-6">
        <p>Bitte wählen Sie einen Klub aus.</p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="p-6">
        <p>Laden...</p>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => navigate("/app/gamedays")}
            >
              <ArrowLeftIcon className="h-4 w-4" />
            </Button>
            <div>
              <h1 className="text-3xl font-bold">
                {isNew ? "Neuer Spieltag" : "Spieltag bearbeiten"}
              </h1>
              <p className="text-muted-foreground">
                {isNew ? "Erstellen Sie einen neuen Spieltag" : "Bearbeiten Sie den Spieltag"}
              </p>
            </div>
          </div>
          <div className="flex gap-2">
            {!isNew && canDelete("game_days") && (
              <Button variant="destructive" onClick={handleDelete}>
                <TrashIcon className="mr-2 h-4 w-4" />
                Löschen
              </Button>
            )}
            {canEditGameday && (
              <Button onClick={handleSave} disabled={isSaving}>
                {isSaving ? (
                  <Loader2Icon className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <SaveIcon className="mr-2 h-4 w-4" />
                )}
                Speichern
              </Button>
            )}
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Grunddaten</CardTitle>
            <CardDescription>
              Datum und Notizen für diesen Spieltag
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="date">Datum</Label>
              <Input
                id="date"
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                required
                disabled={!canEditGameday}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="notes">Notizen</Label>
              <Textarea
                id="notes"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                placeholder="Optionale Notizen zum Spieltag..."
                rows={3}
                disabled={!canEditGameday}
              />
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox
                id="gameday-is-draft"
                checked={isDraft}
                onCheckedChange={(checked) => setIsDraft(checked === true)}
                disabled={!canEditGameday}
              />
              <Label htmlFor="gameday-is-draft" className="text-sm font-normal cursor-pointer">
                Vorläufig
              </Label>
            </div>
            {isDraft && (
              <p className="text-sm text-muted-foreground">
                Bei vorläufigen Spieltagen werden keine Gebühren (Grundgebühr, Strafen) gebucht. Heben Sie „Vorläufig“ beim Bearbeiten auf, um alle Gebühren anzulegen.
              </p>
            )}
          </CardContent>
        </Card>

        {!isNew && canUpdate("game_days") && (
          <Card>
            <CardHeader>
              <CardTitle>Teilnehmer hinzufügen</CardTitle>
              <CardDescription>
                Wählen Sie Spieler aus der Liste aus. Die Reihenfolge wird in der Tabelle unten beibehalten.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-2">
                <Checkbox
                  id="gameday-hide-inactive"
                  checked={hideInactive}
                  onCheckedChange={(checked) => {
                    const value = checked === true;
                    setHideInactive(value);
                    if (activeClub) saveGamedayHideInactive(activeClub.id, value);
                  }}
                />
                <Label htmlFor="gameday-hide-inactive" className="cursor-pointer text-sm font-normal">
                  Inaktive ausblenden
                </Label>
              </div>
              {/* Add Player Section with Combobox */}
              <div className="space-y-2">
                <Label htmlFor="player-select">Spieler hinzufügen</Label>
                <Popover open={comboboxOpen} onOpenChange={setComboboxOpen}>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      role="combobox"
                      aria-expanded={comboboxOpen}
                      className="w-auto justify-between min-w-[200px]"
                      disabled={isLoadingPlayers || availablePlayers.length === 0}
                    >
                      {selectedPlayerToAdd
                        ? allPlayers.find((player) => player.id === selectedPlayerToAdd)?.name
                        : "Spieler auswählen..."}
                      <ChevronsUpDownIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-[300px] p-0" align="start">
                    <Command>
                      <CommandInput placeholder="Spieler suchen..." />
                      <CommandList>
                        <CommandEmpty>
                          {selectedPlayerIds.size === allPlayers.length
                            ? "Alle Spieler wurden bereits hinzugefügt"
                            : "Keine Spieler gefunden"}
                        </CommandEmpty>
                        <CommandGroup>
                          {availablePlayers.map((player) => (
                            <CommandItem
                              key={player.id}
                              value={player.name}
                              onSelect={() => {
                                handleAddPlayer(player.id);
                                setComboboxOpen(false);
                              }}
                            >
                              <CheckIcon
                                className={cn(
                                  "mr-2 h-4 w-4",
                                  selectedPlayerToAdd === player.id ? "opacity-100" : "opacity-0"
                                )}
                              />
                              {player.name}
                            </CommandItem>
                          ))}
                        </CommandGroup>
                      </CommandList>
                    </Command>
                  </PopoverContent>
                </Popover>
              </div>
            </CardContent>
          </Card>
        )}

        {!isNew && (
          <Card>
            <CardHeader>
              <CardTitle>Strafen & Wettbewerbe</CardTitle>
              <CardDescription>
                {canUpdate("game_days")
                  ? "Strafen und Wettbewerbs-Werte pro Teilnehmer erfassen"
                  : "Sie haben keine Berechtigung, Strafen oder Wettbewerbe zu bearbeiten."}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {orderedParticipants.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <p>Noch keine Teilnehmer hinzugefügt</p>
                  <p className="text-sm mt-1">
                    {canUpdate("game_days")
                      ? "Wählen Sie oben Spieler aus, um sie hinzuzufügen"
                      : "Für diesen Spieltag sind keine Teilnehmer erfasst."}
                  </p>
                </div>
              ) : canUpdate("game_days") ? (
                <Tabs defaultValue="strafen" className="w-full">
                  <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="strafen">Strafen</TabsTrigger>
                    <TabsTrigger value="wettbewerbe">Wettbewerbe</TabsTrigger>
                  </TabsList>
                  <TabsContent value="strafen" className="space-y-4 mt-4">
                    {isLoadingPenaltyTypes ? (
                      <p className="text-muted-foreground">Lade Strafarten...</p>
                    ) : allPenaltyTypes.length === 0 ? (
                      <p className="text-muted-foreground">Keine Strafarten vorhanden</p>
                    ) : (
                      <>
                        <div className="overflow-x-auto">
                          <table className="w-full border-collapse">
                            <thead>
                              <tr className="border-b">
                                <th className="text-left p-2 font-semibold">Spieler</th>
                                {allPenaltyTypes.map(pt => (
                                  <th key={pt.id} className="text-center p-2 font-semibold min-w-[100px]">
                                    <div className="text-sm">{pt.name}</div>
                                    <div className="text-xs text-muted-foreground font-normal">
                                      {(pt.price / 100).toFixed(2)} €
                                    </div>
                                  </th>
                                ))}
                                <th className="text-right p-2 font-semibold">Gesamt</th>
                                <th className="text-center p-2 font-semibold w-[60px]"></th>
                              </tr>
                            </thead>
                            <tbody>
                              {orderedParticipants.map((participant) => {
                                const playerFees = feeInputs.get(participant.player_id) || new Map();
                                const total = allPenaltyTypes.reduce((sum, pt) => {
                                  const qty = playerFees.get(pt.id) || 0;
                                  return sum + (qty * pt.price);
                                }, 0);
                                return (
                                  <tr key={participant.player_id} className="border-b hover:bg-muted/50">
                                    <td className="p-2 font-medium">{participant.player_name}</td>
                                    {allPenaltyTypes.map(pt => {
                                      const currentCount = playerFees.get(pt.id) ?? 0;
                                      const existingFee = detail?.participants
                                        .find(p => p.participant.player_id === participant.player_id)
                                        ?.fees.find(f => f.penalty_type_id === pt.id);
                                      const hasSnapshot = existingFee && existingFee.penalty_type_price !== pt.price;
                                      const allowDecimal = pt.allows_decimal_quantity === true;
                                      return (
                                        <td key={pt.id} className="p-2 text-center">
                                          <div className="flex flex-col items-center gap-1">
                                            <Input
                                              type="number"
                                              min="0"
                                              step={allowDecimal ? "0.01" : "1"}
                                              value={currentCount === 0 ? "" : currentCount}
                                              onChange={(e) => handleFeeChange(participant.player_id, pt.id, parseFloat(e.target.value) || 0)}
                                              className="w-20 text-center"
                                            />
                                            {hasSnapshot && (
                                              <span className="text-xs text-orange-600" title="Preis hat sich geändert">
                                                ⚠ {(existingFee!.penalty_type_price / 100).toFixed(2)} €
                                              </span>
                                            )}
                                          </div>
                                        </td>
                                      );
                                    })}
                                    <td className="p-2 text-right font-semibold">
                                      {(total / 100).toFixed(2)} €
                                    </td>
                                    <td className="p-2 text-center">
                                      <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => handleRemovePlayer(participant.player_id)}
                                        className="h-8 w-8 p-0"
                                        title="Teilnehmer entfernen"
                                      >
                                        <XIcon className="h-4 w-4" />
                                      </Button>
                                    </td>
                                  </tr>
                                );
                              })}
                            </tbody>
                          </table>
                        </div>
                        <div className="flex justify-end pt-4">
                          <Button onClick={handleSaveFees} disabled={isSavingFees}>
                            {isSavingFees ? (
                              <Loader2Icon className="mr-2 h-4 w-4 animate-spin" />
                            ) : (
                              <SaveIcon className="mr-2 h-4 w-4" />
                            )}
                            Strafen speichern
                          </Button>
                        </div>
                      </>
                    )}
                  </TabsContent>
                  <TabsContent value="wettbewerbe" className="space-y-4 mt-4">
                    {isLoadingCompetitions ? (
                      <p className="text-muted-foreground">Lade Wettbewerbe...</p>
                    ) : allCompetitions.length === 0 ? (
                      <p className="text-muted-foreground">Keine Wettbewerbe vorhanden</p>
                    ) : (
                      <>
                        <div className="overflow-x-auto">
                          <table className="w-full border-collapse">
                            <thead>
                              <tr className="border-b">
                                <th className="text-left p-2 font-semibold">Spieler</th>
                                {allCompetitions.map(c => (
                                  <th key={c.id} className="text-center p-2 font-semibold min-w-[80px]">
                                    <div className="text-sm">{c.name}</div>
                                  </th>
                                ))}
                                <th className="text-center p-2 font-semibold w-[60px]"></th>
                              </tr>
                            </thead>
                            <tbody>
                              {orderedParticipants.map((participant) => {
                                const playerValues = competitionValueInputs.get(participant.player_id) || new Map();
                                return (
                                  <tr key={participant.player_id} className="border-b hover:bg-muted/50">
                                    <td className="p-2 font-medium">{participant.player_name}</td>
                                    {allCompetitions.map(c => {
                                      const value = playerValues.get(c.id) ?? "";
                                      return (
                                        <td key={c.id} className="p-2 text-center">
                                          <Input
                                            type="number"
                                            min="0"
                                            step="1"
                                            value={value}
                                            onChange={(e) => handleCompetitionValueChange(
                                              participant.player_id,
                                              c.id,
                                              parseInt(e.target.value, 10) || 0
                                            )}
                                            className="w-20 text-center"
                                          />
                                        </td>
                                      );
                                    })}
                                    <td className="p-2 text-center">
                                      <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => handleRemovePlayer(participant.player_id)}
                                        className="h-8 w-8 p-0"
                                        title="Teilnehmer entfernen"
                                      >
                                        <XIcon className="h-4 w-4" />
                                      </Button>
                                    </td>
                                  </tr>
                                );
                              })}
                            </tbody>
                          </table>
                        </div>
                        <div className="flex justify-end pt-4">
                          <Button onClick={handleSaveCompetitionValues} disabled={isSavingCompetitionValues}>
                            {isSavingCompetitionValues ? (
                              <Loader2Icon className="mr-2 h-4 w-4 animate-spin" />
                            ) : (
                              <SaveIcon className="mr-2 h-4 w-4" />
                            )}
                            Wettbewerbs-Werte speichern
                          </Button>
                        </div>
                      </>
                    )}
                  </TabsContent>
                </Tabs>
              ) : (
                <Tabs defaultValue="strafen" className="w-full">
                  <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="strafen">Strafen</TabsTrigger>
                    <TabsTrigger value="wettbewerbe">Wettbewerbe</TabsTrigger>
                  </TabsList>
                  <TabsContent value="strafen" className="space-y-4 mt-4">
                    {isLoadingPenaltyTypes ? (
                      <p className="text-muted-foreground">Lade Strafarten...</p>
                    ) : allPenaltyTypes.length === 0 ? (
                      <p className="text-muted-foreground">Keine Strafarten vorhanden</p>
                    ) : (
                      <div className="overflow-x-auto">
                        <table className="w-full border-collapse">
                          <thead>
                            <tr className="border-b">
                              <th className="text-left p-2 font-semibold">Spieler</th>
                              {allPenaltyTypes.map(pt => (
                                <th key={pt.id} className="text-center p-2 font-semibold min-w-[100px]">
                                  <div className="text-sm">{pt.name}</div>
                                  <div className="text-xs text-muted-foreground font-normal">
                                    {(pt.price / 100).toFixed(2)} €
                                  </div>
                                </th>
                              ))}
                              <th className="text-right p-2 font-semibold">Gesamt</th>
                            </tr>
                          </thead>
                          <tbody>
                            {orderedParticipants.map((participant) => {
                              const partData = detail?.participants.find(
                                p => p.participant.player_id === participant.player_id
                              );
                              const total = (partData?.fees ?? []).reduce(
                                (sum, f) => sum + f.quantity * f.penalty_type_price,
                                0
                              );
                              return (
                                <tr key={participant.player_id} className="border-b hover:bg-muted/50">
                                  <td className="p-2 font-medium">{participant.player_name}</td>
                                  {allPenaltyTypes.map(pt => {
                                    const fee = partData?.fees.find(f => f.penalty_type_id === pt.id);
                                    const qty = fee?.quantity ?? 0;
                                    return (
                                      <td key={pt.id} className="p-2 text-center">
                                        {qty === 0 ? "–" : qty}
                                      </td>
                                    );
                                  })}
                                  <td className="p-2 text-right font-semibold">
                                    {(total / 100).toFixed(2)} €
                                  </td>
                                </tr>
                              );
                            })}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </TabsContent>
                  <TabsContent value="wettbewerbe" className="space-y-4 mt-4">
                    {isLoadingCompetitions ? (
                      <p className="text-muted-foreground">Lade Wettbewerbe...</p>
                    ) : allCompetitions.length === 0 ? (
                      <p className="text-muted-foreground">Keine Wettbewerbe vorhanden</p>
                    ) : (
                      <div className="overflow-x-auto">
                        <table className="w-full border-collapse">
                          <thead>
                            <tr className="border-b">
                              <th className="text-left p-2 font-semibold">Spieler</th>
                              {allCompetitions.map(c => (
                                <th key={c.id} className="text-center p-2 font-semibold min-w-[80px]">
                                  <div className="text-sm">{c.name}</div>
                                </th>
                              ))}
                            </tr>
                          </thead>
                          <tbody>
                            {orderedParticipants.map((participant) => {
                              const partData = detail?.participants.find(
                                p => p.participant.player_id === participant.player_id
                              );
                              return (
                                <tr key={participant.player_id} className="border-b hover:bg-muted/50">
                                  <td className="p-2 font-medium">{participant.player_name}</td>
                                  {allCompetitions.map(c => {
                                    const cv = partData?.competition_values.find(v => v.competition_id === c.id);
                                    const value = cv?.value ?? 0;
                                    return (
                                      <td key={c.id} className="p-2 text-center">
                                        {value === 0 ? "–" : value}
                                      </td>
                                    );
                                  })}
                                </tr>
                              );
                            })}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </TabsContent>
                </Tabs>
              )}
            </CardContent>
          </Card>
        )}

        {!isNew && transactionSummary && (
          <Card>
            <CardHeader>
              <CardTitle>Transaktionsübersicht</CardTitle>
              <CardDescription>
                Zusammenfassung aller Einnahmen für diesen Spieltag
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoadingTransactions ? (
                <p className="text-muted-foreground">Laden...</p>
              ) : (
                <div className="space-y-3">
                  <div className="flex items-center justify-between py-2 border-b">
                    <div>
                      <p className="font-medium">Grundgebühren</p>
                      <p className="text-sm text-muted-foreground">
                        {transactionSummary.base_fee_count} Transaktionen
                      </p>
                    </div>
                    <p className="text-lg font-semibold">
                      {(transactionSummary.base_fee_total / 100).toFixed(2)} €
                    </p>
                  </div>
                  <div className="flex items-center justify-between py-2 border-b">
                    <div>
                      <p className="font-medium">Strafgebühren</p>
                      <p className="text-sm text-muted-foreground">
                        {transactionSummary.penalty_fee_count} Transaktionen
                      </p>
                    </div>
                    <p className="text-lg font-semibold">
                      {(transactionSummary.penalty_fee_total / 100).toFixed(2)} €
                    </p>
                  </div>
                  <div className="flex items-center justify-between py-2 pt-4">
                    <p className="text-lg font-bold">Gesamteinnahmen</p>
                    <p className="text-2xl font-bold text-green-600">
                      {(transactionSummary.total / 100).toFixed(2)} €
                    </p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        )}
    </div>
  );
}

export default GameDayDetailPage;
