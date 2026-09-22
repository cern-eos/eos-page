package web

import "embed"

//go:embed public
var Public embed.FS

//go:embed controller
var Controller embed.FS
