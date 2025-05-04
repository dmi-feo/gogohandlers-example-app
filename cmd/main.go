package main

import (
	app "gogohandlers-example-app/internal"
	"log"
	"net/http"
)

func main() {
	mux := app.GetRouter()

	if err := http.ListenAndServe(":7777", mux); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
