package email

import (
	"strings"
	"testing"
	"time"
)

func TestFormatEuro(t *testing.T) {
	cases := map[int]string{
		0:       "0,00 €",
		5:       "0,05 €",
		-350:    "-3,50 €",
		123456:  "1.234,56 €",
		-100000: "-1.000,00 €",
	}
	for cents, want := range cases {
		if got := FormatEuro(cents); got != want {
			t.Errorf("FormatEuro(%d) = %q, want %q", cents, got, want)
		}
	}
}

func TestRenderGameDaySettled(t *testing.T) {
	mail := GameDaySettledEmail{
		ClubName: "KC <Alle> Neune",
		Date:     time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC),
		Players:  []GameDaySettledPlayer{{Name: "Anna", GameDayAmount: -750, Balance: -1250}},
		AppURL:   "https://kegel.example.org/app",
	}
	html, err := RenderGameDaySettled(mail)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{"14.11.2026", "KC &lt;Alle&gt; Neune", "Anna", "-7,50 €", "-12,50 €", `href="https://kegel.example.org/app"`} {
		if !strings.Contains(html, want) {
			t.Errorf("body misses %q", want)
		}
	}
	if strings.Contains(html, "<Alle>") {
		t.Error("club name must be escaped")
	}
	if got := GameDaySettledSubject(mail); got != "Spieltag vom 14.11.2026 abgerechnet – KC <Alle> Neune" {
		t.Errorf("subject = %q", got)
	}
}

func TestDisabledServiceDoesNotSend(t *testing.T) {
	svc := NewService("", "noreply@example.com")
	if svc.Enabled() {
		t.Fatal("service without API key must be disabled")
	}
	if err := svc.SendGameDaySettledEmail("a@example.com", GameDaySettledEmail{}); err == nil {
		t.Fatal("expected error")
	}
}
