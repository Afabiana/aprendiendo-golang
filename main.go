package main

import (
	"fmt"
	"log"
	"net/http"
	"api-libros/db"
	"api-libros/handlers"
)


func main() {
	database := db.New()
	defer database.Close()

	librosHandler := &handlers.LibrosHandler{ //uso puntero porque el handler es un servicio y puede llevar metricas globales que no me serviria copiar
		DB : database,
	}

	http.HandleFunc("/libros", librosHandler.Libros)
	http.HandleFunc("/libros/", librosHandler.LibrosByID)

	fmt.Println("Servidor REST corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
