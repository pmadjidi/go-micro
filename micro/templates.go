package main

import (
	"html/template"
)

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.WsAddr}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))

var passwordTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
</head>
<body>
<div>
  <form method="POST" action="/auth">   
      <p>
      <label>UserName</label><input name="uname" type="text" value="payam" />
      </p>
      <p>
      <label>Password</label><input name="password" type="text" value="simsalabim" />
      </p>
      <input type="submit" value="submit" />
  </form>
</div>
</body>
</html>
`))

var registerPersonTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
<style> 
input[type=text] {
  width: 100%;
  padding: 12px 20px;
  margin: 8px 0;
  box-sizing: border-box;
}
</style>
</head>
<body>
<div>
<h1>{{.Title}}</h1>
  <form method="POST" action={{.Action}}>   
      <p>
      <label>UserName</label><input name="uname" type="text" value="payam" />
      </p>
      <p>
      <label>Password</label><input name="password" type="text" value="simsalabim" />
      </p>
       <p>
      <label>DisplayName</label><input name="dname" type="text" value="payam" />
      </p>
		<p>
      <label>Email</label><input name="email" type="text" value="payam@gmail.com" />
      </p>
		<p>
      <label>GivenName</label><input name="gname" type="text" value="payam" />
      </p>
		<p>
      <label>SirName</label><input name="sname" type="text" value="payam" />
      </p>
		<p>
      <label>Avatar</label><input name="avatar" type="text" value="" />
      </p>

      <input type="submit" value="submit" />
  </form>
</div>
</body>
</html>
`))


var regSlackSalesForceIntegrationTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
<style> 
input[type=text] {
  width: 100%;
  padding: 12px 20px;
  margin: 8px 0;
  box-sizing: border-box;
}
</style>
</head>
<body>
<div>
<h1>{{.Title}}</h1>
  <form method="POST" action={{.Action}}>   
      <p>
      <label>UserName</label><input name="uname" type="text" value="payam" />
      </p>
      <p>
      <label>Password</label><input name="password" type="text" value="simsalabim" />
      </p>
       <p>
      <label>DisplayName</label><input name="dname" type="text" value="payam" />
      </p>
		<p>
      <label>Slack API Key</label><input name="SlackApiKey" type="text" value="Slack API Key" />
      </p>
		<p>
      <label>SalesForce Client Id</label><input name="SFClientId" type="text" value="SF client id" />
      </p>
		<p>
      <label>SalesForce Client Secret</label><input name="SFclientSecret" type="text" value="SF client secret" />
      </p>
		<p>
      <label>SalesForce User Name</label><input name="SFusername" type="text" value="SF user name" />
      </p>
	<p>
      <label>SalesForce User Password</label><input name="SFuserpass" type="text" value="SF user password" />
      </p>

		<p>
      <input type="submit" value="submit" />
  </form>
</div>
</body>
</html>
`))

