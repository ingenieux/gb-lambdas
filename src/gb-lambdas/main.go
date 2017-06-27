package main

import (
	"archive/zip"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"shim"
	"strings"
)

var filesToCopy = []string{"__init__.pyc", "proxy.pyc", "runtime.so"}

func main() {
	lambdaMap := BuildLambdaMap()

	log.Debugf("lambdaMap: %s", lambdaMap)

	for src, name := range lambdaMap {
		generateZip(src, name)
	}

}

func generateZip(src string, name string) {
	goBinary, err := exec.LookPath("go")

	if nil != err {
		panic(err)
	}

	fileList := buildFileList(src)

	buildArgs := make([]string, 0)

	buildArgs = append(buildArgs, "build", "-buildmode=plugin", `-ldflags=-w -s`, "-o", "pkg/"+name+".so", "-tags", `'+lambda'`)

	buildArgs = append(buildArgs, fileList...)

	buildCmd := exec.Command(goBinary, buildArgs...)

	buildCmd.Dir = os.Getenv("GB_PROJECT_DIR")

	log.Debugf("Project Dir: %s", buildCmd.Dir)

	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Stdin = os.Stdin

	buildCmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"GOROOT=" + os.Getenv("GOROOT"),
		"GOPATH=" + fmt.Sprintf("%s:%s/vendor", buildCmd.Dir, buildCmd.Dir),
	}

	log.Debugf("Running: %s with env: %s", buildCmd.Args, buildCmd.Env)

	err = buildCmd.Run()

	if nil != err {
		log.Warnf("err: %v", err)

		panic(err)
	}

	outputFile, err := os.OpenFile("bin/"+name+".zip", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0664))

	if nil != err {
		log.Warnf("err: %v", err)

		panic(err)
	}

	zipWriter := zip.NewWriter(outputFile)

	for _, f := range filesToCopy {
		obj, err := shim.Assets.Open(f)

		if nil != err {
			log.Warnf("err: %v", err)

			panic(err)
		}

		stat, err := obj.Stat()

		if nil != err {
			log.Warnf("err: %v", err)

			panic(err)
		}

		fileInfoHeader, err := zip.FileInfoHeader(stat)

		if nil != err {
			log.Warnf("err: %v", err)

			panic(err)
		}

		fileInfoHeader.Name = "handler/" + fileInfoHeader.Name

		fileWriter, err := zipWriter.CreateHeader(fileInfoHeader)

		if nil != err {
			log.Warnf("err: %v", err)

			panic(err)
		}

		_, err = io.Copy(fileWriter, obj)

		if nil != err {
			log.Warnf("err: %v", err)

			panic(err)
		}
	}

	stat, err := os.Stat("pkg/" + name + ".so")

	if nil != err {
		log.Warnf("err: %v", err)

		panic(err)
	}

	fileInfoHeader, err := zip.FileInfoHeader(stat)

	if nil != err {
		log.Warnf("err: %v", err)

		panic(err)
	}

	fileInfoHeader.Name = "handler.so"

	fileWriter, err := zipWriter.CreateHeader(fileInfoHeader)

	if nil != err {
		log.Warnf("err: %v", err)

		panic(err)
	}

	soFileReader, err := os.OpenFile("pkg/"+name+".so", os.O_RDONLY, os.FileMode(0x660))

	if nil != err {
		log.Warnf("err: %v", err)

		panic(err)
	}

	_, err = io.Copy(fileWriter, soFileReader)

	if nil != err {
		log.Warnf("err: %v", err)

		panic(err)
	}

	zipWriter.Close()
}

func BuildLambdaMap() map[string]string {
	lambdaMap := make(map[string]string, 0)

	filepath.Walk("src", func(path string, info os.FileInfo, err error) error {
		if _, ok := lambdaMap[path]; ok {
			return nil
		}

		fset := token.NewFileSet()

		pkgs, err := parser.ParseDir(fset, path, nil, 0)

		if nil != err {
			return nil
		}

	OUTER:
		for packageName, v := range pkgs {

			log.Debugf("Processing package %s", packageName)

			for file, _ := range v.Files {
				curFset := token.NewFileSet()

				ast, err := parser.ParseFile(curFset, file, nil, 0)

				if nil != err {
					return nil
				}

				for _, i := range ast.Imports {
					pathValue := strings.Trim(i.Path.Value, `"`)
					if strings.HasPrefix(pathValue, "github.com/eawsy/aws-lambda-go-core") {
						key := path
						value := filepath.Base(path)

						lambdaMap[key] = value
						continue OUTER
					}
				}
			}
		}

		return nil
	})

	return lambdaMap
}

func buildFileList(path string) []string {
	fileList, err := filepath.Glob(path + "/*.go")

	if nil != err {
		panic(err)
	}

	return fileList
}
