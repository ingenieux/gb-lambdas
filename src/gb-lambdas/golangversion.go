package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"shim"
	"strings"
)

func checkGolangVersion() {
	goBinary, err := exec.LookPath("go")

	if nil != err {
		log.Warnf("Oops: %v", err)

		panic(err)
	}

	outputBuf := new(bytes.Buffer)

	versionCommand := exec.Command(goBinary, "version")

	versionCommand.Stdout = outputBuf

	log.Debugf("Running command: %+v", versionCommand)

	err = versionCommand.Run()

	if nil != err {
		log.Warnf("Oops: %v", err)

		panic(err)
	}

	versionString := strings.TrimSpace(outputBuf.String())

	log.Debugf("Current golang version string: %s (expected: %s)", versionString, shim.GoVersionPrefix)

	if !strings.HasPrefix(versionString, shim.GoVersionPrefix) {
		panic(fmt.Errorf("Incompatible golang version: Current: %s Expected: %s", versionString, shim.GoVersionPrefix))
	}
}
