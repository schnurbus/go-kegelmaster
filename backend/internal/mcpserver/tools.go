package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
)

func registerTools(s *mcp.Server, api API) {
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: boolPtr(false)}

	addTool(s, api, &mcp.Tool{
		Name:        "klubs_auflisten",
		Title:       "Klubs auflisten",
		Description: "Listet alle Kegelklubs, in denen du Eigentümer oder Spieler bist, mit Kassenstand, Grundgebühr und deinen Rechten im Klub. Ausgangspunkt für alle anderen Tools (liefert die club_id).",
		Annotations: readOnly,
	}, listClubs)

	addTool(s, api, &mcp.Tool{
		Name:  "klub_ueberblick",
		Title: "Klub-Überblick",
		Description: "Stammdaten eines Klubs in einem Aufruf: Spieler (mit Saldo, Rolle, aktiv/inaktiv), Rollen (zahlt Grundgebühr ja/nein), Strafentypen (Preis, Dezimalmengen erlaubt) und Wettbewerbe. " +
			"Nutze die IDs daraus für spieltag_erfassen und buchung_erfassen. Bereiche ohne Leserecht werden als nicht_verfuegbar gemeldet.",
		Annotations: readOnly,
	}, clubOverview)

	addTool(s, api, &mcp.Tool{
		Name:        "spieltage_auflisten",
		Title:       "Spieltage auflisten",
		Description: "Listet Spieltage (Spielabende) eines Klubs, neueste zuerst, mit Teilnehmerzahl, Status (Entwurf/veröffentlicht) und gebuchten Strafen. Optional nach Zeitraum oder nur Entwürfen filtern.",
		Annotations: readOnly,
	}, listGameDays)

	addTool(s, api, &mcp.Tool{
		Name:  "spieltag_details",
		Title: "Spieltag-Details",
		Description: "Details eines Spieltags: Teilnehmer mit Strafen und Wettbewerbswerten. Bei Entwürfen zusätzlich eine Vorschau, was spieltag_veroeffentlichen buchen würde " +
			"(Strafen der Teilnehmer und Grundgebühr aller aktiven Spieler mit zahlender Rolle). Bei veröffentlichten Spieltagen die tatsächlich gebuchten Summen.",
		Annotations: readOnly,
	}, gameDayDetails)

	addTool(s, api, &mcp.Tool{
		Name:        "spieler_konto",
		Title:       "Spielerkonto",
		Description: "Saldo und die letzten Buchungen eines Spielers. Negativer Saldo = Spieler schuldet der Kasse Geld.",
		Annotations: readOnly,
	}, playerAccount)

	addTool(s, api, &mcp.Tool{
		Name:  "transaktionen_suchen",
		Title: "Transaktionen suchen",
		Description: "Buchungen der Klubkasse, neueste zuerst. Filter: typ, spieler_id oder spieltag_id (höchstens einer von spieler_id/spieltag_id). " +
			"Typen: grundgebuehr, strafe, einzahlung, trinkgeld, ausgabe. Seitenweise (seite, pro_seite ≤ 100).",
		Annotations: readOnly,
	}, searchTransactions)

	addTool(s, api, &mcp.Tool{
		Name:        "mein_spieler",
		Title:       "Mein Spieler",
		Description: "Dein eigener Spieler im Klub: Saldo sowie Strafen- und Wettbewerbsverlauf pro Spieltag (standardmäßig das letzte Jahr). Funktioniert auch ohne besondere Rechte.",
		Annotations: readOnly,
	}, myPlayer)

	addTool(s, api, &mcp.Tool{
		Name:  "spieltag_erfassen",
		Title: "Spieltag als Entwurf erfassen",
		Description: "Legt einen Spieltag als ENTWURF an, inklusive aller Teilnehmer mit ihren Strafen und Wettbewerbswerten, in einem Aufruf. Es wird noch nichts gebucht. " +
			"IDs aus klub_ueberblick verwenden. anzahl ist die Menge einer Strafe (Dezimalwerte nur bei Strafentypen mit dezimal_erlaubt). " +
			"wert ist das Ergebnis im Wettbewerb als Ganzzahl (höchster Wert gewinnt, niedrigster verliert). Schlägt ein Teil fehl, wird der Entwurf wieder gelöscht. " +
			"Danach Vorschau mit spieltag_details zeigen und erst nach Bestätigung spieltag_veroeffentlichen aufrufen.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(false), OpenWorldHint: boolPtr(false)},
	}, createGameDay)

	addTool(s, api, &mcp.Tool{
		Name:  "spieltag_veroeffentlichen",
		Title: "Spieltag veröffentlichen (bucht Gebühren)",
		Description: "Veröffentlicht einen Spieltag-Entwurf und BUCHT dabei Geld: die Strafen aller Teilnehmer und die Grundgebühr für alle aktiven Spieler mit zahlender Rolle (auch wer nicht teilgenommen hat). " +
			"Nur nach ausdrücklicher Bestätigung durch den Nutzer aufrufen. Liefert die gebuchten Summen.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true), IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	}, publishGameDay)

	addTool(s, api, &mcp.Tool{
		Name:  "buchung_erfassen",
		Title: "Buchung erfassen",
		Description: "Bucht eine manuelle Kassenbewegung: einzahlung (Spieler zahlt Geld in die Kasse, erhöht seinen Saldo und den Kassenstand), " +
			"trinkgeld (Spieler gibt der Kasse freiwillig Geld, erhöht nur den Kassenstand, sein Saldo bleibt gleich) " +
			"oder ausgabe (Ausgabe aus der Klubkasse ohne Spieler, verringert den Kassenstand). betrag_cent ist immer positiv. Einzahlung und Trinkgeld brauchen spieler_id; im Paar-Modus kann eine Einzahlung mit spieler_ids auf mehrere Spieler verteilt werden. " +
			"Ist im Klub Auto-Trinkgeld aktiv, wird bei einer Einzahlung der Teil, der über die Schulden des Spielers hinausgeht, automatisch als Trinkgeld gebucht. " +
			"Strafen und Grundgebühren entstehen nur über Spieltage. Bei Erfolg nicht erneut aufrufen.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true), OpenWorldHint: boolPtr(false)},
	}, createTransaction)
}

// ---------- shared helpers ----------

type clubIDInput struct {
	ClubID string `json:"club_id" jsonschema:"ID des Klubs (aus klubs_auflisten)"`
}

func clubPath(clubID string, parts ...string) string {
	p := "/api/clubs/" + pathEsc(clubID)
	for _, part := range parts {
		p += "/" + part
	}
	return p
}

func requireID(name, v string) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("%s ist erforderlich", name)
	}
	return nil
}

