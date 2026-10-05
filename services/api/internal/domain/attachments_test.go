package domain

import (
	"strings"
	"testing"
)

func TestSecureAttachmentValidation(t *testing.T) {
	validHash := strings.Repeat("a", 64)
	validPat := "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name    string
		input   CreateSecureAttachmentInput
		wantErr bool
	}{
		{
			name: "valid pdf attachment",
			input: CreateSecureAttachmentInput{
				FileName:      "lab_result.pdf",
				MimeType:      "application/pdf",
				FileSizeBytes: 1024 * 1024,
				StoragePath:   "/storage/attachments/lab_result.pdf",
				Sha256Hash:    validHash,
				PatientID:     &validPat,
			},
			wantErr: false,
		},
		{
			name: "valid image attachment",
			input: CreateSecureAttachmentInput{
				FileName:      "xray.png",
				MimeType:      "image/png",
				FileSizeBytes: 5 * 1024 * 1024,
				StoragePath:   "/storage/attachments/xray.png",
				Sha256Hash:    validHash,
			},
			wantErr: false,
		},
		{
			name: "disallowed mime type executable",
			input: CreateSecureAttachmentInput{
				FileName:      "script.exe",
				MimeType:      "application/x-msdownload",
				FileSizeBytes: 1024,
				StoragePath:   "/storage/attachments/script.exe",
				Sha256Hash:    validHash,
			},
			wantErr: true,
		},
		{
			name: "exceeds 25MB max size limit",
			input: CreateSecureAttachmentInput{
				FileName:      "huge_scan.dicom",
				MimeType:      "application/dicom",
				FileSizeBytes: 27 * 1024 * 1024, // 27MB
				StoragePath:   "/storage/attachments/huge_scan.dicom",
				Sha256Hash:    validHash,
			},
			wantErr: true,
		},
		{
			name: "zero byte file rejected",
			input: CreateSecureAttachmentInput{
				FileName:      "empty.pdf",
				MimeType:      "application/pdf",
				FileSizeBytes: 0,
				StoragePath:   "/storage/attachments/empty.pdf",
				Sha256Hash:    validHash,
			},
			wantErr: true,
		},
		{
			name: "invalid sha256 hash length",
			input: CreateSecureAttachmentInput{
				FileName:      "doc.pdf",
				MimeType:      "application/pdf",
				FileSizeBytes: 1000,
				StoragePath:   "/storage/attachments/doc.pdf",
				Sha256Hash:    "short-hash",
			},
			wantErr: true,
		},
		{
			name: "empty filename rejected",
			input: CreateSecureAttachmentInput{
				FileName:      "   ",
				MimeType:      "application/pdf",
				FileSizeBytes: 1000,
				StoragePath:   "/storage/attachments/doc.pdf",
				Sha256Hash:    validHash,
			},
			wantErr: true,
		},
		{
			name: "empty storage path rejected",
			input: CreateSecureAttachmentInput{
				FileName:      "doc.pdf",
				MimeType:      "application/pdf",
				FileSizeBytes: 1000,
				StoragePath:   "  ",
				Sha256Hash:    validHash,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
