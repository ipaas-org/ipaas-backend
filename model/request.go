package model

type DbPost struct {
	DbDescription     string `json:"dbDescription,omitempty"`
	DbName            string `json:"databaseName"`
	DbType            string `json:"databaseType"`
	DbVersion         string `json:"databaseVersion"`
	DbTableCollection string `json:"databaseTable"`
}

type AppPost struct {
	GithubRepoUrl string `json:"github-repo"`
	GithubBranch  string `json:"github-branch"`
	Language      string `json:"language"`
	Port          string `json:"port"`
	Description   string `json:"description,omitempty"`
	Envs          []Env  `json:"envs,omitempty"`
}