func validDate(name, v string) error {
	if _, err := time.Parse("2006-01-02", v); err != nil {
		return fmt.Errorf("%s muss im Format JJJJ-MM-TT sein", name)
	}
	return nil
}

type money struct {
	Cent int    `json:"cent"`
	Text string `json:"text"`
}

func m(cents int) money { return money{Cent: cents, Text: euro(cents)} }

var txTypeLabels = map[string]string{
	"base_fee": "grundgebuehr",
	"fee":      "strafe",
	"deposit":  "einzahlung",
	"tip":      "trinkgeld",
	"expense":  "ausgabe",
}

func txTypeFromLabel(label string) (string, bool) {
	for k, v := range txTypeLabels {
		if v == label {
			return k, true
		}
	}
	return "", false
}

type txOut struct {
	ID          string  `json:"transaktion_id"`
	Datum       string  `json:"datum"`
	Typ         string  `json:"typ"`
	Betrag      money   `json:"betrag"`
	Beschreib   string  `json:"beschreibung,omitempty"`
	SpielerID   *string `json:"spieler_id,omitempty"`
	SpielerName string  `json:"spieler_name,omitempty"`
	SpieltagID  *string `json:"spieltag_id,omitempty"`
	SaldoDanach *money  `json:"spieler_saldo_danach,omitempty"`
	KasseDanach money   `json:"kasse_danach"`
}

func toTxOut(t apiTransaction) txOut {
	out := txOut{
		ID: t.ID, Datum: t.TransactionDate, Typ: txTypeLabels[t.TransactionType], Betrag: m(t.Amount),
		Beschreib: t.Description, SpielerID: t.PlayerID, SpielerName: t.PlayerName, SpieltagID: t.GameDayID,
		KasseDanach: m(t.ClubBalanceAfter),
	}
	if out.Typ == "" {
		out.Typ = t.TransactionType
	}
	if t.PlayerBalanceAfter != nil {
		v := m(*t.PlayerBalanceAfter)
		out.SaldoDanach = &v
	}
	return out
}

func toTxOuts(ts []apiTransaction) []txOut {
	out := make([]txOut, 0, len(ts))
	for _, t := range ts {
		out = append(out, toTxOut(t))
	}
	return out
}

func unavailable(err error) map[string]string {
	return map[string]string{"nicht_verfuegbar": err.Error()}
}

// ---------- klubs_auflisten ----------

type noInput struct{}

func listClubs(ctx context.Context, c *client, _ noInput) (any, error) {
	var clubs []apiClub
	if err := c.get(ctx, "/api/clubs", nil, &clubs); err != nil {
		return nil, err
	}
	type clubOut struct {
		ID           string              `json:"club_id"`
		Name         string              `json:"name"`
		Kasse        money               `json:"kassenstand"`
		Grundgebuehr money               `json:"grundgebuehr"`
		PaarModus    bool                `json:"paar_modus"`
		Eigentuemer  bool                `json:"bin_eigentuemer"`
		Rechte       map[string][]string `json:"rechte,omitempty"`
	}
	out := make([]clubOut, 0, len(clubs))
	for _, cl := range clubs {
		co := clubOut{ID: cl.ID, Name: cl.Name, Kasse: m(cl.Balance), Grundgebuehr: m(cl.BaseFee), PaarModus: cl.CouplesModeEnabled}
		var perms apiPermissions
		if err := c.get(ctx, clubPath(cl.ID, "permissions", "me"), nil, &perms); err == nil {
			co.Eigentuemer = perms.IsOwner
			if !perms.IsOwner {
				co.Rechte = map[string][]string{}
				for _, p := range perms.Permissions {
					co.Rechte[p.EntityType] = append(co.Rechte[p.EntityType], p.PermissionType)
				}
			}
		}
		out = append(out, co)
	}
	return map[string]any{"klubs": out}, nil
}

// ---------- klub_ueberblick ----------

