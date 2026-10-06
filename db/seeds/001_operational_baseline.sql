-- 001_operational_baseline.sql
-- Realistic Ethiopian Hospital Operational Baseline Seed Data
-- Provides live records for all operational workspaces without mock fallbacks.

BEGIN;

-- 1. Doctor Departments
INSERT INTO doctor_department (id, title, description, archived) VALUES
  ('a1111111-1111-1111-1111-111111111111', 'Cardiology', 'Heart and cardiovascular care', false),
  ('a2222222-2222-2222-2222-222222222222', 'General Surgery', 'Elective and acute surgical procedures', false),
  ('a3333333-3333-3333-3333-333333333333', 'Pediatrics', 'Infant and child wellness and pathology', false),
  ('a4444444-4444-4444-4444-444444444444', 'Obstetrics & Gynecology', 'Maternity and womens health', false),
  ('a5555555-5555-5555-5555-555555555555', 'Orthopedics', 'Bones, joints and musculoskeletal care', false),
  ('a6666666-6666-6666-6666-666666666666', 'General Medicine', 'Primary care and internal medicine', false)
ON CONFLICT DO NOTHING;

-- 2. Staff Users Across Roles
INSERT INTO "user" (id, name, email) VALUES
  ('usr-adm-01', 'Hospital Administrator', 'admin@hospital.et'),
  ('usr-adm-02', 'System Superuser', 'superadmin@hospital.et'),
  ('usr-doc-01', 'Dr. Dawit Haile', 'dawit.haile@hospital.et'),
  ('usr-doc-02', 'Dr. Tigist Mengistu', 'tigist.mengistu@hospital.et'),
  ('usr-doc-03', 'Dr. Yohannes Berhanu', 'yohannes.berhanu@hospital.et'),
  ('usr-doc-04', 'Dr. Meron Tesfaye', 'meron.tesfaye@hospital.et'),
  ('usr-doc-05', 'Dr. Henok Tadesse', 'henok.tadesse@hospital.et'),
  ('usr-doc-06', 'Dr. Selamawit Assefa', 'selamawit.assefa@hospital.et'),
  ('usr-nur-01', 'Sister Aster Aweke', 'aster.aweke@hospital.et'),
  ('usr-nur-02', 'Nurse Bethelhem Desta', 'bethelhem.desta@hospital.et'),
  ('usr-acc-01', 'Kassahun Bekele', 'kassahun.bekele@hospital.et'),
  ('usr-phr-01', 'Rahel Kebede', 'rahel.kebede@hospital.et'),
  ('usr-lab-01', 'Ephrem Girma', 'ephrem.girma@hospital.et'),
  ('usr-rec-01', 'Marta Wolde', 'marta.wolde@hospital.et')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email;

INSERT INTO staff_access (user_id, role, active) VALUES
  ('usr-adm-01', 'admin', true),
  ('usr-adm-02', 'admin', true),
  ('usr-doc-01', 'doctor', true),
  ('usr-doc-02', 'doctor', true),
  ('usr-doc-03', 'doctor', true),
  ('usr-doc-04', 'doctor', true),
  ('usr-doc-05', 'doctor', true),
  ('usr-doc-06', 'doctor', true),
  ('usr-nur-01', 'nurse', true),
  ('usr-nur-02', 'nurse', true),
  ('usr-acc-01', 'accountant', true),
  ('usr-phr-01', 'pharmacist', true),
  ('usr-lab-01', 'lab_technician', true),
  ('usr-rec-01', 'receptionist', true)
ON CONFLICT (user_id) DO UPDATE SET role = EXCLUDED.role, active = true;

-- 3. Doctor Profiles
INSERT INTO doctor_profile (user_id, department, department_id, slot_minutes, description, photo_url, opd_charge, appointment_charge) VALUES
  ('usr-doc-01', 'Cardiology', 'a1111111-1111-1111-1111-111111111111', 30, 'Senior Consultant Cardiologist', '', 500.00, 600.00),
  ('usr-doc-02', 'General Surgery', 'a2222222-2222-2222-2222-222222222222', 45, 'Chief of Surgery and Trauma', '', 700.00, 850.00),
  ('usr-doc-03', 'Pediatrics', 'a3333333-3333-3333-3333-333333333333', 30, 'Pediatric Consultant & Neonatal Care', '', 400.00, 500.00),
  ('usr-doc-04', 'Obstetrics & Gynecology', 'a4444444-4444-4444-4444-444444444444', 30, 'Maternal Care & Gynecological Surgery', '', 600.00, 750.00),
  ('usr-doc-05', 'Orthopedics', 'a5555555-5555-5555-5555-555555555555', 30, 'Orthopedic Surgeon & Sports Injuries', '', 650.00, 800.00),
  ('usr-doc-06', 'General Medicine', 'a6666666-6666-6666-6666-666666666666', 20, 'Internal Medicine Specialist', '', 350.00, 450.00)
