package main

import "fmt"

var (
	packageName string = "marid"
	versionTag  string = "no version"
	versionHash string = "no hash"
	versionDate string = "no date"
)

func fmtVersion() string {
	return fmt.Sprintf("%s - %s(%s %s)", packageName, versionTag, versionHash, versionDate)
}
