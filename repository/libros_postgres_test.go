package repository

import (
	"context"
	"testing"
	"github.com/jackc/pgx/v5/pgxpool"
	"api-libros/models"
)

func setupTestRepo(t *testing.T) (*pgxpool.Pool, *PostgresLibrosRepo) {
	t.Helper()

	connString := "postgres://postgres:postgres123@localhost:5432/biblioteca_test?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		t.Fatalf("no se pudo conectar a la DB: %v", err)
	}

	repo := NewPostgresLibrosRepo(pool)

	return pool, repo
}

func cleanLibrosTable(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(), "TRUNCATE TABLE libros RESTART IDENTITY")
	if err != nil {
		t.Fatalf("error limpiando tabla libros: %v", err)
	}
}

func TestLibrosRepo_GetAll_OK(t *testing.T) {
	pool, repo := setupTestRepo(t)
	defer pool.Close()

	cleanLibrosTable(t, pool)

	// arrange: insertamos data real
	_, err := pool.Exec(context.Background(), `
		INSERT INTO libros (titulo, autor, ano)
		VALUES 
			('Dune', 'Frank Herbert', 1965),
			('1984', 'George Orwell', 1949)
	`)
	if err != nil {
		t.Fatalf("error insertando libros: %v", err)
	}

	// act
	libros, err := repo.GetAll(context.Background())

	// assert
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if len(libros) != 2 {
		t.Fatalf("esperaba 2 libros, vinieron %d", len(libros))
	}
}

func TestLibrosRepo_GetByID_OK(t *testing.T) {
	pool, repo := setupTestRepo(t)
	defer pool.Close()

	cleanLibrosTable(t, pool)

	var id int
	err := pool.QueryRow(context.Background(), `
		INSERT INTO libros (titulo, autor, ano)
		VALUES ('Dune', 'Frank Herbert', 1965)
		RETURNING id
	`).Scan(&id)

	if err != nil {
		t.Fatalf("error insertando libro: %v", err)
	}

	libro, err := repo.GetByID(context.Background(), id)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if libro.ID != id {
		t.Fatalf("ID incorrecto: %+v", libro)
	}
}


func TestLibrosRepo_Create_OK(t *testing.T) {
	pool, repo := setupTestRepo(t)
	defer pool.Close()

	cleanLibrosTable(t, pool)

	in := models.LibroInput{
		Titulo: "Fahrenheit 451",
		Autor:  "Ray Bradbury",
		Ano:    1953,
	}

	libro, err := repo.Create(context.Background(), in)

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if libro.ID == 0 {
		t.Fatal("se esperaba ID generado")
	}

	if libro.Titulo != in.Titulo {
		t.Fatalf("titulo incorrecto: %+v", libro)
	}
}

func TestLibrosRepo_GetByID_NotFound(t *testing.T) {
	pool, repo := setupTestRepo(t)
	defer pool.Close()

	cleanLibrosTable(t, pool)

	_, err := repo.GetByID(context.Background(), 999)

	if err != ErrNotFound {
		t.Fatalf("esperaba ErrNotFound, vino %v", err)
	}
}

func TestLibrosRepo_Patch(t *testing.T) {
	tests := []struct {
		name  string
		patch models.LibroPatch
	}{
		{"solo titulo", models.LibroPatch{Titulo: ptr("Nuevo")}},
		{"solo autor", models.LibroPatch{Autor: ptr("Autor")}},
		{"todo", models.LibroPatch{
			Titulo: ptr("X"),
			Autor:  ptr("Y"),
			Ano:    ptrInt(2000),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, repo := setupTestRepo(t)
			defer pool.Close()
			cleanLibrosTable(t, pool)

			var id int
			pool.QueryRow(context.Background(), `
				INSERT INTO libros (titulo, autor, ano)
				VALUES ('A', 'B', 1990)
				RETURNING id
			`).Scan(&id)

			_, err := repo.Patch(context.Background(), id, tt.patch)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
		})
	}
}

func TestLibrosRepo_Delete_OK(t *testing.T) {
	pool, repo := setupTestRepo(t)
	defer pool.Close()

	cleanLibrosTable(t, pool)

	var id int
	pool.QueryRow(context.Background(), `
		INSERT INTO libros (titulo, autor, ano)
		VALUES ('X', 'Y', 2000)
		RETURNING id
	`).Scan(&id)

	err := repo.Delete(context.Background(), id)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	_, err = repo.GetByID(context.Background(), id)
	if err != ErrNotFound {
		t.Fatalf("el libro no fue eliminado")
	}
}

// helpers
func ptr(s string) *string { return &s }
func ptrInt(i int) *int    { return &i }