ON CONFLICT (user_id) DO UPDATE SET 
  department = EXCLUDED.department,
  department_id = EXCLUDED.department_id,
  opd_charge = EXCLUDED.opd_charge,
  appointment_charge = EXCLUDED.appointment_charge;

-- Doctor Hours (Mon-Fri 08:30 to 17:00)
INSERT INTO doctor_hours (doctor_id, weekday, start_minute, end_minute) VALUES
  ('usr-doc-01', 1, 510, 1020), ('usr-doc-01', 2, 510, 1020), ('usr-doc-01', 3, 510, 1020), ('usr-doc-01', 4, 510, 1020), ('usr-doc-01', 5, 510, 1020),
  ('usr-doc-02', 1, 510, 1020), ('usr-doc-02', 2, 510, 1020), ('usr-doc-02', 3, 510, 1020), ('usr-doc-02', 4, 510, 1020), ('usr-doc-02', 5, 510, 1020),
  ('usr-doc-03', 1, 510, 1020), ('usr-doc-03', 2, 510, 1020), ('usr-doc-03', 3, 510, 1020), ('usr-doc-03', 4, 510, 1020), ('usr-doc-03', 5, 510, 1020)
ON CONFLICT (doctor_id, weekday, start_minute) DO NOTHING;

-- 4. Ethiopian Patients
INSERT INTO patient (id, given_name, family_name, date_of_birth, phone, created_at) VALUES
  ('b1111111-1111-1111-1111-111111111111', 'Abebe', 'Bikila', '1985-04-12', '+251911223344', now() - interval '20 days'),
  ('b2222222-2222-2222-2222-222222222222', 'Almaz', 'Ayana', '1992-06-25', '+251911334455', now() - interval '18 days'),
  ('b3333333-3333-3333-3333-333333333333', 'Haile', 'Gebrselassie', '1978-11-03', '+251911445566', now() - interval '15 days'),
  ('b4444444-4444-4444-4444-444444444444', 'Tirunesh', 'Dibaba', '1989-08-19', '+251911556677', now() - interval '12 days'),
  ('b5555555-5555-5555-5555-555555555555', 'Kenenisa', 'Bekele', '1982-10-14', '+251911667788', now() - interval '10 days'),
  ('b6666666-6666-6666-6666-666666666666', 'Derartu', 'Tulu', '1975-03-21', '+251911778899', now() - interval '8 days'),
  ('b7777777-7777-7777-7777-777777777777', 'Mamo', 'Wolde', '1968-07-09', '+251911889900', now() - interval '5 days'),
  ('b8888888-8888-8888-8888-888888888888', 'Meseret', 'Defar', '1987-12-30', '+251911990011', now() - interval '2 days')
ON CONFLICT (id) DO NOTHING;

INSERT INTO patient_profile (patient_id, email, gender, blood_group, father_name, city, region, country, active) VALUES
  ('b1111111-1111-1111-1111-111111111111', 'abebe.bikila@ethiopia.et', 'male', 'O+', 'Bikila', 'Addis Ababa', 'Bole', 'Ethiopia', true),
  ('b2222222-2222-2222-2222-222222222222', 'almaz.ayana@ethiopia.et', 'female', 'A+', 'Ayana', 'Addis Ababa', 'Kirkos', 'Ethiopia', true),
  ('b3333333-3333-3333-3333-333333333333', 'haile.g@ethiopia.et', 'male', 'B+', 'Gebrselassie', 'Addis Ababa', 'Yeka', 'Ethiopia', true),
  ('b4444444-4444-4444-4444-444444444444', 'tirunesh.d@ethiopia.et', 'female', 'AB+', 'Dibaba', 'Addis Ababa', 'Arada', 'Ethiopia', true),
  ('b5555555-5555-5555-5555-555555555555', 'kenenisa.b@ethiopia.et', 'male', 'O-', 'Bekele', 'Addis Ababa', 'Nifas Silk', 'Ethiopia', true),
  ('b6666666-6666-6666-6666-666666666666', 'derartu.t@ethiopia.et', 'female', 'A-', 'Tulu', 'Addis Ababa', 'Akaky Kaliti', 'Ethiopia', true),
  ('b7777777-7777-7777-7777-777777777777', 'mamo.w@ethiopia.et', 'male', 'B-', 'Wolde', 'Addis Ababa', 'Gullele', 'Ethiopia', true),
  ('b8888888-8888-8888-8888-888888888888', 'meseret.d@ethiopia.et', 'female', 'AB-', 'Defar', 'Addis Ababa', 'Kolfe Keranio', 'Ethiopia', true)
