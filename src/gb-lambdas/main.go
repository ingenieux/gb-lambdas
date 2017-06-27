package main

import (
	"archive/zip"
	"errors"
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
	"time"
)

//2013 02 21 0800
var (
	fixedDate = time.Date(2013, 02, 21, 8, 00, 00, 00, time.UTC)

	filesToCopy = []string{"__init__.pyc", "proxy.pyc", "runtime.so"}

	EInvalidProjectDir = errors.New("Invalid Project Dir (variable: $GB_PROJECT_DIR)")
	EInvalidGoRoot     = errors.New("Invalid GOROOT (variable: $GOROOT)")
)

func main() {
	log.SetFormatter(&log.TextFormatter{})

	if lambdasLogLevel := os.Getenv("GB_LAMBDAS_LOGLEVEL"); "" != lambdasLogLevel {
		if newLogLevel, err := log.ParseLevel(lambdasLogLevel); nil == err {
			log.SetLevel(newLogLevel)
		}
	}

	lambdaMap := BuildLambdaMap()

	log.Debugf("lambdaMap: %s", lambdaMap)

	for src, name := range lambdaMap {
		generateZip(src, name)
	}

}

func generateZip(src string, name string) {
	log.Infof("Package: %s @ %s", name, src)

	goBinary, err := exec.LookPath("go")

	if nil != err {
		log.Warnf("Oops: %v", err)

		panic(err)
	}

	goRoot := os.Getenv("GOROOT")

	if "" == goRoot {
		log.Warnf("Oops: %v", EInvalidGoRoot)

		panic(EInvalidGoRoot)
	}

	projectDir := os.Getenv("GB_PROJECT_DIR")

	if "" == projectDir {
		log.Warnf("Oops: %v", EInvalidProjectDir)

		panic(EInvalidProjectDir)
	}

	log.Debugf("projectDir: %s", projectDir)

	projectDir, err = filepath.Abs(projectDir)

	if nil != err {
		log.Warnf("Oops: %v", err)

		panic(err)
	}

	goPath := make([]string, 0)

	for keyToAppend, childPath := range map[string]string{".": "src", "vendor": "vendor/src"} {
		pathToTest := filepath.Join(projectDir, childPath)
		pathToAppend, _ := filepath.Abs(filepath.Join(projectDir, keyToAppend))

		if stat, err := os.Stat(pathToTest); nil == err && stat.IsDir() {
			log.Debugf("Appending path: %s", pathToAppend)

			goPath = append(goPath, pathToAppend)
		} else {
			log.Debugf("Invalid path: %s (stat: %+v reason: %v)", pathToTest, stat, err)
		}
	}

	goPathAsString := strings.Join(goPath, string(os.PathListSeparator))

	log.Debugf("Using gopath: %s", goPathAsString)

	fileList := buildFileList(src)

	buildArgs := make([]string, 0)

	buildArgs = append(buildArgs, "build", "-buildmode=plugin", `-ldflags=-w -s`, "-o", "pkg/"+name+".so", "-tags", `'+lambda'`)

	buildArgs = append(buildArgs, fileList...)

	buildCmd := exec.Command(goBinary, buildArgs...)

	buildCmd.Dir = projectDir

	log.Debugf("Project Dir: %s", buildCmd.Dir)

	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Stdin = os.Stdin

	buildCmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"GOROOT=" + goRoot,
		"GOPATH=" + goPathAsString,
		"GOOS=linux",
		"GOARCH=amd64",
	}

	log.Debugf("Running: %s with env: %s", buildCmd.Args, buildCmd.Env)

	err = buildCmd.Run()

	if nil != err {
		log.Warnf("err: %v", err)

		panic(err)
	}

	binDir := filepath.Join(projectDir, "bin")

	if stat, err := os.Stat(binDir); nil != err && os.IsNotExist(err) {
		log.Debugf("Creating directory %s", binDir)

		err = os.MkdirAll(binDir, 0777)

		if nil != err {
			log.Warnf("Oops: %v", err)

			panic(err)
		}
	} else if stat.IsDir() {
		err = fmt.Errorf("Invalid Binary Dir not a dir: %s", binDir)

		log.Warnf("Oops: %v", err)

		panic(err)
	}

	zipFile := filepath.Join(binDir, name+".zip")

	outputFile, err := os.OpenFile(zipFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0664))

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

		fileInfoHeader.SetModTime(fixedDate)

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
	fileInfoHeader.SetModTime(fixedDate)

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

	log.Infof("Created zip file '%s'", zipFile)
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
		log.Warnf("Oops: %v", err)

		panic(err)
	}

	return fileList
}
