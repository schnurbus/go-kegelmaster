# Kegelmaster

Verwaltung von Kegelclubs: Spieler, Spieltage mit Strafen und Wettbewerben sowie die Clubkasse.

## Language

### Personen und Zugehörigkeit

**Club**:
Ein Kegelclub mit eigenen Spielern, Rollen, Strafentypen und eigener Kasse.
_Avoid_: Klub, Verein

**Benutzer**:
Ein Konto, mit dem sich ein Mensch anmeldet; gehört zu keinem Club, sondern ist über Spieler mit Clubs verbunden.
_Avoid_: User, Account, Mitglied

**Spieler**:
Eine Person innerhalb eines Clubs, auf die Strafen, Einzahlungen und ein Saldo gebucht werden; hat immer genau eine Rolle und kann, muss aber nicht mit einem Benutzer verknüpft sein.
_Avoid_: Mitglied, Kegler

**Rolle**:
Eine clubeigene Gruppe von Rechten (z. B. Kassenwart, Mitglied, Gast); jeder Spieler eines Clubs hat genau eine.
_Avoid_: Gruppe, Rang

**Besitzer**:
Der Benutzer, dem ein Club aktuell gehört; seine Rechte im Club hängen nicht an einer Rolle, und der Besitz lässt sich übertragen.
_Avoid_: Gründer, Owner, Admin

**Einladung**:
Die Aufforderung an eine E-Mail-Adresse, einen bestimmten, bereits angelegten Spieler mit einem Benutzer zu verknüpfen; sie gilt für den Spieler, nicht für den Club.
_Avoid_: Club-Einladung, Beitrittsanfrage

### Einstieg in die Weboberfläche

**Rundgang**:
Eine Folge von Sprechblasen, die einem Benutzer zeigt, was er an welcher Stelle der Oberfläche findet; er erklärt nur und verlangt keine Eingaben.
_Avoid_: Tour, Tutorial, Onboarding
_Im Code_: `tour` (englische Bezeichner wie bei `user` für Benutzer; „Avoid“ gilt für deutsche Texte und Gespräche)

**Ersteinrichtung**:
Die Liste der Schritte, die der Besitzer erledigen muss, bis sein Club benutzbar ist; ein Schritt gilt als erledigt, sobald die zugehörigen Daten im Club vorhanden sind.
_Avoid_: Setup-Wizard, Onboarding, Checkliste
_Im Code_: `setup`; Frontend-Modul für beides: `lib/onboarding.ts`
