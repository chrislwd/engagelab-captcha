package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/engagelab/captcha/internal/model"
)

// PGAppStore implements AppStore backed by PostgreSQL.
type PGAppStore struct {
	pool *pgxpool.Pool
}

// NewPGAppStore creates a new PostgreSQL-backed app store.
func NewPGAppStore(pool *pgxpool.Pool) *PGAppStore {
	return &PGAppStore{pool: pool}
}

// CreateApp inserts a new app into the database.
func (s *PGAppStore) CreateApp(app *model.App) error {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO apps (id, tenant_id, name, site_key, secret_key, allowed_domains, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		app.ID, app.TenantID, app.Name, app.SiteKey, app.SecretKey,
		app.AllowedDomains, string(app.Status), app.CreatedAt, app.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}
	return nil
}

// GetApp retrieves an app by its ID.
func (s *PGAppStore) GetApp(id string) (*model.App, error) {
	ctx := context.Background()
	return s.scanApp(s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, site_key, secret_key, allowed_domains, status, created_at, updated_at
		 FROM apps WHERE id = $1`, id,
	))
}

// GetAppBySiteKey retrieves an app by its site key.
func (s *PGAppStore) GetAppBySiteKey(siteKey string) (*model.App, error) {
	ctx := context.Background()
	return s.scanApp(s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, site_key, secret_key, allowed_domains, status, created_at, updated_at
		 FROM apps WHERE site_key = $1`, siteKey,
	))
}

// GetAppBySecretKey retrieves an app by its secret key.
func (s *PGAppStore) GetAppBySecretKey(secretKey string) (*model.App, error) {
	ctx := context.Background()
	return s.scanApp(s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, name, site_key, secret_key, allowed_domains, status, created_at, updated_at
		 FROM apps WHERE secret_key = $1`, secretKey,
	))
}

// ListAppsByTenant returns all apps belonging to a tenant.
func (s *PGAppStore) ListAppsByTenant(tenantID string) []*model.App {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx,
		`SELECT id, tenant_id, name, site_key, secret_key, allowed_domains, status, created_at, updated_at
		 FROM apps WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*model.App
	for rows.Next() {
		app, err := s.scanAppRow(rows)
		if err != nil {
			continue
		}
		result = append(result, app)
	}
	return result
}

// DeleteApp removes an app by its ID.
func (s *PGAppStore) DeleteApp(id string) error {
	ctx := context.Background()
	tag, err := s.pool.Exec(ctx, `DELETE FROM apps WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete app: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("app not found: %s", id)
	}
	return nil
}

func (s *PGAppStore) scanApp(row pgx.Row) (*model.App, error) {
	var app model.App
	var status string
	err := row.Scan(
		&app.ID, &app.TenantID, &app.Name, &app.SiteKey, &app.SecretKey,
		&app.AllowedDomains, &status, &app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}
	app.Status = model.AppStatus(status)
	return &app, nil
}

func (s *PGAppStore) scanAppRow(rows pgx.Rows) (*model.App, error) {
	var app model.App
	var status string
	err := rows.Scan(
		&app.ID, &app.TenantID, &app.Name, &app.SiteKey, &app.SecretKey,
		&app.AllowedDomains, &status, &app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	app.Status = model.AppStatus(status)
	return &app, nil
}