func clubOverview(ctx context.Context, c *client, in clubIDInput) (any, error) {
	if err := requireID("club_id", in.ClubID); err != nil {
		return nil, err
	}
	var cl apiClub
	if err := c.get(ctx, clubPath(in.ClubID), nil, &cl); err != nil {
		return nil, err
	}
	result := map[string]any{
		"klub": map[string]any{
			"club_id": cl.ID, "name": cl.Name, "kassenstand": m(cl.Balance), "grundgebuehr": m(cl.BaseFee),
			"paar_modus": cl.CouplesModeEnabled, "auto_trinkgeld": cl.AutoTipEnabled,
		},
	}

	roleNames := map[string]string{}
	var roles []apiRole
	if err := c.get(ctx, clubPath(in.ClubID, "roles"), nil, &roles); err != nil {
		result["rollen"] = unavailable(err)
	} else {
		type roleOut struct {
			ID   string `json:"rolle_id"`
			Name string `json:"name"`
			Zahl bool   `json:"zahlt_grundgebuehr"`
		}
		out := make([]roleOut, 0, len(roles))
		for _, r := range roles {
			roleNames[r.ID] = r.Name
			out = append(out, roleOut{r.ID, r.Name, r.PaysBaseFee})
		}
		result["rollen"] = out
	}

	var players []apiPlayer
	if err := c.get(ctx, clubPath(in.ClubID, "players"), nil, &players); err != nil {
		result["spieler"] = unavailable(err)
	} else {
		type playerOut struct {
			ID         string  `json:"spieler_id"`
			Name       string  `json:"name"`
			Saldo      money   `json:"saldo"`
			PaarSaldo  *money  `json:"paar_saldo,omitempty"`
			Rolle      string  `json:"rolle,omitempty"`
			RolleID    *string `json:"rolle_id,omitempty"`
			Geschlecht *string `json:"geschlecht,omitempty"`
			Inaktiv    bool    `json:"inaktiv"`
			PartnerID  *string `json:"partner_id,omitempty"`
		}
		sort.Slice(players, func(i, j int) bool { return strings.ToLower(players[i].Name) < strings.ToLower(players[j].Name) })
		out := make([]playerOut, 0, len(players))
		for _, p := range players {
			po := playerOut{ID: p.ID, Name: p.Name, Saldo: m(p.Balance), RolleID: p.RoleID, Geschlecht: p.Gender, Inaktiv: p.Inactive, PartnerID: p.PartnerID}
			if p.RoleID != nil {
				po.Rolle = roleNames[*p.RoleID]
			}
			if p.PairBalance != nil {
				v := m(*p.PairBalance)
				po.PaarSaldo = &v
			}
			out = append(out, po)
		}
		result["spieler"] = out
	}

	var pts []apiPenaltyType
	if err := c.get(ctx, clubPath(in.ClubID, "penalty-types"), nil, &pts); err != nil {
		result["strafentypen"] = unavailable(err)
	} else {
		type ptOut struct {
			ID      string `json:"strafentyp_id"`
			Name    string `json:"name"`
			Beschr  string `json:"beschreibung,omitempty"`
			Preis   money  `json:"preis"`
			Dezimal bool   `json:"dezimal_erlaubt"`
		}
		out := make([]ptOut, 0, len(pts))
		for _, p := range pts {
			if p.ReplacedByID != nil {
				continue
			}
			out = append(out, ptOut{p.ID, p.Name, p.Description, m(p.Price), p.AllowsDecimalQuantity})
		}
		result["strafentypen"] = out
	}

	var comps []apiCompetition
	if err := c.get(ctx, clubPath(in.ClubID, "competitions"), nil, &comps); err != nil {
		result["wettbewerbe"] = unavailable(err)
	} else {
		type compOut struct {
			ID     string `json:"wettbewerb_id"`
			Name   string `json:"name"`
			Wert   string `json:"wertung"`
			Gender bool   `json:"getrennt_nach_geschlecht"`
		}
		out := make([]compOut, 0, len(comps))
		for _, cp := range comps {
			out = append(out, compOut{cp.ID, cp.Name, scoringLabel(cp.ScoringType), cp.IsGenderSpecific})
		}
		result["wettbewerbe"] = out
	}
	return result, nil
}

func scoringLabel(s string) string {
	switch s {
	case "winner":
		return "Gewinner (höchster Wert)"
	case "loser":
		return "Verlierer (niedrigster Wert)"
	case "both":
		return "Gewinner (höchster) und Verlierer (niedrigster Wert)"
	}
	return s
}

// ---------- spieltage_auflisten ----------

type listGameDaysInput struct {
	ClubID       string `json:"club_id" jsonschema:"ID des Klubs"`
	Von          string `json:"von,omitempty" jsonschema:"frühestes Datum JJJJ-MM-TT (optional)"`
	Bis          string `json:"bis,omitempty" jsonschema:"spätestes Datum JJJJ-MM-TT (optional)"`
	NurEntwuerfe bool   `json:"nur_entwuerfe,omitempty" jsonschema:"nur noch nicht veröffentlichte Spieltage"`
	Anzahl       int    `json:"anzahl,omitempty" jsonschema:"maximale Anzahl (Standard 20, höchstens 200)"`
}

