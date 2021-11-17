// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"
	"flag"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/tools/present"
)

//go:embed templates/*.tmpl
var templates embed.FS

func init() {
	initTemplates(".")
}

var (
	in  = flag.String("in", "", "The `.slide` to be processed.")
	out = flag.String("out", "", "The file to output the rendered slide to.")
)

func main() {
	flag.Parse()
	contentPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if out == nil || *out == "" {
		input := *in
		output := input[:len(input)-len(filepath.Ext(input))] + ".html"
		out = &output
	}

	out := filepath.Join(contentPath, *out)
	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	name := filepath.Join(contentPath, *in)
	if err := renderDoc(f, name); err != nil {
		panic(err)
	}
}

var (
	// contentTemplate maps the presentable file extensions to the
	// template to be executed.
	contentTemplate map[string]*template.Template
)

func initTemplates(base string) error {
	// Locate the template file.
	actionTmpl := filepath.Join(base, "templates/action.tmpl")

	contentTemplate = make(map[string]*template.Template)

	for ext, contentTmpl := range map[string]string{
		".slide":   "slides.tmpl",
		".article": "article.tmpl",
	} {
		contentTmpl = filepath.Join(base, "templates", contentTmpl)

		// Read and parse the input.
		tmpl := present.Template()
		tmpl = tmpl.Funcs(template.FuncMap{"playable": playable})
		if _, err := tmpl.ParseFS(templates, actionTmpl, contentTmpl); err != nil {
			return err
		}
		contentTemplate[ext] = tmpl
	}
	return nil
}

// renderDoc reads the present file, gets its template representation,
// and executes the template, sending output to w.
func renderDoc(w io.Writer, docFile string) error {
	// Read the input and build the doc structure.
	doc, err := parse(docFile, 0)
	if err != nil {
		return err
	}

	// Find which template should be executed.
	tmpl := contentTemplate[filepath.Ext(docFile)]

	// Execute the template.
	return doc.Render(w, tmpl)
}

func parse(name string, mode present.ParseMode) (*present.Doc, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return present.Parse(f, name, mode)
}

func playable(c present.Code) bool {
	play := present.PlayEnabled && c.Play

	// Restrict playable files to only Go source files when using play.golang.org,
	// since there is no method to execute shell scripts there.
	if true {
		return play && c.Ext == ".go"
	}
	return play
}
