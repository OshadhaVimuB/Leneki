package binaries

import "embed"

//go:embed all:payload/windows-amd64
var payloadFS embed.FS

const payloadDir = "payload/windows-amd64"
