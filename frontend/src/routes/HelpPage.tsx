import { AppLayout } from "@/components/AppLayout";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
function HelpPage() {
  return (
    <AppLayout title="Hilfe">
      <div className="flex flex-col gap-6 px-4 py-4 md:px-6 md:py-6">
        <p className="text-muted-foreground max-w-3xl">
          Diese Seite erklärt alle Funktionen der Kegelmaster-App in einfachen
          Worten. Sie können auch den Hilfe-Chat (Button unten rechts) nutzen,
          um Fragen zur Bedienung zu stellen.
        </p>

        <Card>
          <CardHeader>
            <CardTitle>Einstieg</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              <strong>Anmeldung:</strong> Unter „Login“ melden Sie sich mit
              E-Mail und Passwort an. Wenn Sie noch keinen Account haben, können
              Sie sich unter „Registrieren“ anlegen.
            </p>
            <p>
              <strong>Passwort vergessen:</strong> Auf der Login-Seite gibt es
              einen Link „Passwort vergessen“. Sie erhalten dann eine E-Mail mit
              einem Link zum Zurücksetzen des Passworts.
            </p>
            <p>
              <strong>Club auswählen:</strong> Nach dem Login sehen Sie oben in
              der linken Leiste den „Club-Switcher“. Wählen Sie dort den Club,
              mit dem Sie arbeiten möchten. Alle folgenden Seiten (Spieler,
              Spieltage, Transaktionen usw.) beziehen sich auf diesen Club.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Dashboard</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              Das Dashboard ist Ihre Übersicht: Es zeigt Karten mit wichtigen
              Kennzahlen (z. B. Klub-Saldo, Anzahl Spieler), die letzten
              Transaktionen und Diagramme zu Strafen und Wettbewerben. So
              behalten Sie den Überblick, ohne in die einzelnen Bereiche
              wechseln zu müssen.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Spieler</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              <strong>Spieler anlegen:</strong> Über „Spieler hinzufügen“ legen
              Sie einen neuen Spieler an. Dabei müssen Sie einen Namen und
              eine Rolle (z. B. Mitglied, Gast) vergeben. Jeder Spieler hat
              genau eine Rolle.
            </p>
            <p>
              <strong>Spieler bearbeiten oder löschen:</strong> In der
              Spielerliste können Sie einen Spieler auswählen, um ihn zu
              bearbeiten oder zu löschen.
            </p>
            <p>
              <strong>Spieler mit Account verknüpfen:</strong> Wenn ein Spieler
              sich in der App anmelden soll, können Sie ihm eine Einladung
              senden („Einladen“). Der Spieler erhält einen Link und kann nach
              dem Anmelden seine eigenen Statistiken (Strafen, Wettbewerbe)
              einsehen.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Spieltage</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              <strong>Spielabend anlegen:</strong> Unter „Spieltage“ legen Sie
              einen neuen Spielabend mit Datum an.
            </p>
            <p>
              <strong>Teilnehmer:</strong> Beim Spielabend fügen Sie die
              anwesenden Spieler als Teilnehmer hinzu. Pro Teilnehmer können
              Sie die Teilnahmegebühr und optional Trinkgeld eintragen.
            </p>
            <p>
              <strong>Strafen und Wettbewerbe:</strong> Für jeden Teilnehmer
              tragen Sie pro Strafentyp die Anzahl der Strafen ein sowie die
              Werte für die Wettbewerbe. Daraus berechnet die App automatisch
              die Salden (was der Spieler dem Club schuldet bzw. umgekehrt).
            </p>
            <p>
              <strong>Auswirkung:</strong> Die erfassten Gebühren und Strafen
              fließen in die Spieler- und Klub-Salden ein. Über Transaktionen
              (z. B. Einzahlungen) werden die Salden wieder ausgeglichen.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Transaktionen</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              <strong>Was sind Transaktionen?</strong> Transaktionen sind
              Geldbewegungen: z. B. Einzahlungen von Spielern (Saldo-Ausgleich),
              Trinkgeld, sonstige Einnahmen oder Ausgaben des Clubs.
            </p>
            <p>
              <strong>Transaktion anlegen:</strong> Sie wählen den Typ (z. B.
              Einzahlung, Trinkgeld), den Betrag und optional den Spieler und
              den Spielabend. Die App aktualisiert automatisch die Salden von
              Spieler und Club.
            </p>
            <p>
              In der Transaktionsliste können Sie alle Vorgänge einsehen und
              bei Bedarf filtern (z. B. nach Spieler oder Spieltag).
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Rollen</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              Rollen erscheinen nur, wenn Sie die nötige Berechtigung haben
              (z. B. als Club-Besitzer oder mit entsprechender Rolle).
            </p>
            <p>
              <strong>Was sind Rollen?</strong> Rollen gruppieren Spieler (z. B.
              Kassenwart, Mitglied, Gast) und legen fest, wer in der App was
              darf: z. B. Spieler anlegen, Spielabende verwalten, Strafentypen
              bearbeiten.
            </p>
            <p>
              Sie können Rollen anlegen, bearbeiten und jeder Rolle bestimmte
              Berechtigungen zuweisen oder entziehen.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Strafentypen</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              Strafentypen legen fest, welche „Strafen“ es in Ihrem Club gibt
              (z. B. „Kegel umgeworfen“) und welche Gebühr pro Stück fällig
              wird. Sie können Strafentypen anlegen, bearbeiten (Name, Gebühr)
              und die Reihenfolge der Anzeige ändern. Diese Typen werden dann
              bei den Spielabenden pro Spieler ausgewählt und die Anzahl
              eingetragen.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Wettbewerbe</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              Wettbewerbe sind z. B. „Höchste Serie“, „Bester Wurf“. Sie legen
              sie pro Club an und können sie bearbeiten. Bei jedem Spielabend
              können Sie pro Teilnehmer die zugehörigen Wettbewerbs-Werte
              eintragen (z. B. Punkte oder Rang).
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Club bearbeiten</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              Den Link „Club bearbeiten“ sehen nur Sie, wenn Sie der Besitzer
              des aktuell ausgewählten Clubs sind.
            </p>
            <p>
              Dort können Sie den Club-Namen, die Basisgebühr, die
              Start-Balance und Optionen wie „Auto-Trinkgeld“ oder „Paar-Modus“
              (Einzahlung auf mehrere Spieler verteilen) einstellen. Sie können
              den Besitzer an einen anderen Nutzer übertragen oder den Club
              löschen.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Hilfe-Chat</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              Unten rechts auf der Seite finden Sie einen schwebenden Button
              („Hilfe zur App“). Wenn Sie ihn öffnen, können Sie Fragen zur
              Bedienung der Kegelmaster-App stellen. Der Assistent antwortet
              auf Deutsch und hilft nur bei Themen rund um diese App.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Für technische Nutzer / Entwickler</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>
              Die API der Anwendung ist per OpenAPI beschrieben und kann
              interaktiv in Swagger genutzt werden:
            </p>
            <p>
              <a
                href="/api/docs"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary underline hover:no-underline"
              >
                API-Dokumentation (Swagger) öffnen
              </a>
            </p>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

export default HelpPage;
