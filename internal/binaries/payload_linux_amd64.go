package binaries

import "embed"

//go:embed all:payload/linux-amd64
var payloadFS embed.FS

const payloadDir = "payload/linux-amd64"
