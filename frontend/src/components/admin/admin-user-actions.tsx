import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { BanIcon, LogOutIcon, ShieldCheckIcon, Trash2Icon } from "lucide-react";

import { AdminActionDialog } from "@/components/admin/admin-action-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/context/AuthContext";
import { adminMutation } from "@/lib/admin";
import type { AdminUserDetail } from "@/types/admin";

type UserDialog = "disable" | "enable" | "revoke" | "delete" | null;

function userUrl(id: string, suffix = "") {
  return `/api/admin/users/${encodeURIComponent(id)}${suffix}`;
}

/**
 * Interventions on an account: disable/enable, revoke browser sessions, delete.
 * The API refuses the own account and other instance admins as well; the
 * buttons are disabled here with an explanation.
 */
export function AdminUserActions({ user, onChanged }: { user: AdminUserDetail; onChanged: () => void }) {
  const { user: me } = useAuth();
  const navigate = useNavigate();
  const [dialog, setDialog] = useState<UserDialog>(null);

  const isSelf = me?.id === user.id;
  const disabled = user.disabled_at !== null;
  const lockReason = isSelf
    ? "Das eigene Konto kann hier nicht gesperrt oder gelöscht werden."
    : user.is_instance_admin
      ? "Instanz-Admins können nicht gesperrt oder gelöscht werden. Die Admin-Rolle wird per CLI entzogen (kegelmaster admin revoke)."
      : null;
  const ownsClubs = user.owned_clubs.length > 0;

  const close = (open: boolean) => {
    if (!open) setDialog(null);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Eingriffe</CardTitle>
        <CardDescription>Alle Eingriffe werden protokolliert.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex flex-wrap gap-2">
          {disabled ? (
            <Button variant="outline" onClick={() => setDialog("enable")}>
              <ShieldCheckIcon />
              Entsperren
            </Button>
          ) : (
            <Button variant="outline" disabled={lockReason !== null} onClick={() => setDialog("disable")}>
              <BanIcon />
              Sperren
            </Button>
          )}
          <Button variant="outline" disabled={isSelf} onClick={() => setDialog("revoke")}>
            <LogOutIcon />
            Sessions widerrufen
          </Button>
          <Button
            variant="destructive"
            disabled={lockReason !== null || ownsClubs}
            onClick={() => setDialog("delete")}
          >
            <Trash2Icon />
            Benutzer löschen
          </Button>
        </div>
        {lockReason && <p className="text-sm text-muted-foreground">{lockReason}</p>}
        {!lockReason && ownsClubs && (
          <p className="text-sm text-muted-foreground">
            Löschen ist erst möglich, wenn der Owner aller eigenen Klubs übertragen wurde (im Klub-Detail).
          </p>
        )}
      </CardContent>

      <AdminActionDialog
        open={dialog === "disable"}
        onOpenChange={close}
        title="Benutzer sperren?"
        description={
          <>
            <p>
              <strong className="break-all">{user.email}</strong> kann sich nicht mehr anmelden. Alle Browser-Sessions
              werden beendet und API-Tokens abgelehnt, bis das Konto entsperrt wird.
            </p>
            <p>Spieler-Verknüpfungen und Klubdaten bleiben unverändert.</p>
          </>
        }
        confirmLabel="Sperren"
        destructive
        successMessage="Benutzer gesperrt"
        onConfirm={async () => {
          await adminMutation("POST", userUrl(user.id, "/disable"));
          onChanged();
        }}
      />
      <AdminActionDialog
        open={dialog === "enable"}
        onOpenChange={close}
        title="Benutzer entsperren?"
        description={
          <p>
            <strong className="break-all">{user.email}</strong> kann sich wieder anmelden; vorhandene API-Tokens gelten
            wieder. Frühere Browser-Sessions bleiben beendet.
          </p>
        }
        confirmLabel="Entsperren"
        successMessage="Benutzer entsperrt"
        onConfirm={async () => {
          await adminMutation("POST", userUrl(user.id, "/enable"));
          onChanged();
        }}
      />
      <AdminActionDialog
        open={dialog === "revoke"}
        onOpenChange={close}
        title="Sessions widerrufen?"
        description={
          <p>
            <strong className="break-all">{user.email}</strong> wird in allen Browsern abgemeldet. API-Tokens bleiben
            gültig und können einzeln widerrufen werden.
          </p>
        }
        confirmLabel="Sessions widerrufen"
        successMessage="Sessions widerrufen"
        onConfirm={async () => {
          await adminMutation("POST", userUrl(user.id, "/revoke-sessions"));
          onChanged();
        }}
      />
      <AdminActionDialog
        open={dialog === "delete"}
        onOpenChange={close}
        title="Benutzer endgültig löschen?"
        description={
          <>
            <p>
              Das Konto <strong className="break-all">{user.email}</strong> wird mit Passkeys und API-Tokens gelöscht.
              Verknüpfte Spieler bleiben in ihren Klubs, nur die Verknüpfung wird gelöst. Das kann nicht rückgängig
              gemacht werden.
            </p>
            <p>Zur Bestätigung die E-Mail-Adresse eingeben.</p>
          </>
        }
        input={{ label: "E-Mail-Adresse", type: "email", mustMatch: user.email, ignoreCase: true, placeholder: user.email }}
        confirmLabel="Endgültig löschen"
        destructive
        successMessage="Benutzer gelöscht"
        onConfirm={async (value) => {
          await adminMutation("DELETE", userUrl(user.id), { confirm_email: value });
          navigate("/admin/users", { replace: true });
        }}
      />
    </Card>
  );
}

/** Revokes one API token of the account after confirmation. */
export function AdminRevokeTokenButton({
  userId,
  token,
  onChanged,
}: {
  userId: string;
  token: { id: string; name: string };
  onChanged: () => void;
}) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button variant="ghost" size="sm" onClick={() => setOpen(true)} aria-label={`Token ${token.name} widerrufen`}>
        <Trash2Icon />
        Widerrufen
      </Button>
      <AdminActionDialog
        open={open}
        onOpenChange={setOpen}
        title="API-Token widerrufen?"
        description={
          <p>
            Das Token <strong>{token.name}</strong> wird gelöscht und ist sofort ungültig. Skripte und Agenten, die es
            nutzen, verlieren den Zugriff.
          </p>
        }
        confirmLabel="Widerrufen"
        destructive
        successMessage="API-Token widerrufen"
        onConfirm={async () => {
          await adminMutation("DELETE", userUrl(userId, `/tokens/${encodeURIComponent(token.id)}`));
          onChanged();
        }}
      />
    </>
  );
}
