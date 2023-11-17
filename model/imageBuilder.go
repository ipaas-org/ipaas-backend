package model

type (
	BuildResponse struct {
		ApplicationID string             `json:"applicationID"`
		Repo          string             `json:"repo"`
		Status        ResponseStatus     `json:"status"` // success | failed
		ImageID       string             `json:"imageID"`
		ImageName     string             `json:"imageName"`
		BuiltCommit   string             `json:"buildCommit"`
		IsError       bool               `json:"isError"`
		Fault         ResponseErrorFault `json:"fault"` // service | user
		Message       string             `json:"message"`
	}

	ResponseStatus     string
	ResponseErrorFault string

	BuildRequest struct {
		ApplicationID string `json:"applicationID"`
		Token         string `json:"token"`
		UserID        string `json:"userID"`
		Type          string `json:"type"`      // repo, tag, release, ...
		Connector     string `json:"connector"` //github, gitlab, ...
		Repo          string `json:"repo,omitempty"`
		Branch        string `json:"branch,omitempty"`
		Tag           string `json:"tag,omitempty"`
		Release       string `json:"release,omitempty"`
		// Binary     string `json:"binary, omitempty"`
	}
)

const (
	ResponseStatusSuccess ResponseStatus = "success"
	ResponseStatusFailed  ResponseStatus = "failed"

	ResponseErrorFaultService ResponseErrorFault = "service"
	ResponseErrorFaultUser    ResponseErrorFault = "user"
)
