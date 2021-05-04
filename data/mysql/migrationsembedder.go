package mysql

import (
	"embed"
	"fmt"
	"net/http"
)

//go:embed migrations/*
var migrations embed.FS

const migrationsDir = "migrations"

type embedder func()

var MigrationsEmbedder embedder

func (m embedder) GetDir() http.FileSystem {
	return httpFileSystemRelativePathAdapter{fs: http.FS(migrations)}
}

type httpFileSystemRelativePathAdapter struct {
	fs http.FileSystem
}

func (receiver httpFileSystemRelativePathAdapter) Open(name string) (http.File, error) {
	decoratedPath := fmt.Sprintf("%s%s", migrationsDir, name)
	if name == "." || name == "/" {
		decoratedPath = migrationsDir
	}
	return receiver.fs.Open(decoratedPath)
}
