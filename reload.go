package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/websocket"
)

const url = "localhost:3000"

const script = `<script>
var ___ = function() {
  websocket = new WebSocket('ws://` + url + `/ws');
  websocket.onmessage = function(evt) {
    console.log(evt);
    var message = JSON.parse(evt.data);
    if (message.reload) {
      console.log("Reloading. Bye");
      window.location.reload();
    }
  };
}();
</script>`

func main() {
	var ping = make(chan chan bool, 1)
	go monitorFiles(ping)

	wsHandler := func(ws *websocket.Conn) {
		log.Println("Hello ws")
		me := make(chan bool, 1)
		ping <- me
		// Since we send a reload message, this
		// is going to close the connection. So only
		// consume a single message.
		<-me
		log.Println("Sending reload message")
		ws.Write([]byte(`{"reload":true}`))
	}

	http.HandleFunc("/", serveHot)
	http.Handle("/ws", websocket.Handler(wsHandler))
	http.ListenAndServe(url, nil)
}

func monitorFiles(ping chan chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				x: for {
					select {
					case conn := <-ping:
						conn <- true
					default:
						break x
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		log.Println("Watching " + path)
		watcher.Add(path)
		return nil
	})
	<-done
}

func serveHot(w http.ResponseWriter, r *http.Request) {
	f, err := ioutil.ReadFile(r.URL.Path[1:])
	if err != nil {
		http.Error(w, "Can't find "+r.URL.Path, http.StatusNotFound)
		return
	}

	w.Write(f)
	if strings.HasSuffix(r.URL.Path, ".html") {
		w.Write([]byte(script))
	}
}
