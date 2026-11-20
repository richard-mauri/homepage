package ui

import _ "embed"

// DesignCSS defines opt-in suite typography and semantic colors.
// It does not alter the legacy ButtonCSS contract.
//go:embed design.css
var DesignCSS []byte
