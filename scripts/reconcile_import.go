package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReconciliationReport struct {
	Timestamp      time.Time                  `json:"timestamp"`
	Database       string                     `json:"database"`
	Status         string                     `json:"status"` // "PASS" or "FAIL"
	TableCounts    map[string]int64           `json:"tableCounts"`
	FinancialSums  map[string]int64           `json:"financialSums"`
	IntegrityRules map[string]RuleCheckResult `json:"integrityRules"`
	Errors         []string                   `json:"errors,omitempty"`
}

type RuleCheckResult struct {
	Description string `json:"description"`
	Passed      bool   `json:"passed"`
	Violations  int64  `json:"violations"`
}

func main() {
	dbURL := flag.String("db", os.Getenv("DATABASE_URL"), "PostgreSQL target database connection string")
	reportFile := flag.String("out", "reconciliation_report.json", "Output JSON path for reconciliation report")
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("DATABASE_URL must be specified via flag -db or environment variable")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, *dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		log.Fatal("Unable to open reconciliation snapshot")
	}
	defer tx.Rollback(ctx)

	report := ReconciliationReport{
		Timestamp:      time.Now().UTC(),
		Database:       "HMS ET PostgreSQL",
		Status:         "PASS",
		TableCounts:    make(map[string]int64),
		FinancialSums:  make(map[string]int64),
		IntegrityRules: make(map[string]RuleCheckResult),
	}

	// 1. Table Counts
	tables := []string{
		"patient",
		"\"user\"",
		"appointment",
		"encounter",
		"invoice",
		"invoice_line",
		"invoice_payment",
		"medicine_batch",
		"blood_bank",
		"hospital_bed",
		"bed_assignment",
		"secure_attachment",
	}

	for _, table := range tables {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if err := tx.QueryRow(ctx, query).Scan(&count); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Count failed for %s: %v", table, err))
			report.Status = "FAIL"
		} else {
			report.TableCounts[table] = count
		}
	}

	// 2. Financial Sums (in cents)
	financialQueries := map[string]string{
		"total_invoiced_minor": "SELECT COALESCE(SUM(total_minor), 0) FROM invoice",
		"net_collected_minor":  "SELECT COALESCE(SUM(CASE WHEN direction='refund' THEN -amount_minor ELSE amount_minor END),0) FROM invoice_payment",
	}
	for key, query := range financialQueries {
		var sum int64
		if err := tx.QueryRow(ctx, query).Scan(&sum); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Financial sum failed for %s: %v", key, err))
			report.Status = "FAIL"
		} else {
			report.FinancialSums[key] = sum
		}
	}

	// 3. Referential and Logical Integrity Checks
	rules := map[string]struct {
		desc  string
		query string
	}{
		"zero_orphan_appointments": {
			desc:  "All appointments must reference existing patients",
			query: "SELECT COUNT(*) FROM appointment a LEFT JOIN patient p ON a.patient_id = p.id WHERE p.id IS NULL",
		},
		"zero_orphan_encounters": {
			desc:  "All encounters must reference existing patients",
			query: "SELECT COUNT(*) FROM encounter e LEFT JOIN patient p ON e.patient_id = p.id WHERE p.id IS NULL",
		},
		"non_negative_invoices": {
			desc:  "Invoice total must be non-negative",
			query: "SELECT COUNT(*) FROM invoice WHERE total_minor < 0",
		},
		"non_negative_payments": {
			desc:  "Payment amounts must be positive",
			query: "SELECT COUNT(*) FROM invoice_payment WHERE amount_minor <= 0",
		},
		"valid_patient_names": {
			desc:  "Patient records must not have blank names",
			query: "SELECT COUNT(*) FROM patient WHERE trim(given_name) = '' OR trim(family_name) = ''",
		},
		"no_duplicate_active_beds": {
			desc: "No bed can have multiple active un-discharged assignments",
			query: `SELECT COUNT(*) FROM (
				SELECT bed_id FROM encounter WHERE status='active' AND bed_id IS NOT NULL GROUP BY bed_id HAVING COUNT(*) > 1
			) dupes`,
		},
	}

	for ruleKey, rule := range rules {
		var violations int64
		if err := tx.QueryRow(ctx, rule.query).Scan(&violations); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Integrity check failed for %s: %v", ruleKey, err))
			report.Status = "FAIL"
			report.IntegrityRules[ruleKey] = RuleCheckResult{
				Description: rule.desc,
				Passed:      false,
				Violations:  -1,
			}
		} else {
			passed := violations == 0
			if !passed {
				report.Status = "FAIL"
			}
			report.IntegrityRules[ruleKey] = RuleCheckResult{
				Description: rule.desc,
				Passed:      passed,
				Violations:  violations,
			}
		}
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatalf("Failed to format reconciliation report: %v", err)
	}

	if err := os.WriteFile(*reportFile, data, 0600); err != nil {
		log.Fatalf("Failed to write report file: %v", err)
	}

	fmt.Printf("Reconciliation completed with status: %s\nReport written to: %s\n", report.Status, *reportFile)
	if report.Status != "PASS" {
		os.Exit(1)
	}
}
