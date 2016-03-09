package main

import "fmt"

func newVersion(p, t, h, d string) *version {
	return &version{p, t, h, d}
}

type version struct {
	p, t, h, d string
}

func (v *version) Fmt() string {
	var msg string
	if v.h != "" && v.d != "" {
		msg = fmt.Sprintf("%s version %s(%s %s)\n", v.p, v.t, v.h, v.d)
	} else {
		msg = fmt.Sprintf("%s %s\n", v.p, v.t)
	}
	return msg
}
