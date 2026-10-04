# Patient profile backend

Backend-only Section 3C increment; frontend integration is deferred.

## Source mapping

Reviewed Patient model rules/fillable, CreatePatientRequest, User demographic fields, Address fields and the patient form. The original patient identity combines login credentials and demographic records; this implementation keeps Better Auth account identities separate. Changing a patient's contact email/phone does not change their sign-in email or Firebase phone binding.

Supported fields: given/family names, required date of birth, international phone, contact email, gender (unknown/female/male/other), blood group, two address lines, city, region, country, string postal code, emergency contact name/phone/relationship, administrative active flag, and stored SMS/email preferences. Postal codes retain leading zeros. Existing and newly registered patients receive a default profile without overwriting their demographics. Full legacy custom fields/photos/smart cards, guardian authority/consent evidence and optional/unknown DOB policy remain open.

The active flag is administrative metadata; it does not deactivate a login or hide clinical history. Stored message preferences are not yet applied to every message-producing workflow; consent enforcement and preference verification remain in Section 3I.

## Routes and authorization

- GET `/v1/patients/{id}`: patient demographics and profile version. Admin/reception may read; doctors only their assigned patients; patients only their linked record. Scope denial returns 404. Billing/pharmacy roles do not gain unrestricted demographic access.
- PATCH `/v1/patients/{id}/profile`: complete profile replacement by admin/reception, including the current version and a required reason. Unrecognized fields are rejected; identity linkage and generated MRN cannot be changed here. Use the existing administrator access-link endpoint for authorized linkage.
- GET `/v1/patients/{id}/revisions?page=1`: administrator-only retained before/after snapshots, actor, reason, version and timestamp; 25 rows per page.

All requests use the existing Better Auth session checks. Mutation Origin validation, 32 KiB body limit and strict JSON apply. The new profile/revision routes are not added to the frontend proxy in this backend-first increment.

## Integrity and QA

Patient/profile rows are locked in a transaction. A stale version returns conflict, without partially changing base demographics. An accepted edit increments the profile version and records both snapshots plus an audit event in the same transaction. Revision updates/deletions are rejected by the database. No patient delete API is provided, and retained revisions reference the patient record.

Tests cover profile initialization, persistence, leading-zero postal codes, invalid email/gender/blood group/phone/version/reason, patient/doctor record scope, edit permission denial, concurrent edits, revision completeness/immutability, retained-patient deletion rejection, HTTP strict input, origin checks and revision authorization. Go tests with isolated PostgreSQL enabled and vet passed. Photos/uploads, duplicate resolution/merges, import identifiers, staff profiles and full care-team assignments remain open.

## Unknown date of birth

The original `review/legacy/app/Models/Patient.php` declares DOB nullable. Migration 025 now permits unknown DOB: omit `dateOfBirth` or send an empty string; PostgreSQL stores NULL and API reads return an empty string. No fabricated birth date or estimated age is generated. A supplied date must be a real calendar date between today and 150 years ago, inclusive, using the hospital's Africa/Addis_Ababa calendar day. Normal profile version/reason and retained before/after history apply when supplying or clearing a DOB. Existing connected forms still need optional-date presentation in Section 4.

Tests cover unknown-date create/list/detail, NULL persistence, known-to-unknown correction history, invalid leap dates and the EAT midnight/150-year boundaries. Existing dated records remain unchanged.
