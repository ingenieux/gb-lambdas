package main

import (
	log "github.com/sirupsen/logrus"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

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

			for file := range v.Files {
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
		log.Warnf("Oops: %v", err)

		panic(err)
	}

	return fileList
}
