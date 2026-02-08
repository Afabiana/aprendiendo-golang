package repository

import (
	"api-libros/models"
	"context"
)

type LibrosRepository interface {
	GetAll(ctx context.Context, filter models.LibroFilter) ([]models.Libro, error)
	GetByID(ctx context.Context, id int) (*models.Libro, error)
	Create(ctx context.Context, in models.LibroInput) (*models.Libro, error)
	Update(ctx context.Context, id int, upd models.LibroInput) (*models.Libro, error)
	Patch(ctx context.Context, id int, p models.LibroPatch) (*models.Libro, error)
	Delete(ctx context.Context, id int) error
}
