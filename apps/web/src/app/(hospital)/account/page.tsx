"use client";
import { ShieldCheck, Smartphone } from "lucide-react";
import { PhoneAuth } from "@/components/phone-auth";
import { useIdentity } from "@/components/workspace";
export default function Account() {
  const identity = useIdentity();
  return (
    <div className="max-w-3xl space-y-6">
      <div>
        <p className="eyebrow text-brand">Your profile</p>
        <h1 className="mt-2 text-[28px] font-semibold tracking-tight">
          Account & security
        </h1>
        <p className="mt-2 text-muted">
          Manage how you access your hospital workspace.
        </p>
      </div>
      <div className="card flex items-center gap-4 p-6">
        <ShieldCheck className="text-brand" />
        <div>
          <p className="font-semibold">{identity?.user.name}</p>
          <p className="mt-1 text-xs capitalize text-muted">
            {identity?.user.role.replaceAll("_", " ")} · permissions managed by
            your hospital
          </p>
        </div>
      </div>
      <div className="card max-w-lg p-7">
        <h2 className="mb-5 flex items-center gap-2 font-semibold">
          <Smartphone size={19} className="text-brand" />
          Link phone sign-in
        </h2>
        <PhoneAuth link />
      </div>
    </div>
  );
}
