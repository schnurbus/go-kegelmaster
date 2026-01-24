import type { ReactNode } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

type Club = {
  id: string;
  name: string;
  balance: number;
  base_fee: number;
  user_id: string;
  created_at: string;
  updated_at: string;
};

type ClubContextValue = {
  activeClub: Club | null;
  clubs: Club[];
  setActiveClub: (club: Club | null) => void;
  setClubs: (clubs: Club[]) => void;
  refreshClubs: () => Promise<void>;
  isLoading: boolean;
};

const ClubContext = createContext<ClubContextValue | undefined>(undefined);

const ACTIVE_CLUB_STORAGE_KEY = "kegelmaster-active-club-id";

export function ClubProvider({ children }: { children: ReactNode }) {
  const [clubs, setClubs] = useState<Club[]>([]);
  const [activeClub, setActiveClubState] = useState<Club | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Lade aktiven Club aus localStorage beim Start
  useEffect(() => {
    const savedClubId = localStorage.getItem(ACTIVE_CLUB_STORAGE_KEY);
    if (savedClubId && clubs.length > 0) {
      const club = clubs.find((c) => c.id === savedClubId);
      if (club) {
        setActiveClubState(club);
      }
    }
  }, [clubs]);

  // Speichere aktiven Club in localStorage
  const setActiveClub = useCallback((club: Club | null) => {
    setActiveClubState(club);
    if (club) {
      localStorage.setItem(ACTIVE_CLUB_STORAGE_KEY, club.id);
    } else {
      localStorage.removeItem(ACTIVE_CLUB_STORAGE_KEY);
    }
  }, []);

  const refreshClubs = useCallback(async () => {
    try {
      setIsLoading(true);
      const resp = await fetch("/api/clubs", {
        credentials: "include",
      });
      if (!resp.ok) {
        throw new Error("Clubs konnten nicht geladen werden");
      }
      const data: Club[] = await resp.json();
      setClubs(data);

      // Wenn kein aktiver Club gesetzt ist, setze den ersten oder den gespeicherten
      if (data.length > 0) {
        const savedClubId = localStorage.getItem(ACTIVE_CLUB_STORAGE_KEY);
        const clubToSet = savedClubId
          ? data.find((c) => c.id === savedClubId) || data[0]
          : data[0];
        setActiveClubState(clubToSet);
        localStorage.setItem(ACTIVE_CLUB_STORAGE_KEY, clubToSet.id);
      } else {
        setActiveClubState(null);
        localStorage.removeItem(ACTIVE_CLUB_STORAGE_KEY);
      }
    } catch (error) {
      console.error("Error fetching clubs:", error);
      setClubs([]);
      setActiveClubState(null);
      localStorage.removeItem(ACTIVE_CLUB_STORAGE_KEY);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Initiales Laden der Clubs
  useEffect(() => {
    refreshClubs();
  }, [refreshClubs]);

  const value = useMemo(
    () => ({
      activeClub,
      clubs,
      setActiveClub,
      setClubs,
      refreshClubs,
      isLoading,
    }),
    [activeClub, clubs, setActiveClub, refreshClubs, isLoading]
  );

  return <ClubContext.Provider value={value}>{children}</ClubContext.Provider>;
}

export function useClub() {
  const ctx = useContext(ClubContext);
  if (!ctx) {
    throw new Error("useClub must be used within ClubProvider");
  }
  return ctx;
}

