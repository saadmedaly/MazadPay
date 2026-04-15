package repository

import (
    "github.com/jmoiron/sqlx"
)

// WithTransaction est un helper pour exécuter une fonction dans une transaction SQL.
// Rollback automatique si la fonction retourne une erreur.
func WithTransaction(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
    tx, err := db.Beginx()
    if err != nil {
        return err
    }
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()

    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }
    return tx.Commit()
}
