package main

import (
	"context" //cancelación, timeout, valores. El "mensajero" lleva solo lo necesario: "para ya", "tenés 5 segundos", "este es el user 12345"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func initDB() {
	// Cadena de conexión (ajustala si cambiaste user/pass/db)
	connString := "postgres://postgres:postgres123@localhost:5432/biblioteca?sslmode=disable"

	var err error
	db, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("No se pudo conectar a PostgreSQL: %v", err)
	}

	// Prueba rápida de conexión
	err = db.Ping(context.Background())
	if err != nil {
		log.Fatalf("Ping falló: %v", err)
	}

	log.Println("¡Conexión exitosa a PostgreSQL!")
}

type Libro struct {
	ID     int    `json:"id"`
	Titulo string `json:"titulo"`
	Autor  string `json:"autor"`
	Ano    int    `json:"ano"`
}

var (
	libros   []Libro
	ultimoID int = 0
)

func librosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Consulta a Postgres
		rows, err := db.Query(context.Background(),
			"SELECT id, titulo, autor, ano FROM libros ORDER BY id") //espcifico las columnas para asegurar el orden en que Postgres devuelve los atributos
		if err != nil {
			http.Error(w, `{"error": "Error al consultar la base"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close() //ejecutá esta línea cuando la función termine y libera los recursos del cursor (muy importante para no tener leaks)

		var libros []Libro
		for rows.Next() {
			var l Libro
			err := rows.Scan(&l.ID, &l.Titulo, &l.Autor, &l.Ano) //El orden debe coincidir exactamente con el SELECT
			if err != nil {
				http.Error(w, `{"error": "Error al leer filas"}`, http.StatusInternalServerError)
				return
			}
			libros = append(libros, l)
		}

		// Si no hay error, devolvemos el JSON
		json.NewEncoder(w).Encode(libros)

	case http.MethodPost:
		var nuevo Libro

		if err := json.NewDecoder(r.Body).Decode(&nuevo); err != nil {
			http.Error(w, `{"Error": "Formato incorrecto}`, http.StatusBadRequest)
		}

		//podria validar info vacia o años > hoy

		err := db.QueryRow(r.Context(),
			"INSERT INTO libros (titulo, autor, ano) VALUES ($1, $2, $3) RETURNING id",
			nuevo.Titulo, nuevo.Autor, nuevo.Ano).
			Scan(&nuevo.ID) //scan no deja de ser una funcion, si no paso puntero, recibe una copia de nuevo.ID

		if err != nil {
			http.Error(w, `"Error": "Error al crear nuevo libro"`, http.StatusInternalServerError)
			return
		}

		// Respondemos con 201 Created + el libro completo (incluyendo ID generado)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(nuevo)
	default:
		http.Error(w, `{"error":"metodo no permitido"}`, http.StatusMethodNotAllowed)

	}
}

func librosHandlerByID(w http.ResponseWriter, r *http.Request) {

	// Siempre seteamos el Content-Type para que el cliente sepa que devolvemos JSON
	w.Header().Set("Content-type", "application/json")

	// 1. Extraer el ID de la URL
	// r.URL.Path es algo como "/tareas/123" o "/tareas/abc"
	// TrimPrefix quita el prefijo "/tareas/" → queda "123"
	idStr := strings.TrimPrefix(r.URL.Path, "/libros/")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		// 400 Bad Request + JSON de error
		// http.Error escribe status, header y body automáticamente
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	//----CASO: Se encontro libro buscado
	switch r.Method {
	case http.MethodGet:
		var l Libro

		//queryRow para consultas que devuelven una sola fila
		err := db.QueryRow(r.Context(),
			"SELECT id, titulo, autor, ano FROM libros WHERE id = $1",
			id).
			Scan(&l.ID, &l.Titulo, &l.Autor, &l.Ano)

		if err == pgx.ErrNoRows {
			http.Error(w, `{"Eror":"No se encontraron resultados"}`, http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, `{"error": "Error al consultar"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(l)

	case http.MethodPut:

		var actualizado Libro

		if err := json.NewDecoder(r.Body).Decode(&actualizado); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		//podria agregar chequeos aca para no recibir vacio

		//DB.EXEC para INSERT/UPDATE/DELETE
		_, err := db.Exec(r.Context(),
			"UPDATE libros SET titulo = $1, autor = $2, ano = $3 WHERE id = $4",
			actualizado.Titulo, actualizado.Autor, actualizado.Ano, id)

		if err != nil {
			http.Error(w, `{"Error": informacion incorrecta}`, http.StatusBadRequest)
			return
		}

		actualizado.ID = id
		json.NewEncoder(w).Encode(actualizado)

	case http.MethodDelete:

		result, err := db.Exec(r.Context(), "DELETE FROM libros WHERE id = $1", id)

		if err != nil {
			http.Error(w, `{"Error": no se puedo eliminar}`, http.StatusInternalServerError)
			return
		}

		if result.RowsAffected() == 0 { //si 0 → 404
			http.Error(w, `{"Error": libro no encontrado}`, http.StatusNotFound)
			return
		}

		// libros = append(libros[:idx], libros[idx+1:]...)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Libro %d eliminado"}`, id)

	default:
		http.Error(w, `{"error":"metodo no permitido"}`, http.StatusMethodNotAllowed)

	}

}

func main() {
	initDB()

	// Carga inicial de libros de ejemplo (para que no arranque vacío)
	// libros = []Libro{
	// 	{ID: 1, Titulo: "El Principito", Autor: "Antoine de Saint-Exupéry", Ano: 1943},
	// 	{ID: 2, Titulo: "Cien años de soledad", Autor: "Gabriel García Márquez", Ano: 1967},
	// 	{ID: 3, Titulo: "1984", Autor: "George Orwell", Ano: 1949},
	// 	{ID: 4, Titulo: "Harry Potter y la piedra filosofal", Autor: "J.K. Rowling", Ano: 1997},
	// 	{ID: 5, Titulo: "Don Quijote de la Mancha", Autor: "Miguel de Cervantes", Ano: 1605},
	// }

	// Actualizamos el contador de ID esto nomas es de prueba, despues en el addLibro iteramos el contador
	// ultimoID = 5

	http.HandleFunc("/libros", librosHandler)
	http.HandleFunc("/libros/", librosHandlerByID)

	fmt.Println("Servidor REST corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
