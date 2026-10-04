import { Plus } from "lucide-react";
export function Brand({ light = false }: { light?: boolean }) {
  return (
    <div className="flex items-center gap-3">
      <span
        className={`flex h-10 w-10 items-center justify-center rounded-xl ${light ? "bg-white/15 text-white" : "bg-brand text-white"}`}
      >
        <Plus size={29} strokeWidth={3} />
      </span>
      <div>
        <div
          className={`text-xl font-bold tracking-tight ${light ? "text-white" : "text-ink"}`}
        >
          ULSHMS
          <span className={light ? "text-emerald-200" : "text-brand"}>.</span>
        </div>
        <div
          className={`text-[9px] tracking-[.15em] uppercase ${light ? "text-emerald-100/65" : "text-muted"}`}
        >
          Hospital management
        </div>
      </div>
    </div>
  );
}
