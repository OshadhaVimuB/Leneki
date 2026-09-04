package binaries

import "embed"

//go:embed all:payload/darwin-arm64
var payloadFS embed.FS

const payloadDir = "payload/darwin-arm64"
