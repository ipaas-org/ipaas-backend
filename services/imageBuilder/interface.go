package imageBuilder

import "github.com/ipaas-org/ipaas-backend/model"

/*

BuildRequest struct {
		UUID      string `json:"uuid"` // given by the client
		Token     string `json:"token"`
		UserID    string `json:"userID"`
		Type      string `json:"type"`      // repo, tag, release, ...
		Connector string `json:"connector"` //github, gitlab, ...
		Repo      string `json:"repo,omitempty"`
		Branch    string `json:"branch,omitempty"`
		Tag       string `json:"tag,omitempty"`
		Release   string `json:"release,omitempty"`
		// Binary     string `json:"binary, omitempty"`
	}*/

type ImageBuilder interface {
	BuildImage(buildInfo model.BuildRequest) error
	ValidateImageResponse(response model.BuildResponse) (string, error)
}
