package service

import "embed"

type BuildInfo struct {
	Version string
	Date    string
	Time    string
	Commit  string
	Release string

	EmbedFS *embed.FS
}
