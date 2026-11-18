package main

import (
 "strings"
 "testing"

 suiteui "github.com/richard-mauri/homepage/shared/ui"
)

func TestSuiteAdminUsesCanonicalButtons(t *testing.T) {
 page,err:=assets.ReadFile("static/suite-admin.html")
 if err!=nil {t.Fatal(err)}
 html:=string(page)
 for _,required:=range []string{`href="/suite-ui/buttons.css"`,`body class="suite-admin"`,`--soft:#203c45`,`--accent:#7ce3d7`} {
  if !strings.Contains(html,required) {t.Errorf("missing shared button integration %q",required)}
 }
 for _,obsolete:=range []string{`button{cursor:pointer;background:#194d53`,`button:hover{background:#28646b}`} {
  if strings.Contains(html,obsolete) {t.Errorf("obsolete button override %q",obsolete)}
 }
 css:=string(suiteui.ButtonCSS)
 for _,required:=range []string{"button.ui-button", "body.suite-admin button", ":focus-visible", ":disabled", ":hover", "prefers-reduced-motion"} {
  if !strings.Contains(css,required) {t.Errorf("shared CSS missing %q",required)}
 }
 if !strings.Contains(adminPageCSP,"style-src 'self' 'unsafe-inline'") {t.Error("same-origin CSS blocked by admin CSP")}
}

func TestSuiteButtonCSSServedByHomepage(t *testing.T) {
 // Keep the implementation embedded and self-contained, with no remote fetches.
 if len(suiteui.ButtonCSS)==0 {t.Fatal("embedded canonical CSS is empty")}
 if strings.Contains(string(suiteui.ButtonCSS),"@import") {t.Fatal("shared CSS must not load remote styles")}
}
