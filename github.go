package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

func getUsers(url string) *[]User {
	var username string = os.Getenv("GITHUB_USERNAME")
	if username == "" {
		panic(appConfigErrorMessage)
	}

	var passwd string = os.Getenv("GITHUB_PAT")
	if passwd == "" {
		panic(appConfigErrorMessage)
	}

	client := &http.Client{}

	var users []User
	nextUrl := url

	for nextUrl != "" {
		fmt.Println("URL: " + nextUrl)
		req, err := http.NewRequest("GET",
			nextUrl,
			nil)
		if err != nil {
			panic(gitHubErrorMessage)
		}
		req.SetBasicAuth(username, passwd)
		resp, err := client.Do(req)
		if err != nil {
			panic(gitHubErrorMessage)
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		body, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			panic(gitHubErrorMessage)
		}
		var userList []User = []User{}
		jsonErr := json.Unmarshal(body, &userList)
		if jsonErr != nil {
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

	print(len(users))
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
