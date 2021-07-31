package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type User struct {
	Login string `json:"login"`
}

type Users struct {
	Active  []User `json:"active"`
	Pending []User `json:"pending"`
}

const appConfigErrorMessage = "App configuration error"
const gitHubErrorMessage = "Github API error"

func createHttpClient() *http.Client {
	return &http.Client{}
}

func addGithubLoginCredentialsHeader(client *http.Client, req *http.Request) {
	var username string = os.Getenv("GITHUB_USERNAME")
	if username == "" {
		panic(appConfigErrorMessage)
	}

	var passwd string = os.Getenv("GITHUB_PAT")
	if passwd == "" {
		panic(appConfigErrorMessage)
	}

	req.SetBasicAuth(username, passwd)
}

func getBodyAsString(resp *http.Response) []byte {
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		panic(gitHubErrorMessage)
	}
	return body
}

func getUrlJson(client *http.Client, url string) (*http.Response, []byte) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(gitHubErrorMessage)
	}

	addGithubLoginCredentialsHeader(client, req)
	resp, err := client.Do(req)
	if err != nil {
		panic(gitHubErrorMessage)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body := getBodyAsString(resp)
	return resp, body
}

func getUsers(url string) *[]User {
	client := createHttpClient()

	var users []User
	nextUrl := url

	for nextUrl != "" {
		resp, body := getUrlJson(client, nextUrl)

		var userList []User
		err := json.Unmarshal(body, &userList)
		if err != nil {
			panic(gitHubErrorMessage)
		}

		users = append(users, userList...)

		nextUrl = ""
		headerStr := resp.Header.Get("Link")
		links := strings.Split(headerStr, ",")
		for _, link := range links {
			fields := strings.Split(link, ";")
			if len(fields) < 2 {
				continue
			}
			relField := strings.TrimSpace(fields[1])
			if relField == "rel=\"next\"" {
				nextUrl = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(fields[0]), "<", ""), ">", "")
			}
		}
	}

	return &users
}

func getActiveUsersFromGithub() *[]User {
	return getUsers("https://api.github.com/orgs/programmers-from-the-same-company/members")
}

func getPendingInviteUsers() *[]User {
	return getUsers("https://api.github.com/orgs/programmers-from-the-same-company/invitations")
}

func GetUserLists() *Users {
	activeUsers := getActiveUsersFromGithub()
	pendingInviteUsers := getPendingInviteUsers()
	return &Users{Active: *activeUsers, Pending: *pendingInviteUsers}
}

func InviteUser(login string) error {
	client := createHttpClient()

	userResponse := struct {
		Id int `json:"id"`
	}{}
	getResp, body := getUrlJson(client, `https://api.github.com/users/`+login)
	err := json.Unmarshal(body, &userResponse)
	if err != nil {
		panic(gitHubErrorMessage)
	}
	if getResp.StatusCode == http.StatusNotFound {
		error := "GitHub user not found: " + login
		log.Println(error)
		return errors.New(error)
	}

	log.Println("Inviting user: " + strconv.Itoa(userResponse.Id))

	invite := struct {
		Invitee_id int    `json:"invitee_id"`
		Role       string `json:"role"`
	}{
		Invitee_id: userResponse.Id,
		Role:       "direct_member",
	}
	json, err := json.Marshal(invite)
	if err != nil {
		panic(gitHubErrorMessage)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.github.com/orgs/programmers-from-the-same-company/invitations",
		strings.NewReader(string(json)))
	if err != nil {
		log.Print(err.Error())
		panic(gitHubErrorMessage)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	addGithubLoginCredentialsHeader(client, req)

	resp, postErr := client.Do(req)
	if postErr != nil {
		log.Panic(postErr.Error())
	}
	if resp.StatusCode != http.StatusCreated {
		log.Println("Invite unsuccessful")
		log.Println(resp)
		log.Println(err)
		return errors.New("Invite unsuccessful")
	}

	return nil
}
