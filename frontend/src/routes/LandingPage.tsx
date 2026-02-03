import { Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CalendarIcon, ReceiptIcon, UsersIcon } from "lucide-react";

function LandingPage() {
  const { status } = useAuth();
  const isAuthed = status === "authenticated";

  return (
    <div className="min-h-[80vh] flex flex-col">
      {/* Hero */}
      <section className="py-16 md:py-24 px-4">
        <div className="max-w-3xl mx-auto text-center space-y-6">
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl md:text-6xl text-foreground">
            Kegelabende mühelos verwalten
          </h1>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
            Erfasse Spielabende, Strafen und Einzahlungen für mehrere Klubs in
            einer zentralen Oberfläche.
          </p>
          <div className="flex flex-wrap gap-4 justify-center pt-4">
            {isAuthed ? (
              <Button asChild size="lg" className="rounded-full px-8">
                <Link to="/app">Zum Dashboard</Link>
              </Button>
            ) : (
              <>
                <Button asChild size="lg" variant="default" className="rounded-full px-8">
                  <Link to="/register">Account erstellen</Link>
                </Button>
                <Button asChild size="lg" variant="outline" className="rounded-full px-8">
                  <Link to="/login">Ich habe bereits einen Account</Link>
                </Button>
              </>
            )}
          </div>
        </div>
      </section>

      {/* Feature cards */}
      <section className="py-12 px-4 flex-1" aria-label="Funktionen">
        <div className="max-w-5xl mx-auto">
          <h2 className="sr-only">Funktionen</h2>
          <div className="grid gap-6 sm:grid-cols-3">
            <Card className="border-border/50 bg-card/50">
              <CardHeader>
                <div className="flex size-10 items-center justify-center rounded-lg bg-primary/10 text-primary mb-2">
                  <CalendarIcon className="size-5" aria-hidden />
                </div>
                <CardTitle className="text-lg">Spieltage verwalten</CardTitle>
                <CardDescription>
                  Erfasse Teilnehmer, Strafen und Wettbewerbe pro Spielabend –
                  alles an einem Ort.
                </CardDescription>
              </CardHeader>
            </Card>
            <Card className="border-border/50 bg-card/50">
              <CardHeader>
                <div className="flex size-10 items-center justify-center rounded-lg bg-primary/10 text-primary mb-2">
                  <ReceiptIcon className="size-5" aria-hidden />
                </div>
                <CardTitle className="text-lg">Transaktionen im Blick</CardTitle>
                <CardDescription>
                  Einnahmen, Ausgaben und Spieler-Guthaben übersichtlich
                  geführt und nachvollziehbar.
                </CardDescription>
              </CardHeader>
            </Card>
            <Card className="border-border/50 bg-card/50">
              <CardHeader>
                <div className="flex size-10 items-center justify-center rounded-lg bg-primary/10 text-primary mb-2">
                  <UsersIcon className="size-5" aria-hidden />
                </div>
                <CardTitle className="text-lg">Mehrere Klubs</CardTitle>
                <CardDescription>
                  Verwalte mehrere Vereine oder Klubs in einem Account und
                  wechsle schnell zwischen ihnen.
                </CardDescription>
              </CardHeader>
            </Card>
          </div>
        </div>
      </section>
    </div>
  );
}

export default LandingPage;
