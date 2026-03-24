package main

import "github.com/drossan/http2postman/cmd"

var (
	version = "dev"
	commit  = "none"
)

func main() {
	cmd.SetVersionInfo(version, commit)
	cmd.Execute()
}
