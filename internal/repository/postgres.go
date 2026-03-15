package repository

import (
	"database/sql"
	"shorturl/internal/model"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository[T Entity](db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(entity model.Link) error {
	_, err := r.db.Exec("INSERT INTO links VALUES ($1, $2)", entity.ID, entity.URL)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) GetByID(id string) (model.Link, error) {
	var link model.Link
	row := r.db.QueryRow("SELECT id, url from links WHERE id=$1", id)
	err := row.Scan(&link.ID, &link.URL)
	if err != nil {
		var zero model.Link
		return zero, err
	}

	return link, nil

}
