// Trivial HTTP <-> MPD gateway.

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/fhs/gompd/mpd"
)

//go:embed web/index.html
var indexTemplate string

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
	mux["/"] = IndexHandler

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

	// Forward to the handler, if it exists
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}

	// Otherwise we've hit a route we don't know.
	http.Redirect(w, r, "/", http.StatusFound)
}

// IndexHandler shows the current state of the server, and returns
// a basic GUI.
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	// Pagedata is a structure which is used to add
	// dynamic data to our template.
	type Pagedata struct {
		// Name of artist
		Artist string

		// Are we playing?
		Playing bool

		// Song title
		Title string
	}

	// Create an instance of the pagedata to populate our
	// template with.
	x := Pagedata{Artist: "", Title: "", Playing: false}

	// Try to get the currently-playing track, and
	// update our structure with it.
	client, err := mpd.Dial("tcp", "localhost:6600")
	if err == nil {

		// If we can connect
		defer client.Close()

		var status mpd.Attrs
		status, err = client.Status()
		if err == nil {

			// If we got the status of the server
			var song mpd.Attrs
			song, err = client.CurrentSong()
			if err == nil {

				// If we found a current song
				if status["state"] == "play" {

					// Populate details
					x.Artist = song["Artist"]
					x.Title = song["Title"]
					x.Playing = true
				}
			}
		}
	}

	// Parse our (embedded) template.
	t := template.Must(template.New("tmpl").Parse(indexTemplate))

	// Execute the template into a buffer.
	buf := &bytes.Buffer{}
	err = t.Execute(buf, x)

	// If there were errors, then show them.
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// Otherwise serve the result to the client.
	buf.WriteTo(w)
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
