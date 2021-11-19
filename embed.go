// This file just contains the embedded HTML-template from within
// our `web` directory.

package main

import (
	_ "embed"
)

//go:embed web/index.html
var indexTemplate string
