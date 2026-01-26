package db

import (
	"context"
	"log"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func New() *pgxpool.Pool{
	// Cadena de conexión (user/pass/db)
	connString := "postgres://postgres:postgres123@localhost:5432/biblioteca?sslmode=disable"

	var err error
	db, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("No se pudo conectar a PostgreSQL: %v", err)
	}

	// // Prueba rápida de conexión
	// err = db.Ping(context.Background())
	// if err != nil {
	// 	log.Fatalf("Ping falló: %v", err)
	// }

	// log.Println("¡Conexión exitosa a PostgreSQL!")
	return db
}
