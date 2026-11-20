(function (global) {
  'use strict';
  // Display only. The consuming application verifies the identity first.
  function signedInText(username) {
    return 'Signed in as ' + String(username);
  }
  function setText(text) {
    const node = global.document.getElementById('suite-auth-status');
    if (!node) return;
    node.textContent = text;
    node.hidden = false;
  }
  global.SuiteBannerStatus = Object.freeze({
    signedInText,
    signedIn(username) { setText(signedInText(username)); },
    setText
  });
})(window);
