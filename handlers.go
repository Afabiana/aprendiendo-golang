package main

import(
	"fmt"
	"context" //cancelación, timeout, valores. El "mensajero" lleva solo lo necesario: "para ya", "tenés 5 segundos", "este es el user 12345"
	"encoding/json"
	"strconv"
	"strings"
	"net/http"
	"github.com/jackc/pgx/v5"
)

func librosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Consulta a Postgres
		rows, err := db.Query(context.Background(),
			"SELECT id, titulo, autor, ano FROM libros ORDER BY id") //espcifico las columnas para asegurar el orden en que Postgres devuelve los atributos
		if err != nil {
			respondError(w, "Error al consultar la base", http.StatusInternalServerError)
			return
		}
		defer rows.Close() //ejecutá esta línea cuando la función termine y libera los recursos del cursor (muy importante para no tener leaks)

		var libros []Libro
		for rows.Next() {
			var l Libro
			err := rows.Scan(&l.ID, &l.Titulo, &l.Autor, &l.Ano) //El orden debe coincidir exactamente con el SELECT
			if err != nil {
				respondError(w, "Error al leer filas", http.StatusInternalServerError)
				return
			}
			libros = append(libros, l)
		}

		// Si no hay error, devolvemos el JSON
		json.NewEncoder(w).Encode(libros)

	case http.MethodPost:
		var nuevo Libro

		if err := json.NewDecoder(r.Body).Decode(&nuevo); err != nil {
			respondError(w, "Formato incorrecto", http.StatusBadRequest)
		}

		//podria validar info vacia o años > hoy

		err := db.QueryRow(r.Context(),
			"INSERT INTO libros (titulo, autor, ano) VALUES ($1, $2, $3) RETURNING id",
			nuevo.Titulo, nuevo.Autor, nuevo.Ano).
			Scan(&nuevo.ID) //scan no deja de ser una funcion, si no paso puntero, recibe una copia de nuevo.ID

		if err != nil {
			respondError(w, "Error al crear nuevo libro", http.StatusInternalServerError)
			return
		}

		// Respondemos con 201 Created + el libro completo (incluyendo ID generado)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(nuevo)
	default:
		respondError(w, "metodo no permitido", http.StatusMethodNotAllowed)
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
		respondError(w, "ID inválido", http.StatusBadRequest)
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
			respondError(w, "No se encontraron resultados", http.StatusNotFound)
			return
		}

		if err != nil {
			respondError(w, "Error al consultar", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(l)

	case http.MethodPut:

		var actualizado Libro
		var upd LibroUpdate

		if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
			respondError(w, "json inválido", http.StatusBadRequest)
			return
		}

		setClauses := []string{} //voy a ir formando la query
		args := []interface{}{}  //aca voy a guardar los valores de cada atributo
		i := 1                   //contador de args

		if upd.Titulo != nil {
			setClauses = append(setClauses, fmt.Sprintf("titulo = $%d", i))
			args = append(args, *upd.Titulo)
			i++
		}

		if upd.Autor != nil {
			setClauses = append(setClauses, fmt.Sprintf("Autor = $%d", i))
			args = append(args, *upd.Autor) //upd.autor es un puntero, por eso uso & para pasar el valor y no la direccion
			i++
		}

		if upd.Ano != nil {
			setClauses = append(setClauses, fmt.Sprintf("ano = $%d", i))
			args = append(args, *upd.Ano)
			i++
		}

		if len(setClauses) == 0 { //si no recibi ningun valor
			respondError(w, "no hay campos para actualizar", http.StatusBadRequest)
			return
		}

		//aca formo la query
		query := fmt.Sprintf(
			"UPDATE libros SET %s WHERE id = $%d RETURNING id, titulo, autor, ano",
			strings.Join(setClauses, ", "),
			i,
		)

		fmt.Println("Sin join", setClauses)
		fmt.Println("query", query)

		//cuando ya hice todos los chequeos agrego el id como ultimo arg
		args = append(args, id)

		//DB.EXEC para INSERT/UPDATE/DELETE
		err := db.QueryRow(r.Context(), query, args...). //args... expande el slice como parámetros individuales
									Scan(&actualizado.ID,
				&actualizado.Titulo,
				&actualizado.Autor,
				&actualizado.Ano)

		if err != nil {
			respondError(w, "error al actualizar", http.StatusBadRequest)
			return
		}

		actualizado.ID = id
		json.NewEncoder(w).Encode(actualizado)

	case http.MethodDelete:

		result, err := db.Exec(r.Context(), "DELETE FROM libros WHERE id = $1", id)

		if err != nil {
			respondError(w, "no se puedo eliminar", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected() == 0 { //si 0 → 404
			respondError(w, "libro no encontrado", http.StatusNotFound)
			return
		}

		// libros = append(libros[:idx], libros[idx+1:]...)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Libro %d eliminado"}`, id)

	default:
		respondError(w, "metodo no permitido", http.StatusMethodNotAllowed)

	}

}

func respondError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}
