package storage

import "errors"

// TODO: В дальнейшем попробовать реализовать данный функционал используя другую СУБД
var (
	ErrURLNotFound = errors.New("url не найден")
	ErrURLExists   = errors.New("url уже существует")
)
