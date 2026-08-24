// Package web embeds the server-rendered HTML templates and static assets
// bundled into the Lambda deployment package.
package web

import "embed"

//go:embed templates/*.html
var Templates embed.FS

//go:embed static
var Static embed.FS
