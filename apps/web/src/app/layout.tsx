import type { Metadata } from "next";
import "./globals.css";
import { LanguageProvider } from "@/components/language";
export const metadata: Metadata = {
  title: { default: "ULSHMS · Hospital workspace", template: "%s · ULSHMS" },
  description: "One connected workspace for your hospital.",
  icons: { icon: "/favicon.png" },
};
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <LanguageProvider>{children}</LanguageProvider>
      </body>
    </html>
  );
}
