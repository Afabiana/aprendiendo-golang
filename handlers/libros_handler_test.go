package handlers

import (
	"api-libros/models"
	"api-libros/repository"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


type FakeLibrosRepo struct {
	libros map[int]models.Libro
}

func NewFakeLibrosRepo() *FakeLibrosRepo {
	return &FakeLibrosRepo{
		libros: map[int]models.Libro{
			1: {ID: 1, Titulo: "Dune", Autor: "Frank Herbert", Ano: 1965},
			2: {ID: 2, Titulo: "1984", Autor: "George Orwell", Ano: 1949},
			3: {ID: 3, Titulo: "Fahrenheit 451", Autor: "Ray Bradbury", Ano: 1953},
		},
	}
}

func (f *FakeLibrosRepo) GetAll(ctx context.Context) ([]models.Libro, error) {
	res := []models.Libro{}
	for _, l := range f.libros {
		res = append(res, l)
	}
	return res, nil
}

func (f *FakeLibrosRepo) GetByID(ctx context.Context, id int) (*models.Libro, error) {
	l, ok := f.libros[id]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return &l, nil
}

func (f *FakeLibrosRepo) Create(ctx context.Context, in models.LibroInput) (*models.Libro, error) {
	id := len(f.libros) + 1
	l := models.Libro{
		ID:     id,
		Titulo: in.Titulo,
		Autor:  in.Autor,
		Ano:    in.Ano,
	}
	f.libros[id] = l
	return &l, nil
}

func (f *FakeLibrosRepo) Update(ctx context.Context, id int, in models.LibroInput) (*models.Libro, error) {
	if _, ok := f.libros[id]; !ok {
		return nil, repository.ErrNotFound
	}

	l := models.Libro{
		ID:     id,
		Titulo: in.Titulo,
		Autor:  in.Autor,
		Ano:    in.Ano,
	}

	f.libros[id] = l
	return &l, nil

}

func (f *FakeLibrosRepo) Patch(ctx context.Context, id int, in models.LibroPatch) (*models.Libro, error) {
	existing, ok := f.libros[id]
	if !ok {
		return nil, repository.ErrNotFound
	}

	if in.Titulo != nil {
		existing.Titulo = *in.Titulo
	}
	if in.Autor != nil {
		existing.Autor = *in.Autor
	}
	if in.Ano != nil {
		existing.Ano = *in.Ano
	}

	f.libros[id] = existing
	return &existing, nil
}

func (f *FakeLibrosRepo) Delete(ctx context.Context, id int) error {

	// if borrado == nil { //el id NUNCA VA A SER 0 a menos que no se encuentre un libro con ese id
	// 	return repository.ErrNotFound
	// }
	_, ok := f.libros[id]
	if !ok {
		return repository.ErrNotFound
	}

	delete(f.libros, id)
	return nil
}

// --------------------- METODOS DE PRUEBA ---------------------

func TestLibros_GET_All(t *testing.T) {
	repo := NewFakeLibrosRepo()
	handler := NewLibrosHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/libros", nil)
	rr := httptest.NewRecorder()

	handler.Libros(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, vino %d", rr.Code)
	}

	var resp []models.Libro
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("json invalido: %v", err)
	}

	if len(resp) != 3 {
		t.Fatalf("esperaba 3 libros, vinieron %d", len(resp))
	}
}

func TestLibros_GET_ByID_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantTitulo string
	}{
		{
			name:       "existe",
			id:         "2",
			wantStatus: http.StatusOK,
			wantTitulo: "1984",
		},
		{
			name:       "no existe",
			id:         "999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "id invalido",
			id:         "abc",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewFakeLibrosRepo()
			handler := NewLibrosHandler(repo)

			req := httptest.NewRequest(
				http.MethodGet,
				"/libros/"+tt.id,
				nil,
			)
			rr := httptest.NewRecorder()

			handler.LibrosByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status esperado %d, vino %d", tt.wantStatus, rr.Code)
			}

			if tt.wantStatus == http.StatusOK {
				resp := decodeJSON[models.Libro](t, rr)

				if tt.wantTitulo != "" && resp.Titulo != tt.wantTitulo {
					t.Fatalf("titulo esperado %q, vino %q", tt.wantTitulo, resp.Titulo)
				}
			}
		})
	}
}

