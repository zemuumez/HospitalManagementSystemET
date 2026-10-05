package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// Serialize matching retries before reading or mutating attendance. The response
// is committed in the same transaction as the mutation, including its audit event.
func attendanceReplay[T any](ctx context.Context, tx pgx.Tx, actor, operation, key string, input any, out *T) (bool, error) {
	if key == "" {
		return false, nil
	}
	if len(key) > 200 {
		return false, domain.ErrValidation
	}
	body, err := json.Marshal(input)
	if err != nil {
		return false, err
	}
	lockID, _ := json.Marshal([]string{actor, operation, key})
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, string(lockID)); err != nil {
		return false, err
	}
	var matches bool
	var response []byte
	err = tx.QueryRow(ctx, `SELECT input=$4::jsonb,response FROM attendance_request WHERE actor_id=$1 AND operation=$2 AND request_key=$3`, actor, operation, key, body).Scan(&matches, &response)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !matches {
		return false, domain.ErrConflict
	}
	return true, json.Unmarshal(response, out)
}

func attendanceRemember(ctx context.Context, tx pgx.Tx, actor, operation, key string, input, response any) error {
	if key == "" {
		return nil
	}
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}
	result, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO attendance_request(actor_id,operation,request_key,input,response) VALUES($1,$2,$3,$4,$5)`, actor, operation, key, body, result)
	return err
}
