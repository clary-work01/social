package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/chainflow/chainflow-api/internal/store"
	"github.com/go-chi/chi/v5"
)

// 用 typed key
const userCtxKey contextKey = "user"

// GetUser godoc
//
//	@Summary		查詢用戶詳情
//	@Description	查詢用戶詳情
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "userID")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	// user, err := app.store.User.GetByID(ctx, id)
	user, err := app.getUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSONData(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// FollowUser godoc
//
//	@Summary		追蹤用戶
//	@Description	追蹤用戶
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//
//	@Success		204	{string}	string
//	@Failure		400	{object}	error	"User payload missing"
//	@Failure		404	{object}	error	"User Not Found"
//	@Security		ApiKeyAuth
//
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := app.getUserFromCtx(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := app.store.Follower.Follow(r.Context(), followerUser.ID, followedID); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSONData(w, http.StatusNoContent, followerUser); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// UnfollowUser godoc
//
//	@Summary		退追用戶
//	@Description	退追用戶
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//
//	@Success		204	{string}	string
//	@Failure		400	{object}	error	"User payload missing"
//	@Failure		404	{object}	error	"User Not Found"
//	@Security		ApiKeyAuth
//
//	@Router			/users/{userID}/unfollow [put]
func (app *application) unFollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := app.getUserFromCtx(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := app.store.Follower.UnFollow(r.Context(), followerUser.ID, unfollowedID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSONData(w, http.StatusNoContent, followerUser); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// ActivateUser godoc
//
//	@Summary		啟用用戶
//	@Description	啟用用戶
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitaion Token"
//
//	@Success		204		{string}	string	true	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.User.Activate(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSONData(w, http.StatusNoContent, ""); err != nil {
		app.internalServerError(w, r, err)
	}
}

// func (app *application) userContextMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		idParam := chi.URLParam(r, "userID")

// 		id, err := strconv.ParseInt(idParam, 10, 64)
// 		if err != nil {
// 			app.badRequestError(w, r, err)
// 			return
// 		}

// 		ctx := r.Context()
// 		user, err := app.store.User.GetByID(ctx, id)
// 		if err != nil {
// 			switch {
// 			case errors.Is(err, store.ErrNotFound):
// 				app.notFoundError(w, r, err)
// 			default:
// 				app.internalServerError(w, r, err)
// 			}
// 			return
// 		}

// 		ctx = context.WithValue(ctx, userCtxKey, user)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func (app *application) getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtxKey).(*store.User)
	return user
}