func listGameDays(ctx context.Context, c *client, in listGameDaysInput) (any, error) {
	if err := requireID("club_id", in.ClubID); err != nil {
		return nil, err
	}
	for name, v := range map[string]string{"von": in.Von, "bis": in.Bis} {
		if v != "" {
			if err := validDate(name, v); err != nil {
				return nil, err
			}
		}
	}
	limit := in.Anzahl
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	var days []apiGameDaySummary
	if err := c.get(ctx, clubPath(in.ClubID, "gamedays", "summaries"), nil, &days); err != nil {
		return nil, err
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Date > days[j].Date })
	type dayOut struct {
		ID       string `json:"spieltag_id"`
		Datum    string `json:"datum"`
		Notizen  string `json:"notizen,omitempty"`
		Status   string `json:"status"`
		Teiln    int    `json:"teilnehmer"`
		Gebuchte *money `json:"gebuchte_strafen,omitempty"`
	}
	out := []dayOut{}
	total := 0
	for _, d := range days {
		if (in.Von != "" && d.Date < in.Von) || (in.Bis != "" && d.Date > in.Bis) || (in.NurEntwuerfe && !d.IsDraft) {
			continue
		}
		total++
		if len(out) >= limit {
			continue
		}
		do := dayOut{ID: d.ID, Datum: d.Date, Notizen: d.Notes, Status: statusLabel(d.IsDraft), Teiln: d.ParticipantCount}
		if !d.IsDraft {
			v := m(-d.PenaltyFeeTotal)
			do.Gebuchte = &v
		}
		out = append(out, do)
	}
	return map[string]any{"spieltage": out, "treffer_gesamt": total}, nil
}

func statusLabel(draft bool) string {
	if draft {
		return "entwurf"
	}
	return "veroeffentlicht"
}

// ---------- spieltag_details ----------

type gameDayInput struct {
	ClubID     string `json:"club_id" jsonschema:"ID des Klubs"`
	SpieltagID string `json:"spieltag_id" jsonschema:"ID des Spieltags"`
}

func gameDayDetails(ctx context.Context, c *client, in gameDayInput) (any, error) {
	if err := requireID("club_id", in.ClubID); err != nil {
		return nil, err
	}
	if err := requireID("spieltag_id", in.SpieltagID); err != nil {
		return nil, err
	}
	return loadGameDay(ctx, c, in.ClubID, in.SpieltagID)
}

func loadGameDay(ctx context.Context, c *client, clubID, gameDayID string) (map[string]any, error) {
	var detail apiGameDayDetail
	if err := c.get(ctx, clubPath(clubID, "gamedays", pathEsc(gameDayID)), nil, &detail); err != nil {
		return nil, err
	}

	compNames := map[string]string{}
	var comps []apiCompetition
	if err := c.get(ctx, clubPath(clubID, "competitions"), nil, &comps); err == nil {
		for _, cp := range comps {
			compNames[cp.ID] = cp.Name
		}
	}

	type feeOut struct {
		ID     string  `json:"strafentyp_id"`
		Name   string  `json:"name"`
		Anzahl float64 `json:"anzahl"`
		Preis  money   `json:"preis"`
		Summe  money   `json:"summe"`
	}
	type compOut struct {
		ID   string `json:"wettbewerb_id"`
		Name string `json:"name,omitempty"`
		Wert int    `json:"wert"`
	}
	type partOut struct {
		ID      string    `json:"spieler_id"`
		Name    string    `json:"name"`
		Strafen []feeOut  `json:"strafen"`
		Summe   money     `json:"strafen_summe"`
		Wettb   []compOut `json:"wettbewerbe"`
	}
	parts := make([]partOut, 0, len(detail.Participants))
	penaltyTotal := 0
	for _, p := range detail.Participants {
		po := partOut{ID: p.Participant.PlayerID, Name: p.Participant.PlayerName, Strafen: []feeOut{}, Wettb: []compOut{}}
		sum := 0
		for _, f := range p.Fees {
			amount := -transaction.FeeAmount(f.PenaltyTypePrice, f.Count, f.QuantityScale)
			if amount == 0 && f.Count == 0 {
				continue
			}
			sum += amount
			po.Strafen = append(po.Strafen, feeOut{f.PenaltyTypeID, f.PenaltyTypeName, f.Quantity, m(f.PenaltyTypePrice), m(amount)})
		}
		for _, cv := range p.CompetitionValues {
			po.Wettb = append(po.Wettb, compOut{cv.CompetitionID, compNames[cv.CompetitionID], cv.Value})
		}
		po.Summe = m(sum)
		penaltyTotal += sum
		parts = append(parts, po)
	}

	gd := detail.GameDay
	result := map[string]any{
		"spieltag":      map[string]any{"spieltag_id": gd.ID, "datum": gd.Date, "notizen": gd.Notes, "status": statusLabel(gd.IsDraft)},
		"teilnehmer":    parts,
		"strafen_summe": m(penaltyTotal),
	}

	if gd.IsDraft {
		result["vorschau_veroeffentlichung"] = publishPreview(ctx, c, clubID, penaltyTotal)
	} else {
		var sum apiTxSummary
		if err := c.get(ctx, clubPath(clubID, "gamedays", pathEsc(gameDayID), "transaction-summary"), nil, &sum); err != nil {
			result["gebucht"] = unavailable(err)
		} else {
			// The summary reports absolute amounts charged to the players.
			result["gebucht"] = map[string]any{
				"grundgebuehren": m(sum.BaseFeeTotal), "anzahl_grundgebuehren": sum.BaseFeeCount,
				"strafen": m(sum.PenaltyFeeTotal), "anzahl_strafen": sum.PenaltyFeeCount,
				"gesamt": m(sum.BaseFeeTotal + sum.PenaltyFeeTotal),
			}
		}
	}
	return result, nil
}

