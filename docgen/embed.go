package docgen

import (
	"embed"
)

//go:embed *.tmpl
var DefaultDocTemplates embed.FS
