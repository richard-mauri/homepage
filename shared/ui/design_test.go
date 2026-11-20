package ui

import (
 "regexp"
 "strings"
 "testing"
)

func TestDesignTokensAreOptIn(t *testing.T) {
 css:=string(DesignCSS)
 for _,want:=range []string{".suite-ui", "--suite-font", ".suite-type-title", "data-suite-theme", "prefers-color-scheme: light", "tabular-nums", ".suite-ambient-background", "radial-gradient", "background-attachment: fixed"} {
  if !strings.Contains(css,want) {t.Errorf("missing %s",want)}
 }
 if strings.Contains(css,"@import") {t.Error("unexpected runtime stylesheet import")}
 if regexp.MustCompile(`(?m)^\s*(body|html)\s*\{`).MatchString(css) {t.Error("unexpected global element selector")}
 if len(ButtonCSS)==0 {t.Fatal("legacy ButtonCSS was removed")}
}
