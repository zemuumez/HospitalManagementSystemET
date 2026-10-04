import { LegacyScreen } from "@/components/legacy-screen";
import { screens } from "@/lib/legacy";
export default function Patients() {
  return <LegacyScreen screen={screens.find((s) => s.id === "patients")!} />;
}
