package main

import (
	"bytes"
	"embed"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/bundler"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/pb33f/libopenapi/index"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

//go:embed specs/*.yaml
var apiYamls embed.FS

func main() {
	spec, _ := filepath.Abs(filepath.Join("specs", "main.yaml"))
	specBytes, _ := os.ReadFile(spec)

	realFSdoc := merged(specBytes, "specs", nil)
	save("merged-real-fs.yaml", realFSdoc)

	lfs, err := index.NewLocalFSWithConfig(&index.LocalFSConfig{
		BaseDirectory: "/",
		DirFS:         apiYamls,
	})
	if err != nil {
		panic(err)
	}
	embedFSdoc := merged(specBytes, "/specs", lfs)
	save("merged-embed-fs.yaml", embedFSdoc)

	if bytes.Compare(realFSdoc, embedFSdoc) != 0 {
		panic("mismatch")
	}
}

func save(name string, data []byte) {
	err := os.WriteFile(name, data, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func merged(specBytes []byte, basePath string, lfs fs.FS) []byte {
	doc, err := libopenapi.NewDocumentWithConfiguration(specBytes, &datamodel.DocumentConfiguration{
		BasePath:                basePath,
		LocalFS:                 lfs,
		ExtractRefsSequentially: true,
		Logger:                  slog.Default(),
	})
	if err != nil {
		panic(err)
	}

	v3Doc, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		panic(errs)
	}

	bytes, err := bundler.BundleDocument(&v3Doc.Model)
	if err != nil {
		panic(err)
	}
	return bytes
}
