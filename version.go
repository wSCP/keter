package main

import "fmt"

var (
	pkgVersion  *version
	packageName string = "keter"
	versionTag  string = "No version tag supplied with compilation"
	versionHash string = "No hash"
	versionDate string = "No date"
)

func newVersion(p, t, h, d string) *version {
	return &version{p, t, h, d}
}

type version struct {
	p, t, h, d string
}

func (v *version) Fmt() string {
	var msg string
	msg = fmt.Sprintf("%s version %s(%s %s)\n", v.p, v.t, v.h, v.d)
	return msg
}
