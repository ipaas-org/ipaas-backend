package model

type (
	BuildResponse struct {
		UUID         string      `json:"uuid"`
		Repo         string      `json:"repo"`
		Status       string      `json:"status"` // success | failed
		ImageID      string      `json:"imageID"`
		ImageName    string      `json:"imageName"`
		LatestCommit string      `json:"latestCommit"`
		Error        *BuildError `json:"error"`
		Metadata     map[string][]string
	}

	BuildError struct {
		Fault string `json:"fault"` // service | user
		// if user's fault this message will be the reason why the image didnt compile
		//otherwise it will be the service error
		Message string `json:"message"`
	}

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
	}
)
