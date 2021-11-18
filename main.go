// Trivial HTTP <-> MPD gateway
package main

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/fhs/gompd/mpd"
)

//Define a map to implement routing table.
var mux map[string]func(http.ResponseWriter, *http.Request)

// Entry Point
func main() {
	server := http.Server{
		Addr:        ":8888",
		Handler:     &myHandler{},
		ReadTimeout: 5 * time.Second,
	}

	//
	// Setup our routes
	//
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["/next"] = NextHandler
	mux["/play"] = PlayHandler
	mux["/prev"] = PrevHandler
	mux["/stop"] = StopHandler

	//
	// Start the server
	//
	fmt.Printf("http://localhost:8888/\n")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

//
// Handler structure
//
type myHandler struct{}

// ServeHTTP delegates an incoming request to the appropriate handler,
// if one has been setup, otherwise shows the server-status
func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}

	htmlHeader := `
<!DOCTYPE html>
 <html lang="en">
  <meta charset="UTF-8">
  <title>MPD</title>
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <meta http-equiv="refresh" content="3">
 </head>
 <body>
  <h1>mpd</h1>`

	htmlMiddle := ``

	client, err := mpd.Dial("tcp", "localhost:6600")
	if err == nil {
		defer client.Close()

		status, err := client.Status()
		if err == nil {
			song, err := client.CurrentSong()
			if err == nil {
				if status["state"] == "play" {
					htmlMiddle = fmt.Sprintf("Playing %s - %s", html.EscapeString(song["Artist"]), html.EscapeString(song["Title"]))
				} else {
					htmlMiddle = fmt.Sprintf("State: %s", status["state"])
				}
			}
		}
	}

	htmlFooter := `
<p><ul>
<li><a href="/prev">Previous</a></li>
<li><a href="/next">Next</a></li>
<li><a href="/play">Play</a></li>
<li><a href="/stop">Stop</a></li>
</ul></p>
</body></html>`

	// Unhandled - show the defualt
	io.WriteString(w, htmlHeader+htmlMiddle+htmlFooter)
}

// handle returns an MPD client handle
func handle() (*mpd.Client, error) {
	client, err := mpd.Dial("tcp", "localhost:6600")
	return client, err
}

// NextHandler is invoked at /next, and moves to the next track in the playlist.
func NextHandler(w http.ResponseWriter, r *http.Request) {
	client, err := handle()
	if err != nil {
		io.WriteString(w, fmt.Sprintf("error %s", err.Error()))
		return
	}
	defer client.Close()

	client.Next()
	http.Redirect(w, r, "/", http.StatusFound)
}

// PlayHandler is invoked at /play, and starts playing, if stopped.
func PlayHandler(w http.ResponseWriter, r *http.Request) {
	client, err := handle()
	if err != nil {
		io.WriteString(w, fmt.Sprintf("error %s", err.Error()))
		return
	}
	defer client.Close()

	client.Play(-1)
	http.Redirect(w, r, "/", http.StatusFound)
}

// PrevHandler is invoked at /prev, and moves to the previous track in the playlist.
func PrevHandler(w http.ResponseWriter, r *http.Request) {
	client, err := handle()
	if err != nil {
		io.WriteString(w, fmt.Sprintf("error %s", err.Error()))
		return
	}
	defer client.Close()

	client.Previous()
	http.Redirect(w, r, "/", http.StatusFound)
}

// StopHandler is invoked at /stop, and stops playback.
func StopHandler(w http.ResponseWriter, r *http.Request) {
	client, err := handle()
	if err != nil {
		io.WriteString(w, fmt.Sprintf("error %s", err.Error()))
		return
	}
	defer client.Close()

	client.Stop()
	http.Redirect(w, r, "/", http.StatusFound)
}
