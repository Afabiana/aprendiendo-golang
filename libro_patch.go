package main

import (
	"errors"
	"strings"
)

type LibroPatch struct {
	Titulo *string `json:"titulo"`
	Autor  *string `json:"autor"`
	Ano    *int    `json:"ano"`
}

func (u *LibroPatch) Validate() error {
	if u.Titulo != nil && strings.TrimSpace(*u.Titulo) == "" {
		return errors.New("titulo vacio")
	}

	if u.Autor != nil && strings.TrimSpace(*u.Autor) == "" {
		return errors.New("Autor vacio")
	}

	if u.Ano != nil && *u.Ano <= 0 {
		return errors.New("año invalido")
	}

	return nil
}
