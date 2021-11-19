// http.go - This file contains HTTP-handlers
//
// These are bound in `main.go`

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/fhs/gompd/mpd"
)

// IndexHandler shows the current state of the server, as a simple HTML
// page.
//
// The output comes from the text/template file which is located beneath
// `web/` in our source-repository, and embedded into a binary in `embed.go`.
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	// Pagedata is a structure which is used to add
	// dynamic data to our template.
	type Pagedata struct {

		// Name of album
		Album string

		// Name of artist
		Artist string

		// Filename of track
		File string

		// Are we playing?
		Playing bool

		// Do we have a valid artist/title tag?
		Populated bool

		// Song title
		Title string
	}

	// Create an instance of the pagedata to populate our
	// template with.
	x := Pagedata{}

	// Invoke the callback, and try to update our instance
	// with the appropriate data.
	_ = invokeMPD(func(c *mpd.Client) error {

		defer c.Close()

		// Get the status of the server
		status, err := c.Status()
		if err != nil {
			return err
		}

		// If we got the status of the server then
		// get the current song next.
		var song mpd.Attrs
		song, err = c.CurrentSong()
		if err != nil {
			return err
		}

		// If we're playing then we can populate
		// the track details.
		if status["state"] == "play" {

			x.Album = song["Album"]
			x.Artist = song["Artist"]
			x.File = song["file"]
			x.Title = song["Title"]

			x.Populated = (len(x.Artist) > 0 && len(x.Title) > 0)
			x.Playing = true
		}

		return nil
	})

	// Parse our (embedded) template.
	t := template.Must(template.New("tmpl").Parse(indexTemplate))

	// Execute the template into a buffer.
	buf := &bytes.Buffer{}
	err := t.Execute(buf, x)

	// If there were errors, then show them.
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// Otherwise serve the result to the client.
	buf.WriteTo(w)
}

// NextHandler is invoked at /next, and moves to the next track in the playlist.
func NextHandler(w http.ResponseWriter, r *http.Request) {

	// Next track
	err := invokeMPD(func(c *mpd.Client) error {
		return c.Next()
	})

	// Error in the MPD connection/action?
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// No error?  Return to the server root
	http.Redirect(w, r, "/", http.StatusFound)
}

// PlayHandler is invoked at /play, and starts playing, if stopped.
func PlayHandler(w http.ResponseWriter, r *http.Request) {

	// Play music.
	err := invokeMPD(func(c *mpd.Client) error {
		return c.Play(-1)
	})

	// Error in the MPD connection/action?
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// No error?  Return to the server root
	http.Redirect(w, r, "/", http.StatusFound)
}

// PrevHandler is invoked at /prev, and moves to the previous track in the playlist.
func PrevHandler(w http.ResponseWriter, r *http.Request) {

	// Previous track
	err := invokeMPD(func(c *mpd.Client) error {
		return c.Previous()
	})

	// Error in the MPD connection/action?
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// No error?  Return to the server root
	http.Redirect(w, r, "/", http.StatusFound)
}

// StopHandler is invoked at /stop, and stops playback.
func StopHandler(w http.ResponseWriter, r *http.Request) {

	// Stop music
	err := invokeMPD(func(c *mpd.Client) error {
		return c.Stop()
	})

	// Error in the MPD connection/action?
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// No error?  Return to the server root
	http.Redirect(w, r, "/", http.StatusFound)
}
