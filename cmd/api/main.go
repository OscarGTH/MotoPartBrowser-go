package main

import (
	"Crawler/internal/helpers"
)

func main() {
	helpers.ReadConfig()
	a := App{}
	a.Initialize()
	a.Run("127.0.0.1:8010")
}