// publishPreview mirrors transaction.gameDayCharges: the base fee goes to every
// active player whose role pays it, penalties to the participants.
func publishPreview(ctx context.Context, c *client, clubID string, penaltyTotal int) any {
	var cl apiClub
	if err := c.get(ctx, clubPath(clubID), nil, &cl); err != nil {
		return unavailable(err)
	}
	preview := map[string]any{"strafen": m(penaltyTotal)}
	if cl.BaseFee <= 0 {
		preview["grundgebuehr"] = "keine (Grundgebühr des Klubs ist 0)"
		preview["gesamt"] = m(penaltyTotal)
		return preview
	}
	var roles []apiRole
	var players []apiPlayer
	if err := c.get(ctx, clubPath(clubID, "roles"), nil, &roles); err != nil {
		preview["grundgebuehr"] = unavailable(err)
		return preview
	}
	if err := c.get(ctx, clubPath(clubID, "players"), nil, &players); err != nil {
		preview["grundgebuehr"] = unavailable(err)
		return preview
	}
	pays := map[string]bool{}
	for _, r := range roles {
		pays[r.ID] = r.PaysBaseFee
	}
	names := []string{}
	for _, p := range players {
		if !p.Inactive && p.RoleID != nil && pays[*p.RoleID] {
			names = append(names, p.Name)
		}
	}
	sort.Strings(names)
	base := cl.BaseFee * len(names)
	preview["grundgebuehr"] = map[string]any{"pro_spieler": m(cl.BaseFee), "anzahl_spieler": len(names), "summe": m(base), "spieler": names}
	preview["gesamt"] = m(penaltyTotal + base)
	return preview
}

// ---------- spieler_konto ----------

type playerAccountInput struct {
	ClubID    string `json:"club_id" jsonschema:"ID des Klubs"`
	SpielerID string `json:"spieler_id" jsonschema:"ID des Spielers (aus klub_ueberblick)"`
	Anzahl    int    `json:"anzahl,omitempty" jsonschema:"Anzahl der letzten Buchungen (Standard 20, höchstens 100)"`
}

func playerAccount(ctx context.Context, c *client, in playerAccountInput) (any, error) {
	if err := requireID("club_id", in.ClubID); err != nil {
		return nil, err
	}
	if err := requireID("spieler_id", in.SpielerID); err != nil {
		return nil, err
	}
	var p apiPlayer
	if err := c.get(ctx, clubPath(in.ClubID, "players", pathEsc(in.SpielerID)), nil, &p); err != nil {
		return nil, err
	}
	result := map[string]any{"spieler": map[string]any{"spieler_id": p.ID, "name": p.Name, "saldo": m(p.Balance), "inaktiv": p.Inactive}}
	if p.PairBalance != nil {
		result["paar_saldo"] = m(*p.PairBalance)
	}
	var page apiTxPage
	if err := c.get(ctx, clubPath(in.ClubID, "players", pathEsc(in.SpielerID), "transactions"), map[string]string{"limit": strconv.Itoa(clampLimit(in.Anzahl, 20))}, &page); err != nil {
		result["buchungen"] = unavailable(err)
	} else {
		result["buchungen"] = toTxOuts(page.Data)
		result["buchungen_gesamt"] = page.Total
	}
	return result, nil
}

func clampLimit(v, def int) int {
	if v <= 0 {
		return def
	}
	if v > 100 {
		return 100
	}
	return v
}

// ---------- transaktionen_suchen ----------

type searchTxInput struct {
	ClubID     string `json:"club_id" jsonschema:"ID des Klubs"`
	Typ        string `json:"typ,omitempty" jsonschema:"grundgebuehr, strafe, einzahlung, trinkgeld oder ausgabe (optional)"`
	SpielerID  string `json:"spieler_id,omitempty" jsonschema:"nur Buchungen dieses Spielers (optional)"`
	SpieltagID string `json:"spieltag_id,omitempty" jsonschema:"nur Buchungen dieses Spieltags (optional)"`
	Seite      int    `json:"seite,omitempty" jsonschema:"Seite ab 1 (Standard 1)"`
	ProSeite   int    `json:"pro_seite,omitempty" jsonschema:"Einträge pro Seite (Standard 50, höchstens 100)"`
}

func searchTransactions(ctx context.Context, c *client, in searchTxInput) (any, error) {
	if err := requireID("club_id", in.ClubID); err != nil {
		return nil, err
	}
	if in.SpielerID != "" && in.SpieltagID != "" {
		return nil, errors.New("bitte entweder spieler_id oder spieltag_id angeben, nicht beides")
	}
	apiType := ""
	if in.Typ != "" {
		t, ok := txTypeFromLabel(in.Typ)
		if !ok {
			return nil, fmt.Errorf("unbekannter typ %q (erlaubt: grundgebuehr, strafe, einzahlung, trinkgeld, ausgabe)", in.Typ)
		}
		apiType = t
	}

	if in.SpieltagID != "" {
		var txs []apiTransaction
		if err := c.get(ctx, clubPath(in.ClubID, "gamedays", pathEsc(in.SpieltagID), "transactions"), nil, &txs); err != nil {
			return nil, err
		}
		filtered := txs[:0]
		for _, t := range txs {
			if apiType == "" || t.TransactionType == apiType {
				filtered = append(filtered, t)
			}
		}
		return map[string]any{"buchungen": toTxOuts(filtered), "treffer_gesamt": len(filtered)}, nil
	}

	query := map[string]string{"type": apiType, "limit": strconv.Itoa(clampLimit(in.ProSeite, 50))}
	if in.Seite > 0 {
		query["page"] = strconv.Itoa(in.Seite)
	}
	path := clubPath(in.ClubID, "transactions")
	if in.SpielerID != "" {
		path = clubPath(in.ClubID, "players", pathEsc(in.SpielerID), "transactions")
	}
	var page apiTxPage
	if err := c.get(ctx, path, query, &page); err != nil {
		return nil, err
	}
	data := page.Data
	if in.SpielerID != "" && apiType != "" {
		// the per-player endpoint has no type filter
		filtered := data[:0]
		for _, t := range data {
			if t.TransactionType == apiType {
				filtered = append(filtered, t)
			}
		}
		data = filtered
	}
	return map[string]any{
		"buchungen": toTxOuts(data), "seite": page.Page, "seiten_gesamt": page.TotalPages, "treffer_gesamt": page.Total,
	}, nil
}

