package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/engagelab/captcha/internal/model"
)

// PGTenantStore implements TenantStore backed by PostgreSQL.
type PGTenantStore struct {
	pool *pgxpool.Pool
}

// NewPGTenantStore creates a new PostgreSQL-backed tenant store.
func NewPGTenantStore(pool *pgxpool.Pool) *PGTenantStore {
	return &PGTenantStore{pool: pool}
}

// CreateTenant inserts a new tenant into the database.
func (s *PGTenantStore) CreateTenant(t *model.Tenant) {
	ctx := context.Background()
	_, _ = s.pool.Exec(ctx,
		`INSERT INTO tenants (id, name, api_key, plan, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO NOTHING`,
		t.ID, t.Name, t.APIKey, string(t.Plan), t.CreatedAt, t.UpdatedAt,
	)
}

// GetTenantByAPIKey retrieves a tenant by its API key.
func (s *PGTenantStore) GetTenantByAPIKey(apiKey string) (*model.Tenant, error) {
	ctx := context.Background()
	row := s.pool.QueryRow(ctx,
		`SELECT id, name, api_key, plan, created_at, updated_at
		 FROM tenants WHERE api_key = $1`,
		apiKey,
	)

	var t model.Tenant
	var plan string
	err := row.Scan(&t.ID, &t.Name, &t.APIKey, &plan, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("tenant not found for api key: %w", err)
	}
	t.Plan = model.Plan(plan)
	return &t, nil
}
