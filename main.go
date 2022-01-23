// Trivial HTTP <-> MPD gateway.

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fhs/gompd/mpd"
)

// Signature for a callback function we can invoke against the MPD client
type Signature func(client *mpd.Client) error

// invokeMPD invokes a given function with the MPD client object.
//
// This is used to trigger calls in a simple way.
func invokeMPD(fnc Signature) error {
	client, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		return err
	}

	// Invoke the callback function, then close the connection
	err = fnc(client)
	client.Close()
	return err
}

// Entry Point
func main() {

	// Bind our handlers
	http.HandleFunc("/next", NextHandler)
	http.HandleFunc("/play", PlayHandler)
	http.HandleFunc("/prev", PrevHandler)
	http.HandleFunc("/stop", StopHandler)
	http.HandleFunc("/", IndexHandler)

	// Start the server
	fmt.Println("Server started at http://0.0.0.0:8888/")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
