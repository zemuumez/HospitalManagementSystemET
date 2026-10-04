import {
  ConnectedUsers,
  ConnectedSchedules,
  ConnectedAppointments,
} from "@/components/connected-scheduling";
import { ConnectedClinical } from "@/components/connected-clinical";
import {
  ConnectedAccounts,
  ConnectedInvoices,
} from "@/components/connected-billing";
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
  if (slug === "accounts") return <ConnectedAccounts />;
  if (slug === "invoices") return <ConnectedInvoices />;
  if (slug === "beds" || slug === "bed-status")
    return <ConnectedClinical mode="beds" />;
  if (slug === "patient-cases") return <ConnectedClinical mode="cases" />;
  if (slug === "ipd-patient-departments")
    return <ConnectedClinical mode="ipd" />;
  if (slug === "opd-patient-departments")
    return <ConnectedClinical mode="opd" />;
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