// ---------- mein_spieler ----------

type myPlayerInput struct {
	ClubID string `json:"club_id" jsonschema:"ID des Klubs"`
	Seit   string `json:"seit,omitempty" jsonschema:"Verlauf ab Datum JJJJ-MM-TT (Standard: vor einem Jahr)"`
}

func myPlayer(ctx context.Context, c *client, in myPlayerInput) (any, error) {
	if err := requireID("club_id", in.ClubID); err != nil {
		return nil, err
	}
	if in.Seit != "" {
		if err := validDate("seit", in.Seit); err != nil {
			return nil, err
		}
	}
	var p apiPlayer
	if err := c.get(ctx, clubPath(in.ClubID, "players", "me"), nil, &p); err != nil {
		if isStatus(err, http.StatusNotFound) {
			return nil, errors.New("du bist in diesem Klub keinem Spieler zugeordnet")
		}
		return nil, err
	}
	result := map[string]any{"spieler": map[string]any{"spieler_id": p.ID, "name": p.Name, "saldo": m(p.Balance)}}
	if p.PairBalance != nil {
		result["paar_saldo"] = m(*p.PairBalance)
	}
	q := map[string]string{"since": in.Seit}
	var penalties []apiPenaltyHistoryDay
	if err := c.get(ctx, clubPath(in.ClubID, "players", "me", "penalty-history"), q, &penalties); err != nil {
		result["strafen_verlauf"] = unavailable(err)
	} else {
		type entry struct {
			Name   string  `json:"name"`
			Anzahl float64 `json:"anzahl"`
		}
		type day struct {
			ID      string  `json:"spieltag_id"`
			Datum   string  `json:"datum"`
			Strafen []entry `json:"strafen"`
		}
		out := make([]day, 0, len(penalties))
		for _, d := range penalties {
			dd := day{ID: d.GameDayID, Datum: d.Date, Strafen: []entry{}}
			for _, pe := range d.Penalties {
				dd.Strafen = append(dd.Strafen, entry{pe.PenaltyTypeName, pe.Quantity})
			}
			out = append(out, dd)
		}
		result["strafen_verlauf"] = out
	}
	var comps []apiCompetitionHistoryDay
	if err := c.get(ctx, clubPath(in.ClubID, "players", "me", "competition-history"), q, &comps); err != nil {
		result["wettbewerbs_verlauf"] = unavailable(err)
	} else {
		type entry struct {
			Name string `json:"name"`
			Wert int    `json:"wert"`
		}
		type day struct {
			ID    string  `json:"spieltag_id"`
			Datum string  `json:"datum"`
			Werte []entry `json:"werte"`
		}
		out := make([]day, 0, len(comps))
		for _, d := range comps {
			dd := day{ID: d.GameDayID, Datum: d.Date, Werte: []entry{}}
			for _, v := range d.Values {
				dd.Werte = append(dd.Werte, entry{v.CompetitionName, v.Value})
			}
			out = append(out, dd)
		}
		result["wettbewerbs_verlauf"] = out
	}
	return result, nil
}

// ---------- spieltag_erfassen ----------

type penaltyInput struct {
	StrafentypID string  `json:"strafentyp_id" jsonschema:"ID des Strafentyps"`
	Anzahl       float64 `json:"anzahl" jsonschema:"Menge der Strafe, größer als 0"`
}

type competitionInput struct {
	WettbewerbID string `json:"wettbewerb_id" jsonschema:"ID des Wettbewerbs"`
	Wert         int    `json:"wert" jsonschema:"Ergebnis als Ganzzahl (höchster Wert gewinnt)"`
}

type participantInput struct {
	SpielerID   string             `json:"spieler_id" jsonschema:"ID des Spielers"`
	Strafen     []penaltyInput     `json:"strafen,omitempty" jsonschema:"Strafen des Spielers an diesem Abend"`
	Wettbewerbe []competitionInput `json:"wettbewerbe,omitempty" jsonschema:"Wettbewerbsergebnisse des Spielers"`
}

type createGameDayInput struct {
	ClubID     string             `json:"club_id" jsonschema:"ID des Klubs"`
	Datum      string             `json:"datum" jsonschema:"Datum des Spieltags, JJJJ-MM-TT"`
	Notizen    string             `json:"notizen,omitempty" jsonschema:"freie Notizen (optional)"`
	Teilnehmer []participantInput `json:"teilnehmer" jsonschema:"alle Spieler, die an diesem Abend dabei waren"`
}

