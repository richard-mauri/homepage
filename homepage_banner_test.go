package main

import (
 "bytes"
 "strings"
 "testing"
)

func TestHomepageUsesSharedBannerWithoutChangingPortal(t *testing.T) {
 source,err:=assets.ReadFile("static/index.html")
 if err!=nil {t.Fatal(err)}
 page:=renderHomepage(source)
 for _,want:=range []string{
  `class="suite-ui suite-banner"`,
  `id="suite-auth-status"`,
  `src="/suite-ui/tanner.jpg"`,
  `href="/suite-ui/banner.css?v=`,
  `class="suite-ambient suite-ambient-one"`,
  `class="suite-ambient suite-ambient-two"`,
  `class="suite-toolbar"`,
  `class="topright"`,
  `id="clock"`,
  `id="themeToggle"`,
  `One place for`,
  `id="appsHeading"`,
 } {
  if !strings.Contains(page,want) {t.Errorf("Homepage missing %q",want)}
 }
 if strings.Contains(page,`<header class="topbar">`) {t.Fatal("legacy header retained in rendered Homepage")}
 css,err:=assets.ReadFile("static/styles.css")
 if err!=nil {t.Fatal(err)}
 for _,obsolete:=range []string{".topbar{", ".ambient-one{", ".ambient-two{"} {
  if strings.Contains(string(css),obsolete) {t.Errorf("obsolete Homepage styling retained: %s",obsolete)}
 }
 if strings.Count(page,`id="suite-auth-status"`)!=1 {t.Fatal("expected exactly one account identity status")}
 if !strings.Contains(page,`</div><div class="shell">`) || !strings.Contains(page,`<div class="homepage-shared-chrome"><header class="suite-ui suite-banner"`) {t.Fatal("shared banner must mount outside Homepage's app-specific shell")}
 if strings.Contains(page,"Build Build") {t.Fatal("duplicated build label")}
}

func TestTannerImageReuse(t *testing.T) {
 jpeg,err:=homepageTannerPhoto()
 if err!=nil {t.Fatal(err)}
 if len(jpeg)<1000 || !bytes.HasPrefix(jpeg,[]byte{0xff,0xd8,0xff}) {t.Fatal("Tanner source is not the expected JPEG")}
}

func TestHomepageAccountPresentationUsesVerifiedState(t *testing.T) {
 script,err:=assets.ReadFile("static/app.js")
 if err!=nil {t.Fatal(err)}
 js:=string(script)
 for _,want:=range []string{
  `document.querySelector('#suite-auth-status')`,
  `bannerIdentity.textContent='Account status unavailable'`,
  `bannerIdentity.textContent='Browser-only mode'`,
  `SuiteBannerStatus.signedIn(state.session.user)`,
  `signedIn.append(manage)`,
 } {
  if !strings.Contains(js,want) {t.Errorf("missing account-state contract %q",want)}
 }
}
