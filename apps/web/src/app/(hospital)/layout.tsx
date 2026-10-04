import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { auth } from "@/lib/auth";
import { Workspace } from "@/components/workspace";
export const dynamic = "force-dynamic";
export default async function HospitalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const session = await auth.api.getSession({ headers: await headers() });
  if (!session) redirect("/login");
  return <Workspace>{children}</Workspace>;
}
