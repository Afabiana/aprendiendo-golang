package main

import (
	"api-libros/models"
	"context"
	"errors"
)

var ErrNotFound = errors.New("libro not found")

type FakeLibrosRepo struct {
	libros map[int]models.Libro
}

func NewFakeLibrosRepo() *FakeLibrosRepo {
	return &FakeLibrosRepo{
		libros: map[int]models.Libro{
			1: {ID: 1, Titulo: "Dune", Autor: "Frank Herbert", Ano: 1965},
		},
	}
}

func (f *FakeLibrosRepo) getAll(ctx context.Context) ([]models.Libro, error){
	res := []models.Libro{}
	for _, l := range f.libros{
		res  = append(res, l)
	}
	return res, nil
}

func (f *FakeLibrosRepo) getById(ctx context.Context, id int) (*models.Libro, error){
	l, ok := f.libros[id]
	if !ok{
		return nil, ErrNotFound
	}
	return &l, nil
}

func (f *FakeLibrosRepo) Create(ctx context.Context, in models.LibroInput)(*models.Libro, error){
	id := len(f.libros) + 1
	l := models.Libro{
		ID: id,
		Titulo: in.Titulo,
		Autor: in.Autor,
		Ano: in.Ano,
	}
	f.libros[id] = l
	return  &l, nil
}