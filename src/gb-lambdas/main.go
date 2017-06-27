package main

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	log.SetFormatter(&log.TextFormatter{})

	if lambdasLogLevel := os.Getenv("GB_LAMBDAS_LOGLEVEL"); "" != lambdasLogLevel {
		if newLogLevel, err := log.ParseLevel(lambdasLogLevel); nil == err {
			log.SetLevel(newLogLevel)
		}
	}

	checkGolangVersion()

	lambdaMap := BuildLambdaMap()

	log.Debugf("lambdaMap: %s", lambdaMap)

	for src, name := range lambdaMap {
		generateZip(src, name)
	}
}
