import * as React from "react";
import { PlusIcon, UsersIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { CreateClubDialog } from "@/components/create-club-dialog";

/** Willkommensseite für Benutzer ohne Club, statt des leeren Dashboards. */
export function WelcomeCard() {
  const [open, setOpen] = React.useState(false);

  return (
    <div className="flex flex-1 items-center justify-center px-4 py-12">
      <Card className="w-full max-w-lg" aria-labelledby="welcome-title">
        <CardHeader className="text-center">
          <div className="mx-auto mb-2 flex size-12 items-center justify-center rounded-full bg-muted">
            <UsersIcon className="size-6 text-muted-foreground" aria-hidden />
          </div>
          <CardTitle id="welcome-title" className="text-2xl">
            Willkommen beim Kegelmaster
          </CardTitle>
          <CardDescription>
            Sie sind noch in keinem Club. Legen Sie zuerst einen Club an. Danach zeigen wir Ihnen in einem kurzen
            Rundgang, wo Sie was finden, und die Ersteinrichtung führt Sie bis zum ersten Spieltag.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-3">
          <Button size="lg" onClick={() => setOpen(true)}>
            <PlusIcon />
            Club anlegen
          </Button>
          <p className="text-center text-sm text-muted-foreground">
            Wurden Sie in einen Club eingeladen? Dann öffnen Sie den Link aus der Einladungs-E-Mail.
          </p>
        </CardContent>
      </Card>
      <CreateClubDialog open={open} onOpenChange={setOpen} onClubCreated={() => setOpen(false)} />
    </div>
  );
}
