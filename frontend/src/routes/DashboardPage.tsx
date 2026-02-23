import {
  DashboardLastGameDayResults,
  DashboardWinnersLosers,
} from "@/components/dashboard-competition-cards";
import { useDashboardCompetitionData } from "@/hooks/use-dashboard-competition-data";
import { DashboardCompetitionChart } from "@/components/dashboard-competition-chart";
import { DashboardPenaltyChart } from "@/components/dashboard-penalty-chart";
import { DashboardTransactionTable } from "@/components/dashboard-transaction-table";
import { SectionCards } from "@/components/section-cards";

function DashboardPage() {
  const competitionData = useDashboardCompetitionData();

  return (
    <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
      <div className="*:data-[slot=card]:shadow-xs @xl/main:grid-cols-2 @5xl/main:grid-cols-4 grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:from-primary/5 *:data-[slot=card]:to-card dark:*:data-[slot=card]:bg-card lg:px-6">
        <SectionCards />
        <DashboardLastGameDayResults data={competitionData} />
        <DashboardWinnersLosers data={competitionData} />
      </div>
      <DashboardPenaltyChart />
      <DashboardCompetitionChart />
      <DashboardTransactionTable />
    </div>
  );
}

export default DashboardPage;
