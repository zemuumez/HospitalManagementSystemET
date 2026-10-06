"use client";

import { useState, useMemo, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Search,
  Plus,
  Filter,
  Edit2,
  Trash2,
  X,
  ChevronDown,
  Download,
  Activity,
  Bed as BedIcon,
  Check,
  Eye,
  AlertCircle,
  Loader2,
  RefreshCw,
  CheckCircle2,
} from "lucide-react";
import { useLanguage } from "./language";
import { api, type Patient } from "@/lib/api";

export type BedTab = "bed-status" | "bed-assigns" | "beds" | "bed-types";

export interface ApiBed {
  id: string;
  name: string;
  type: string;
  typeId?: string;
  chargeMinor: number;
  available: boolean;
  state?: string;
  version: number;
}

export interface ApiBedType {
  id: string;
  name: string;
  description: string;
  active: boolean;
  version: number;
}

export interface ApiBedOccupancyReport {
  totalBeds: number;
  occupiedBeds: number;
  availableBeds: number;
  occupancyRate: number;
  activeAdmissions: number;
}

interface BedStatusItem {
  id: string;
  name: string;
  isAvailable: boolean;
  patientName?: string;
  phone?: string;
  admissionDate?: string;
  gender?: "Male" | "Female";
}

interface WardGroup {
  id: string;
  name: string;
  beds: BedStatusItem[];
}

interface BedAssignRow {
  id: string;
  ipdNo: string;
  patientName: string;
  patientEmail: string;
  initials: string;
  avatarColor: string;
  bedName: string;
  assignDate: string;
  dischargeDate: string;
  status: boolean; // true = active, false = discharged/inactive
  caseId?: string;
  description?: string;
}

interface BedRow {
  id: string;
  bedId: string;
  bedName: string;
  bedType: string;
  charge: number;
  available: boolean;
  description?: string;
}

interface BedTypeRow {
  id: string;
  title: string;
  description?: string;
  bedsCount?: number;
}

// Initial wards & beds strictly matching screenshots 1–6
const initialWards: WardGroup[] = [
  {
    id: "icu1",
    name: "ICU1",
    beds: [
      {
        id: "b1",
        name: "11222",
        isAvailable: false,
        patientName: "Bahrom Parpiyev",
        phone: "+3519253417748",
        admissionDate: "20th Sep, 2026 12:00:00 AM",
        gender: "Male",
      },
      {
        id: "b2",
        name: "ICU-102",
        isAvailable: false,
        patientName: "Zaki Zaki",
        phone: "+971501234567",
        admissionDate: "18th Sep, 2026 09:30:00 AM",
        gender: "Male",
      },
      {
        id: "b3",
        name: "ICU-103",
        isAvailable: false,
        patientName: "Ashish Chaudhary",
        phone: "+919876543210",
        admissionDate: "15th Sep, 2026 11:15:00 AM",
        gender: "Male",
      },
      {
        id: "b4",
        name: "ICU-104",
        isAvailable: false,
        patientName: "Aadarsha Subedi",
        phone: "+9779812345678",
        admissionDate: "12th Sep, 2026 04:45:00 PM",
        gender: "Male",
      },
      {
        id: "b5",
        name: "11222",
        isAvailable: false,
        patientName: "A B",
        phone: "+3519253417748",
        admissionDate: "20th Sep, 2026 12:00:00 AM",
        gender: "Male",
      },
    ],
  },
  {
    id: "nicu",
    name: "NICU",
    beds: [{ id: "b6", name: "bed11231", isAvailable: true }],
  },
  {
    id: "vip-ward",
    name: "VIP Ward",
    beds: [], // Empty -> displays "No Bed Available"
  },
  {
    id: "private-ward",
    name: "Private Ward",
    beds: [
      {
        id: "b7",
        name: "PW-201",
        isAvailable: false,
        patientName: "345345 2423432",
        phone: "+12025550143",
        admissionDate: "10th Aug, 2026 08:00:00 AM",
        gender: "Male",
      },
      {
        id: "b8",
        name: "PW-202",
        isAvailable: false,
        patientName: "Abdul Basheer",
        phone: "+966501234567",
        admissionDate: "05th Aug, 2026 02:20:00 PM",
        gender: "Male",
      },
    ],
  },
  {
    id: "general-ward-female",
    name: "General Ward Female",
    beds: [
      { id: "b9", name: "bed6", isAvailable: true },
      {
        id: "b10",
        name: "GWF-02",
        isAvailable: false,
        patientName: "34 3432",
        phone: "+12025550199",
        admissionDate: "14th Jul, 2026 01:10:00 PM",
        gender: "Female",
      },
      {
        id: "b11",
        name: "GWF-03",
        isAvailable: false,
        patientName: "Nita Black",
        phone: "+14155552671",
        admissionDate: "12th Jul, 2026 10:00:00 AM",
        gender: "Female",
      },
      {
        id: "b12",
        name: "GWF-04",
        isAvailable: false,
        patientName: "Ttt 123",
        phone: "+447700900077",
        admissionDate: "10th Jul, 2026 03:30:00 PM",
        gender: "Female",
      },
      {
        id: "b13",
        name: "GWF-05",
        isAvailable: false,
        patientName: "AMINUR ISLAM",
        phone: "+8801712345678",
        admissionDate: "08th Jul, 2026 11:20:00 AM",
        gender: "Female",
      },
      {
        id: "b14",
        name: "GWF-06",
        isAvailable: false,
        patientName: "Muhammad Umair Aslam",
        phone: "+923001234567",
        admissionDate: "06th Jul, 2026 09:00:00 AM",
        gender: "Female",
      },
      {
        id: "b15",
        name: "GWF-07",
        isAvailable: false,
        patientName: "Visdhal Vishal",
        phone: "+919811223344",
        admissionDate: "04th Jul, 2026 04:15:00 PM",
        gender: "Female",
      },
      {
        id: "b16",
        name: "GWF-08",
        isAvailable: false,
        patientName: "Ashish Chaudhary",
        phone: "+919876543210",
        admissionDate: "02nd Jul, 2026 08:30:00 AM",
        gender: "Female",
      },
      {
        id: "b17",
        name: "GWF-09",
        isAvailable: false,
        patientName: "Pasien Satu",
        phone: "+628123456789",
        admissionDate: "01st Jul, 2026 05:40:00 PM",
        gender: "Female",
      },
      {
        id: "b18",
        name: "GWF-10",
        isAvailable: false,
        patientName: "PST GLOBAL SENTOSA",
        phone: "+628198765432",
        admissionDate: "28th Jun, 2026 07:15:00 AM",
        gender: "Female",
      },
      {
        id: "b19",
        name: "GWF-11",
        isAvailable: false,
        patientName: "Jayme Kirby",
        phone: "+12125550188",
        admissionDate: "25th Jun, 2026 12:00:00 PM",
        gender: "Female",
      },
      {
        id: "b20",
        name: "GWF-12",
        isAvailable: false,
        patientName: "Saleem Basha",
        phone: "+966551122334",
        admissionDate: "20th Jun, 2026 03:00:00 PM",
        gender: "Female",
      },
    ],
  },
  {
    id: "general-ward-male",
    name: "General Ward Male",
    beds: [
      { id: "b21", name: "bed4", isAvailable: true },
      {
        id: "b22",
        name: "GWM-02",
        isAvailable: false,
        patientName: "AHMED ALL",
        phone: "+201012345678",
        admissionDate: "18th Jun, 2026 02:45:00 PM",
        gender: "Male",
      },
      {
        id: "b23",
        name: "GWM-03",
        isAvailable: false,
        patientName: "Ayush Agrawal",
        phone: "+919988776655",
        admissionDate: "15th Jun, 2026 10:15:00 AM",
        gender: "Male",
      },
      { id: "b24", name: "new3", isAvailable: true },
      {
        id: "b25",
        name: "GWM-05",
        isAvailable: false,
        patientName: "PST GLOBAL SENTOSA 2",
        phone: "+628198765432",
        admissionDate: "12th Jun, 2026 09:30:00 AM",
        gender: "Male",
      },
      {
        id: "b26",
        name: "GWM-06",
        isAvailable: false,
        patientName: "Abdullah Geyik",
        phone: "+905321234567",
        admissionDate: "10th Jun, 2026 04:00:00 PM",
        gender: "Male",
      },
      {
        id: "b27",
        name: "GWM-07",
        isAvailable: false,
        patientName: "Ayush Agrawal",
        phone: "+919988776655",
        admissionDate: "08th Jun, 2026 11:20:00 AM",
        gender: "Male",
      },
      { id: "b28", name: "numb3", isAvailable: true },
      {
        id: "b29",
        name: "GWM-09",
        isAvailable: false,
        patientName: "Faridz Test",
        phone: "+60123456789",
        admissionDate: "05th Jun, 2026 01:10:00 PM",
        gender: "Male",
      },
    ],
  },
  {
    id: "nulla-assumenda",
    name: "Nulla assumenda ex u",
    beds: [],
  },
  {
    id: "john-kennedy",
    name: "John Kennedy Beds",
    beds: [],
  },
  {
    id: "pd-ward",
    name: "PD Ward",
    beds: [
      {
        id: "b30",
        name: "PD-101",
        isAvailable: false,
        patientName: "AHMED ALL",
        phone: "+201012345678",
        admissionDate: "01st Jun, 2026 08:30:00 AM",
        gender: "Male",
      },
      {
        id: "b31",
        name: "PD-102",
        isAvailable: false,
        patientName: "Abdo Tester",
        phone: "+201122334455",
        admissionDate: "28th May, 2026 12:45:00 PM",
        gender: "Male",
      },
      {
        id: "b32",
        name: "PD-103",
        isAvailable: false,
        patientName: "Steve Rogers",
        phone: "+12125550100",
        admissionDate: "25th May, 2026 03:15:00 PM",
        gender: "Male",
      },
    ],
  },
  {
    id: "asas",
    name: "ASAs",
    beds: [
      {
        id: "b33",
        name: "ASA-01",
        isAvailable: false,
        patientName: "Ibrahim Abdullahi Mo...",
        phone: "+254711223344",
        admissionDate: "22nd May, 2026 10:00:00 AM",
        gender: "Male",
      },
      {
        id: "b34",
        name: "ASA-02",
        isAvailable: false,
        patientName: "Bipin Sahani",
        phone: "+919765432100",
        admissionDate: "20th May, 2026 04:30:00 PM",
        gender: "Male",
      },
      { id: "b35", name: "Test1155456", isAvailable: true },
      { id: "b36", name: "number2", isAvailable: true },
      { id: "b37", name: "Cama Prueba 2022 04 14", isAvailable: true },
      {
        id: "b38",
        name: "ASA-06",
        isAvailable: false,
        patientName: "Noel Whitaker",
        phone: "+14155550144",
        admissionDate: "18th May, 2026 09:15:00 AM",
        gender: "Male",
      },
      {
        id: "b39",
        name: "ASA-07",
        isAvailable: false,
        patientName: "Trith Shah",
        phone: "+919822334455",
        admissionDate: "15th May, 2026 02:00:00 PM",
        gender: "Male",
      },
      {
        id: "b40",
        name: "ASA-08",
        isAvailable: false,
        patientName: "Giancarlo Alfau",
        phone: "+18095551234",
        admissionDate: "12th May, 2026 11:45:00 AM",
        gender: "Male",
      },
      { id: "b41", name: "gdfg", isAvailable: true },
      {
        id: "b42",
        name: "ASA-10",
        isAvailable: false,
        patientName: "Asmara Simbachawene",
        phone: "+255712345678",
        admissionDate: "10th May, 2026 08:30:00 AM",
        gender: "Female",
      },
      {
        id: "b43",
        name: "ASA-11",
        isAvailable: false,
        patientName: "Ateeb Afzal",
        phone: "+923211234567",
        admissionDate: "08th May, 2026 01:20:00 PM",
        gender: "Male",
      },
    ],
  },
  {
    id: "regular",
    name: "Regular",
    beds: [
      {
        id: "b44",
        name: "REG-01",
        isAvailable: false,
        patientName: "Ayush Agrawal",
        phone: "+919988776655",
        admissionDate: "05th May, 2026 10:00:00 AM",
        gender: "Male",
      },
    ],
  },
  {
    id: "al-funduqiyah",
    name: "الفندقية",
    beds: [],
  },
  {
    id: "general-ward",
    name: "General Ward",
    beds: [
      {
        id: "b45",
        name: "GW-01",
        isAvailable: false,
        patientName: "AHMED ALL",
        phone: "+201012345678",
        admissionDate: "01st May, 2026 03:00:00 PM",
        gender: "Male",
      },
      { id: "b46", name: "hr", isAvailable: true },
      {
        id: "b47",
        name: "GW-03",
        isAvailable: false,
        patientName: "111111 111111",
        phone: "+12025550177",
        admissionDate: "28th Apr, 2026 09:30:00 AM",
        gender: "Male",
      },
      {
        id: "b48",
        name: "GW-04",
        isAvailable: false,
        patientName: "Bahrom Parpiyev",
        phone: "+3519253417748",
        admissionDate: "25th Apr, 2026 02:15:00 PM",
        gender: "Male",
      },
      { id: "b49", name: "G2", isAvailable: true },
      { id: "b50", name: "numb4", isAvailable: true },
      {
        id: "b51",
        name: "GW-07",
        isAvailable: false,
        patientName: "AHMED ALL",
        phone: "+201012345678",
        admissionDate: "22nd Apr, 2026 11:00:00 AM",
        gender: "Male",
      },
      { id: "b52", name: "uzel", isAvailable: true },
      {
        id: "b53",
        name: "GW-09",
        isAvailable: false,
        patientName: "Vishwjeet Kumar",
        phone: "+919833445566",
        admissionDate: "20th Apr, 2026 08:45:00 AM",
        gender: "Male",
      },
    ],
  },
  {
    id: "mind-upset",
    name: "MIND UPSET",
    beds: [
      {
        id: "b54",
        name: "MU-01",
        isAvailable: false,
        patientName: "Ashish Chaudhary",
        phone: "+919876543210",
        admissionDate: "15th Apr, 2026 04:00:00 PM",
        gender: "Male",
      },
    ],
  },
  {
    id: "bore-adikuithu",
    name: "BORE ADIKUITHU",
    beds: [
      {
        id: "b55",
        name: "BA-01",
        isAvailable: false,
        patientName: "Ayush Agrawal",
        phone: "+919988776655",
        admissionDate: "12th Apr, 2026 01:30:00 PM",
        gender: "Male",
      },
    ],
  },
  {
    id: "tension-agathu",
    name: "TENSION AGATHU",
    beds: [
      {
        id: "b56",
        name: "TA-01",
        isAvailable: false,
        patientName: "Balaji Bhushan",
        phone: "+919711223344",
        admissionDate: "10th Apr, 2026 10:15:00 AM",
        gender: "Male",
      },
    ],
  },
  {
    id: "tamil",
    name: "TAMIL",
    beds: [
      {
        id: "b57",
        name: "TM-01",
        isAvailable: false,
        patientName: "Pankaj Kumar",
        phone: "+919822110099",
        admissionDate: "05th Apr, 2026 03:45:00 PM",
        gender: "Male",
      },
    ],
  },
];

