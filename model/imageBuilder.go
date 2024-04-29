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
		BuildOutput   string             `json:"buildOutput"`
	}

	ResponseStatus     string
	ResponseErrorFault string

	// BuildRequest struct {
	// 	ApplicationID string `json:"applicationID"`
	// 	Token         string `json:"token"`
	// 	UserID        string `json:"userID"`
	// 	Type          string `json:"type"`      // repo, tag, release, ...
	// 	Connector     string `json:"connector"` //github, gitlab, ...
	// 	Repo          string `json:"repo,omitempty"`
	// 	Branch        string `json:"branch,omitempty"`
	// 	Tag           string `json:"tag,omitempty"`
	// 	Release       string `json:"release,omitempty"`
	// 	// Binary     string `json:"binary, omitempty"`
	// }

	Request struct {
		ApplicationID string           `json:"applicationID"`
		PullInfo      *PullInfoRequest `json:"pullInfo"`
		BuildPlan     *BuildConfig     `json:"buildPlan"`
	}

	BuildConfig struct {
		Builder        string     `json:"builder"`
		RootDirectory  string     `json:"rootDirectory"`
		DockerfilePath string     `json:"dockerfilePath"`
		NixpacksPath   string     `json:"nixpacksPath"`
		Envs           []KeyValue `json:"envs"`
		NixPkgs        []string   `json:"nixPkgs"`
		NixLibs        []string   `json:"nixLibs"`
		AptPkgs        []string   `json:"aptPkgs"`
		InstallCommand string     `json:"installCommand"`
		BuildCommand   string     `json:"buildCommand"`
		StartCommand   string     `json:"startCommand"`
	}

	PullInfoRequest struct {
		UserID    string `json:"userID"`
		Token     string `json:"token"`
		Repo      string `json:"repo"`
		Connector string `json:"connector"`
		Branch    string `json:"branch"`
	}
)

const (
	ResponseStatusSuccess ResponseStatus = "success"
	ResponseStatusFailed  ResponseStatus = "failed"

	ResponseErrorFaultService ResponseErrorFault = "service"
	ResponseErrorFaultUser    ResponseErrorFault = "user"
)
