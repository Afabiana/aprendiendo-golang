# 📚 API REST de Libros (Go puro)

API REST para gestionar una colección de libros, desarrollada en **Go usando net/http**, sin frameworks externos.
El objetivo del proyecto es aprender y aplicar correctamente HTTP, REST y buenas prácticas desde Go “crudo”.

---

## 🚀 Cómo ejecutar

```bash
go run main.go
```

Servidor disponible en:

```
http://localhost:8080
```

---

## 📌 Modelo de datos

```json
{
  "id": 1,
  "titulo": "Dune",
  "autor": "Frank Herbert",
  "ano": 1965
}
```

---

## 📖 Endpoints

### 🔹 Obtener todos los libros

**Request**

```bash
curl http://localhost:8080/libros
```

**Response**

```json
[
  {
    "id": 1,
    "titulo": "Clean Code (2nd Edition)",
    "autor": "Robert C. Martin",
    "ano": 2021
  },
  {
    "id": 4,
    "titulo": "Clean Code",
    "autor": "Robert C. Martin",
    "ano": 2008
  },
  {
    "id": 5,
    "titulo": "Dune",
    "autor": "Frank Herbert",
    "ano": 1965
  }
]
```

---

### 🔹 Obtener un libro por ID

**Request**

```bash
curl http://localhost:8080/libros/5
```

**Response**

```json
{
  "id": 5,
  "titulo": "Dune",
  "autor": "Frank Herbert",
  "ano": 1965
}
```

---

### 🔹 Crear un libro

**Request**

```bash
curl -X POST http://localhost:8080/libros \
  -H "Content-Type: application/json" \
  -d '{
    "titulo": "Crimen y castigo",
    "autor": "Fiódor Dostoievski",
    "ano": 1866
  }'
```

**Response**

```json
{
  "id": 12,
  "titulo": "Crimen y castigo",
  "autor": "Fiódor Dostoievski",
  "ano": 1866
}
```

---

### 🔹 Reemplazar un libro completo (PUT)

**Request**

```bash
curl -X PUT http://localhost:8080/libros/1 \
  -H "Content-Type: application/json" \
  -d '{
    "titulo": "El Principito Edición Especial",
    "autor": "Antoine de Saint-Exupéry",
    "ano": 1943
  }'
```

**Response**

```json
{
  "id": 1,
  "titulo": "El Principito Edición Especial",
  "autor": "Antoine de Saint-Exupéry",
  "ano": 1943
}
```

---

### 🔹 Actualización parcial de un libro (PATCH)

**Request**

```bash
curl -X PATCH http://localhost:8080/libros/1 \
  -H "Content-Type: application/json" \
  -d '{
    "titulo": "El Principito Versión Final"
  }'
```

**Response**

```json
{
  "id": 1,
  "titulo": "El Principito Versión Final",
  "autor": "Antoine de Saint-Exupéry",
  "ano": 1943
}
```

---

### 🔹 Eliminar un libro

**Request**

```bash
curl -X DELETE http://localhost:8080/libros/11
```

**Response**

```
204 No Content
```

---

## ⚠️ Manejo de errores

Las respuestas de error se devuelven en formato JSON:

```json
{
  "error": "mensaje descriptivo del error"
}
```

Se utilizan códigos HTTP adecuados (`400`, `404`, `405`, `500`).

---

## 🛠️ Tecnologías usadas

- Go (`net/http`)
- `encoding/json`
- PostgreSQL
- `curl` para pruebas manuales

Sin frameworks externos.

---

## 🎯 Objetivo del proyecto

- Comprender HTTP y REST desde la base
- Implementar handlers claros y consistentes
- Separar responsabilidades (decode, validate, respond)
- Construir una base sólida antes de usar frameworks

---

## 📝 Próximos pasos

- Tests automatizados
- Logging
- Middleware
- Router externo (chi / gin) cuando se termine la versión cruda
