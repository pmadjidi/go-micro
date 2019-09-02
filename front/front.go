package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"html/template"
	"time"
	"github.com/hashicorp/consul/api"
)

func StartListenToEndPoint() {
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	addr := "ws://" + r.Host + "/echo"
	log.Printf("Listening to addrs: %s\n", addr)
	homeTemplate.Execute(w, addr)
}

func echo(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		err = decodeWebSocket(message)
		if err != nil {
			log.Println("Decode:", err)
			c.WriteMessage(mt, []byte(err.Error()))
		} else {
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

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
        ws = new WebSocket("{{.}}");
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

type Envelope struct {
	SourceId string
	T        []int64
	App      string `json:"instance"`
	Cmd      string `json:"cmd"`
	Payload  string `json:"payload"`
}

type Login struct {
	User   string `json:"user"`
	Secret string `json:"secret"`
}

type Logout struct {
	User string `json:"user"`
}

type Delete struct {
	User string `json:"user"`
	Secret string `json:"secret"`
}

type Update struct {
	Givenname    string   `json:"givenname"`
	Sirname   int64    `json:"sirname"`
	Displayname  string   `json:"displayname"`
	Username    string `json:"username"`
	Email string   `json:"email"`
	Location string   `json:"location"`
	Avatar	string   `json:"avatar"`
}

type Register struct {
	Givenname    string   `json:"givenname"`
	Sirname   int64    `json:"sirname"`
	Displayname  string   `json:"displayname"`
	Username    string `json:"username"`
	Email string   `json:"email"`
	Location string   `json:"location"`
	Avatar	string   `json:"avatar"`
}







func stamp() int64 {
	return time.Now().UnixNano()
}

func decodeWebSocket(m []byte) error {
	var env Envelope
	if err := json.Unmarshal(m, &env); err != nil {
		return err
	}
	msg.Stamp = stamp()
	fmt.Printf("Recieved message from WebSocket: %+v\n", env)
	fmt.Printf("Message size is: %+v\n", len(m))
	return decode(&env)
}

func Initkafka() {

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{"onacci", "^aRegex.*[Tt]opic"}, nil)

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	c.Close()
}

func main() {
	var appName string
	flag.StringVar(&appName, "app", "", "name of the app...")
	flag.Parse()

	if appName == "" {
		panic("Application name is mandatory....")
	}
	StartListenToEndPoint()
}
