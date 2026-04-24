package storage

import "errors"

// TODO: В дальнейшем попробовать реализовать данный функционал используя другую СУБД
var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
)
