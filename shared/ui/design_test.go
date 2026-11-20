package ui

import (
 "strings"
 "testing"
)

func TestDesignTokensAreOptIn(t *testing.T) {
 css:=string(DesignCSS)
 for _,want:=range []string{".suite-ui", "--suite-font", ".suite-type-title", "data-suite-theme", "tabular-nums"} {
  if !strings.Contains(css,want) {t.Errorf("missing %s",want)}
 }
 for _,forbidden:=range []string{"@import", "body{", "body {", "html{", "html {"} {
  if strings.Contains(css,forbidden) {t.Errorf("unexpected global/network rule %q",forbidden)}
 }
 if len(ButtonCSS)==0 {t.Fatal("legacy ButtonCSS was removed")}
}
