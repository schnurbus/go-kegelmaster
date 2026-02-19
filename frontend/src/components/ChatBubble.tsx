import * as React from "react";
import { MessageCircle, Send } from "lucide-react";
import { useLocation } from "react-router-dom";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { useAuth } from "@/context/AuthContext";
import { cn } from "@/lib/utils";

type Message = { role: "user" | "assistant"; text: string };

export function ChatBubble() {
  const { status, csrfToken, refreshCsrf } = useAuth();
  const location = useLocation();
  const [open, setOpen] = React.useState(false);
  const [messages, setMessages] = React.useState<Message[]>([]);
  const [input, setInput] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const messagesEndRef = React.useRef<HTMLDivElement>(null);
  const inputRef = React.useRef<HTMLInputElement>(null);

  React.useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  React.useEffect(() => {
    if (!loading) {
      inputRef.current?.focus();
    }
  }, [loading]);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();
    const text = input.trim();
    if (!text || loading) return;

    const historyToSend = messages.slice(-5);
    setInput("");
    setMessages((prev) => [...prev, { role: "user", text }]);
    setLoading(true);

    try {
      const token = csrfToken ?? (await refreshCsrf());
      const response = await fetch("/api/chat", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": token,
        },
        body: JSON.stringify({
          message: text,
          history: historyToSend,
          current_page: location.pathname,
        }),
      });

      const data = await response.json().catch(() => ({}));

      if (!response.ok) {
        const msg =
          response.status === 429
            ? "Zu viele Anfragen. Bitte kurz warten."
            : (data as { message?: string }).message ?? "Fehler beim Senden";
        toast.error(msg);
        setMessages((prev) => prev.slice(0, -1));
        return;
      }

      const reply = (data as { reply?: string }).reply ?? "";
      setMessages((prev) => [...prev, { role: "assistant", text: reply }]);
    } catch (error) {
      console.error("Chat request failed:", error);
      toast.error("Assistent vorübergehend nicht erreichbar");
      setMessages((prev) => prev.slice(0, -1));
    } finally {
      setLoading(false);
    }
  };

  if (status !== "authenticated") {
    return null;
  }

  return (
    <>
      <Button
        type="button"
        variant="default"
        size="icon-lg"
        className="fixed bottom-6 right-6 z-50 size-14 rounded-full shadow-lg"
        aria-label="Hilfe-Chat öffnen"
        onClick={() => setOpen(true)}
      >
        <MessageCircle className="size-6" />
      </Button>

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent
          side="right"
          className="flex w-full flex-col border-l p-0 sm:max-w-md"
        >
          <SheetHeader className="border-b px-4 py-3">
            <SheetTitle>Hilfe zur App</SheetTitle>
            <SheetDescription>
              Stelle Fragen zur Bedienung der App.
            </SheetDescription>
          </SheetHeader>

          <div className="flex min-h-0 flex-1 flex-col">
            <div className="flex-1 overflow-y-auto p-4 space-y-3">
              {messages.length === 0 && (
                <p className="text-muted-foreground text-sm">
                  Schreib eine Frage zur Kegelmaster-App, z.B. wie du einen
                  Spielabend anlegst oder Strafen einträgst.
                </p>
              )}
              {messages.map((m, i) => (
                <div
                  key={i}
                  className={cn(
                    "rounded-lg px-3 py-2 text-sm",
                    m.role === "user"
                      ? "ml-8 bg-primary text-primary-foreground"
                      : "mr-8 bg-muted"
                  )}
                >
                  {m.text}
                </div>
              ))}
              {loading && (
                <div className="mr-8 rounded-lg bg-muted px-3 py-2 text-sm text-muted-foreground">
                  …
                </div>
              )}
              <div ref={messagesEndRef} />
            </div>

            <form
              onSubmit={handleSend}
              className="flex gap-2 border-t p-4"
            >
              <Input
                ref={inputRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="Frage zur App…"
                disabled={loading}
                maxLength={2000}
                className="min-w-0 flex-1"
              />
              <Button type="submit" size="icon" disabled={loading || !input.trim()}>
                <Send className="size-4" />
                <span className="sr-only">Senden</span>
              </Button>
            </form>
          </div>
        </SheetContent>
      </Sheet>
    </>
  );
}