ON CONFLICT (patient_id) DO UPDATE SET 
  email = EXCLUDED.email, 
  blood_group = EXCLUDED.blood_group,
  father_name = EXCLUDED.father_name;

-- 5. Bed Types & Wards
INSERT INTO bed_type (id, name, description, active) VALUES
  ('c1111111-1111-1111-1111-111111111111', 'Intensive Care Unit', 'Critical medical care with ventilators', true),
  ('c2222222-2222-2222-2222-222222222222', 'Neonatal Intensive Care', 'Infant incubators and phototherapy', true),
  ('c3333333-3333-3333-3333-333333333333', 'VIP Suite', 'Executive private inpatient suite', true),
  ('c4444444-4444-4444-4444-444444444444', 'General Male Ward', 'Standard shared male inpatient care', true),
  ('c5555555-5555-5555-5555-555555555555', 'General Female Ward', 'Standard shared female inpatient care', true)
ON CONFLICT (id) DO NOTHING;

-- 12 Distinct Hospital Beds across 5 Wards
INSERT INTO hospital_bed (id, name, ward_name, bed_type, type_id, charge_minor, active, created_by) VALUES
  ('d1111111-1111-1111-1111-111111111111', 'ICU-101', 'ICU1', 'Intensive Care Unit', 'c1111111-1111-1111-1111-111111111111', 250000, true, 'usr-adm-01'),
  ('d2222222-2222-2222-2222-222222222222', 'ICU-102', 'ICU1', 'Intensive Care Unit', 'c1111111-1111-1111-1111-111111111111', 250000, true, 'usr-adm-01'),
  ('d3333333-3333-3333-3333-333333333333', 'NICU-201', 'NICU', 'Neonatal Intensive Care', 'c2222222-2222-2222-2222-222222222222', 180000, true, 'usr-adm-01'),
  ('d4444444-4444-4444-4444-444444444444', 'NICU-202', 'NICU', 'Neonatal Intensive Care', 'c2222222-2222-2222-2222-222222222222', 180000, true, 'usr-adm-01'),
  ('d5555555-5555-5555-5555-555555555555', 'VIP-301', 'VIP Ward', 'VIP Suite', 'c3333333-3333-3333-3333-333333333333', 350000, true, 'usr-adm-01'),
  ('d6666666-6666-6666-6666-666666666666', 'VIP-302', 'VIP Ward', 'VIP Suite', 'c3333333-3333-3333-3333-333333333333', 350000, true, 'usr-adm-01'),
  ('d7777777-7777-7777-7777-777777777777', 'GWM-401', 'General Ward Male', 'General Male Ward', 'c4444444-4444-4444-4444-444444444444', 80000, true, 'usr-adm-01'),
  ('d8888888-8888-8888-8888-888888888888', 'GWM-402', 'General Ward Male', 'General Male Ward', 'c4444444-4444-4444-4444-444444444444', 80000, true, 'usr-adm-01'),
  ('d9999999-9999-9999-9999-999999999999', 'GWM-403', 'General Ward Male', 'General Male Ward', 'c4444444-4444-4444-4444-444444444444', 80000, true, 'usr-adm-01'),
  ('daaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'GWF-501', 'General Ward Female', 'General Female Ward', 'c5555555-5555-5555-5555-555555555555', 80000, true, 'usr-adm-01'),
  ('dbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'GWF-502', 'General Ward Female', 'General Female Ward', 'c5555555-5555-5555-5555-555555555555', 80000, true, 'usr-adm-01'),
  ('dccccccc-cccc-cccc-cccc-cccccccccccc', 'GWF-503', 'General Ward Female', 'General Female Ward', 'c5555555-5555-5555-5555-555555555555', 80000, true, 'usr-adm-01')
ON CONFLICT (id) DO NOTHING;

-- 6. Active Inpatient Admissions (Cases & Encounters)
-- Occupy 2 beds: ICU-101 with Abebe Bikila, and VIP-301 with Haile Gebrselassie
INSERT INTO patient_case (id, patient_id, doctor_id, description, created_by) VALUES
  ('e1111111-1111-1111-1111-111111111111', 'b1111111-1111-1111-1111-111111111111', 'usr-doc-01', 'Acute coronary observation and monitoring', 'usr-adm-01'),
  ('e2222222-2222-2222-2222-222222222222', 'b3333333-3333-3333-3333-333333333333', 'usr-doc-02', 'Post-surgical orthopedics rehabilitation', 'usr-adm-01')
ON CONFLICT (id) DO NOTHING;

INSERT INTO encounter (id, kind, case_id, patient_id, doctor_id, bed_id, admitted_at, status, bed_charge_minor, created_by, request_key) VALUES
  ('f1111111-1111-1111-1111-111111111111', 'ipd', 'e1111111-1111-1111-1111-111111111111', 'b1111111-1111-1111-1111-111111111111', 'usr-doc-01', 'd1111111-1111-1111-1111-111111111111', now() - interval '2 days', 'active', 250000, 'usr-adm-01', 'req-adm-enc-001'),
  ('f2222222-2222-2222-2222-222222222222', 'ipd', 'e2222222-2222-2222-2222-222222222222', 'b3333333-3333-3333-3333-333333333333', 'usr-doc-02', 'd5555555-5555-5555-5555-555555555555', now() - interval '1 day', 'active', 350000, 'usr-adm-01', 'req-adm-enc-002')
ON CONFLICT (created_by, request_key) DO NOTHING;

INSERT INTO bed_assignment (encounter_id, bed_id, assigned_at, bed_charge_minor, active) VALUES
  ('f1111111-1111-1111-1111-111111111111', 'd1111111-1111-1111-1111-111111111111', now() - interval '2 days', 250000, true),
  ('f2222222-2222-2222-2222-222222222222', 'd5555555-5555-5555-5555-555555555555', now() - interval '1 day', 350000, true)
ON CONFLICT DO NOTHING;

-- 7. Invoices & Payments
INSERT INTO invoice (id, patient_id, invoice_date, currency, subtotal_minor, discount_basis_points, total_minor, paid_minor, request_hash, request_key, created_by) VALUES
  ('11111111-2222-3333-4444-555555555551', 'b1111111-1111-1111-1111-111111111111', CURRENT_DATE - 3, 'ETB', 350000, 0, 350000, 350000, 'hash-inv-001', 'key-inv-001', 'usr-acc-01'),
  ('11111111-2222-3333-4444-555555555552', 'b2222222-2222-2222-2222-222222222222', CURRENT_DATE - 2, 'ETB', 150000, 0, 150000, 150000, 'hash-inv-002', 'key-inv-002', 'usr-acc-01'),
  ('11111111-2222-3333-4444-555555555553', 'b3333333-3333-3333-3333-333333333333', CURRENT_DATE - 1, 'ETB', 485000, 0, 485000, 200000, 'hash-inv-003', 'key-inv-003', 'usr-acc-01')
ON CONFLICT (created_by, request_key) DO NOTHING;

-- 8. Advance Payments
INSERT INTO patient_advance_payment (patient_id, amount_minor, payment_date, payment_method, notes, received_by) VALUES
  ('b1111111-1111-1111-1111-111111111111', 120000, CURRENT_DATE - 4, 'cash', 'Pre-admission deposit for ICU stay', 'usr-acc-01'),
  ('b3333333-3333-3333-3333-333333333333', 250000, CURRENT_DATE - 2, 'bank', 'VIP suite security deposit', 'usr-acc-01')
ON CONFLICT DO NOTHING;

-- 9. Attendance Shifts & Records
INSERT INTO attendance_shift (id, name, start_time, end_time, grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active) VALUES
  ('11111111-1111-1111-1111-111111111111', 'Day Shift', '08:00:00', '17:00:00', 15, 60, 240, 480, false, true),
  ('22222222-2222-2222-2222-222222222222', 'Night Shift', '20:00:00', '05:00:00', 15, 60, 240, 480, true, true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO attendance_record (staff_id, work_date, shift_id, check_in_at, check_out_at, status, approval_status, worked_minutes) VALUES
  ('usr-doc-01', CURRENT_DATE, '11111111-1111-1111-1111-111111111111', CURRENT_DATE + time '07:55:00', CURRENT_DATE + time '17:05:00', 'present', 'approved', 490),
  ('usr-doc-02', CURRENT_DATE, '11111111-1111-1111-1111-111111111111', CURRENT_DATE + time '08:05:00', CURRENT_DATE + time '17:00:00', 'present', 'approved', 475),
  ('usr-nur-01', CURRENT_DATE, '11111111-1111-1111-1111-111111111111', CURRENT_DATE + time '07:50:00', NULL, 'checked_in', 'draft', 320),
  ('usr-acc-01', CURRENT_DATE, '11111111-1111-1111-1111-111111111111', CURRENT_DATE + time '08:00:00', CURRENT_DATE + time '17:00:00', 'present', 'approved', 480)
ON CONFLICT (staff_id, work_date) DO NOTHING;

-- 10. Blood Bank Stock Across 8 Groups
INSERT INTO blood_stock (blood_group, units_available, updated_at) VALUES
  ('A+', 15, now()), ('A-', 8, now()),
  ('B+', 12, now()), ('B-', 6, now()),
  ('AB+', 9, now()), ('AB-', 4, now()),
  ('O+', 24, now()), ('O-', 11, now())
ON CONFLICT (blood_group) DO UPDATE SET units_available = EXCLUDED.units_available;

COMMIT;
