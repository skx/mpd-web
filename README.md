# mpd-web

This repository contains a simple HTTP-server which allows *basic* control of an [mpd](https://www.musicpd.org/).

* Move to the next track
* Move to the previous track
* Stop playback, and restart it.



## Rational

I have `mpd` installed upon my desktop, and I use it for playing music when I'm in front of my PC.

When I'm outside the office I often have playback sent through some bluetooth speakers - and I don't want to have to go back to my PC to skip to the next track.

> I can stop playback by turning off the speaker, of course!

So I make sure that I launch this deamon too, and then I can skip tracks with my mobile phone.



## Installation

Assuming you have a working go compiler, & etc, you can install via:

    $ go install github.com/skx/mpd-web@latest

Otherwise clone [this repository](https://github.com/skx/mpd-web) and build/install like so:

    $ cd mpd-web
    $ go build .
    $ go install .



## Configuration

There are zero configuration options; the server defaults to serving on port 8888 (all interfaces), and contacting the `mpd` server on `localhost:6600`.



## Bugs?

Report as you see them :)



Steve
