// Trivial HTTP <-> MPD gateway.
package main

import (
	"bytes"
	"fmt"
	"html/template"
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

	src := `<!DOCTYPE html>
 <html lang="en">
  <meta charset="UTF-8">
  <title>MPD</title>
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <meta http-equiv="refresh" content="3">
 </head>
 <body>
  <h1>mpd</h1>
{{if .Playing }}
   <p>Currently playing {{.Artist}} {{.Title}}.</p>
{{end}}
  <p><ul>
   <li><a href="/prev">Previous</a></li>
   <li><a href="/next">Next</a></li>
   {{if .Playing}}
     <li>Play</li>
     <li><a href="/stop">Stop</a></li>
   {{else}}
     <li><a href="/play">Play</a></li>
     <li>Stop</li>
   {{end}}
  </ul></p>
 </body>
</html>`

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

	// Create an instance of the pagedata, object, and
	// try to fill with the currently-playing details.
	x := Pagedata{Artist: "", Title: "", Playing: false}

	// Get the connection
	client, err := mpd.Dial("tcp", "localhost:6600")
	if err == nil {
		defer client.Close()

		status, err := client.Status()
		if err == nil {
			song, err := client.CurrentSong()
			if err == nil {
				if status["state"] == "play" {
					x.Artist = song["Artist"]
					x.Title = song["Title"]
					x.Playing = true
				} else {
					x.Artist = ""
					x.Title = ""
				}
			}
		}
	}

	t := template.Must(template.New("tmpl").Parse(src))

	buf := &bytes.Buffer{}
	err = t.Execute(buf, x)

	//
	// If there were errors, then show them.
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	//
	// Otherwise write the result.
	//
	buf.WriteTo(w)

	return

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
