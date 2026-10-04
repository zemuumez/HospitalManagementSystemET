import { ConnectedAppointments } from "@/components/connected-scheduling";
import { ConnectedClinical } from "@/components/connected-clinical";
import { notFound } from "next/navigation";
import { portalSections } from "@/lib/role-preview";
import { RolePortal } from "@/components/role-portal";
export default async function PortalPage({
  params,
}: {
  params: Promise<{ section: string }>;
}) {
  const { section } = await params;
  if (section === "appointments") return <ConnectedAppointments />;
  if (section === "ipd" || section === "opd" || section === "cases")
    return <ConnectedClinical mode={section} />;
  if (!portalSections[section]) notFound();
  return <RolePortal key={section} section={section} />;
}
