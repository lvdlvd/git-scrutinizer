//
// Including this script ensures that $.ajax() includes the X-XSRF-TOKEN header
// set from the XSRF-TOKEN cookie as managed by auth.go.
//
function getToken() {
  var key = "XSRF-TOKEN=";
  var ca = document.cookie.split(';');
  for(var i=0; i<ca.length; i++) {
      var c = ca[i];
      while (c.charAt(0)==' ')
        c = c.substring(1);
      if (c.indexOf(key) == 0)
        return c.substring(key.length, c.length);
  }
  return "";
};

$.ajaxPrefilter(function(options, originalOptions, jqXHR) {
  var t = options['type'];
  if (t === "POST" || t === "PUT" || t === "DELETE") {
    jqXHR.setRequestHeader("X-XSRF-TOKEN", getToken());
  }
});
