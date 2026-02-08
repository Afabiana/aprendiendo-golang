package repository

import (
	"api-libros/models"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("libro not found")

type PostgresLibrosRepo struct {
	DB *pgxpool.Pool
}

func NewPostgresLibrosRepo(db *pgxpool.Pool) *PostgresLibrosRepo {
	return &PostgresLibrosRepo{ //por que pasa un puntero funciona como un factory? o no? porque cada vez que se hace un new se hace uno nuevo?
		DB: db,
	}
}

func (repo *PostgresLibrosRepo) GetAll(ctx context.Context, f models.LibroFilter) ([]models.Libro, error) {

	query := `SELECT id, titulo, autor, ano FROM libros WHERE 1=1`
	args := []any{}
	i := 1

	if f.Autor != nil {
		query += fmt.Sprintf(" AND autor ILIKE $%d", i)
		args = append(args, "%"+*f.Autor+"%")
		i++
	}

	if f.From != nil {
		query += fmt.Sprintf(" AND ano >= $%d", i)
		args = append(args, *f.From)
		i++
	}

	if f.To != nil {
		query += fmt.Sprintf(" AND ano <= $%id", i)
		args = append(args, *f.To)
		i++
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := repo.DB.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []models.Libro

	for rows.Next() {
		var l models.Libro
		if err := rows.Scan(&l.ID, &l.Titulo, &l.Autor, &l.Ano); err != nil {
			return nil, err
		}

		result = append(result, l)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *PostgresLibrosRepo) GetByID(ctx context.Context, id int) (*models.Libro, error) {
	var result models.Libro

	err := repo.DB.QueryRow(ctx,
		"SELECT id, titulo, autor, ano FROM libros WHERE id = $1",
		id).
		Scan(&result.ID, &result.Titulo, &result.Autor, &result.Ano)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (repo *PostgresLibrosRepo) Create(ctx context.Context, in models.LibroInput) (*models.Libro, error) {
	var salida models.Libro

	err := repo.DB.QueryRow(ctx,
		"INSERT INTO libros (titulo, autor, ano) VALUES ($1, $2, $3) RETURNING id, titulo, autor, ano",
		in.Titulo, in.Autor, in.Ano).
		Scan(&salida.ID,
			&salida.Titulo,
			&salida.Autor,
			&salida.Ano) //scan no deja de ser una funcion, si no paso puntero, recibe una copia de nuevo.ID

	if err != nil {
		return nil, err
	}

	return &salida, nil
}

func (repo *PostgresLibrosRepo) Update(ctx context.Context, id int, upd models.LibroInput) (*models.Libro, error) {
	var salida models.Libro

	//DB.EXEC para INSERT/UPDATE/DELETE
	err := repo.DB.QueryRow(ctx,
		`UPDATE libros
			SET titulo = $1, autor = $2, ano = $3 WHERE id = $4
			RETURNING id, titulo, autor, ano`,
		upd.Titulo,
		upd.Autor,
		upd.Ano,
		id,
	).Scan(
		&salida.ID,
		&salida.Titulo,
		&salida.Autor,
		&salida.Ano,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &salida, nil

}

func (repo *PostgresLibrosRepo) Patch(ctx context.Context, id int, patch models.LibroPatch) (*models.Libro, error) {
	var salida models.Libro

	// armo la query dinamicamente
	setClauses := []string{} //voy a ir formando la query
	args := []interface{}{}  //aca voy a guardar los valores de cada atributo //esta bien o deberia ir any?
	argsPos := 1             //contador de args

	if patch.Titulo != nil {
		setClauses = append(setClauses, fmt.Sprintf("titulo = $%d", argsPos))
		args = append(args, *patch.Titulo)
		argsPos++
	}

	if patch.Autor != nil {
		setClauses = append(setClauses, fmt.Sprintf("autor = $%d", argsPos))
		args = append(args, *patch.Autor) //patch.autor es un puntero, por eso uso * para pasar el valor y no la direccion
		argsPos++
	}

	if patch.Ano != nil {
		setClauses = append(setClauses, fmt.Sprintf("ano = $%d", argsPos))
		args = append(args, *patch.Ano)
		argsPos++
	}

	if len(setClauses) == 0 { //si no recibi ningun valor
		return repo.GetByID(ctx, id)
	}

	//aca formo la query
	query := fmt.Sprintf(
		"UPDATE libros SET %s WHERE id = $%d RETURNING id, titulo, autor, ano",
		strings.Join(setClauses, ", "),
		argsPos,
	)

	//cuando ya hice todos los chequeos agrego el id como ultimo arg
	args = append(args, id)

	err := repo.DB.QueryRow(ctx, query, args...). //args... expande el slice como parámetros individuales
							Scan(&salida.ID,
			&salida.Titulo,
			&salida.Autor,
			&salida.Ano)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &salida, nil
}

func (repo *PostgresLibrosRepo) Delete(ctx context.Context, id int) error {
	result, err := repo.DB.Exec(ctx, "DELETE FROM libros WHERE id = $1", id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 { //si 0 → 404
		return ErrNotFound
	}
	return nil
}
