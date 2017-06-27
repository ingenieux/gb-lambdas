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
	versionString, err := Assets.String("version")

	if nil != err {
		log.Warnf("Oops: %v", err)
		panic(err)
	}

	elements := strings.SplitN(versionString, " ", 3)

	GoVersionPrefix = strings.Join(elements, " ") + " "
}
