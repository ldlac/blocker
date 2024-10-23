package main

import (
	_ "github.com/coredns/coredns/core/plugin"
	_ "github.com/ldlac/blocker"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/coremain"
)

func init() {
	dnsserver.Directives = append(
		[]string{"log", "blocker"},
		dnsserver.Directives...,
	)
}

func main() {
	coremain.Run()
}