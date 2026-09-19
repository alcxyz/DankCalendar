package main

import (
	"runtime/debug"
	"testing"
)

func TestDevelopmentBuildIdentity(t *testing.T) {
	info := &debug.BuildInfo{Settings: []debug.BuildSetting{
		{Key: "vcs.modified", Value: "true"},
		{Key: "vcs.revision", Value: "abcdef1234567890"},
	}}
	if got := resolveBuildVersion("dev", info); got != "dev-abcdef123456-dirty" {
		t.Fatal(got)
	}
	if got := resolveBuildVersion("1.2.3", info); got != "1.2.3" {
		t.Fatal(got)
	}
	if got := resolveBuildVersion("dev", nil); got != "dev" {
		t.Fatal(got)
	}
	info.Settings[1].Value = "private-path"
	if got := resolveBuildVersion("dev", info); got != "dev" {
		t.Fatal(got)
	}
}
