package repository

import (
	"shorturl/internal/model"

	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository[T Entity](db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(entity model.Link) error {
	_, err := r.db.Exec("INSERT INTO links (id, url, created_by) VALUES ($1, $2, $3)", entity.ID, entity.URL, entity.CreatedBy)
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
			"INSERT INTO links (id, url, created_by) VALUES ($1, $2, $3)",
			value.ID,
			value.URL,
			value.CreatedBy,
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
	err := r.db.Get(&link, "SELECT id, url, created_by, is_deleted FROM links WHERE id=$1", id)
	if err != nil {
		var zero model.Link
		return zero, err
	}

	return link, nil
}

func (r *PostgresRepository) GetByURL(url string) (model.Link, error) {
	var link model.Link
	err := r.db.Get(&link, "SELECT id, url, created_by, is_deleted FROM links WHERE url = $1", url)
	if err != nil {
		var zero model.Link
		return zero, err
	}
	return link, nil
}

func (r *PostgresRepository) GetAllByUserID(userID string) ([]model.Link, error) {
	var links []model.Link
	err := r.db.Select(&links, `SELECT id, url, created_by, is_deleted FROM links WHERE created_by = $1 AND is_deleted = FALSE`, userID)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (r *PostgresRepository) SoftDeleteByIDs(ids []string, userID string) error {
	_, err := r.db.Exec(
		`UPDATE links
		 SET is_deleted = TRUE
		 WHERE created_by = $1
		   AND id = ANY($2)`,
		userID,
		ids,
	)
	return err
}
