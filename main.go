package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"regexp"
)

const makefileTemplate = `.DEFAULT_GOAL := help

BIN = $(CURDIR)/bin
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo v0)

$(BIN):
	@mkdir -p $@

.PHONY:phony

fmt: phony ## format the codes
	@go fmt ./...

lint: phony fmt ## lint the codes
	@golint ./...

vet: phony lint ## vet the codes
	@go vet ./...
{{- if .shadow}}	@shadow ./...{{end}}

{{ if not .library}}
build: phony vet | $(BIN) ## build the binary
	@go build \
		-tags release \
		-ldflags '-X main.Version=$(VERSION)' \
		-o $(BIN)/ ./...

run: phony vet ## run the binary
	@go run main.go
{{ else}}
build: phony vet ## build the library
	@go build ./...
{{end}}

clean: phony
	rm -rf $(BIN)

{{- if .test}}
test: phony vet ## test the codes
	@go test -v ./...
{{ end }}

{{- if .bench}}
bench: phony vet ## test with benchmarks
	@go test -v -bench=. -benchmem ./...
{{ end }}

{{- if and .test .cover}}
test-cover: phony vet ## test with coverage
	@go test -v -cover ./...
{{ end }}

{{- if and .test .coverHTML}}
test-cover-html: phony vet ## test with coverage in an HTML view
	@go test -v -cover -coverprofile=c.out ./...
	@go tool cover -html=c.out
{{ end }}

{{- if .testRace}}
test-race: phony vet ## test and check for race conditions
	@go test -race ./...
{{ end }}

{{- if .race}}
build-race: phony vet ## build and check for race conditions
	@go build -race
{{ end }}

{{- if .cpuProfile}}
test-cpu: phony vet ## test and profile CPU
	@go test {{if .bench}}-bench=. -benchmem{{end}} -cpuprofile cpu.out ./...
	@go tool pprof cpu.out
{{ end }}

{{- if .memProfile}}
test-mem: phony vet ## test and profile memory
	@go test {{if .bench}}-bench=. -benchmem{{end}} -memprofile mem.out ./...
	@go tool pprof mem.out
{{ end }}

GREEN  := $(shell tput -Txterm setaf 2)
RESET  := $(shell tput -Txterm sgr0)

help: phony ## print this help message
	@awk -F ':|##' '/^[^\t].+?:.*?##/ { printf "${GREEN}%-20s${RESET}%s\n", $$1, $$NF }' $(MAKEFILE_LIST)
`

// Version is the version of the binary. This is set by -ldflags during the build.
var Version = "dev"

func main() {
	t := flag.Bool("test", false, "Adds test to makefile")
	b := flag.Bool("bench", false, "Adds bench to makefile")
	s := flag.Bool("shadow", false, "Adds shadow to makefile")
	c := flag.Bool("cover", false, "Adds cover to makefile")
	ch := flag.Bool("coverHTML", false, "Adds cover HTML to makefile")
	cp := flag.Bool("cpuProfile", false, "Adds CPU profiling to makefile")
	mp := flag.Bool("memProfile", false, "Adds Memory profiling to makefile")
	r := flag.Bool("race", false, "Adds race checking to makefile")
	tr := flag.Bool("testRace", false, "Adds race checking tests to makefile")
	l := flag.Bool("library", false, "Creates a library makefile")
	m := flag.String("mod", "", "Creates a mod file. Specify the source control path (github.com/user/project).")
	v := flag.Bool("version", false, "Displays the version of this binary")

	flag.Parse()

	if *v {
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	if len(flag.Args()) != 1 {
		fmt.Println("Expected use: maker DIRNAME")
		os.Exit(1)
	}
	dirName := flag.Arg(0)

	templ := template.Must(template.New("makefile").Parse(makefileTemplate))

	var buffer bytes.Buffer
	err := templ.Execute(&buffer, map[string]interface{}{
		"test":       *t,
		"bench":      *b,
		"shadow":     *s,
		"cover":      *c,
		"coverHTML":  *ch,
		"cpuProfile": *cp,
		"memProfile": *mp,
		"race":       *r,
		"testRace":   *tr,
		"library":    *l,
	})
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(dirName, os.ModePerm)
	if err != nil {
		panic(err)
	}
	regex, err := regexp.Compile("\n\n+")
	if err != nil {
		panic(err)
	}
	cleanBuf := regex.ReplaceAll(buffer.Bytes(), []byte("\n\n"))
	err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"Makefile", cleanBuf, 0744)
	if err != nil {
		panic(err)
	}
	if !(*l) {
		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"main.go", []byte(`package main

func main() {
}
`), 0744)
	} else {
		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+dirName+".go", []byte("package "+dirName+"\n"), 0744)
	}
	if err != nil {
		panic(err)
	}
	if *m != "" {
		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"go.mod", []byte(fmt.Sprintf(`module %s

go 1.14
`, *m)), 0744)
		if err != nil {
			panic(err)
		}
	}
	err = ioutil.WriteFile(dirName+string(os.PathSeparator)+".gitignore", []byte(`bin/`), 0644)
	if err != nil {
		panic(err)
	}
}
