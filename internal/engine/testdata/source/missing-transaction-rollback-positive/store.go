package store

import (
	"database/sql"
	"fmt"
)

func TransferBalance(db *sql.DB, from, to int64, amount float64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if _, err := tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from); err != nil {
		return fmt.Errorf("debit: %w", err)
	}
	if _, err := tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to); err != nil {
		return fmt.Errorf("credit: %w", err)
	}
	return tx.Commit()
}
