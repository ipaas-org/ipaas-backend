package model

type GithubCommit struct {
	SHA    string               `json:"sha"`
	Commit GithubCommitInternal `json:"commit"`
}

type GithubCommitInternal struct {
	Message string `json:"message"`
}
