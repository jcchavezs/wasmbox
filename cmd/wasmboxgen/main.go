package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/base64"
	"flag"
	"io"
	"os"
	"os/exec"
	"text/template"
)

type ConfigMap struct {
	Name      string
	Namespace string
	Content   string
}

var (
	//go:embed configmap.yaml.tmpl
	defaultTmpl string
	tmpl        string
	o           string
)

func main() {
	if len(os.Args) < 2 {
		panic("usage: wasmbox <configmap-name>")
	}

	cf := ConfigMap{}
	flag.StringVar(&cf.Namespace, "namespace", "default", "namespace for the configmap")
	flag.StringVar(&tmpl, "template", "", "template for the configmap")
	flag.StringVar(&o, "o", "/dev/stdout", "output for the generated configmap")
	flag.Parse()

	cf.Name = flag.Arg(0)

	dir := os.TempDir()
	defer os.RemoveAll(dir)

	// flags come from https://www.fermyon.com/blog/optimizing-tinygo-wasm
	buildCMD := exec.Command("tinygo", "build", "-o", dir+"/main.wasm", "-target=wasi", "-scheduler=none", "-gc=leaking", "-no-debug", "main.go")
	buildCMD.Stderr = os.Stderr
	buildCMD.Stdout = os.Stdout
	if err := buildCMD.Run(); err != nil {
		panic(err)
	}

	var (
		cfTmpl *template.Template
		err    error
	)

	if tmpl == "" {
		cfTmpl, err = template.New("configmap").Parse(defaultTmpl)
	} else {
		cfTmpl, err = template.New("configmap").ParseFiles(tmpl)
	}
	if err != nil {
		panic(err)
	}

	var output io.Writer
	switch o {
	case "/dev/stdout":
		output = os.Stdout
	case "/dev/stderr":
		output = os.Stderr
	default:
		output, err = os.Create(o)
		if err != nil {
			panic(err)
		}
	}

	wasmBuf, err := os.ReadFile(dir + "/main.wasm")
	if err != nil {
		panic(err)
	}

	var compressedWatBuf bytes.Buffer
	gz := gzip.NewWriter(&compressedWatBuf)
	if _, err := gz.Write(wasmBuf); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	cf.Content = base64.StdEncoding.EncodeToString(compressedWatBuf.Bytes())

	err = cfTmpl.Execute(output, cf)
	if err != nil {
		panic(err)
	}
}
