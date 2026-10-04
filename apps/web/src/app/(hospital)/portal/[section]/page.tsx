import { notFound } from "next/navigation";
import { portalSections } from "@/lib/role-preview";
import { RolePortal } from "@/components/role-portal";
export default async function PortalPage({
  params,
}: {
  params: Promise<{ section: string }>;
}) {
  const { section } = await params;
  if (!portalSections[section]) notFound();
  return <RolePortal key={section} section={section} />;
}
