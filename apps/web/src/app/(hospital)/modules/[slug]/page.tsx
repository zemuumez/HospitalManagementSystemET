import {
  ConnectedUsers,
  ConnectedSchedules,
  ConnectedAppointments,
} from "@/components/connected-scheduling";
import { notFound } from "next/navigation";
import { screens } from "@/lib/legacy";
import { LegacyScreen } from "@/components/legacy-screen";
import { LegacyExtras } from "@/components/legacy-extras";
export default async function ModulePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const screen = screens.find((s) => s.id === slug);
  if (!screen) notFound();
  if (slug === "users") return <ConnectedUsers />;
  if (slug === "schedules") return <ConnectedSchedules />;
  if (slug === "appointments") return <ConnectedAppointments />;
  if (
    [
      "front-settings",
      "settings",
      "modules-setting",
      "patient-queue-theme",
      "attendance",
      "manage-attendance",
      "patient-id-card-template",
      "generate-patient-id-card",
    ].includes(slug)
  )
    return <LegacyExtras key={slug} id={slug} />;
  return <LegacyScreen key={screen.id} screen={screen} />;
}
