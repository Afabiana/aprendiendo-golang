package main

type LibroUpdate struct {
	Titulo *string `json:"titulo"`
	Autor  *string `json:"autor"`
	Ano    *int    `json:"ano"`
}
