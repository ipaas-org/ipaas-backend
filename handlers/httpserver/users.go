package httpserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type (
	// Here we declare a user model that will be returned by the api
	// to any unauthorized user as some informations should be visible only to admins
	HttpUnauthenticatedUser struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Pfp       string `json:"pfp"`
		Email     string `json:"email"`
	}
	// We are not going to declare a model for the authorized request as we will just return the model

	HttpNewUserPost struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	HttpUserSettingsPost struct {
		Theme string `json:"theme"`
	}

	HttpUpdateUserPost struct {
		ID        int    `json:"id,omitempty" validate:"optional"`
		FirstName string `json:"first_name,omitempty" validate:"optional"`
		LastName  string `json:"last_name,omitempty" validate:"optional"`
		Email     string `json:"email,omitempty" validate:"optional"`
		Password  string `json:"password,omitempty" validate:"optional"`
	}

	HttpLoginUserPost struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
)

func (h *httpHandler) UserInfo(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}

	return respSuccess(c, http.StatusOK, "user info", user)
}

func (h *httpHandler) UpdateUser(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}
	ctx := c.Request().Context()

	//get HttpUserSettingsPost from request body
	var post HttpUserSettingsPost
	err := c.Bind(&post)
	if err != nil {
		return respError(c, http.StatusBadRequest, "invalid request body", "invalid request body", "invalid_request_body")
	}
	if user.UserSettings.Theme == post.Theme {
		return respSuccess(c, http.StatusOK, "user info not changed", post.Theme)
	}

	user.UserSettings.Theme = post.Theme
	err = h.controller.UpdateUser(ctx, user)
	if err != nil {
		return respError(c, http.StatusNotImplemented, "unexpected error", "unexpected error trying to update user", "unexpected_error")
	}

	return respSuccess(c, http.StatusOK, "user info updated correctly", post.Theme)
}
