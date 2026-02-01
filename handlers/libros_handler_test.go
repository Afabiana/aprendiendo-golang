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

// var ErrNotFound = errors.New("libro not found")

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
func TestGetLibro_OK(t *testing.T) {
	repo := NewFakeLibrosRepo()
	handler := NewLibrosHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/libros/2", nil)
	rr := httptest.NewRecorder()

	//act
	handler.LibrosByID(rr, req)

	//assert
	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200 pero vino %d", rr.Code)
	}

	var resp models.Libro
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("json invalido: %v", err)
	}

	if  resp.Titulo != "1984" {
		t.Fatalf("respuesta incorrecta: %+v", resp)
	}
}

func TestLibros_GET_ByID_NotFound(t *testing.T) {
	repo := NewFakeLibrosRepo()
	handler := NewLibrosHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/libros/999", nil)
	rr := httptest.NewRecorder()

	handler.LibrosByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status esperado 404, vino %d", rr.Code)
	}
}

func TestLibros_POST_OK(t *testing.T){
	repoLibros := NewFakeLibrosRepo()
	handlerLibros := NewLibrosHandler(repoLibros)

	input := models.LibroInput{
		Titulo: "Rebelión en la granja", 
		Autor: "George Orwell", 
		Ano: 1945,
	}

	body, err := json.Marshal(input)

	if err != nil{
		t.Fatalf("Error serializando json: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/libros/", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handlerLibros.Libros(rr, req)

	if rr.Code != http.StatusCreated{
		t.Fatalf("status esperado 201, vino %d", rr.Code)
	}

	var resp models.Libro
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("json invalido: %v", err)
	}

	// 3️⃣ contenido
	if resp.ID != 4 {
		t.Fatalf("id esperado 4, vino %d", resp.ID)
	}

	if resp.Titulo != input.Titulo {
		t.Fatalf("titulo incorrecto: %s", resp.Titulo)
	}

	// 4️⃣ repo modificado
	if len(repoLibros.libros) != 4 {
		t.Fatalf("esperaba 4 libros en el repo, hay %d", len(repoLibros.libros))
	}
}

func TestLibros_POST_JSON_Invalido(t *testing.T) {
	fakeRepo := NewFakeLibrosRepo()
	handler := NewLibrosHandler(fakeRepo)

	body := strings.NewReader(`{titulo:}`) // JSON roto
	req := httptest.NewRequest(http.MethodPost, "/libros", body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.Libros(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status esperado 400, vino %d", rr.Code)
	}
}

func TestLibros_PUT_OK(t *testing.T) {
	repo := NewFakeLibrosRepo()
	handler := NewLibrosHandler(repo)

	input := models.LibroInput{
		Titulo: "Dune Messiah",
		Autor:  "Frank Herbert",
		Ano:    1969,
	}

	req := newJSONRequest(http.MethodPut, "/libros/1", input)
	rr := httptest.NewRecorder()

	handler.LibrosByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, vino %d", rr.Code)
	}

	resp := decodeJSON[models.Libro](t, rr)

	if resp.Titulo != "Dune Messiah" {
		t.Fatalf("titulo incorrecto: %s", resp.Titulo)
	}
}

func TestLibros_PUT_NotFound(t *testing.T) {
	repo := NewFakeLibrosRepo()
	handler := NewLibrosHandler(repo)

	input := models.LibroInput{
		Titulo: "X",
		Autor:  "Y",
		Ano:    2000,
	}

	req := newJSONRequest(http.MethodPut, "/libros/999", input)
	rr := httptest.NewRecorder()

	handler.LibrosByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status esperado 404, vino %d", rr.Code)
	}
}

func TestLibros_PATCH_OK(t *testing.T) {
	repo := NewFakeLibrosRepo()
	handler := NewLibrosHandler(repo)

	patch := models.LibroPatch{
		Titulo: ptr("Nuevo titulo"),
	}

	req := newJSONRequest(http.MethodPatch, "/libros/1", patch)
	rr := httptest.NewRecorder()

	handler.LibrosByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, vino %d", rr.Code)
	}

	resp := decodeJSON[models.Libro](t, rr)

	if resp.Titulo != "Nuevo titulo" {
		t.Fatalf("patch no aplicado")
	}
}

func TestLibros_DELETE_OK(t *testing.T) {
	repo := NewFakeLibrosRepo()
	handler := NewLibrosHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/libros/1", nil)
	rr := httptest.NewRecorder()

	handler.LibrosByID(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status esperado 204, vino %d", rr.Code)
	}

	if _, ok := repo.libros[1]; ok {
		t.Fatalf("el libro no fue eliminado")
	}
}

func TestLibros_DELETE_NotFound(t *testing.T) {
	repo := NewFakeLibrosRepo()
	handler := NewLibrosHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/libros/999", nil)
	rr := httptest.NewRecorder()

	handler.LibrosByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status esperado 404, vino %d", rr.Code)
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
