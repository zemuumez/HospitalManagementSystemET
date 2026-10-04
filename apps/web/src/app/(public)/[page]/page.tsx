import { notFound } from "next/navigation";
import { PublicSite } from "@/components/public-site";
export default async function Page({
  params,
}: {
  params: Promise<{ page: string }>;
}) {
  const { page } = await params;
  if (
    ![
      "about-us",
      "our-services",
      "doctors",
      "appointment",
      "working-hours",
      "testimonials",
      "contact-us",
      "privacy-policy",
      "terms-of-service",
      "register",
    ].includes(page)
  )
    notFound();
  return <PublicSite page={page} />;
}
