package main

import (
	"fmt"
	"log"
	"net/http"
)


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
