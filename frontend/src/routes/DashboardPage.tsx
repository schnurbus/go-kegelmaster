import {
  DashboardLastGameDayResults,
  DashboardWinnersLosers,
} from "@/components/dashboard-competition-cards";
import { useDashboardCompetitionData } from "@/hooks/use-dashboard-competition-data";
import { DashboardCompetitionChart } from "@/components/dashboard-competition-chart";
import { DashboardPenaltyChart } from "@/components/dashboard-penalty-chart";
import { DashboardTransactionTable } from "@/components/dashboard-transaction-table";
import { SectionCards } from "@/components/section-cards";
import { SetupCard } from "@/components/setup-card";
import { TourAutoStart } from "@/components/tour/TourAutoStart";
import { WelcomeCard } from "@/components/welcome-card";
import { useAuth } from "@/context/AuthContext";
import { useClub } from "@/context/ClubContext";
import { useClubSetup } from "@/hooks/use-club-setup";

function DashboardPage() {
  const { user } = useAuth();
  const { activeClub, clubs, isLoading } = useClub();

  if (!isLoading && clubs.length === 0) {
    return <WelcomeCard />;
  }
  if (!activeClub || !user) {
    return null;
  }
  return <ClubDashboard key={activeClub.id} clubId={activeClub.id} isOwner={activeClub.user_id === user.id} />;
}

function ClubDashboard({ clubId, isOwner }: { clubId: string; isOwner: boolean }) {
  const competitionData = useDashboardCompetitionData();
  const { setup, ready, visible, reload } = useClubSetup(clubId, isOwner);

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      {visible && setup && <SetupCard clubId={clubId} setup={setup} onChanged={reload} />}
      <div
        data-tour="dashboard-cards"
        className="*:data-[slot=card]:shadow-xs @xl/main:grid-cols-2 @5xl/main:grid-cols-4 grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:from-primary/5 *:data-[slot=card]:to-card dark:*:data-[slot=card]:bg-card lg:px-6"
      >
        <SectionCards />
        <DashboardLastGameDayResults data={competitionData} />
        <DashboardWinnersLosers data={competitionData} />
      </div>
      <DashboardPenaltyChart />
      <DashboardCompetitionChart />
      <DashboardTransactionTable />
      <TourAutoStart setupReady={ready} showSetupCard={visible} />
    </div>
  );
}

export default DashboardPage;
