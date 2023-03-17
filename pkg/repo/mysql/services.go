package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/repo"
)

func (r *Repo) SaveService(ctx context.Context, service *model.Service) error {
	query := `INSERT INTO service (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = VALUES(name)`

	_, err := r.db.ExecContext(ctx, query, service.ID, service.Name)
	if err != nil {
		return fmt.Errorf("db: insert service: %w", err)
	}

	return nil
}

func (r *Repo) LoadService(ctx context.Context, serviceID uuid.UUID) (*model.Service, error) {
	query := `SELECT id, name FROM service WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, serviceID)

	var service model.Service

	err := row.Scan(&service.ID, &service.Name)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, repo.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("db: select service: %w", err)
	}

	return &service, nil
}

func (r *Repo) LoadServices(ctx context.Context) ([]model.Service, error) {
	query := `SELECT id, name FROM service`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("db: select services: %w", err)
	}
	defer rows.Close()

	var services []model.Service

	for rows.Next() {
		var service model.Service

		err = rows.Scan(&service.ID, &service.Name)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		services = append(services, service)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return services, nil
}

func (r *Repo) DeleteService(ctx context.Context, serviceID uuid.UUID) error {
	query := `DELETE FROM service WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, serviceID)
	if err != nil {
		return fmt.Errorf("db: delete service: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("retrieve affected rows: %w", err)
	}

	if count == 0 {
		return repo.ErrNotFound
	}

	return nil
}
