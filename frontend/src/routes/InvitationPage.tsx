import * as React from "react";
import { useNavigate, useParams } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/context/AuthContext";
import { toast } from "sonner";
import type { InvitationResponse } from "@/types/invitation";
import { MailIcon, UserIcon, UsersIcon } from "lucide-react";

function InvitationPage() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const { user, status } = useAuth();
  const [invitation, setInvitation] = React.useState<InvitationResponse | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isAccepting, setIsAccepting] = React.useState(false);

  React.useEffect(() => {
    if (!token) {
      toast.error("Ungültiger Einladungslink");
      navigate("/");
      return;
    }

    const fetchInvitation = async () => {
      try {
        toast.loading("Einladung wird geladen...", { id: "load-invitation" });
        
        const response = await fetch(`/api/invitations/${token}`, {
          credentials: "include",
        });

        if (!response.ok) {
          if (response.status === 404) {
            toast.error("Einladung nicht gefunden. Der Link ist möglicherweise ungültig.", { 
              id: "load-invitation",
              duration: 6000 
            });
            navigate("/");
            return;
          }
          if (response.status === 410) {
            const errorData = await response.json().catch(() => ({}));
            toast.error(
              errorData.message || "Einladung ist abgelaufen oder wurde bereits akzeptiert",
              { id: "load-invitation", duration: 6000 }
            );
            navigate("/");
            return;
          }
          const errorData = await response.json().catch(() => ({}));
          throw new Error(errorData.message || "Fehler beim Laden der Einladung");
        }

        const data: InvitationResponse = await response.json();
        setInvitation(data);
        toast.success("Einladung geladen", { id: "load-invitation", duration: 2000 });
      } catch (error) {
        console.error("Error fetching invitation:", error);
        toast.error(
          error instanceof Error ? error.message : "Fehler beim Laden der Einladung",
          { id: "load-invitation", duration: 6000 }
        );
        navigate("/");
      } finally {
        setIsLoading(false);
      }
    };

    fetchInvitation();
  }, [token, navigate]);

  const handleAccept = async () => {
    if (!token) {
      toast.error("Token fehlt");
      return;
    }

    setIsAccepting(true);
    try {
      toast.loading("Einladung wird akzeptiert...", { id: "accept-invitation" });
      
      const csrfToken = document.cookie
        .split("; ")
        .find((row) => row.startsWith("csrf_token="))
        ?.split("=")[1] || "";

      if (!csrfToken) {
        throw new Error("CSRF-Token fehlt. Bitte Seite neu laden.");
      }

      const response = await fetch(`/api/invitations/${token}/accept`, {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrfToken,
        },
        credentials: "include",
      });

      const responseData = await response.json().catch(() => ({}));

      if (!response.ok) {
        const errorMessage = responseData.message || "Fehler beim Akzeptieren der Einladung";
        
        if (response.status === 404) {
          throw new Error("Einladung nicht gefunden");
        }
        if (response.status === 410) {
          throw new Error("Einladung ist abgelaufen oder wurde bereits akzeptiert");
        }
        
        throw new Error(errorMessage);
      }

      toast.success(
        `Einladung erfolgreich akzeptiert! Sie sind jetzt mit ${invitation?.player_name} verbunden.`,
        { id: "accept-invitation", duration: 5000 }
      );
      
      // Small delay to show success message
      setTimeout(() => {
        if (invitation) {
          navigate(`/app/players/${invitation.player_id}`);
        } else {
          navigate("/app/players");
        }
      }, 1000);
    } catch (error) {
      console.error("Error accepting invitation:", error);
      toast.error(
        error instanceof Error ? error.message : "Fehler beim Akzeptieren der Einladung",
        { id: "accept-invitation", duration: 6000 }
      );
    } finally {
      setIsAccepting(false);
    }
  };

  const handleLogin = () => {
    navigate(`/login?invite_token=${token}`);
  };

  const handleRegister = () => {
    navigate(`/register?invite_token=${token}`);
  };

  if (isLoading) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-screen p-4">
          <Card className="w-full max-w-md">
            <CardHeader>
              <Skeleton className="h-6 w-48" />
              <Skeleton className="h-4 w-64 mt-2" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-32 w-full" />
            </CardContent>
          </Card>
        </div>
      </AppLayout>
    );
  }

  if (!invitation) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-screen p-4">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>Einladung nicht gefunden</CardTitle>
              <CardDescription>
                Die Einladung konnte nicht geladen werden.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button onClick={() => navigate("/")} className="w-full">
                Zur Startseite
              </Button>
            </CardContent>
          </Card>
        </div>
      </AppLayout>
    );
  }

  const isAuthenticated = status === "authenticated" && user !== null;

  return (
    <AppLayout>
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <MailIcon className="size-5" />
              Einladung
            </CardTitle>
            <CardDescription>
              Sie wurden eingeladen, einem Spieler beizutreten
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-4">
              <div className="flex items-start gap-3 p-4 bg-muted rounded-lg">
                <UserIcon className="size-5 mt-0.5 text-muted-foreground" />
                <div className="flex-1">
                  <p className="text-sm font-medium">Spieler</p>
                  <p className="text-sm text-muted-foreground">{invitation.player_name}</p>
                </div>
              </div>
              <div className="flex items-start gap-3 p-4 bg-muted rounded-lg">
                <UsersIcon className="size-5 mt-0.5 text-muted-foreground" />
                <div className="flex-1">
                  <p className="text-sm font-medium">Club</p>
                  <p className="text-sm text-muted-foreground">{invitation.club_name}</p>
                </div>
              </div>
              <div className="flex items-start gap-3 p-4 bg-muted rounded-lg">
                <MailIcon className="size-5 mt-0.5 text-muted-foreground" />
                <div className="flex-1">
                  <p className="text-sm font-medium">E-Mail</p>
                  <p className="text-sm text-muted-foreground">{invitation.email}</p>
                </div>
              </div>
            </div>

            {isAuthenticated ? (
              <div className="space-y-2">
                <Button
                  onClick={handleAccept}
                  disabled={isAccepting}
                  className="w-full"
                >
                  {isAccepting ? "Wird akzeptiert..." : "Einladung akzeptieren"}
                </Button>
                <p className="text-xs text-center text-muted-foreground">
                  Als {user?.email} angemeldet
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                <p className="text-sm text-center text-muted-foreground mb-4">
                  Bitte melden Sie sich an oder registrieren Sie sich, um die Einladung anzunehmen.
                </p>
                <Button onClick={handleLogin} variant="default" className="w-full">
                  Anmelden
                </Button>
                <Button onClick={handleRegister} variant="outline" className="w-full">
                  Registrieren
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

export default InvitationPage;
