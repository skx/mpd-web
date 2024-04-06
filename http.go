// http.go - This file contains HTTP-handlers
//
// These are bound in `main.go`

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/fhs/gompd/mpd"
)

// IndexHandler shows the current state of the server, as a simple HTML
// page.
//
// The output comes from the text/template file which is located beneath
// `web/` in our source-repository, and embedded into a binary in `embed.go`.
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	// We output the files in the current playlist, which will
	// shown.
	type PlaylistEntry struct {
		// pos holds the playlist position
		Pos string

		// name holds the artist/title, or filename.
		Name string
	}
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

		// Playlist has details of the playlist
		Playlist []PlaylistEntry
	}

	// Create an instance of the pagedata to populate our
	// template with.
	x := Pagedata{}

	// Invoke the callback, and try to update our instance
	// with the appropriate data.
	_ = invokeMPD(func(c *mpd.Client) error {

		// Get the status of the server
		status, err := c.Status()
		if err != nil {
			return fmt.Errorf("failed to get mpd-status %s", err)
		}

		// If we got the status of the server then
		// get the current song next.
		var song mpd.Attrs
		song, err = c.CurrentSong()
		if err != nil {
			return fmt.Errorf("failed to get current song %s", err)
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

		// Get details of the current playlist
		info, err2 := c.PlaylistInfo(-1, -1)
		if err2 != nil {
			return fmt.Errorf("failed to get current playlist %s", err2)
		}

		// For each one
		for _, song := range info {

			// Item we'll add
			var tmp PlaylistEntry

			// Get artist/title
			Artist := song["Artist"]
			Title := song["Title"]

			// If that worked we'll save
			if len(Artist) > 0 && len(Title) > 0 {
				tmp.Name = Artist + " " + Title
			} else {
				tmp.Name = song["file"]
			}
			tmp.Pos = song["Pos"]

			x.Playlist = append(x.Playlist, tmp)
		}

		// Add the list of tracks
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

		// Get the status
		stats, err := c.Status()
		if err != nil {
			return fmt.Errorf("error getting status %s", err.Error())
		}

		// If we're stopped then play before we jump the track
		if stats["state"] == "stop" {
			err = c.Play(-1)

			if err != nil {
				return fmt.Errorf("error starting playback when stopped %s", err.Error())
			}
		}

		// Now move
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

// GotoHandler is invoked at /goto, and starts playing the selected song.
func GotoHandler(w http.ResponseWriter, r *http.Request) {

	queryValues := r.URL.Query()
	id := queryValues.Get("position")
	if id == "" {
		fmt.Fprint(w, "Missing position parameter")
		return
	}

	num, nErr := strconv.Atoi(id)
	if nErr != nil {
		fmt.Fprint(w, "Failed to convert position parameter to integer")
		return
	}

	// Play music.
	err := invokeMPD(func(c *mpd.Client) error {
		return c.Play(num)
	})

	// Error in the MPD connection/action?
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	// No error?  Return to the server root
	http.Redirect(w, r, "/", http.StatusFound)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {

	// Pagedata is a structure which is used to add
	// dynamic data to our template.
	type Pagedata struct {
		Data map[string]string
	}

	// Create an instance of the pagedata to populate our
	// template with.
	x := Pagedata{}

	err := invokeMPD(func(c *mpd.Client) error {
		status, err := c.Status()
		if err != nil {
			return fmt.Errorf("failed to get status %s", err)
		}

		x.Data = status

		// Parse our (embedded) template.
		t := template.Must(template.New("tmpl").Parse(statusTemplate))

		// Execute the template into a buffer.
		buf := &bytes.Buffer{}
		err = t.Execute(buf, x)
		if err != nil {
			return fmt.Errorf("failed to render status template %s", err)
		}

		fmt.Fprint(w, buf)
		return nil
	})

	// Error in the MPD connection/action?
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

}

// PrevHandler is invoked at /prev, and moves to the previous track in the playlist.
func PrevHandler(w http.ResponseWriter, r *http.Request) {

	// Previous track
	err := invokeMPD(func(c *mpd.Client) error {

		// Get the status
		stats, err := c.Status()
		if err != nil {
			return fmt.Errorf("error getting status %s", err.Error())
		}

		// If we're stopped then play before we jump the track
		if stats["state"] == "stop" {
			err = c.Play(-1)

			if err != nil {
				return fmt.Errorf("error starting playback when stopped %s", err.Error())
			}
		}

		// Now move
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
