import { Outlet, useLocation } from "react-router-dom";
import { AppFooter } from "./components/AppFooter";
import { ChatBubble } from "./components/ChatBubble";
import { PublicSiteHeader } from "./components/site-header-public";
import "./App.css";

function App() {
  const location = useLocation();
  const isDashboard =
    location.pathname.startsWith("/app") || location.pathname.startsWith("/admin");

  // Dashboard hat seine eigene Sidebar, daher keinen Header anzeigen
  if (isDashboard) {
    return (
      <>
        <Outlet />
        <ChatBubble />
      </>
    );
  }

  return (
    <div className="flex min-h-screen flex-col">
      <PublicSiteHeader />
      <main className="flex-1">
        <Outlet />
      </main>
      <AppFooter className="border-t border-border px-4 py-6 text-center text-sm" />
      <ChatBubble />
    </div>
  );
}

export default App;
