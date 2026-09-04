package binaries

import "embed"

//go:embed all:payload/darwin-amd64
var payloadFS embed.FS

const payloadDir = "payload/darwin-amd64"
