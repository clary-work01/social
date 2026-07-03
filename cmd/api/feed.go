package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chainflow/chainflow-api/internal/store"
)

// GetUserFeed godoc
//
//	@Summary		取得用戶跟用戶追蹤的人的貼文
//	@Description	取得用戶跟用戶追蹤的人的貼文
//	@Tags			feeds
//	@Accept			json
//	@Produce		json
//
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			search	query		string	false	"Search"
//	@Param			tags	query		string	false	"Tags"
//	@Param			since	query		string	false	"Since"
//	@Param			until	query		string	false	"Until"
//
//	@Success		200		{object}	[]store.UserFeed
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//
//	@Router			/users/feed [get]
func (app *application) getUserFeedQueryParams(r *http.Request) (store.FeedQueryParam, error) {
	q := r.URL.Query()

	params := store.FeedQueryParam{
		Offset: 0,
		Limit:  10,
		Sort:   "desc",
	}

	if v := q.Get("offset"); v != "" {
		offset, err := strconv.Atoi(v)
		if err != nil {
			return params, fmt.Errorf("invalid offset %w", err)
		}
		params.Offset = offset
	}

	if v := q.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {
			return params, fmt.Errorf("invalid limit %w", err)
		}
		params.Limit = limit
	}

	if v := q.Get("sort"); v != "" {
		params.Sort = v
	}

	if v := q.Get("search"); v != "" {
		params.Search = v
	}

	if v := q.Get("tags"); v != "" {
		tags := strings.Split(v, ",")
		params.Tags = tags
	}

	if v := q.Get("since"); v != "" {
		since, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return params, fmt.Errorf("invalid since format, use RFC3339: %w", err)
		}
		params.Since = &since
	}

	if v := q.Get("until"); v != "" {
		until, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return params, fmt.Errorf("invalid until format, use RFC3339: %w", err)
		}
		params.Until = &until
	}

	return params, nil

}
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	params, err := app.getUserFeedQueryParams(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(params); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	feed, err := app.store.Post.GetUserFeed(r.Context(), 51, params)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSONData(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
