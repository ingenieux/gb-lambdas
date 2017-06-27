package shim

import "github.com/GeertJohan/go.rice"

//go:generate rice embed-go

// CLI: $ GOPATH=$PWD:$PWD/vendor rice embed-go -i shim
var Assets = rice.MustFindBox("assets")
