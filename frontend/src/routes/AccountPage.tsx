import * as React from "react";
import { Link } from "react-router-dom";
import { ChevronRightIcon, KeyRoundIcon } from "lucide-react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { PasswordInput } from "@/components/ui/password-input";
import { Label } from "@/components/ui/label";
import { PasskeysCard } from "@/components/account/PasskeysCard";
import { useAuth } from "@/context/AuthContext";
import { useClub } from "@/context/ClubContext";

const MIN_PASSWORD_LENGTH = 8;
const MAX_PASSWORD_LENGTH = 128;

function formatDate(value: string | null | undefined): string {
  if (!value) return "–";
  return new Date(value).toLocaleString("de-DE", {
    dateStyle: "long",
    timeStyle: "short",
  });
}

function ChangePasswordCard() {
  const { changePassword } = useAuth();
  const [currentPassword, setCurrentPassword] = React.useState("");
  const [newPassword, setNewPassword] = React.useState("");
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [error, setError] = React.useState<string | null>(null);
  const [isSaving, setIsSaving] = React.useState(false);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);

    if (newPassword.trim().length < MIN_PASSWORD_LENGTH) {
      setError(`Das neue Passwort muss mindestens ${MIN_PASSWORD_LENGTH} Zeichen lang sein.`);
      return;
    }
    if (newPassword !== confirmPassword) {
      setError("Die Passwörter stimmen nicht überein.");
      return;
    }
    if (newPassword === currentPassword) {
      setError("Das neue Passwort muss sich vom aktuellen unterscheiden.");
      return;
    }

    setIsSaving(true);
    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      toast.success("Passwort geändert");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Passwort konnte nicht geändert werden.");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Card>
      <form onSubmit={handleSubmit} className="flex flex-col gap-6">
        <CardHeader>
          <CardTitle>Passwort ändern</CardTitle>
          <CardDescription className="max-w-2xl">
            Nach der Änderung werden alle anderen Sitzungen abgemeldet und alle
            API-Tokens widerrufen. In diesem Browser bleiben Sie angemeldet.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid max-w-md gap-4">
          <div className="space-y-2">
            <Label htmlFor="current-password">Aktuelles Passwort</Label>
            <PasswordInput
              id="current-password"
              autoComplete="current-password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              maxLength={MAX_PASSWORD_LENGTH}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="new-password">Neues Passwort</Label>
            <PasswordInput
              id="new-password"
              autoComplete="new-password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              minLength={MIN_PASSWORD_LENGTH}
              maxLength={MAX_PASSWORD_LENGTH}
              required
            />
            <p className="text-xs text-muted-foreground">
              Mindestens {MIN_PASSWORD_LENGTH} Zeichen.
            </p>
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirm-password">Neues Passwort bestätigen</Label>
            <PasswordInput
              id="confirm-password"
              autoComplete="new-password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              minLength={MIN_PASSWORD_LENGTH}
              maxLength={MAX_PASSWORD_LENGTH}
              required
            />
          </div>
          {error && (
            <p role="alert" className="text-sm text-destructive">
              {error}
            </p>
          )}
        </CardContent>
        <CardFooter>
          <Button type="submit" disabled={isSaving} className="w-full sm:w-auto">
            {isSaving ? "Speichert …" : "Passwort ändern"}
          </Button>
        </CardFooter>
      </form>
    </Card>
  );
}

function NotificationsCard() {
  const { user, updateSettings } = useAuth();
  const [isSaving, setIsSaving] = React.useState(false);

  const handleChange = async (checked: boolean) => {
    setIsSaving(true);
    try {
      await updateSettings({ notify_gameday_settled: checked });
      toast.success(
        checked
          ? "Benachrichtigung für abgerechnete Spieltage aktiviert"
          : "Benachrichtigung für abgerechnete Spieltage deaktiviert",
      );
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Einstellung konnte nicht gespeichert werden.");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Benachrichtigungen</CardTitle>
        <CardDescription className="max-w-2xl">
          Benachrichtigungen werden an {user?.email ?? "Ihre E-Mail-Adresse"} geschickt.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex items-start gap-3">
          <Checkbox
            id="notify-gameday-settled"
            className="mt-0.5"
            checked={user?.notify_gameday_settled ?? false}
            disabled={!user || isSaving}
            onCheckedChange={(checked) => handleChange(checked === true)}
          />
          <div className="grid gap-1">
            <Label htmlFor="notify-gameday-settled">
              E-Mail, wenn ein Spieltag abgerechnet wurde
            </Label>
            <p className="text-sm text-muted-foreground">
              Sobald ein Spieltag in einem Ihrer Clubs veröffentlicht wird, erhalten Sie eine
              E-Mail mit Ihren Buchungen des Abends und Ihrem aktuellen Saldo. Voraussetzung:
              Ihr Account ist mit einem Spieler im Club verknüpft.
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function AccountPage() {
  const { user } = useAuth();
  const { clubs, isLoading: clubsLoading } = useClub();

  return (
    <div className="flex flex-col gap-6 px-4 py-4 md:px-6 md:py-6">
      <Card>
        <CardHeader>
          <CardTitle>Mein Account</CardTitle>
          <CardDescription>Angaben zu Ihrem Benutzerkonto.</CardDescription>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-x-6 gap-y-3 text-sm sm:grid-cols-[auto_1fr]">
            <dt className="text-muted-foreground">E-Mail</dt>
            <dd className="break-all font-medium">{user?.email}</dd>
            <dt className="text-muted-foreground">Registriert am</dt>
            <dd>{formatDate(user?.created_at)}</dd>
            <dt className="text-muted-foreground">Passwort zuletzt geändert</dt>
            <dd>{user?.password_changed_at ? formatDate(user.password_changed_at) : "Noch nie"}</dd>
          </dl>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Meine Clubs</CardTitle>
          <CardDescription>Clubs, auf die Sie Zugriff haben.</CardDescription>
        </CardHeader>
        <CardContent>
          {clubsLoading ? (
            <p className="text-sm text-muted-foreground">Lädt …</p>
          ) : clubs.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Sie sind noch in keinem Club.
            </p>
          ) : (
            <ul className="divide-y">
              {clubs.map((club) => (
                <li
                  key={club.id}
                  className="flex items-center justify-between gap-2 py-2 text-sm first:pt-0 last:pb-0"
                >
                  <span className="min-w-0 truncate font-medium">{club.name}</span>
                  <Badge variant={club.user_id === user?.id ? "default" : "secondary"}>
                    {club.user_id === user?.id ? "Eigentümer" : "Mitglied"}
                  </Badge>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <NotificationsCard />

      <ChangePasswordCard />

      <PasskeysCard />

      <Card>
        <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-1.5">
            <CardTitle>API-Tokens</CardTitle>
            <CardDescription>
              Zugriff für Skripte und KI-Agenten verwalten.
            </CardDescription>
          </div>
          <Button asChild variant="outline" className="w-full shrink-0 sm:w-auto">
            <Link to="/app/api-tokens">
              <KeyRoundIcon />
              API-Tokens verwalten
              <ChevronRightIcon />
            </Link>
          </Button>
        </CardHeader>
      </Card>
    </div>
  );
}

export default AccountPage;
