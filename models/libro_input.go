package models

import (
	"errors"
	"strings"
)

type LibroInput struct {
	Titulo string `json:"titulo"`
	Autor  string `json:"autor"`
	Ano    int    `json:"ano"`
}

// aca no chequeo si es nil porque no uso punteros
func (l LibroInput) Validate() error {
	if strings.TrimSpace(l.Titulo) == "" {
		return errors.New("titulo requerido")
	}
	if strings.TrimSpace(l.Autor) == "" {
		return errors.New("autor requerido")
	}
	if l.Ano <= 0 {
		return errors.New("año inválido")
	}
	return nil
}
