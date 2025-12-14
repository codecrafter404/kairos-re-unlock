package common

import "runtime/debug"

func GetVersionInformation() string {
	rev := "no build info available"
	mod := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				rev = s.Value
			}
			if s.Key == "vcs.modified" {
				mod = " (dirty)"
			}
		}
	}
	return rev + mod
}
