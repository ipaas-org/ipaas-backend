package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/services/oauth"
	"github.com/tidwall/gjson"
)

const (
	userInfo      = "https://api.github.com/user"
	emailInfo     = "https://api.github.com/user/emails"
	rateLimitInfo = "https://api.github.com/rate_limit"
	repoInfo      = "https://api.github.com/user/repos"
)

var _ oauth.Oauther = new(GithubOauth)

type OauthUser struct {
	Login        string `json:"login"`
	ID           int64  `json:"id"`
	NodeID       string `json:"node_id"`
	AvatarURL    string `json:"avatar_url"`
	URL          string `json:"url"`
	HTMLURL      string `json:"html_url"`
	FollowersURL string `json:"followers_url"`
	FollowingURL string `json:"following_url"`
	Name         string `json:"name"`
	Blog         string `json:"blog"`
	Location     string `json:"location"`
	Email        string `json:"email"`
	Bio          string `json:"bio"`
}

type GithubOauth struct {
	clientID     string
	clientSecret string
	callbackUri  string
}

func NewGithubOauth(clientID, clientSecret, callbackUri string) *GithubOauth {
	return &GithubOauth{
		clientID:     clientID,
		clientSecret: clientSecret,
		callbackUri:  callbackUri,
	}
}

func (g GithubOauth) GenerateLoginRedirectUri(state string) string {
	return fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user user:email repo write:repo_hook&state=%s",
		g.clientID,
		g.callbackUri,
		state,
	)
}

func (g GithubOauth) GetAccessTokenFromCode(code string) (string, error) {
	requestBodyMap := map[string]string{
		"client_id":     g.clientID,
		"client_secret": g.clientSecret,
		"code":          code,
	}

	requestJSON, err := json.Marshal(requestBodyMap)
	if err != nil {
		return "", err
	}

	// POST request to set URL
	req, err := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		bytes.NewBuffer(requestJSON),
	)

	if err != nil {
		return "", fmt.Errorf("unable to generate access token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	// Get the response
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	var ghresp githubAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghresp); err != nil {
		return "", fmt.Errorf("unable to decode response body: %w", err)
	}

	return ghresp.AccessToken, nil
}

func (g GithubOauth) GetUserInfo(accessToken string) (model.User, error) {
	var user model.User

	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req, err := http.NewRequest("GET", userInfo, nil)
	if err != nil {
		return user, fmt.Errorf("unable to generate user info request: %w", err)
	}
	req.Header.Set("Authorization", authorizationHeaderValue)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return user, fmt.Errorf("unable to get user info: %w", err)
	}

	var githubUser OauthUser
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return user, fmt.Errorf("unable to unmarshal user info response: %w", err)
	}

	user.Username = githubUser.Login
	user.FullName = githubUser.Name
	user.GithubUrl = githubUser.URL
	user.Pfp = githubUser.AvatarURL
	user.GithubAccessToken = accessToken

	//set req to get email
	req.URL.Parse(emailInfo)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return user, fmt.Errorf("unable to get user email: %w", err)
	}

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return user, fmt.Errorf("unable to read user email response: %w", err)
	}

	result := gjson.Get(string(respbody), "#(primary==true).email")

	user.Email = result.String()
	return user, nil
}
