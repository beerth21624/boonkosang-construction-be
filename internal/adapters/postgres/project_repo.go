package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type projectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) repositories.ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (r *projectRepository) Create(ctx context.Context, req requests.CreateProjectRequest) (*models.Project, error) {
	project := &models.Project{
		ProjectID:   uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Status:      models.ProjectStatusPlanning,
		ClientID:    req.ClientID,
		CreatedAt:   time.Now(),
	}

	query := `
        INSERT INTO Project (
            project_id, name, description, address, status, 
            client_id, created_at
        ) VALUES (
            :project_id, :name, :description, :address, :status,
            :client_id, :created_at
        ) RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, query, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(project)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		return project, nil
	}
	return nil, errors.New("failed to create project: no rows returned")
}

func (r *projectRepository) Update(ctx context.Context, id uuid.UUID, req requests.UpdateProjectRequest) error {
	query := `
        UPDATE Project SET 
            name = :name,
            description = :description,
            address = :address,
			client_id = :client_id,
            updated_at = :updated_at
        WHERE project_id = :project_id`

	params := map[string]interface{}{
		"project_id":  id,
		"name":        req.Name,
		"description": req.Description,
		"address":     req.Address,
		"client_id":   req.ClientID,
		"updated_at":  time.Now(),
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("project not found")
	}

	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM Project WHERE project_id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return errors.New("project not found")
	}
	return nil
}

func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	project := &models.Project{}
	query := `SELECT * FROM Project WHERE project_id = $1`

	err := r.db.GetContext(ctx, project, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

func (r *projectRepository) GetByIDWithClient(ctx context.Context, id uuid.UUID) (*models.Project, *models.Client, error) {
	project := &models.Project{}
	client := &models.Client{}

	query := `
        SELECT 
            p.*,
            c.client_id as "client.client_id",
            c.name as "client.name",
            c.email as "client.email",
            c.tel as "client.tel",
            c.address as "client.address",
            c.tax_id as "client.tax_id"
        FROM Project p
        LEFT JOIN Client c ON p.client_id = c.client_id
        WHERE p.project_id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ProjectID, &project.Name, &project.Description,
		&project.Address, &project.Status, &project.ClientID,
		&project.CreatedAt, &project.UpdatedAt,
		&client.ClientID, &client.Name, &client.Email,
		&client.Tel, &client.Address, &client.TaxID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("project not found")
		}
		return nil, nil, fmt.Errorf("failed to get project with client: %w", err)
	}

	return project, client, nil
}

func (r *projectRepository) List(ctx context.Context) ([]models.Project, error) {
	var projects []models.Project

	query := `
		SELECT * FROM Project 
	`

	err := r.db.SelectContext(ctx, &projects, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}