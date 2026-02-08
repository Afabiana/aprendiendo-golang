package models

import (
	"errors"
)
type LibroFilter struct {
	Autor  *string
	From   *int
	To     *int
	Limit  int
	Offset int
}

//uso punteros para poder distinguir "no vino el filtro" vs "vino vacio"

func (f *LibroFilter) Validate() error {

	if f.Limit < 0 {
		return errors.New("limit no puede ser negativo")
	}

	if f.Offset < 0 {
		return errors.New("offset no puede ser negativo")
	}

	if f.From != nil && f.To != nil && *f.From > *f.To {
		return errors.New("from no puede ser mayor que to")
	}

	return nil
}