// Initial Bed Assigns strictly matching Screenshot 7
const initialBedAssigns: BedAssignRow[] = [
  {
    id: "ba-1",
    ipdNo: "HMS14",
    patientName: "A B",
    patientEmail: "ab@gmail.com",
    initials: "AB",
    avatarColor: "#f59e0b",
    bedName: "mohamed",
    assignDate: "8th Aug,2026",
    dischargeDate: "23rd September 2026",
    status: false,
  },
  {
    id: "ba-2",
    ipdNo: "HMS13",
    patientName: "Vinay Grover",
    patientEmail: "vinaygrover14365@gmail.com",
    initials: "VG",
    avatarColor: "#3b82f6",
    bedName: "lit 14",
    assignDate: "5th Aug,2026",
    dischargeDate: "N/A",
    status: true,
  },
  {
    id: "ba-3",
    ipdNo: "HMS12",
    patientName: "111111 111111",
    patientEmail: "111111@qwe.com",
    initials: "11",
    avatarColor: "#10b981",
    bedName: "Muhammad Haseeb",
    assignDate: "5th Aug,2026",
    dischargeDate: "N/A",
    status: true,
  },
  {
    id: "ba-4",
    ipdNo: "HMS11",
    patientName: "Aljun Cardona",
    patientEmail: "cardona.aljun@gmail.com",
    initials: "AC",
    avatarColor: "#f97316",
    bedName: "Student Visa USA",
    assignDate: "23rd Jul,2026",
    dischargeDate: "N/A",
    status: true,
  },
  {
    id: "ba-5",
    ipdNo: "HMS10",
    patientName: "123 123",
    patientEmail: "oshannimesh@gmail.com",
    initials: "11",
    avatarColor: "#f59e0b",
    bedName: "VVIP -1",
    assignDate: "16th Jul,2026",
    dischargeDate: "N/A",
    status: false,
  },
  {
    id: "ba-6",
    ipdNo: "HMS08",
    patientName: "Jyothi Laxmi",
    patientEmail: "jyothi@gmail.com",
    initials: "JL",
    avatarColor: "#3b82f6",
    bedName: "nik",
    assignDate: "9th Jun,2026",
    dischargeDate: "N/A",
    status: true,
  },
  {
    id: "ba-7",
    ipdNo: "HMS07",
    patientName: "11111 1111",
    patientEmail: "sulapojim@mailinator.com",
    initials: "11",
    avatarColor: "#10b981",
    bedName: "ICU12",
    assignDate: "9th Jun,2026",
    dischargeDate: "N/A",
    status: true,
  },
  {
    id: "ba-8",
    ipdNo: "HMS06",
    patientName: "Abeer Saleem",
    patientEmail: "mohammadabeer3@gmail.com",
    initials: "AS",
    avatarColor: "#10b981",
    bedName: "mohamed",
    assignDate: "19th Apr,2026",
    dischargeDate: "17th May 2026",
    status: false,
  },
  {
    id: "ba-9",
    ipdNo: "HMS05",
    patientName: "Test User",
    patientEmail: "testyoshita@yahoo.com",
    initials: "TU",
    avatarColor: "#f59e0b",
    bedName: "sfsf",
    assignDate: "18th Apr,2026",
    dischargeDate: "N/A",
    status: true,
  },
  {
    id: "ba-10",
    ipdNo: "HMS04",
    patientName: "Vishwjeet Kumar",
    patientEmail: "vishwjeet@hms.com",
    initials: "VK",
    avatarColor: "#ef4444",
    bedName: "lit 14",
    assignDate: "17th Apr,2026",
    dischargeDate: "N/A",
    status: false,
  },
  {
    id: "ba-11",
    ipdNo: "HMS03",
    patientName: "Sara Jones",
    patientEmail: "sara.jones@example.com",
    initials: "SJ",
    avatarColor: "#8b5cf6",
    bedName: "bed11231",
    assignDate: "10th Apr,2026",
    dischargeDate: "N/A",
    status: true,
  },
  {
    id: "ba-12",
    ipdNo: "HMS02",
    patientName: "Michael Scott",
    patientEmail: "michael@dundermifflin.com",
    initials: "MS",
    avatarColor: "#06b6d4",
    bedName: "ICU-B-100",
    assignDate: "01st Apr,2026",
    dischargeDate: "05th Apr, 2026",
    status: false,
  },
];

