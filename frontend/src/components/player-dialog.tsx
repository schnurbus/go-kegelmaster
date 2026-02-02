import * as React from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import type {
  Player,
  Role,
  Gender,
  CreatePlayerRequest,
  UpdatePlayerRequest,
} from "@/types/player";
import { euroToCents } from "@/types/player";

type PlayerDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  player?: Player | null;
  clubId: string;
  roles: Role[];
  onSuccess: () => void;
};

export function PlayerDialog({
  open,
  onOpenChange,
  player,
  clubId,
  roles,
  onSuccess,
}: PlayerDialogProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [name, setName] = React.useState("");
  const [balance, setBalance] = React.useState("0");
  const [startBalance, setStartBalance] = React.useState("0");
  const [roleId, setRoleId] = React.useState<string | null>(null);
  const [gender, setGender] = React.useState<Gender | null>(null);
  const [inviteEmail, setInviteEmail] = React.useState("");
  const [isSendingInvite, setIsSendingInvite] = React.useState(false);
  const isEdit = !!player;
  const canInvite = isEdit && player && !player.user_id;

  React.useEffect(() => {
    if (open && player) {
      setName(player.name);
      setBalance((player.balance / 100).toString());
      setStartBalance((player.start_balance / 100).toString());
      setRoleId(player.role_id);
      setGender(player.gender ?? null);
      setInviteEmail("");
    } else if (open && !player) {
      setName("");
      setBalance("0");
      setStartBalance("0");
      setGender(null);
      setInviteEmail("");
      // Pre-select first role when creating new player
      setRoleId(roles.length > 0 ? roles[0].id : null);
    }
  }, [open, player, roles]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast.error("Name ist erforderlich");
      return;
    }

    if (!roleId) {
      toast.error("Rolle ist erforderlich");
      return;
    }
    setIsSubmitting(true);

    try {
      const body: CreatePlayerRequest | UpdatePlayerRequest = {
        name: name.trim(),
        balance: euroToCents(parseFloat(balance) || 0),
        start_balance: euroToCents(parseFloat(startBalance) || 0),
        role_id: roleId!, // role_id is required and validated above
        user_id: null,
        gender: gender,
      };

      const url = isEdit
        ? `/api/clubs/${clubId}/players/${player.id}`
        : `/api/clubs/${clubId}/players`;

      const response = await fetch(url, {
        method: isEdit ? "PUT" : "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": document.cookie
            .split("; ")
            .find((row) => row.startsWith("csrf_token="))
            ?.split("=")[1] || "",
        },
        credentials: "include",
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || "Fehler beim Speichern");
      }

      toast.success(
        isEdit ? "Player erfolgreich aktualisiert" : "Player erfolgreich erstellt"
      );
      onSuccess();
      onOpenChange(false);
      // Reset form
      setName("");
      setBalance("0");
      setStartBalance("0");
      setRoleId(roles.length > 0 ? roles[0].id : null);
      setGender(null);
    } catch (error) {
      console.error("Error saving player:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Speichern"
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleInvite = async (e: React.FormEvent | React.KeyboardEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (!inviteEmail.trim()) {
      toast.error("E-Mail-Adresse ist erforderlich");
      return;
    }

    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(inviteEmail.trim())) {
      toast.error("Ungültige E-Mail-Adresse");
      return;
    }

    if (!player) {
      toast.error("Player-Informationen fehlen");
      return;
    }

    setIsSendingInvite(true);
    const emailToInvite = inviteEmail.trim().toLowerCase();

    try {
      toast.loading("Einladung wird vorbereitet...", { id: "invite-loading" });

      const csrfToken = document.cookie
        .split("; ")
        .find((row) => row.startsWith("csrf_token="))
        ?.split("=")[1] || "";

      if (!csrfToken) {
        throw new Error("CSRF-Token fehlt. Bitte Seite neu laden.");
      }

      const response = await fetch(`/api/clubs/${clubId}/players/${player.id}/invite`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrfToken,
        },
        credentials: "include",
        body: JSON.stringify({ email: emailToInvite }),
      });

      const responseData = await response.json().catch(() => ({}));

      if (!response.ok) {
        const errorMessage = responseData.message || responseData.warning || "Fehler beim Senden der Einladung";
        
        // Check if invitation was created but email failed
        if (responseData.warning || responseData.invitation_id) {
          const warningMsg = responseData.warning || "E-Mail konnte nicht versendet werden";
          toast.warning(
            `Einladung wurde erstellt (ID: ${responseData.invitation_id || "unbekannt"}), aber: ${warningMsg}. ` +
            `Bitte überprüfen Sie die E-Mail-Konfiguration im Backend.`,
            { id: "invite-loading", duration: 10000 }
          );
          console.warn("Invitation created but email failed:", responseData);
          setInviteEmail("");
          onSuccess(); // Refresh player data
          return;
        }
        
        throw new Error(errorMessage);
      }

      // Check if response includes a warning (even with 200 status)
      if (responseData.warning) {
        toast.warning(
          `Einladung erstellt, aber: ${responseData.warning}`,
          { id: "invite-loading", duration: 8000 }
        );
        console.warn("Invitation created with warning:", responseData);
      } else {
        toast.success(
          `Einladung wurde erfolgreich an ${emailToInvite} versendet`,
          { id: "invite-loading", duration: 5000 }
        );
      }
      setInviteEmail("");
      onSuccess(); // Refresh player data
    } catch (error) {
      console.error("Error sending invitation:", error);
      const errorMessage = error instanceof Error ? error.message : "Fehler beim Senden der Einladung";
      toast.error(errorMessage, { id: "invite-loading", duration: 6000 });
    } finally {
      setIsSendingInvite(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? "Player bearbeiten" : "Neuen Player erstellen"}
          </DialogTitle>
          <DialogDescription>
            {isEdit
              ? "Ändern Sie die Daten des Players."
              : "Fügen Sie einen neuen Player zum Club hinzu."}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              placeholder="Max Mustermann"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="balance">Balance (€)</Label>
              <Input
                id="balance"
                type="number"
                step="0.01"
                placeholder="0.00"
                value={balance}
                onChange={(e) => setBalance(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="start_balance">Start-Balance (€)</Label>
              <Input
                id="start_balance"
                type="number"
                step="0.01"
                placeholder="0.00"
                value={startBalance}
                onChange={(e) => setStartBalance(e.target.value)}
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="role_id">
              Rolle <span className="text-red-500">*</span>
            </Label>
            <Select
              value={roleId || undefined}
              onValueChange={(value) => {
                setRoleId(value);
              }}
              required
            >
              <SelectTrigger id="role_id">
                <SelectValue placeholder="Rolle auswählen" />
              </SelectTrigger>
              <SelectContent>
                {roles.map((role) => (
                  <SelectItem key={role.id} value={role.id}>
                    {role.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-sm text-muted-foreground">
              Jeder Spieler muss einer Rolle zugeordnet sein
            </p>
          </div>
          <div className="space-y-2">
            <Label htmlFor="gender">Geschlecht</Label>
            <Select
              value={gender ?? "none"}
              onValueChange={(value) =>
                setGender(value === "none" ? null : (value as Gender))
              }
            >
              <SelectTrigger id="gender">
                <SelectValue placeholder="Keine Angabe" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">Keine Angabe</SelectItem>
                <SelectItem value="male">Männlich</SelectItem>
                <SelectItem value="female">Weiblich</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-sm text-muted-foreground">
              Für die Auswertung geschlechtsspezifischer Wettbewerbe
            </p>
          </div>
          {canInvite && (
            <div className="space-y-4 border-t pt-4">
              <div>
                <h3 className="text-sm font-medium mb-2">User einladen</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Dieser Player hat noch keinen zugewiesenen User. Laden Sie einen User per E-Mail ein, um ihn mit diesem Player zu verbinden.
                </p>
                <div className="flex gap-2">
                  <div className="flex-1">
                    <Input
                      type="email"
                      placeholder="user@example.com"
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                      disabled={isSendingInvite}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" && inviteEmail.trim() && !isSendingInvite) {
                          e.preventDefault();
                          handleInvite(e);
                        }
                      }}
                    />
                  </div>
                  <Button 
                    type="button"
                    onClick={handleInvite}
                    disabled={isSendingInvite || !inviteEmail.trim()}
                  >
                    {isSendingInvite ? "Wird gesendet..." : "Einladen"}
                  </Button>
                </div>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Abbrechen
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting
                ? "Speichere..."
                : isEdit
                  ? "Speichern"
                  : "Erstellen"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