func TestLibros_POST_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantCount  int
	}{
		{
			name: "ok",
			body: models.LibroInput{
				Titulo: "Neuromancer",
				Autor:  "William Gibson",
				Ano:    1984,
			},
			wantStatus: http.StatusCreated,
			wantCount:  4,
		},
		{
			name:       "json invalido",
			body:       "{titulo:}",
			wantStatus: http.StatusBadRequest,
			wantCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewFakeLibrosRepo()
			handler := NewLibrosHandler(repo)

			var req *http.Request

			switch v := tt.body.(type) {
			case string:
				req = httptest.NewRequest(http.MethodPost, "/libros", strings.NewReader(v))
			default:
				req = newJSONRequest(http.MethodPost, "/libros", v)
			}

			rr := httptest.NewRecorder()
			handler.Libros(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status esperado %d, vino %d", tt.wantStatus, rr.Code)
			}

			if len(repo.libros) != tt.wantCount {
				t.Fatalf("esperaba %d libros, hay %d", tt.wantCount, len(repo.libros))
			}
		})
	}
}


func TestLibros_PUT_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		input      models.LibroInput
		wantStatus int
	}{
		{
			name: "ok",
			id:   "1",
			input: models.LibroInput{
				Titulo: "Dune Messiah",
				Autor:  "Frank Herbert",
				Ano:    1969,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "no existe",
			id:   "999",
			input: models.LibroInput{
				Titulo: "X",
				Autor:  "Y",
				Ano:    2000,
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewFakeLibrosRepo()
			handler := NewLibrosHandler(repo)

			req := newJSONRequest(http.MethodPut, "/libros/"+tt.id, tt.input)
			rr := httptest.NewRecorder()

			handler.LibrosByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status esperado %d, vino %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestLibros_PATCH_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		patch      models.LibroPatch
		wantStatus int
		wantTitulo string
	}{
		{
			name: "patch titulo",
			id:   "1",
			patch: models.LibroPatch{
				Titulo: ptr("Nuevo titulo"),
			},
			wantStatus: http.StatusOK,
			wantTitulo: "Nuevo titulo",
		},
		{
			name:       "no existe",
			id:         "999",
			patch:      models.LibroPatch{Titulo: ptr("X")},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewFakeLibrosRepo()
			handler := NewLibrosHandler(repo)

			req := newJSONRequest(http.MethodPatch, "/libros/"+tt.id, tt.patch)
			rr := httptest.NewRecorder()

			handler.LibrosByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status esperado %d, vino %d", tt.wantStatus, rr.Code)
			}

			if tt.wantStatus == http.StatusOK {
				resp := decodeJSON[models.Libro](t, rr)

				if tt.wantTitulo != "" && resp.Titulo != tt.wantTitulo {
					t.Fatalf("titulo incorrecto")
				}
			}
		})
	}
}

func TestLibros_DELETE_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantExists bool
	}{
		{
			name:       "ok",
			id:         "1",
			wantStatus: http.StatusNoContent,
			wantExists: false,
		},
		{
			name:       "no existe",
			id:         "999",
			wantStatus: http.StatusNotFound,
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewFakeLibrosRepo()
			handler := NewLibrosHandler(repo)

			req := httptest.NewRequest(http.MethodDelete, "/libros/"+tt.id, nil)
			rr := httptest.NewRecorder()

			handler.LibrosByID(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status esperado %d, vino %d", tt.wantStatus, rr.Code)
			}

			_, exists := repo.libros[1]
			if exists != tt.wantExists && tt.id == "1" {
				t.Fatalf("estado del repo incorrecto")
			}
		})
	}
}


// ---------- HELPERS ----------

func newJSONRequest(method, url string, body any) *http.Request {
	var reader *strings.Reader

	if body != nil {
		b, _ := json.Marshal(body)
		reader = strings.NewReader(string(b))
	} else {
		reader = strings.NewReader("")
	}

	req := httptest.NewRequest(method, url, reader)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeJSON[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var out T
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("json invalido: %v", err)
	}
	return out
}

func ptr[T any](v T) *T {
	return &v
}
