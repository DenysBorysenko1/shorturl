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
	_, err := r.db.Exec("INSERT INTO links (id, url) VALUES ($1, $2)", entity.ID, entity.URL)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) CreateMany(entities []model.Link) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var totalInsertedCount int

	for _, value := range entities {
		res, err := tx.Exec(
			"INSERT INTO links (id, url) VALUES ($1, $2)",
			value.ID,
			value.URL,
		)
		if err != nil {
			return 0, err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return 0, err
		}

		totalInsertedCount += int(rowsAffected)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}

	return totalInsertedCount, nil
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

func (r *PostgresRepository) GetByURL(url string) (model.Link, error) {
    var link model.Link
    row := r.db.QueryRow("SELECT id, url FROM links WHERE url = $1", url)
    err := row.Scan(&link.ID, &link.URL)
    if err != nil {
        var zero model.Link
        return zero, err
    }
    return link, nil
}