func (in createGameDayInput) validate() error {
	if err := requireID("club_id", in.ClubID); err != nil {
		return err
	}
	if err := validDate("datum", in.Datum); err != nil {
		return err
	}
	if len(in.Teilnehmer) == 0 {
		return errors.New("mindestens ein Teilnehmer ist erforderlich")
	}
	seen := map[string]bool{}
	for i, p := range in.Teilnehmer {
		if strings.TrimSpace(p.SpielerID) == "" {
			return fmt.Errorf("teilnehmer[%d]: spieler_id fehlt", i)
		}
		if seen[p.SpielerID] {
			return fmt.Errorf("spieler_id %s ist doppelt aufgeführt", p.SpielerID)
		}
		seen[p.SpielerID] = true
		types := map[string]bool{}
		for _, s := range p.Strafen {
			if s.StrafentypID == "" || s.Anzahl <= 0 {
				return fmt.Errorf("teilnehmer %s: jede Strafe braucht strafentyp_id und anzahl > 0", p.SpielerID)
			}
			if types[s.StrafentypID] {
				return fmt.Errorf("teilnehmer %s: strafentyp_id %s ist doppelt aufgeführt, bitte Mengen zusammenfassen", p.SpielerID, s.StrafentypID)
			}
			types[s.StrafentypID] = true
		}
		comps := map[string]bool{}
		for _, w := range p.Wettbewerbe {
			if w.WettbewerbID == "" {
				return fmt.Errorf("teilnehmer %s: wettbewerb_id fehlt", p.SpielerID)
			}
			if comps[w.WettbewerbID] {
				return fmt.Errorf("teilnehmer %s: wettbewerb_id %s ist doppelt aufgeführt", p.SpielerID, w.WettbewerbID)
			}
			comps[w.WettbewerbID] = true
		}
	}
	return nil
}

func createGameDay(ctx context.Context, c *client, in createGameDayInput) (any, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	if err := checkPenaltyQuantities(ctx, c, in); err != nil {
		return nil, err
	}
	var gd apiGameDay
	if err := c.send(ctx, http.MethodPost, clubPath(in.ClubID, "gamedays"), map[string]any{
		"date": in.Datum, "notes": in.Notizen, "is_draft": true,
	}, &gd); err != nil {
		return nil, err
	}
	if !gd.IsDraft {
		return nil, errors.New("spieltag wurde nicht als Entwurf angelegt")
	}

	fill := func() error {
		for _, p := range in.Teilnehmer {
			base := clubPath(in.ClubID, "gamedays", pathEsc(gd.ID), "participants")
			if err := c.send(ctx, http.MethodPost, base, map[string]string{"player_id": p.SpielerID}, nil); err != nil {
				return fmt.Errorf("teilnehmer %s: %w", p.SpielerID, err)
			}
			if len(p.Strafen) > 0 {
				fees := make([]map[string]any, 0, len(p.Strafen))
				for _, s := range p.Strafen {
					fees = append(fees, map[string]any{"penalty_type_id": s.StrafentypID, "count": s.Anzahl})
				}
				if err := c.send(ctx, http.MethodPut, base+"/"+pathEsc(p.SpielerID)+"/fees", map[string]any{"fees": fees}, nil); err != nil {
					return fmt.Errorf("strafen für %s: %w", p.SpielerID, err)
				}
			}
			if len(p.Wettbewerbe) > 0 {
				values := make([]map[string]any, 0, len(p.Wettbewerbe))
				for _, w := range p.Wettbewerbe {
					values = append(values, map[string]any{"competition_id": w.WettbewerbID, "value": w.Wert})
				}
				if err := c.send(ctx, http.MethodPut, base+"/"+pathEsc(p.SpielerID)+"/competition-values", map[string]any{"values": values}, nil); err != nil {
					return fmt.Errorf("wettbewerbe für %s: %w", p.SpielerID, err)
				}
			}
		}
		return nil
	}

	if err := fill(); err != nil {
		// The game day is still a draft, so nothing was booked; drop it so no
		// half-filled draft is left behind.
		if delErr := c.send(ctx, http.MethodDelete, clubPath(in.ClubID, "gamedays", pathEsc(gd.ID)), nil, nil); delErr != nil {
			return nil, fmt.Errorf("%v (der unvollständige Entwurf %s konnte nicht gelöscht werden: %v)", err, gd.ID, delErr)
		}
		return nil, fmt.Errorf("%v – der Entwurf wurde verworfen, es wurde nichts gespeichert", err)
	}

	result, err := loadGameDay(ctx, c, in.ClubID, gd.ID)
	if err != nil {
		return map[string]any{"spieltag_id": gd.ID, "hinweis": "Entwurf angelegt, Details konnten nicht geladen werden: " + err.Error()}, nil
	}
	result["hinweis"] = "Entwurf angelegt, noch nichts gebucht. Zum Buchen spieltag_veroeffentlichen aufrufen (nach Bestätigung durch den Nutzer)."
	return result, nil
}

// checkPenaltyQuantities rejects unknown penalty types and quantities the API
// would silently round (non-integers for types without decimal quantities,
// more than two decimal places otherwise).
func checkPenaltyQuantities(ctx context.Context, c *client, in createGameDayInput) error {
	var pts []apiPenaltyType
	if err := c.get(ctx, clubPath(in.ClubID, "penalty-types"), nil, &pts); err != nil {
		return err
	}
	byID := make(map[string]apiPenaltyType, len(pts))
	for _, p := range pts {
		byID[p.ID] = p
	}
	for _, part := range in.Teilnehmer {
		for _, s := range part.Strafen {
			pt, ok := byID[s.StrafentypID]
			if !ok {
				return fmt.Errorf("strafentyp_id %s gibt es in diesem Klub nicht (siehe klub_ueberblick)", s.StrafentypID)
			}
			if !pt.AllowsDecimalQuantity && s.Anzahl != math.Trunc(s.Anzahl) {
				return fmt.Errorf("strafe %q erlaubt nur ganze Mengen, nicht %v", pt.Name, s.Anzahl)
			}
			if scaled := s.Anzahl * 100; math.Abs(scaled-math.Round(scaled)) > 1e-9 {
				return fmt.Errorf("strafe %q: höchstens zwei Nachkommastellen erlaubt (%v)", pt.Name, s.Anzahl)
			}
		}
	}
	return nil
}

