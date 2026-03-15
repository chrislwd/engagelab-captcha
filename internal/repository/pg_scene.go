package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/engagelab/captcha/internal/model"
)

// PGSceneStore implements SceneStore backed by PostgreSQL.
type PGSceneStore struct {
	pool *pgxpool.Pool
}

// NewPGSceneStore creates a new PostgreSQL-backed scene store.
func NewPGSceneStore(pool *pgxpool.Pool) *PGSceneStore {
	return &PGSceneStore{pool: pool}
}

// CreateScene inserts a new scene into the database.
func (s *PGSceneStore) CreateScene(scene *model.Scene) error {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO scenes (id, app_id, scene_type, policy_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		scene.ID, scene.AppID, string(scene.SceneType), scene.PolicyID,
		string(scene.Status), scene.CreatedAt, scene.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create scene: %w", err)
	}
	return nil
}

// GetScene retrieves a scene by its ID.
func (s *PGSceneStore) GetScene(id string) (*model.Scene, error) {
	ctx := context.Background()
	return s.scanScene(s.pool.QueryRow(ctx,
		`SELECT id, app_id, scene_type, policy_id, status, created_at, updated_at
		 FROM scenes WHERE id = $1`, id,
	))
}

// ListScenesByApp returns all scenes belonging to an app.
func (s *PGSceneStore) ListScenesByApp(appID string) []*model.Scene {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx,
		`SELECT id, app_id, scene_type, policy_id, status, created_at, updated_at
		 FROM scenes WHERE app_id = $1 ORDER BY created_at DESC`, appID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*model.Scene
	for rows.Next() {
		sc, err := s.scanSceneRow(rows)
		if err != nil {
			continue
		}
		result = append(result, sc)
	}
	return result
}

func (s *PGSceneStore) scanScene(row pgx.Row) (*model.Scene, error) {
	var sc model.Scene
	var sceneType, status string
	err := row.Scan(&sc.ID, &sc.AppID, &sceneType, &sc.PolicyID, &status, &sc.CreatedAt, &sc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scene not found: %w", err)
	}
	sc.SceneType = model.SceneType(sceneType)
	sc.Status = model.SceneStatus(status)
	return &sc, nil
}

func (s *PGSceneStore) scanSceneRow(rows pgx.Rows) (*model.Scene, error) {
	var sc model.Scene
	var sceneType, status string
	err := rows.Scan(&sc.ID, &sc.AppID, &sceneType, &sc.PolicyID, &status, &sc.CreatedAt, &sc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	sc.SceneType = model.SceneType(sceneType)
	sc.Status = model.SceneStatus(status)
	return &sc, nil
}
