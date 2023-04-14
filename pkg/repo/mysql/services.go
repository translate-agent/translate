package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
)

func (r *Repo) SaveService(ctx context.Context, service *model.Service) error {
	query := `INSERT INTO service (id, name) VALUES (UUID_TO_BIN(?), ?) ON DUPLICATE KEY UPDATE name = VALUES (name)`

	if service.ID == uuid.Nil {
		service.ID = uuid.New()
	}

	_, err := r.db.ExecContext(ctx, query, service.ID, service.Name)
	if err != nil {
		return &repo.DefaultError{Entity: "service", Err: err, Operation: "Insert"}
	}

	return nil
}

func (r *Repo) LoadService(ctx context.Context, serviceID uuid.UUID) (*model.Service, error) {
	query := `SELECT id, name FROM service WHERE id = UUID_TO_BIN(?)`
	row := r.db.QueryRowContext(ctx, query, serviceID)

	var service model.Service

	switch err := row.Scan(&service.ID, &service.Name); {
	default:
		return &service, nil
	case errors.Is(err, sql.ErrNoRows):
		return nil, &repo.NotFoundError{Entity: "service", Fields: map[string]string{"id": serviceID.String()}}
	case err != nil:
		return nil, &repo.DefaultError{Entity: "service", Err: err, Operation: "Select"}
	}
}

func (r *Repo) LoadServices(ctx context.Context) ([]model.Service, error) {
	query := `SELECT id, name FROM service`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, &repo.DefaultError{Entity: "service", Err: err, Operation: "Select"}
	}
	defer rows.Close()

	var services []model.Service

	for rows.Next() {
		var service model.Service

		err = rows.Scan(&service.ID, &service.Name)
		if err != nil {
			return nil, &repo.DefaultError{Entity: "service", Err: err, Operation: "Scan"}
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, &repo.DefaultError{Entity: "service", Err: err, Operation: "Scan"}
	}

	return services, nil
}

func (r *Repo) DeleteService(ctx context.Context, serviceID uuid.UUID) error {
	query := `DELETE FROM service WHERE id = UUID_TO_BIN(?)`

	result, err := r.db.ExecContext(ctx, query, serviceID)
	if err != nil {
		return &repo.DefaultError{Entity: "service", Err: err, Operation: "Delete"}
	}

	switch count, err := result.RowsAffected(); {
	default:
		return nil
	case err != nil:
		return &repo.DefaultError{Entity: "service", Err: err, Operation: "Delete Result"}
	case count == 0:
		return &repo.NotFoundError{Entity: "service", Fields: map[string]string{"id": serviceID.String()}}
	}
}
