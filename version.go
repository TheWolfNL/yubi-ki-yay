package main

// version is set at build time with -ldflags "-X main.version=..."; see build/darwin/Taskfile.yml.
var version = "dev"

func displayVersion() string {
	if version == "dev" {
		return version
	}
	return "v" + version
}
