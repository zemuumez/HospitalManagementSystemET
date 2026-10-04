# Staff profile and role management

Source comparison: `review/legacy/app/Models/User.php`, `Doctor.php`, `Nurse.php` and `Repositories/NurseRepository.php` share first/last name, phone, gender, optional DOB, designation, qualification and address fields; Doctor also stores specialist. Migration 020 adds the corresponding typed Go staff-details contract. Photo upload, payroll, department masters, fees and complete source-form parity remain separate work.

Administrator-only endpoints:

- `GET /v1/staff-profiles/{userId}` returns identity display fields, role/active status, profile version and details.
- `PATCH /v1/staff-profiles/{userId}` accepts `{details, version, reason}`. Details are a complete replacement: givenName, familyName, phone, gender, dateOfBirth, designation, qualification, specialty, address1, address2, city, region, country and postalCode. Names are required; gender is unknown/female/male/other; DOB may be empty but must otherwise be a real date between 1850 and today. Phone is optional E.164. Postal code stays text to preserve leading zeroes. Unknown fields are rejected. Successful writes update the display name and append before/after snapshots. Stale versions return 409.
- `PATCH /v1/staff-profiles/{userId}/role` accepts `{previousRole, role, reason}` using existing nine role identifiers. It refuses self-demotion and stale previous roles. Access changes serialize with existing administrator provisioning/disablement and portal-link operations, recheck the actor's active administrator status and lock the target staff record. The active actor guarantees another administrator remains when changing a different admin.

Role changes refuse accounts with active doctor encounters, booked/arrived appointments, active nurse encounter assignments, portal ownership or patient clinician links. Resolve those assignments before retrying. This is a guard, not an automatic care-work transfer. Existing emergency account disablement remains available through `/api/staff`; a complete reassignment/deprovisioning workflow is still pending. Historical clinical records retain the original author IDs. Role changes revoke sessions and pending MFA/trust challenges and append immutable `staff_role_event` records.

The contact phone field does not change a Firebase identity or authorize phone login. Login email changes are excluded pending a verified email-change workflow. Shared details are administrator-only and never exposed to a public doctor directory or patient account. No staff-delete endpoint exists.

Verification: isolated PostgreSQL tests cover profile persistence, exact postal values, future DOB rejection, concurrent edits, stale versions, eight denied roles, self-demotion, active-work and portal-owner guards, session revocation, role-history/profile-history immutability and HTTP origin/unknown-field handling. Go tests and vet pass.
