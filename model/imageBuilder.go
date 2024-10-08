package model

type (
	BuildRequest struct {
		ApplicationID string           `json:"applicationID"`
		PullInfo      *PullInfoRequest `json:"pullInfo"`
		BuildPlan     *BuildConfig     `json:"buildPlan"`
	}

	BuildConfig struct {
		// must
		RootDirectory string `json:"rootDirectory"`

		// shared
		Builder BuilderKind `json:"builder"`

		// docker
		DockerfilePath string `json:"dockerfilePath"`

		// nixpacks
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
		Commit    string `json:"commit"` //commit to build, use latest to build the latest commit
	}
)

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
		PlanUsed      *BuildConfig       `json:"buildPlan"`
		RepoAnalisys  *RepoAnalisys      `json:"repoAnalysis"`
	}

	ResponseStatus     string
	ResponseErrorFault string
)

const (
	ResponseStatusSuccess ResponseStatus = "success"
	ResponseStatusFailed  ResponseStatus = "failed"

	ResponseErrorFaultService ResponseErrorFault = "service"
	ResponseErrorFaultUser    ResponseErrorFault = "user"
)
