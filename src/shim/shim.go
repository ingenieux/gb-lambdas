package shim

import (
	"github.com/GeertJohan/go.rice"
	log "github.com/sirupsen/logrus"
	"strings"
)

//go:generate rice embed-go

// CLI: $ GOPATH=$PWD:$PWD/vendor rice embed-go -i shim
var Assets = rice.MustFindBox("assets")

var GoVersionPrefix string

func init() {
	versionString, err := Assets.String("goversion")

	if nil != err {
		log.Warnf("Oops: %v", err)
		panic(err)
	}

	versionString = strings.TrimSpace(versionString)

	elements := strings.SplitN(versionString, " ", 4)

	GoVersionPrefix = strings.Join(elements[:3], " ") + " "

	// Those two debug statements won't work. When debugging, change to Infof instead

	debugf := log.Debugf

	debugf("versionString from resource: %s (%q)", versionString, elements)

	debugf("final: %s", GoVersionPrefix)
}
