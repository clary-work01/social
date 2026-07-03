package main

import (
	"net/http"
)

// Health godoc
//
//	@Summary		心跳
//	@Description	心跳
//	@Tags			ops
//	@Accept			json
//	@Produce		json
//
//	@Success		200	{object}	string	"ok"
//
//	@Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"version": version,
		"env":     app.config.env,
	}

	// test graceful shutdown
	// time.Sleep(time.Second * 4)

	if err := writeJSONData(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
	}

}
