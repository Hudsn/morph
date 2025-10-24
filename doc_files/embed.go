package docfiles

import (
	"embed"
)

//go:embed *.tmpl
var DocFS embed.FS
