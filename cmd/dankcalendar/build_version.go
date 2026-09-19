package main

import (
	"runtime/debug"
	"strings"
)

func buildVersion() string {
	info, _ := debug.ReadBuildInfo()
	return resolveBuildVersion(version, info)
}

func resolveBuildVersion(configured string, info *debug.BuildInfo) string {
	if configured != "dev" || info == nil {
		return configured
	}
	revision := ""
	dirty := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}
	if len(revision) < 12 {
		return configured
	}
	for _, char := range revision {
		if !strings.ContainsRune("0123456789abcdefABCDEF", char) {
			return configured
		}
	}
	result := "dev-" + strings.ToLower(revision[:12])
	if dirty {
		result += "-dirty"
	}
	return result
}
