package main

import (
	"html/template"
	"log"
	"os"

	"github.com/hudsn/morph"
	"github.com/hudsn/morph/docgen"
)

func main() {
	generateHTML()
}

func generateHTML() {
	tt, err := template.ParseFS(docgen.DefaultDocTemplates, "*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	templateData := docgen.NewFunctionDocs(morph.DefaultFunctionStore())

	out, err := os.Create("docs/index.html")
	if err != nil {
		os.Exit(1)
	}
	defer out.Close()

	err = tt.ExecuteTemplate(out, "base", templateData)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
