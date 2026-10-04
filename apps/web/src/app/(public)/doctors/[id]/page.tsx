import { notFound } from "next/navigation";
import { PublicSite } from "@/components/public-site";
export default async function DoctorDetails({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  if (!["1", "2", "3"].includes(id)) notFound();
  return <PublicSite page="doctor-details" doctorIndex={Number(id) - 1} />;
}
