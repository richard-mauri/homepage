// Package ui embeds the canonical, build-time bundled suite button CSS.
package ui

import _ "embed"

//go:embed buttons.css
var ButtonCSS []byte