// Initial Beds strictly matching Screenshot 8
const initialBeds: BedRow[] = [
  {
    id: "b-1",
    bedId: "QICRIAOY",
    bedName: "bed11231",
    bedType: "NICU",
    charge: 2000,
    available: true,
  },
  {
    id: "b-2",
    bedId: "WMCYURHP",
    bedName: "ICU-B-100",
    bedType: "ICU1",
    charge: 5000,
    available: false,
  },
  {
    id: "b-3",
    bedId: "HR5ZLBOQ",
    bedName: "VVIP -1",
    bedType: "ICU Bed 1",
    charge: 50000,
    available: false,
  },
  {
    id: "b-4",
    bedId: "IVZHYCTJ",
    bedName: "Student Visa USA",
    bedType: "3 star",
    charge: 111,
    available: false,
  },
  {
    id: "b-5",
    bedId: "VDJIBPLI",
    bedName: "11222",
    bedType: "ICU1",
    charge: 300,
    available: false,
  },
  {
    id: "b-6",
    bedId: "RBOL88ZB",
    bedName: "b0001",
    bedType: "ICU Bed 1",
    charge: 200,
    available: true,
  },
  {
    id: "b-7",
    bedId: "TVNECMNY",
    bedName: "sfsf",
    bedType: "3 star",
    charge: 33,
    available: false,
  },
  {
    id: "b-8",
    bedId: "CP36UEDH",
    bedName: "General Ward1",
    bedType: "General Ward",
    charge: 2500,
    available: false,
  },
  {
    id: "b-9",
    bedId: "H2",
    bedName: "ICU12",
    bedType: "5 starc",
    charge: 25000,
    available: false,
  },
  {
    id: "b-10",
    bedId: "H2",
    bedName: "Muhammad Haseeb",
    bedType: "5 star bed",
    charge: 25000,
    available: false,
  },
  {
    id: "b-11",
    bedId: "BD-0011",
    bedName: "bed4",
    bedType: "General Ward Male",
    charge: 800,
    available: true,
  },
  {
    id: "b-12",
    bedId: "BD-0012",
    bedName: "new3",
    bedType: "General Ward Male",
    charge: 950,
    available: true,
  },
];

// Initial Bed Types strictly matching Screenshots 9 & 10
const initialBedTypes: BedTypeRow[] = [
  {
    id: "bt-1",
    title: "5 sys",
    description: "Standard 5 sys observation category",
  },
  {
    id: "bt-2",
    title: "5 starc",
    description: "High comfort private room beds",
  },
  {
    id: "bt-3",
    title: "Birthing Bed",
    description: "Obstetrics and labor delivery beds",
  },
  {
    id: "bt-4",
    title: "Pediatric Bed",
    description: "Dedicated pediatric safety beds with rails",
  },
  {
    id: "bt-5",
    title: "Geriatric Bed",
    description: "Low height orthopedic geriatric beds",
  },
  {
    id: "bt-6",
    title: "Bariatric Bed",
    description: "High-capacity heavy duty bariatric care beds",
  },
  {
    id: "bt-7",
    title: "Pressure Relief Bed (Anti-Decubitus Bed)",
    description: "Air-alternating pressure mattress beds",
  },
  {
    id: "bt-8",
    title: "Stretcher Bed",
    description: "Mobile emergency and transit stretcher units",
  },
  {
    id: "bt-9",
    title: "Step-Down Bed",
    description: "Intermediate care recovery and monitor beds",
  },
  {
    id: "bt-10",
    title: "Dialysis Bed",
    description: "Reclining multi-position renal dialysis beds",
  },
  {
    id: "bt-11",
    title: "NICU",
    description: "Neonatal intensive care incubator beds",
  },
  {
    id: "bt-12",
    title: "ICU1",
    description: "Level-1 critical care monitoring beds",
  },
];

