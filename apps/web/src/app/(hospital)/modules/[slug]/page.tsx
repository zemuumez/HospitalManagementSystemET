import { EncounterRegister } from "@/components/encounter-register";
import { AttendanceWorkspace } from "@/components/attendance-workspace";
import { OdontogramRegister } from "@/components/odontogram-register";
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
import { BillingWorkspace } from "@/components/billing-workspace";
import { BedManagementWorkspace } from "@/components/bed-management-workspace";
import { BloodBankWorkspace } from "@/components/blood-bank-workspace";
import { DoctorsWorkspace } from "@/components/doctors-workspace";
import { PrescriptionsWorkspace } from "@/components/prescriptions-workspace";
import { DiagnosisWorkspace } from "@/components/diagnosis-workspace";
import { ServicesWorkspace } from "@/components/services-workspace";
import { PathologyWorkspace } from "@/components/pathology-workspace";
import { PatientsWorkspace } from "@/components/patients-workspace";
import { MedicinesWorkspace } from "@/components/medicines-workspace";
import { LiveConsultationWorkspace } from "@/components/live-consultation-workspace";
import { ReviewWorkspace } from "@/components/review-workspace";
import { FrontOfficeWorkspace } from "@/components/front-office-workspace";
import { InventoryWorkspace } from "@/components/inventory-workspace";
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
  if (slug === "accounts") return <ConnectedAccounts key={slug} />;
  if (slug === "invoices") return <ConnectedInvoices key={slug} />;
  if (
    [
      "insurances",
      "packages",
      "services",
      "ambulances",
      "ambulance-calls",
    ].includes(slug)
  )
    return <ServicesWorkspace key={slug} id={slug} />;
  if (
    [
      "pathology-categories",
      "pathology-units",
      "pathology-parameters",
      "pathology-tests",
    ].includes(slug)
  )
    return <PathologyWorkspace key={slug} id={slug} />;
  if (["patients", "case-handlers", "patient-admissions"].includes(slug))
    return <PatientsWorkspace key={slug} id={slug} />;
  if (
    [
      "medicine-categories",
      "brands",
      "medicines",
      "purchase-medicines",
      "used-medicine",
      "medicine-bills",
    ].includes(slug)
  )
    return <MedicinesWorkspace key={slug} id={slug} />;
  if (
    [
      "live-consultations",
      "live-consultations-live-meetings",
      "live-meetings",
    ].includes(slug)
  )
    return <LiveConsultationWorkspace key={slug} id={slug} />;
  if (["reviews", "review"].includes(slug))
    return <ReviewWorkspace key={slug} id={slug} />;
  if (
    [
      "blood-banks",
      "blood-donors",
      "blood-donations",
      "blood-issues",
      "blood-donor-reports",
    ].includes(slug)
  )
    return <BloodBankWorkspace key={slug} id={slug} />;
  if (
    [
      "doctors",
      "doctor-departments",
      "doctor-holidays",
      "holidays",
      "breaks",
    ].includes(slug)
  )
    return <DoctorsWorkspace key={slug} id={slug} />;
  if (slug === "prescriptions") return <PrescriptionsWorkspace key={slug} />;
  if (["diagnosis-categories", "patient-diagnosis-test"].includes(slug))
    return <DiagnosisWorkspace key={slug} id={slug} />;
  if (["bed-status", "bed-assigns", "beds", "bed-types"].includes(slug))
    return <BedManagementWorkspace key={slug} id={slug} />;
  if (
    [
      "manual-billing-payments",
      "manual-bill-payments",
      "advance-payments",
      "advanced-payments",
      "payment-reports",
      "payments",
      "employee-payrolls",
      "bills",
      "expenses",
      "incomes",
      "expense-heads",
      "income-heads",
    ].includes(slug)
  )
    return <BillingWorkspace key={slug} id={slug} />;
  if (
    slug === "attendance" ||
    slug.startsWith("attendance-") ||
    slug === "manage-attendance"
  )
    return <AttendanceWorkspace key={slug} id={slug} />;
  if (
    [
      "call-logs",
      "visitors",
      "postals",
      "postal-receives",
      "postal-dispatches",
      "enquiries",
      "complaints",
    ].includes(slug)
  )
    return <FrontOfficeWorkspace key={slug} id={slug} />;
  if (
    ["items", "item-categories", "item-stocks", "issued-items"].includes(slug)
  )
    return <InventoryWorkspace key={slug} id={slug} />;
  const screen = screens.find((s) => s.id === slug);
  if (!screen) notFound();
  if (slug === "odontogram") return <OdontogramRegister />;
  if (slug === "patient-cases") return <ConnectedClinical mode="cases" />;
  if (
    [
      "ipd-patient-departments",
      "ipd-diagnosis",
      "ipd-consultant-registers",
      "ipd-prescriptions",
      "ipd-charges",
      "ipd-payments",
      "ipd-bills",
      "ipd-timelines",
    ].includes(slug)
  )
    return <EncounterRegister key={slug} kind="ipd" />;
  if (
    ["opd-patient-departments", "opd-diagnosis", "opd-timelines"].includes(slug)
  )
    return <EncounterRegister key={slug} kind="opd" />;
  if (
    [
      "users",
      "admins",
      "receptionists",
      "pharmacists",
      "accountants",
      "case-managers",
      "lab-technicians",
      "nurses",
    ].includes(slug)
  )
    return (
      <ConnectedUsers
        key={slug}
        roleFilter={slug === "users" ? undefined : slug}
      />
    );
  if (slug === "schedules") return <ConnectedSchedules key={slug} />;
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
