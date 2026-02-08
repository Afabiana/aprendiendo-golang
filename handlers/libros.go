package handlers

import (
	//"context" cancelación, timeout, valores. El "mensajero" lleva solo lo necesario: "para ya", "tenés 5 segundos", "este es el user 12345"
	"api-libros/httphelpers"
	"api-libros/models"
	"api-libros/repository"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type LibrosHandler struct {
	repo repository.LibrosRepository
}

func NewLibrosHandler(repo repository.LibrosRepository) *LibrosHandler {
    return &LibrosHandler{
        repo: repo,
    }
}


func (h *LibrosHandler) Libros(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Type", "application/json")

	log.Printf("%s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		filtro, err := parseLibroFilter(r)

		if err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}


		result, err := h.repo.GetAll(r.Context(), filtro)

		if err != nil {
			httphelpers.RespondError(w, "Error al consultar la base", http.StatusInternalServerError)
			return
		}

		// Si no hay error, devolvemos el JSON
		httphelpers.RespondJSON(w, http.StatusOK, result)

	case http.MethodPost:
		var input models.LibroInput

		if err := httphelpers.DecodeJSON(w, r, &input); err != nil {
			httphelpers.RespondError(w, "json invalido", http.StatusBadRequest)
			return
		}

		if err := input.Validate(); err != nil {
			httphelpers.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}

		salida, err := h.repo.Create(r.Context(), input)
		if err != nil {
			httphelpers.RespondError(w, "Error al crear nuevo libro", http.StatusInternalServerError)
			return
		}

		// Respondemos con 201 Created + el libro completo (incluyendo ID generado)
		// w.WriteHeader(http.StatusCreated)
		// json.NewEncoder(w).Encode(nuevo)
		httphelpers.RespondJSON(w, http.StatusCreated, salida)

	default:
		w.Header().Set("Allow", "GET, POST") // XQ PROTOCOLO HTTP dice que servidor debería indicar qué métodos sí están permitidos para ese recurso
		httphelpers.RespondError(w, "metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *LibrosHandler) LibrosByID(w http.ResponseWriter, r *http.Request) {

	log.Printf("%s %s", r.Method, r.URL.Path)

	// Siempre seteamos el Content-Type para que el cliente sepa que devolvemos JSON
	// w.Header().Set("Content-type", "application/json")

	// 1. Extraer el ID de la URL
	// r.URL.Path es algo como "/tareas/123" o "/tareas/abc"
	// TrimPrefix quita el prefijo "/tareas/" → queda "123"
	idStr := strings.TrimPrefix(r.URL.Path, "/libros/")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		// 400 Bad Request + JSON de error
		// http.Error escribe status, header y body automáticamente
		httphelpers.RespondError(w, "ID inválido", http.StatusBadRequest)
		return
	}

	//----CASO: Se encontro libro buscado
	switch r.Method {
	case http.MethodGet:

		salida, err := h.repo.GetByID(r.Context(), id)

		if err == repository.ErrNotFound {
			httphelpers.RespondError(w, "libro no encontrado", http.StatusNotFound)
		}

		if err != nil {
			httphelpers.RespondError(w, "Error al consultar", http.StatusInternalServerError)
			return
		}

		httphelpers.RespondJSON(w, http.StatusOK, salida)

	case http.MethodPut:

		var upd models.LibroInput

		if err := httphelpers.DecodeJSON(w, r, &upd); err != nil {
			httphelpers.RespondError(w, "json invalido", http.StatusBadRequest)
			return
		}

		if err := upd.Validate(); err != nil {
			httphelpers.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}

		//DB.EXEC para INSERT/UPDATE/DELETE
		salida, err := h.repo.Update(r.Context(), id, upd)

		if err == repository.ErrNotFound {
			httphelpers.RespondError(w, "libro no encontrado", http.StatusNotFound)
			return
		}

		if err != nil {
			httphelpers.RespondError(w, "error al actualizar", http.StatusInternalServerError)
			return
		}

		httphelpers.RespondJSON(w, http.StatusOK, salida)

	case http.MethodPatch:
		var patch models.LibroPatch

		if err := httphelpers.DecodeJSON(w, r, &patch); err != nil {
			httphelpers.RespondError(w, "json invalido", http.StatusBadRequest)
			return
		}
		// vuelco la info en el LibroPatch
		// if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		// 	respondError(w, "json invalido", http.StatusBadRequest)
		// 	return
		// }

		// chequeo que los datos que llegaron sean validos
		if err := patch.Validate(); err != nil {
			httphelpers.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}

		salida, err := h.repo.Patch(r.Context(), id, patch)

		if err == repository.ErrNotFound {
			httphelpers.RespondError(w, "libro no encontrado", http.StatusNotFound)
			return
		}

		if err != nil {
			httphelpers.RespondError(w, "error al actualizar", http.StatusInternalServerError)
			return
		}

		//quizas aca podriamos usar el omitempty para retornar solo los campos actualizados
		httphelpers.RespondJSON(w, http.StatusOK, salida)
	case http.MethodDelete:

		err := h.repo.Delete(r.Context(), id)

		if err == repository.ErrNotFound { //si 0 → 404
			httphelpers.RespondError(w, "libro no encontrado", http.StatusNotFound)
			return
		}

		if err != nil {
			httphelpers.RespondError(w, "no se puedo eliminar", http.StatusInternalServerError)
			return
		}

		// libros = append(libros[:idx], libros[idx+1:]...)
		// w.WriteHeader(http.StatusOK)
		// fmt.Fprintf(w, `{"message": "Libro %d eliminado"}`, id)

		// 204 → sin body
		w.WriteHeader(http.StatusNoContent)
		// respondJSON(w, http.StatusNoContent, nil)

	default:
		w.Header().Set("Allow", "GET, PUT, PATCH, DELETE")
		httphelpers.RespondError(w, "metodo no permitido", http.StatusMethodNotAllowed)

	}

}

func parseLibroFilter(r *http.Request) (filter models.LibroFilter, err error){
	q := r.URL.Query()

	var f models.LibroFilter

	if autor := q.Get("autor"); autor != "" {
		f.Autor = &autor
	}


	if from := q.Get("from"); from != "" {
		v,err := strconv.Atoi(from)
		if err != nil{
			return f, fmt.Errorf("from invalido")
		}

		f.From = &v
	}

	if to := q.Get("to"); to != "" {
		v, err := strconv.Atoi(to)
		if err != nil {
			return f, fmt.Errorf("to invalido")
		}
		f.To = &v
	}

	if limit := q.Get("limit"); limit != ""{
		v, err := strconv.Atoi(limit)
		if err != nil{
			return f, fmt.Errorf("limit invalido")
		}
		f.Limit = v
	}else{
		f.Limit = 50 // default
	}

	if offset := q.Get("offset"); offset != ""{
		v, err := strconv.Atoi(offset)
		if err != nil {
			return f, fmt.Errorf("offset invalido")
		}
		f.Offset = v
	}

	return f, nil
}

//TODO: estructurar bien el proyecto, los DTOS de response y de respuesta, packages etc
