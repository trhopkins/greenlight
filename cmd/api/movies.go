package main

import (
	"fmt"
	"net/http"
	"time"

	"greenlight.camphopkins.com/internal/data"
)

func (app *application) createMovieHandler(w http.ResponseWriter, _ *http.Request) {
  fmt.Fprint(w, "create a new movie")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
  id, err := app.readIdParam(r)
  if err != nil {
    app.notFoundResponse(w, r)
    return
  }


  movie := data.Movie{
    Id: id,
    CreatedAt: time.Now(),
    Title: "Casablanca",
    Runtime: 102,
    Genres: []string{"drama", "romance", "war"},
    Version: 1,
  }

  err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
  if err != nil {
    app.serverErrorResponse(w, r, err)
  }
}