export function BedManagementWorkspace({ id }: { id: string }) {
  const { t } = useLanguage();
  const router = useRouter();

  // Normalize slug to tab
  const activeTab: BedTab =
    id === "bed-status"
      ? "bed-status"
      : id === "bed-assigns"
        ? "bed-assigns"
        : id === "beds"
          ? "beds"
          : "bed-types";

  // Data states
  const [wards, setWards] = useState<WardGroup[]>(initialWards);
  const [bedAssigns, setBedAssigns] =
    useState<BedAssignRow[]>(initialBedAssigns);
  const [beds, setBeds] = useState<BedRow[]>(initialBeds);
  const [bedTypes, setBedTypes] = useState<BedTypeRow[]>(initialBedTypes);

  // Persistent API states (Section 4 Integration)
  const [apiConnected, setApiConnected] = useState(false);
  const [isLoadingApi, setIsLoadingApi] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [occupancyReport, setOccupancyReport] =
    useState<ApiBedOccupancyReport | null>(null);
  const [remotePatients, setRemotePatients] = useState<Patient[]>([]);
  const [remoteEncounters, setRemoteEncounters] = useState<any[]>([]);
  const [apiSuccessBanner, setApiSuccessBanner] = useState("");
  const [apiErrorBanner, setApiErrorBanner] = useState("");

  const loadBedManagementData = useCallback(async () => {
    setIsLoadingApi(true);
    setApiErrorBanner("");
    try {
      // 1. Fetch live beds
      const bedsRes = await api<{ beds: ApiBed[] }>("beds?page=1");
      if (
        bedsRes?.beds &&
        Array.isArray(bedsRes.beds) &&
        bedsRes.beds.length > 0
      ) {
        const loadedBeds: BedRow[] = bedsRes.beds.map((b) => ({
          id: b.id,
          bedId: b.id.slice(0, 8).toUpperCase(),
          bedName: b.name,
          bedType: b.type,
          charge: b.chargeMinor / 100,
          available: b.available,
          description: `Version ${b.version} · State: ${b.state || "active"}`,
        }));
        setBeds(loadedBeds);

        // Update ward cards with live bed availability
        setWards((prevWards) =>
          prevWards.map((w) => ({
            ...w,
            beds: w.beds.map((wb) => {
              const matched = bedsRes.beds.find(
                (b) => b.name.toLowerCase() === wb.name.toLowerCase(),
              );
              return matched ? { ...wb, isAvailable: matched.available } : wb;
            }),
          })),
        );
      }

      // 2. Fetch live bed types
      const typesRes = await api<{ bedTypes: ApiBedType[] }>(
        "bed-types?page=1",
      );
      if (
        typesRes?.bedTypes &&
        Array.isArray(typesRes.bedTypes) &&
        typesRes.bedTypes.length > 0
      ) {
        const loadedTypes: BedTypeRow[] = typesRes.bedTypes.map((bt) => ({
          id: bt.id,
          title: bt.name,
          description: bt.description,
          bedsCount: 0,
        }));
        setBedTypes(loadedTypes);
      }

      // 3. Fetch occupancy report
      try {
        const rep = await api<ApiBedOccupancyReport>("bed-occupancy/report");
        if (rep) setOccupancyReport(rep);
      } catch {}

      // 4. Fetch patients for assignment
      try {
        const patRes = await api<{ patients: Patient[] }>("patients?page=1");
        if (patRes?.patients) setRemotePatients(patRes.patients);
      } catch {}

      // 5. Fetch encounters for assignment
      try {
        const encRes = await api<{ encounters: any[] }>(
          "encounters?kind=ipd&page=1",
        );
        if (encRes?.encounters) setRemoteEncounters(encRes.encounters);
      } catch {}

      setApiConnected(true);
    } catch {
      setApiConnected(false);
    } finally {
      setIsLoadingApi(false);
    }
  }, []);

  useEffect(() => {
    loadBedManagementData();
  }, [loadBedManagementData]);

  // UI & Search states
  const [search, setSearch] = useState("");
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);
  const [actionsDropdownOpen, setActionsDropdownOpen] = useState(false);
  const [hoveredBed, setHoveredBed] = useState<BedStatusItem | null>(null);

  // Modals state
  const [bedTypeModalOpen, setBedTypeModalOpen] = useState(false);
  const [bedModalOpen, setBedModalOpen] = useState(false);
  const [bedAssignModalOpen, setBedAssignModalOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<any>(null);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<{
    id: string;
    type: string;
  } | null>(null);

  // Form inputs for Bed Type Modal (matches Screenshot 10)
  const [formBedTypeTitle, setFormBedTypeTitle] = useState("");
  const [formBedTypeDescription, setFormBedTypeDescription] = useState("");

  // Form inputs for Bed Modal
  const [formBedName, setFormBedName] = useState("");
  const [formBedTypeId, setFormBedTypeId] = useState("");
  const [formBedCharge, setFormBedCharge] = useState("");
  const [formBedDescription, setFormBedDescription] = useState("");

  // Form inputs for Bed Assign Modal
  const [formAssignCaseId, setFormAssignCaseId] = useState("");
  const [formAssignPatient, setFormAssignPatient] = useState("");
  const [formAssignBed, setFormAssignBed] = useState("");
  const [formAssignDate, setFormAssignDate] = useState("");
  const [formAssignDischargeDate, setFormAssignDischargeDate] = useState("");
  const [formAssignDescription, setFormAssignDescription] = useState("");
  const [formAssignStatus, setFormAssignStatus] = useState(true);

  // Tab switching
  function navigateToTab(tab: BedTab) {
    setSearch("");
    setPage(1);
    router.push(`/modules/${tab}`);
  }

  // Bed Assign Status Toggle
  function toggleBedAssignStatus(assignId: string) {
    setBedAssigns((prev) =>
      prev.map((row) =>
        row.id === assignId ? { ...row, status: !row.status } : row,
      ),
    );
  }

  // Open Bed Type Modal
  function handleOpenNewBedType(existing?: BedTypeRow) {
    if (existing) {
      setEditingItem(existing);
      setFormBedTypeTitle(existing.title);
      setFormBedTypeDescription(existing.description || "");
    } else {
      setEditingItem(null);
      setFormBedTypeTitle("");
      setFormBedTypeDescription("");
    }
    setBedTypeModalOpen(true);
  }

  // Save Bed Type
  async function handleSaveBedType(e: React.FormEvent) {
    e.preventDefault();
    if (!formBedTypeTitle.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");

    let btId = editingItem ? editingItem.id : `bt-${Date.now()}`;
    try {
      if (editingItem) {
        await api<ApiBedType>(`bed-types/${editingItem.id}`, {
          method: "PATCH",
          body: JSON.stringify({
            name: formBedTypeTitle.trim(),
            description: formBedTypeDescription.trim(),
            active: true,
            version: editingItem.version || 1,
          }),
        });
        setApiSuccessBanner(t("Bed Type updated and persisted successfully."));
      } else {
        const created = await api<ApiBedType>("bed-types", {
          method: "POST",
          body: JSON.stringify({
            name: formBedTypeTitle.trim(),
            description: formBedTypeDescription.trim(),
            active: true,
            version: 1,
          }),
        });
        if (created?.id) btId = created.id;
        setApiSuccessBanner(
          t("Bed Type created and registered in Go database successfully."),
        );
      }
    } catch (err: any) {
      setApiErrorBanner(
        err?.message || t("Could not sync with backend. Saved locally."),
      );
    } finally {
      setIsSubmitting(false);
    }

    if (editingItem) {
      setBedTypes((prev) =>
        prev.map((bt) =>
          bt.id === editingItem.id
            ? {
                ...bt,
                title: formBedTypeTitle.trim(),
                description: formBedTypeDescription.trim(),
              }
            : bt,
        ),
      );
    } else {
      const newBt: BedTypeRow = {
        id: btId,
        title: formBedTypeTitle.trim(),
        description: formBedTypeDescription.trim(),
      };
      setBedTypes((prev) => [newBt, ...prev]);
    }
    setBedTypeModalOpen(false);
  }

  // Open Bed Modal
  function handleOpenNewBed(existing?: BedRow) {
    if (existing) {
      setEditingItem(existing);
      setFormBedName(existing.bedName);
      setFormBedTypeId(existing.bedType);
      setFormBedCharge(String(existing.charge));
      setFormBedDescription(existing.description || "");
    } else {
      setEditingItem(null);
      setFormBedName("");
      setFormBedTypeId(bedTypes[0]?.title || "General Ward");
      setFormBedCharge("500");
      setFormBedDescription("");
    }
    setBedModalOpen(true);
    setActionsDropdownOpen(false);
  }

  // Save Bed
  async function handleSaveBed(e: React.FormEvent) {
    e.preventDefault();
    if (!formBedName.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");

    const chargeNum = parseFloat(formBedCharge) || 0;
    const chargeMinor = Math.round(chargeNum * 100);
    let createdId = `b-${Date.now()}`;

    try {
      const created = await api<ApiBed>("beds", {
        method: "POST",
        body: JSON.stringify({
          name: formBedName.trim(),
          type: formBedTypeId,
          chargeMinor,
        }),
      });
      if (created?.id) {
        createdId = created.id;
        setApiSuccessBanner(
          t("Bed created and registered in Go database successfully."),
        );
      }
    } catch (err: any) {
      setApiErrorBanner(
        err?.message || t("Could not save to backend. Preserved locally."),
      );
    } finally {
      setIsSubmitting(false);
    }

    if (editingItem) {
      setBeds((prev) =>
        prev.map((b) =>
          b.id === editingItem.id
            ? {
                ...b,
                bedName: formBedName.trim(),
                bedType: formBedTypeId,
                charge: chargeNum,
                description: formBedDescription.trim(),
              }
            : b,
        ),
      );
    } else {
      const newBed: BedRow = {
        id: createdId,
        bedId: `BED${Math.floor(1000 + Math.random() * 9000)}`,
        bedName: formBedName.trim(),
        bedType: formBedTypeId,
        charge: chargeNum,
        available: true,
        description: formBedDescription.trim(),
      };
      setBeds((prev) => [newBed, ...prev]);
    }
    setBedModalOpen(false);
  }

  // Open Bed Assign Modal
  function handleOpenNewBedAssign(
    existing?: BedAssignRow,
    prefilledBed?: string,
  ) {
    if (existing) {
      setEditingItem(existing);
      setFormAssignCaseId(existing.ipdNo);
      setFormAssignPatient(existing.patientName);
      setFormAssignBed(existing.bedName);
      setFormAssignDate(existing.assignDate);
      setFormAssignDischargeDate(
        existing.dischargeDate === "N/A" ? "" : existing.dischargeDate,
      );
      setFormAssignDescription(existing.description || "");
      setFormAssignStatus(existing.status);
    } else {
      setEditingItem(null);
      setFormAssignCaseId("HMS" + Math.floor(15 + Math.random() * 85));
      setFormAssignPatient("");
      setFormAssignBed(prefilledBed || beds[0]?.bedName || "bed11231");
      setFormAssignDate(
        new Date().toLocaleDateString("en-GB", {
          day: "numeric",
          month: "short",
          year: "numeric",
        }),
      );
      setFormAssignDischargeDate("");
      setFormAssignDescription("");
      setFormAssignStatus(true);
    }
    setBedAssignModalOpen(true);
  }

  // Save Bed Assign
  async function handleSaveBedAssign(e: React.FormEvent) {
    e.preventDefault();
    if (!formAssignPatient.trim()) return;
    setIsSubmitting(true);
    setApiErrorBanner("");

    const matchedBed = beds.find((b) => b.bedName === formAssignBed);
    const matchedPatient = remotePatients.find(
      (p) =>
        `${p.givenName} ${p.familyName}`.toLowerCase() ===
        formAssignPatient.toLowerCase(),
    );
    const matchedEncounter = remoteEncounters.find(
      (enc) =>
        enc.patientName?.toLowerCase() === formAssignPatient.toLowerCase() ||
        enc.bedName === formAssignBed,
    );

    if (matchedBed && matchedPatient && matchedEncounter) {
      try {
        await api("bed-assignments", {
          method: "POST",
          body: JSON.stringify({
            version: 1,
            bedId: matchedBed.id,
            encounterId: matchedEncounter.id,
            patientId: matchedPatient.id,
            notes: formAssignDescription.trim(),
          }),
        });
        setApiSuccessBanner(
          t("Bed assignment stored in persistent clinical records."),
        );
      } catch (err: any) {
        setApiErrorBanner(
          err?.message ||
            t("Saved locally. Could not link to remote encounter."),
        );
      } finally {
        setIsSubmitting(false);
      }
    } else {
      setIsSubmitting(false);
    }

    if (editingItem) {
      setBedAssigns((prev) =>
        prev.map((ba) =>
          ba.id === editingItem.id
            ? {
                ...ba,
                patientName: formAssignPatient.trim(),
                bedName: formAssignBed,
                assignDate: formAssignDate,
                dischargeDate: formAssignDischargeDate.trim() || "N/A",
                status: formAssignStatus,
                description: formAssignDescription.trim(),
              }
            : ba,
        ),
      );
    } else {
      const names = formAssignPatient.trim().split(" ");
      const initials =
        names.length > 1
          ? `${names[0][0]}${names[1][0]}`.toUpperCase()
          : names[0].slice(0, 2).toUpperCase();
      const newBa: BedAssignRow = {
        id: `ba-${Date.now()}`,
        ipdNo: formAssignCaseId || "HMS20",
        patientName: formAssignPatient.trim(),
        patientEmail: `${formAssignPatient.trim().toLowerCase().replace(/\s+/g, ".")}@example.com`,
        initials,
        avatarColor: "#3b82f6",
        bedName: formAssignBed,
        assignDate: formAssignDate || "5th Oct, 2026",
        dischargeDate: formAssignDischargeDate.trim() || "N/A",
        status: formAssignStatus,
        description: formAssignDescription.trim(),
      };
      setBedAssigns((prev) => [newBa, ...prev]);
    }
    setBedAssignModalOpen(false);
  }

  // Confirm Delete
  function handleDeleteConfirm() {
    if (!itemToDelete) return;
    if (itemToDelete.type === "bed-types") {
      setBedTypes((prev) => prev.filter((bt) => bt.id !== itemToDelete.id));
    } else if (itemToDelete.type === "beds") {
      setBeds((prev) => prev.filter((b) => b.id !== itemToDelete.id));
    } else if (itemToDelete.type === "bed-assigns") {
      setBedAssigns((prev) => prev.filter((ba) => ba.id !== itemToDelete.id));
    }
    setDeleteConfirmOpen(false);
    setItemToDelete(null);
  }

  // Export to Excel / CSV
  function handleExportCsv(type: "beds" | "bed-assigns" | "bed-types") {
    let headers: string[] = [];
    let rows: string[][] = [];

    if (type === "beds") {
      headers = ["BED ID", "BED", "BED TYPE", "CHARGE", "AVAILABLE"];
      rows = beds.map((b) => [
        b.bedId,
        b.bedName,
        b.bedType,
        `$${b.charge.toFixed(2)}`,
        b.available ? "Yes" : "No",
      ]);
    } else if (type === "bed-assigns") {
      headers = [
        "IPD NO",
        "PATIENT",
        "BED",
        "ASSIGN DATE",
        "DISCHARGE DATE",
        "STATUS",
      ];
      rows = bedAssigns.map((ba) => [
        ba.ipdNo,
        ba.patientName,
        ba.bedName,
        ba.assignDate,
        ba.dischargeDate,
        ba.status ? "Active" : "Inactive",
      ]);
    } else if (type === "bed-types") {
      headers = ["BED TYPE", "DESCRIPTION"];
      rows = bedTypes.map((bt) => [bt.title, bt.description || ""]);
    }

    const csvContent =
      "data:text/csv;charset=utf-8," +
      [
        headers.join(","),
        ...rows.map((r) =>
          r.map((cell) => `"${cell.replace(/"/g, '""')}"`).join(","),
        ),
      ].join("\n");
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement("a");
    link.setAttribute("href", encodedUri);
    link.setAttribute("download", `${type}-export.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    setActionsDropdownOpen(false);
  }

  // Filtered rows for current tab
  const filteredBedAssigns = useMemo(() => {
    const q = search.toLowerCase().trim();
    if (!q) return bedAssigns;
    return bedAssigns.filter(
      (ba) =>
        ba.ipdNo.toLowerCase().includes(q) ||
        ba.patientName.toLowerCase().includes(q) ||
        ba.patientEmail.toLowerCase().includes(q) ||
        ba.bedName.toLowerCase().includes(q),
    );
  }, [bedAssigns, search]);

  const filteredBeds = useMemo(() => {
    const q = search.toLowerCase().trim();
    if (!q) return beds;
    return beds.filter(
      (b) =>
        b.bedId.toLowerCase().includes(q) ||
        b.bedName.toLowerCase().includes(q) ||
        b.bedType.toLowerCase().includes(q),
    );
  }, [beds, search]);

  const filteredBedTypes = useMemo(() => {
    const q = search.toLowerCase().trim();
    if (!q) return bedTypes;
    return bedTypes.filter(
      (bt) =>
        bt.title.toLowerCase().includes(q) ||
        (bt.description && bt.description.toLowerCase().includes(q)),
    );
  }, [bedTypes, search]);

  const paginatedBedAssigns = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredBedAssigns.slice(start, start + pageSize);
  }, [filteredBedAssigns, page, pageSize]);

  const paginatedBeds = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredBeds.slice(start, start + pageSize);
  }, [filteredBeds, page, pageSize]);

  const paginatedBedTypes = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredBedTypes.slice(start, start + pageSize);
  }, [filteredBedTypes, page, pageSize]);

  return (
    <div className="bed-workspace-container" data-ready="true">
      {/* Persistent Connection Status & Live Occupancy Report (Section 4 Integration) */}
      <div
        className="mb-3 d-flex flex-column gap-2"
        style={{ padding: "0 4px" }}
      >
        <div
          className="d-flex justify-content-between align-items-center flex-wrap gap-2 px-3 py-2 rounded border"
          style={{
            backgroundColor: apiConnected
              ? "rgba(16, 185, 129, 0.08)"
              : "rgba(245, 158, 11, 0.08)",
            borderColor: apiConnected ? "#10b981" : "#f59e0b",
          }}
        >
          <div className="d-flex align-items-center flex-wrap gap-2">
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                backgroundColor: apiConnected ? "#10b981" : "#f59e0b",
                display: "inline-block",
              }}
            />
            <span className="fs-7 fw-semibold">
              {apiConnected
                ? t(
                    "Connected to Go/PostgreSQL Bed Service (/v1/beds, /v1/bed-types, /v1/bed-occupancy)",
                  )
                : t("Local preview mode · Syncing locally")}
            </span>
            {occupancyReport && (
              <span
                className="badge-available-stock fs-8 py-0 px-2"
                style={{
                  backgroundColor: "#3b82f622",
                  color: "#3b82f6",
                  borderColor: "#3b82f6",
                }}
              >
                {t("Total")}: {occupancyReport.totalBeds} | {t("Occupied")}:{" "}
                {occupancyReport.occupiedBeds} | {t("Available")}:{" "}
                {occupancyReport.availableBeds} (
                {occupancyReport.occupancyRate.toFixed(1)}%)
              </span>
            )}
          </div>
          <button
            type="button"
            className="btn-icon-link fs-7 d-flex align-items-center gap-1"
            onClick={loadBedManagementData}
            disabled={isLoadingApi}
            title={t("Refresh bed records from Go API")}
          >
            <RefreshCw
              size={13}
              className={isLoadingApi ? "animate-spin" : ""}
            />
            <span>{isLoadingApi ? t("Syncing...") : t("Sync Backend")}</span>
          </button>
        </div>

        {apiSuccessBanner && (
          <div
            className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded"
            style={{
              borderColor: "#10b981",
              backgroundColor: "rgba(16, 185, 129, 0.1)",
              color: "#10b981",
            }}
          >
            <div className="d-flex align-items-center gap-2">
              <CheckCircle2 size={16} />
              <span className="fs-7">{apiSuccessBanner}</span>
            </div>
            <button
              type="button"
              onClick={() => setApiSuccessBanner("")}
              className="btn-icon-link"
            >
              <X size={14} />
            </button>
          </div>
        )}

        {apiErrorBanner && (
          <div
            className="alert-notice d-flex align-items-center justify-content-between py-2 px-3 border rounded"
            style={{
              borderColor: "#ef4444",
              backgroundColor: "rgba(239, 68, 68, 0.1)",
              color: "#ef4444",
            }}
          >
            <div className="d-flex align-items-center gap-2">
              <AlertCircle size={16} />
              <span className="fs-7">{apiErrorBanner}</span>
            </div>
            <button
              type="button"
              onClick={() => setApiErrorBanner("")}
              className="btn-icon-link"
            >
              <X size={14} />
            </button>
          </div>
        )}
      </div>

      {/* Tab: Bed Status (Screenshots 1-6) */}
      {activeTab === "bed-status" && (
        <div className="bed-status-view">
          <div className="bed-status-top-bar">
            <h1 className="bed-status-title">{t("Bed Status")}</h1>
            <button
              type="button"
              className="btn-back-outline"
              onClick={() => navigateToTab("bed-assigns")}
            >
              {t("Back")}
            </button>
          </div>

          <div className="bed-status-grid">
            {wards.map((ward) => (
              <div key={ward.id} className="ward-section-card">
                <h2 className="ward-title">{ward.name}</h2>
                <div className="ward-content-box">
                  {ward.beds.length === 0 ? (
                    <div className="no-bed-message">
                      <span>{t("No Bed Available")}</span>
                    </div>
                  ) : (
                    <div className="ward-beds-container">
                      {ward.beds.map((bed) => {
                        const isOccupied = !bed.isAvailable;
                        return (
                          <div
                            key={bed.id}
                            className={`bed-status-item ${isOccupied ? "occupied" : "available"}`}
                            onMouseEnter={() =>
                              isOccupied && setHoveredBed(bed)
                            }
                            onMouseLeave={() => setHoveredBed(null)}
                            onClick={() => {
                              if (!isOccupied) {
                                handleOpenNewBedAssign(undefined, bed.name);
                              }
                            }}
                          >
                            <div className="bed-icon-wrapper">
                              {isOccupied ? (
                                /* Red Pulse Bed Icon matching Screenshots */
                                <svg
                                  className="bed-pulse-svg text-danger"
                                  viewBox="0 0 64 64"
                                  width="48"
                                  height="48"
                                  fill="none"
                                  stroke="#f87171"
                                  strokeWidth="2.5"
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                >
                                  {/* Bed frame & patient */}
                                  <path d="M4 48V26" />
                                  <path d="M60 48V36" />
                                  <path d="M4 42h56" />
                                  <circle
                                    cx="16"
                                    cy="24"
                                    r="5"
                                    fill="#f87171"
                                  />
                                  <path
                                    d="M22 28h16a8 8 0 0 1 8 8v6H14v-6a8 8 0 0 1 8-8z"
                                    fill="#f87171"
                                    opacity="0.3"
                                  />
                                  {/* Pulse / Heartbeat waveform */}
                                  <path
                                    d="M26 18h6l3-6 4 12 3-8 2 4h6"
                                    stroke="#ef4444"
                                    strokeWidth="3"
                                  />
                                </svg>
                              ) : (
                                /* Green Available Bed Icon matching Screenshots */
                                <svg
                                  className="bed-available-svg text-success"
                                  viewBox="0 0 64 64"
                                  width="48"
                                  height="48"
                                  fill="none"
                                  stroke="#34d399"
                                  strokeWidth="2.5"
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                >
                                  <path d="M6 48V28" />
                                  <path d="M58 48V36" />
                                  <path d="M6 42h52" />
                                  <circle
                                    cx="18"
                                    cy="26"
                                    r="6"
                                    fill="#10b981"
                                  />
                                  <path
                                    d="M26 32h26a6 6 0 0 1 6 6v4H16v-4a6 6 0 0 1 10-6z"
                                    fill="#10b981"
                                    opacity="0.3"
                                  />
                                </svg>
                              )}
                            </div>

                            <span
                              className={`bed-status-label ${isOccupied ? "text-danger" : "text-success"}`}
                            >
                              {isOccupied ? bed.patientName : bed.name}
                            </span>

                            {/* Hover popover tooltip matching Screenshot 1 */}
                            {hoveredBed?.id === bed.id && isOccupied && (
                              <div className="bed-hover-popover">
                                <div className="popover-row">
                                  <span className="popover-label">
                                    {t("Bed Name")} :
                                  </span>
                                  <span className="popover-val">
                                    {bed.name}
                                  </span>
                                </div>
                                <div className="popover-row">
                                  <span className="popover-label">
                                    {t("Patient")} :
                                  </span>
                                  <span className="popover-val">
                                    {bed.patientName}
                                  </span>
                                </div>
                                <div className="popover-row">
                                  <span className="popover-label">
                                    {t("Phone")} :
                                  </span>
                                  <span className="popover-val">
                                    {bed.phone || "N/A"}
                                  </span>
                                </div>
                                <div className="popover-row">
                                  <span className="popover-label">
                                    {t("Admission Date")} :
                                  </span>
                                  <span className="popover-val">
                                    {bed.admissionDate || "N/A"}
                                  </span>
                                </div>
                                <div className="popover-row">
                                  <span className="popover-label">
                                    {t("Gender")} :
                                  </span>
                                  <span className="popover-val">
                                    {bed.gender || "Male"}
                                  </span>
                                </div>
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tab: Bed Assigns (Screenshot 7) */}
      {activeTab === "bed-assigns" && (
        <div className="bed-assigns-view">
          <div className="billing-toolbar">
            <div className="billing-search-box">
              <Search size={18} className="search-icon" />
              <input
                type="text"
                placeholder={t("Search")}
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
                aria-label={t("Search")}
              />
            </div>

            <div className="billing-actions">
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Filter")}
                aria-label={t("Filter")}
                onClick={() => setActionsDropdownOpen(!actionsDropdownOpen)}
              >
                <Filter size={18} />
              </button>
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => handleOpenNewBedAssign()}
              >
                {t("New Bed Assign")}
              </button>
            </div>
          </div>

          <div className="billing-card">
            <div className="table-responsive">
              <table className="billing-table">
                <thead>
                  <tr>
                    <th>
                      <div className="th-sort">
                        <span>{t("IPD NO")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("PATIENT")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("ASSIGN DATE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("DISCHARGE DATE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>{t("STATUS")}</th>
                    <th style={{ textAlign: "right" }}>{t("ACTION")}</th>
                  </tr>
                </thead>
                <tbody>
                  {paginatedBedAssigns.map((row) => (
                    <tr key={row.id}>
                      {/* IPD NO pill badge */}
                      <td>
                        <Link
                          href={`/modules/ipd-patient-departments?id=${row.ipdNo}`}
                          className="badge-blue-link"
                        >
                          {row.ipdNo}
                        </Link>
                      </td>
                      {/* Patient avatar + name link + email */}
                      <td>
                        <div className="patient-cell">
                          <div
                            className="patient-avatar"
                            style={{ backgroundColor: row.avatarColor }}
                          >
                            {row.initials}
                          </div>
                          <div className="patient-info">
                            <span className="patient-name">
                              {row.patientName}
                            </span>
                            <span className="patient-email">
                              {row.patientEmail}
                            </span>
                          </div>
                        </div>
                      </td>
                      {/* Bed name link */}
                      <td>
                        <span className="bed-name-link">{row.bedName}</span>
                      </td>
                      {/* Assign Date cyan link */}
                      <td>
                        <span className="date-link">{row.assignDate}</span>
                      </td>
                      {/* Discharge Date */}
                      <td>
                        {row.dischargeDate === "N/A" ? (
                          <span className="text-muted">N/A</span>
                        ) : (
                          <span className="date-link">{row.dischargeDate}</span>
                        )}
                      </td>
                      {/* Status Toggle Switch */}
                      <td>
                        <label
                          className="switch-toggle"
                          aria-label={`Toggle status for ${row.patientName}`}
                        >
                          <input
                            type="checkbox"
                            checked={row.status}
                            onChange={() => toggleBedAssignStatus(row.id)}
                          />
                          <span className="slider round"></span>
                        </label>
                      </td>
                      {/* Actions */}
                      <td style={{ textAlign: "right" }}>
                        <div className="action-buttons-group">
                          <button
                            type="button"
                            className="btn-edit-blue"
                            title={t("Edit")}
                            aria-label={`Edit ${row.patientName}`}
                            onClick={() => handleOpenNewBedAssign(row)}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            type="button"
                            className="btn-delete-red"
                            title={t("Delete")}
                            aria-label={`Delete ${row.patientName}`}
                            onClick={() => {
                              setItemToDelete({
                                id: row.id,
                                type: "bed-assigns",
                              });
                              setDeleteConfirmOpen(true);
                            }}
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {paginatedBedAssigns.length === 0 && (
                    <tr>
                      <td colSpan={7} className="no-data-cell">
                        {t("No records found")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination footer matching screenshot */}
            <div className="billing-footer">
              <div className="footer-left">
                <span className="footer-text">{t("Show")}</span>
                <select
                  className="page-size-select"
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setPage(1);
                  }}
                  aria-label={t("Page size")}
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
                <span className="footer-text">
                  {t("Showing")}{" "}
                  {filteredBedAssigns.length ? (page - 1) * pageSize + 1 : 0}{" "}
                  {t("to")}{" "}
                  {Math.min(page * pageSize, filteredBedAssigns.length)}{" "}
                  {t("of")} {filteredBedAssigns.length} {t("Results")}
                </span>
              </div>

              <div className="footer-pagination">
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                  aria-label={t("Previous")}
                >
                  ‹
                </button>
                <button
                  type="button"
                  className={`pagination-btn ${page === 1 ? "active" : ""}`}
                  onClick={() => setPage(1)}
                >
                  1
                </button>
                {Math.ceil(filteredBedAssigns.length / pageSize) > 1 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 2 ? "active" : ""}`}
                    onClick={() => setPage(2)}
                  >
                    2
                  </button>
                )}
                {Math.ceil(filteredBedAssigns.length / pageSize) > 2 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 3 ? "active" : ""}`}
                    onClick={() => setPage(3)}
                  >
                    3
                  </button>
                )}
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={
                    page >= Math.ceil(filteredBedAssigns.length / pageSize)
                  }
                  onClick={() => setPage(page + 1)}
                  aria-label={t("Next")}
                >
                  ›
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab: Beds (Screenshot 8) */}
      {activeTab === "beds" && (
        <div className="beds-view">
          <div className="billing-toolbar">
            <div className="billing-search-box">
              <Search size={18} className="search-icon" />
              <input
                type="text"
                placeholder={t("Search")}
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
                aria-label={t("Search")}
              />
            </div>

            <div className="billing-actions">
              <button
                type="button"
                className="btn-icon-blue"
                title={t("Filter")}
                aria-label={t("Filter")}
              >
                <Filter size={18} />
              </button>

              <div className="relative-actions-dropdown">
                <button
                  type="button"
                  className="btn-action-blue dropdown-trigger"
                  onClick={() => setActionsDropdownOpen(!actionsDropdownOpen)}
                >
                  {t("Actions")}
                  <ChevronDown size={16} className="ml-1" />
                </button>

                {actionsDropdownOpen && (
                  <div className="dropdown-action-menu">
                    <button
                      type="button"
                      className="dropdown-menu-item"
                      onClick={() => handleOpenNewBed()}
                    >
                      <Plus size={16} className="me-2" />
                      {t("New Bed")}
                    </button>
                    <button
                      type="button"
                      className="dropdown-menu-item"
                      onClick={() => handleExportCsv("beds")}
                    >
                      <Download size={16} className="me-2" />
                      {t("Export to Excel")}
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className="billing-card">
            <div className="table-responsive">
              <table className="billing-table">
                <thead>
                  <tr>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED ID")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED TYPE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>
                      <div className="th-sort">
                        <span>{t("CHARGE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th>{t("AVAILABLE")}</th>
                    <th style={{ textAlign: "right" }}>{t("ACTION")}</th>
                  </tr>
                </thead>
                <tbody>
                  {paginatedBeds.map((row) => (
                    <tr key={row.id}>
                      {/* Bed ID pill badge */}
                      <td>
                        <span className="badge-blue-link">{row.bedId}</span>
                      </td>
                      {/* Bed Name */}
                      <td>
                        <span className="bed-name-text">{row.bedName}</span>
                      </td>
                      {/* Bed Type link */}
                      <td>
                        <Link
                          href={`/modules/bed-types?title=${encodeURIComponent(row.bedType)}`}
                          className="bed-type-link"
                        >
                          {row.bedType}
                        </Link>
                      </td>
                      {/* Charge */}
                      <td>
                        <span className="currency-amount">
                          $
                          {row.charge.toLocaleString("en-US", {
                            minimumFractionDigits: 2,
                            maximumFractionDigits: 2,
                          })}
                        </span>
                      </td>
                      {/* Available Yes / No status badge */}
                      <td>
                        {row.available ? (
                          <span className="badge-available-yes">
                            {t("Yes")}
                          </span>
                        ) : (
                          <span className="badge-available-no">{t("No")}</span>
                        )}
                      </td>
                      {/* Actions */}
                      <td style={{ textAlign: "right" }}>
                        <div className="action-buttons-group">
                          <button
                            type="button"
                            className="btn-edit-blue"
                            title={t("Edit")}
                            aria-label={`Edit ${row.bedName}`}
                            onClick={() => handleOpenNewBed(row)}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            type="button"
                            className="btn-delete-red"
                            title={t("Delete")}
                            aria-label={`Delete ${row.bedName}`}
                            onClick={() => {
                              setItemToDelete({ id: row.id, type: "beds" });
                              setDeleteConfirmOpen(true);
                            }}
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {paginatedBeds.length === 0 && (
                    <tr>
                      <td colSpan={6} className="no-data-cell">
                        {t("No records found")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination footer */}
            <div className="billing-footer">
              <div className="footer-left">
                <span className="footer-text">{t("Show")}</span>
                <select
                  className="page-size-select"
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setPage(1);
                  }}
                  aria-label={t("Page size")}
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
                <span className="footer-text">
                  {t("Showing")}{" "}
                  {filteredBeds.length ? (page - 1) * pageSize + 1 : 0}{" "}
                  {t("to")} {Math.min(page * pageSize, filteredBeds.length)}{" "}
                  {t("of")} {filteredBeds.length} {t("Results")}
                </span>
              </div>

              <div className="footer-pagination">
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                  aria-label={t("Previous")}
                >
                  ‹
                </button>
                <button
                  type="button"
                  className={`pagination-btn ${page === 1 ? "active" : ""}`}
                  onClick={() => setPage(1)}
                >
                  1
                </button>
                {Math.ceil(filteredBeds.length / pageSize) > 1 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 2 ? "active" : ""}`}
                    onClick={() => setPage(2)}
                  >
                    2
                  </button>
                )}
                {Math.ceil(filteredBeds.length / pageSize) > 2 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 3 ? "active" : ""}`}
                    onClick={() => setPage(3)}
                  >
                    3
                  </button>
                )}
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page >= Math.ceil(filteredBeds.length / pageSize)}
                  onClick={() => setPage(page + 1)}
                  aria-label={t("Next")}
                >
                  ›
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Tab: Bed Types (Screenshots 9 & 10) */}
      {activeTab === "bed-types" && (
        <div className="bed-types-view">
          <div className="billing-toolbar">
            <div className="billing-search-box">
              <Search size={18} className="search-icon" />
              <input
                type="text"
                placeholder={t("Search")}
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
                aria-label={t("Search")}
              />
            </div>

            <div className="billing-actions">
              <button
                type="button"
                className="btn-action-blue"
                onClick={() => handleOpenNewBedType()}
              >
                {t("New Bed Type")}
              </button>
            </div>
          </div>

          <div className="billing-card">
            <div className="table-responsive">
              <table className="billing-table">
                <thead>
                  <tr>
                    <th>
                      <div className="th-sort">
                        <span>{t("BED TYPE")}</span>
                        <span className="sort-icon">↕</span>
                      </div>
                    </th>
                    <th style={{ textAlign: "right" }}>{t("ACTION")}</th>
                  </tr>
                </thead>
                <tbody>
                  {paginatedBedTypes.map((row) => (
                    <tr key={row.id}>
                      {/* Bed Type Name in blue link style */}
                      <td>
                        <Link
                          href={`/modules/beds?type=${encodeURIComponent(row.title)}`}
                          className="bed-type-title-link"
                        >
                          {row.title}
                        </Link>
                      </td>
                      {/* Actions */}
                      <td style={{ textAlign: "right" }}>
                        <div className="action-buttons-group">
                          <button
                            type="button"
                            className="btn-edit-blue"
                            title={t("Edit")}
                            aria-label={`Edit ${row.title}`}
                            onClick={() => handleOpenNewBedType(row)}
                          >
                            <Edit2 size={16} />
                          </button>
                          <button
                            type="button"
                            className="btn-delete-red"
                            title={t("Delete")}
                            aria-label={`Delete ${row.title}`}
                            onClick={() => {
                              setItemToDelete({
                                id: row.id,
                                type: "bed-types",
                              });
                              setDeleteConfirmOpen(true);
                            }}
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {paginatedBedTypes.length === 0 && (
                    <tr>
                      <td colSpan={2} className="no-data-cell">
                        {t("No records found")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Pagination footer */}
            <div className="billing-footer">
              <div className="footer-left">
                <span className="footer-text">{t("Show")}</span>
                <select
                  className="page-size-select"
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setPage(1);
                  }}
                  aria-label={t("Page size")}
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
                <span className="footer-text">
                  {t("Showing")}{" "}
                  {filteredBedTypes.length ? (page - 1) * pageSize + 1 : 0}{" "}
                  {t("to")} {Math.min(page * pageSize, filteredBedTypes.length)}{" "}
                  {t("of")} {filteredBedTypes.length} {t("Results")}
                </span>
              </div>

              <div className="footer-pagination">
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                  aria-label={t("Previous")}
                >
                  ‹
                </button>
                <button
                  type="button"
                  className={`pagination-btn ${page === 1 ? "active" : ""}`}
                  onClick={() => setPage(1)}
                >
                  1
                </button>
                {Math.ceil(filteredBedTypes.length / pageSize) > 1 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 2 ? "active" : ""}`}
                    onClick={() => setPage(2)}
                  >
                    2
                  </button>
                )}
                {Math.ceil(filteredBedTypes.length / pageSize) > 2 && (
                  <button
                    type="button"
                    className={`pagination-btn ${page === 3 ? "active" : ""}`}
                    onClick={() => setPage(3)}
                  >
                    3
                  </button>
                )}
                <button
                  type="button"
                  className="pagination-btn"
                  disabled={
                    page >= Math.ceil(filteredBedTypes.length / pageSize)
                  }
                  onClick={() => setPage(page + 1)}
                  aria-label={t("Next")}
                >
                  ›
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Modal: New / Edit Bed Type (Exact match to Screenshot 10) */}
      {bedTypeModalOpen && (
        <div
          className="modal-backdrop-custom"
          onClick={() => setBedTypeModalOpen(false)}
        >
          <div
            className="modal-card-custom"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom" id="bed-type-modal-title">
                {editingItem ? t("Edit Bed Type") : t("New Bed Type")}
              </h3>
              <button
                type="button"
                className="modal-close-btn"
                aria-label={t("Close")}
                onClick={() => setBedTypeModalOpen(false)}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSaveBedType} className="modal-form-custom">
              <div className="form-group-custom">
                <label className="form-label-custom">
                  {t("Bed Type")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Bed Type")}
                  className="form-input-custom"
                  value={formBedTypeTitle}
                  onChange={(e) => setFormBedTypeTitle(e.target.value)}
                  autoFocus
                />
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom">{t("Description")}:</label>
                <textarea
                  rows={3}
                  placeholder={t("Description")}
                  className="form-textarea-custom"
                  value={formBedTypeDescription}
                  onChange={(e) => setFormBedTypeDescription(e.target.value)}
                />
              </div>

              <div className="modal-footer-custom">
                <button
                  type="submit"
                  className="btn-modal-save"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <span className="d-inline-flex align-items-center gap-1">
                      <Loader2 size={14} className="animate-spin" />
                      {t("Saving...")}
                    </span>
                  ) : (
                    t("Save")
                  )}
                </button>
                <button
                  type="button"
                  className="btn-modal-cancel"
                  disabled={isSubmitting}
                  onClick={() => setBedTypeModalOpen(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: New / Edit Bed */}
      {bedModalOpen && (
        <div
          className="modal-backdrop-custom"
          onClick={() => setBedModalOpen(false)}
        >
          <div
            className="modal-card-custom"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom" id="bed-modal-title">
                {editingItem ? t("Edit Bed") : t("New Bed")}
              </h3>
              <button
                type="button"
                className="modal-close-btn"
                aria-label={t("Close")}
                onClick={() => setBedModalOpen(false)}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSaveBed} className="modal-form-custom">
              <div className="form-group-custom">
                <label className="form-label-custom">
                  {t("Bed")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t("Bed")}
                  className="form-input-custom"
                  value={formBedName}
                  onChange={(e) => setFormBedName(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom">
                  {t("Bed Type")}: <span className="text-danger">*</span>
                </label>
                <select
                  required
                  className="form-select-custom"
                  value={formBedTypeId}
                  onChange={(e) => setFormBedTypeId(e.target.value)}
                >
                  {bedTypes.map((bt) => (
                    <option key={bt.id} value={bt.title}>
                      {bt.title}
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom">
                  {t("Charge")}: <span className="text-danger">*</span>
                </label>
                <input
                  type="number"
                  step="0.01"
                  required
                  placeholder={t("Charge")}
                  className="form-input-custom"
                  value={formBedCharge}
                  onChange={(e) => setFormBedCharge(e.target.value)}
                />
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom">{t("Description")}:</label>
                <textarea
                  rows={3}
                  placeholder={t("Description")}
                  className="form-textarea-custom"
                  value={formBedDescription}
                  onChange={(e) => setFormBedDescription(e.target.value)}
                />
              </div>

              <div className="modal-footer-custom">
                <button
                  type="submit"
                  className="btn-modal-save"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <span className="d-inline-flex align-items-center gap-1">
                      <Loader2 size={14} className="animate-spin" />
                      {t("Saving...")}
                    </span>
                  ) : (
                    t("Save")
                  )}
                </button>
                <button
                  type="button"
                  className="btn-modal-cancel"
                  disabled={isSubmitting}
                  onClick={() => setBedModalOpen(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal: New / Edit Bed Assign */}
      {bedAssignModalOpen && (
        <div
          className="modal-backdrop-custom"
          onClick={() => setBedAssignModalOpen(false)}
        >
          <div
            className="modal-card-custom modal-wide"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom" id="bed-assign-modal-title">
                {editingItem ? t("Edit Bed Assign") : t("New Bed Assign")}
              </h3>
              <button
                type="button"
                className="modal-close-btn"
                aria-label={t("Close")}
                onClick={() => setBedAssignModalOpen(false)}
              >
                <X size={18} />
              </button>
            </div>

            <form onSubmit={handleSaveBedAssign} className="modal-form-custom">
              <div className="form-grid-2">
                <div className="form-group-custom">
                  <label className="form-label-custom">
                    {t("IPD No / Case")}:
                  </label>
                  <input
                    type="text"
                    className="form-input-custom"
                    value={formAssignCaseId}
                    onChange={(e) => setFormAssignCaseId(e.target.value)}
                    placeholder="HMS15"
                  />
                </div>

                <div className="form-group-custom">
                  <label className="form-label-custom">
                    {t("Patient")}: <span className="text-danger">*</span>
                  </label>
                  {remotePatients.length > 0 ? (
                    <select
                      required
                      className="form-select-custom"
                      value={formAssignPatient}
                      onChange={(e) => setFormAssignPatient(e.target.value)}
                    >
                      <option value="">{t("Select Patient")}</option>
                      {remotePatients.map((p) => (
                        <option
                          key={p.id}
                          value={`${p.givenName} ${p.familyName}`.trim()}
                        >
                          {p.givenName} {p.familyName} ({p.mrn})
                        </option>
                      ))}
                    </select>
                  ) : (
                    <input
                      type="text"
                      required
                      placeholder={t("Patient Name")}
                      className="form-input-custom"
                      value={formAssignPatient}
                      onChange={(e) => setFormAssignPatient(e.target.value)}
                    />
                  )}
                </div>

                <div className="form-group-custom">
                  <label className="form-label-custom">
                    {t("Bed")}: <span className="text-danger">*</span>
                  </label>
                  <select
                    required
                    className="form-select-custom"
                    value={formAssignBed}
                    onChange={(e) => setFormAssignBed(e.target.value)}
                  >
                    {beds.map((b) => (
                      <option key={b.id} value={b.bedName}>
                        {b.bedName} ({b.bedType})
                      </option>
                    ))}
                  </select>
                </div>

                <div className="form-group-custom">
                  <label className="form-label-custom">
                    {t("Assign Date")}: <span className="text-danger">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. 5th Oct, 2026"
                    className="form-input-custom"
                    value={formAssignDate}
                    onChange={(e) => setFormAssignDate(e.target.value)}
                  />
                </div>

                {editingItem && (
                  <div className="form-group-custom">
                    <label className="form-label-custom">
                      {t("Discharge Date")}:
                    </label>
                    <input
                      type="text"
                      placeholder="e.g. 23rd October 2026"
                      className="form-input-custom"
                      value={formAssignDischargeDate}
                      onChange={(e) =>
                        setFormAssignDischargeDate(e.target.value)
                      }
                    />
                  </div>
                )}

                <div className="form-group-custom">
                  <label className="form-label-custom">{t("Status")}:</label>
                  <div style={{ marginTop: 8 }}>
                    <label className="switch-toggle" aria-label={t("Status")}>
                      <input
                        type="checkbox"
                        checked={formAssignStatus}
                        onChange={(e) => setFormAssignStatus(e.target.checked)}
                      />
                      <span className="slider round"></span>
                    </label>
                  </div>
                </div>
              </div>

              <div className="form-group-custom">
                <label className="form-label-custom">{t("Description")}:</label>
                <textarea
                  rows={3}
                  placeholder={t("Description")}
                  className="form-textarea-custom"
                  value={formAssignDescription}
                  onChange={(e) => setFormAssignDescription(e.target.value)}
                />
              </div>

              <div className="modal-footer-custom">
                <button
                  type="submit"
                  className="btn-modal-save"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <span className="d-inline-flex align-items-center gap-1">
                      <Loader2 size={14} className="animate-spin" />
                      {t("Saving...")}
                    </span>
                  ) : (
                    t("Save")
                  )}
                </button>
                <button
                  type="button"
                  className="btn-modal-cancel"
                  disabled={isSubmitting}
                  onClick={() => setBedAssignModalOpen(false)}
                >
                  {t("Cancel")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteConfirmOpen && (
        <div
          className="modal-backdrop-custom"
          onClick={() => setDeleteConfirmOpen(false)}
        >
          <div
            className="modal-card-custom modal-sm"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="modal-header-custom">
              <h3 className="modal-title-custom">{t("Confirm Delete")}</h3>
              <button
                type="button"
                className="modal-close-btn"
                onClick={() => setDeleteConfirmOpen(false)}
              >
                <X size={18} />
              </button>
            </div>
            <div className="modal-body-custom">
              <p>{t("Are you sure you want to delete this record?")}</p>
            </div>
            <div className="modal-footer-custom">
              <button
                type="button"
                className="btn-delete-confirm-red"
                onClick={handleDeleteConfirm}
              >
                {t("Delete")}
              </button>
              <button
                type="button"
                className="btn-modal-cancel"
                onClick={() => setDeleteConfirmOpen(false)}
              >
                {t("Cancel")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
