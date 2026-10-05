# Private attachment storage

The operational review branch implements local private file storage behind an application storage interface. It is not yet enabled in the running development application.

## Configuration

1. Create a dedicated persistent directory outside the web application's public directory and outside Git. On Windows, restrict its ACL to the API service account and approved backup operators; Unix deployments should use mode 0700.
2. Set `HMS_ATTACHMENT_DIR` to its absolute path in the API's private environment. This is a local path, not a cloud credential. Leave external Firebase/SMS/payment credentials blank until configured separately.
3. Start the API. A configured but inaccessible directory stops startup. An unset directory leaves uploads/downloads unavailable rather than pretending files were stored.
4. Back up this directory together with the database. Rehearse a coordinated restore before production; database-only backups do not restore file contents.

## API contract

`POST /v1/attachments` accepts multipart form data with one `file`, required `patientId`, and optional `encounterId`. A valid authenticated session and allowed Origin are required. JSON storage paths, checksums and client-declared sizes are no longer accepted by this endpoint.

The server bounds file data to 25 MiB, detects the MIME type, calculates SHA-256, assigns an opaque storage key, and writes the private file before saving metadata. Failed metadata persistence attempts remove the newly written file. A process crash between file and database writes can leave an orphan; operational orphan reconciliation remains pending. Local storage is suitable for one API host or a persistent shared volume; multi-host object storage requires another implementation of the storage interface.

Allowed detected types are PDF, JPEG, PNG, WebP and plain text. HTML, SVG, archives, unrecognized binary and DICOM are rejected until specialized validation is implemented. MIME detection is not malware scanning. Malware scanning/quarantine remains unfinished and is not claimed complete.

`GET /v1/attachments/{token}` returns metadata; `GET /v1/attachments/{token}/content` returns a forced download with no-store and nosniff headers. Every request checks current patient ownership or care assignment. Tokens never grant access by themselves. Clinical files cannot be marked public. Administrators, assigned doctors/nurses and the linked patient have scoped read access; patient upload is not enabled. Lab-technician access awaits an explicit diagnostic-record policy.

## Verification

Isolated PostgreSQL plus real temporary-directory tests cover scoped metadata, cross-patient and cross-doctor denial, server-computed checksum/size, byte-for-byte downloads, multipart HTTP upload/download, unsafe filenames, oversized files, HTML content and storage-key traversal. Uncached Go tests and go vet pass. Retention approval, malware scanning, restore rehearsal, upload throttling, and browser integration remain separate acceptance work.
