import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Trash2Icon, UserRoundCogIcon } from "lucide-react";

import { AdminActionDialog } from "@/components/admin/admin-action-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { adminMutation } from "@/lib/admin";
import type { AdminClubDetail } from "@/types/admin";

/** Interventions on a club: transfer the owner, delete the club. */
export function AdminClubActions({ club, onChanged }: { club: AdminClubDetail; onChanged: () => void }) {
  const navigate = useNavigate();
  const [dialog, setDialog] = useState<"transfer" | "delete" | null>(null);
  const url = `/api/admin/clubs/${encodeURIComponent(club.id)}`;

  const close = (open: boolean) => {
    if (!open) setDialog(null);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Eingriffe</CardTitle>
        <CardDescription>Alle Eingriffe werden protokolliert.</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-wrap gap-2">
        <Button variant="outline" onClick={() => setDialog("transfer")}>
          <UserRoundCogIcon />
          Owner übertragen
        </Button>
        <Button variant="destructive" onClick={() => setDialog("delete")}>
          <Trash2Icon />
          Klub löschen
        </Button>
      </CardContent>

      <AdminActionDialog
        open={dialog === "transfer"}
        onOpenChange={close}
        title="Owner übertragen"
        description={
          <>
            <p>
              Aktueller Owner: <strong className="break-all">{club.owner.email}</strong>. Der neue Owner muss ein
              bestehendes, nicht gesperrtes Konto sein und erhält alle Rechte im Klub.
            </p>
            <p>Sie selbst werden dadurch nicht Mitglied des Klubs.</p>
          </>
        }
        input={{ label: "E-Mail des neuen Owners", type: "email", placeholder: "name@example.com" }}
        confirmLabel="Übertragen"
        successMessage="Owner übertragen"
        onConfirm={async (value) => {
          await adminMutation("POST", `${url}/transfer-owner`, { new_owner_email: value });
          onChanged();
        }}
      />
      <AdminActionDialog
        open={dialog === "delete"}
        onOpenChange={close}
        title="Klub endgültig löschen?"
        description={
          <>
            <p>
              Der Klub <strong className="break-words">{club.name}</strong> wird mit allen Spielern, Rollen, Spieltagen
              und Transaktionen gelöscht. Das kann nicht rückgängig gemacht werden.
            </p>
            <p>Zur Bestätigung den Klubnamen eingeben.</p>
          </>
        }
        input={{ label: "Klubname", mustMatch: club.name, placeholder: club.name }}
        confirmLabel="Endgültig löschen"
        destructive
        successMessage="Klub gelöscht"
        onConfirm={async (value) => {
          await adminMutation("DELETE", url, { confirm_name: value });
          navigate("/admin/clubs", { replace: true });
        }}
      />
    </Card>
  );
}