// ---------- spieltag_veroeffentlichen ----------

func publishGameDay(ctx context.Context, c *client, in gameDayInput) (any, error) {
	if err := requireID("club_id", in.ClubID); err != nil {
		return nil, err
	}
	if err := requireID("spieltag_id", in.SpieltagID); err != nil {
		return nil, err
	}
	path := clubPath(in.ClubID, "gamedays", pathEsc(in.SpieltagID))
	var detail apiGameDayDetail
	if err := c.get(ctx, path, nil, &detail); err != nil {
		return nil, err
	}
	if !detail.GameDay.IsDraft {
		return nil, errors.New("der Spieltag ist bereits veröffentlicht, es wurde nichts erneut gebucht")
	}
	if err := c.send(ctx, http.MethodPut, path, map[string]any{
		"date": detail.GameDay.Date, "notes": detail.GameDay.Notes, "is_draft": false,
	}, nil); err != nil {
		return nil, err
	}
	result, err := loadGameDay(ctx, c, in.ClubID, in.SpieltagID)
	if err != nil {
		return map[string]any{"spieltag_id": in.SpieltagID, "hinweis": "Veröffentlicht und gebucht; Details konnten nicht geladen werden: " + err.Error()}, nil
	}
	result["hinweis"] = "Spieltag veröffentlicht, Gebühren wurden gebucht."
	return result, nil
}

// ---------- buchung_erfassen ----------

type createTxInput struct {
	ClubID       string   `json:"club_id" jsonschema:"ID des Klubs"`
	Typ          string   `json:"typ" jsonschema:"einzahlung, trinkgeld oder ausgabe"`
	BetragCent   int      `json:"betrag_cent" jsonschema:"Betrag in Cent, immer positiv (z. B. 2000 für 20,00 €)"`
	SpielerID    string   `json:"spieler_id,omitempty" jsonschema:"Spieler (bei einzahlung und trinkgeld)"`
	SpielerIDs   []string `json:"spieler_ids,omitempty" jsonschema:"mehrere Spieler für eine Paar-Einzahlung (nur im Paar-Modus, mindestens 2)"`
	Beschreibung string   `json:"beschreibung,omitempty" jsonschema:"Buchungstext (optional)"`
	Datum        string   `json:"datum,omitempty" jsonschema:"Buchungsdatum JJJJ-MM-TT (Standard: heute)"`
}

func createTransaction(ctx context.Context, c *client, in createTxInput) (any, error) {
	if err := requireID("club_id", in.ClubID); err != nil {
		return nil, err
	}
	apiType, ok := txTypeFromLabel(in.Typ)
	if !ok || !transaction.TransactionType(apiType).IsManual() {
		return nil, fmt.Errorf("typ muss einzahlung, trinkgeld oder ausgabe sein (nicht %q)", in.Typ)
	}
	if in.BetragCent <= 0 {
		return nil, errors.New("betrag_cent muss größer als 0 sein")
	}
	if in.SpielerID != "" && len(in.SpielerIDs) > 0 {
		return nil, errors.New("bitte entweder spieler_id oder spieler_ids angeben")
	}
	if apiType == "expense" && (in.SpielerID != "" || len(in.SpielerIDs) > 0) {
		return nil, errors.New("eine Ausgabe wird ohne Spieler gebucht")
	}
	if in.Datum != "" {
		if err := validDate("datum", in.Datum); err != nil {
			return nil, err
		}
	}
	body := map[string]any{"transaction_type": apiType, "amount": in.BetragCent, "description": in.Beschreibung}
	if in.SpielerID != "" {
		body["player_id"] = in.SpielerID
	}
	if len(in.SpielerIDs) > 0 {
		body["player_ids"] = in.SpielerIDs
	}
	if in.Datum != "" {
		body["transaction_date"] = in.Datum
	}
	var raw json.RawMessage
	if err := c.send(ctx, http.MethodPost, clubPath(in.ClubID, "transactions"), body, &raw); err != nil {
		var ae *apiError
		if errors.As(err, &ae) {
			return nil, err
		}
		// The API accepted the booking; only reading the answer failed.
		// Never report this as a failure, or the agent might book twice.
		return map[string]any{"hinweis": "Buchung gespeichert (Antwort konnte nicht gelesen werden). Nicht erneut buchen; zur Kontrolle transaktionen_suchen nutzen."}, nil
	}
	// With auto tip enabled a deposit is split into deposit + tip, and the API
	// returns an array.
	var txs []apiTransaction
	if err := json.Unmarshal(raw, &txs); err != nil {
		var single apiTransaction
		if err := json.Unmarshal(raw, &single); err != nil {
			return map[string]any{"hinweis": "Buchung gespeichert (Antwort konnte nicht gelesen werden). Nicht erneut buchen."}, nil
		}
		txs = []apiTransaction{single}
	}
	result := map[string]any{"buchungen": toTxOuts(txs), "hinweis": "Buchung gespeichert. Nicht erneut aufrufen."}
	if len(txs) > 1 && apiType == "deposit" {
		result["hinweis"] = "Buchung gespeichert. Die Einzahlung wurde aufgeteilt (Auto-Trinkgeld bzw. Paar-Modus), siehe buchungen. Nicht erneut aufrufen."
	}
	return result, nil
}
