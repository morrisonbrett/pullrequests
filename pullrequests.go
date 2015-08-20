//
// Brett Morrison, June 2015
//
// A simple program to display a list of open pull requests from BitBucket ✔
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Setup the Authentication info
const bbBaseURL = "https://bitbucket.org/api/2.0"

var bitbucketOwnerName string
var bitbucketUserName string
var bitbucketPassword string

// RepoResponse ...
type RepoResponse struct {
	Repos []Repo `json:"values"`
	Next  string `json:"next"`
}

// Repo ...
type Repo struct {
	Links Links `json:"links"`
}

// Links ...
type Links struct {
	PullRequests PullRequests `json:"pullrequests"`
}

// PullRequests ...
type PullRequests struct {
	Href string `json:"href"`
}

// PRResponse ...
type PRResponse struct {
	PRs  []PR   `json:"values"`
	Next string `json:"next"`
}

// PR ...
type PR struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Source      Source      `json:"source"`
	Destination Destination `json:"destination"`
	Author      Author      `json:"author"`
	PRLinks     PRLinks     `json:"links"`
}

// Destination ...
type Destination struct {
	Branch Branch `json:"branch"`
}

// Source ...
type Source struct {
	Branch Branch `json:"branch"`
}

// Branch ...
type Branch struct {
	Name string `json:"name"`
}

// Author ...
type Author struct {
	DisplayName string `json:"display_name"`
}

// PRLinks ...
type PRLinks struct {
	Self Self `json:"self"`
}

// Self ...
type Self struct {
	Href string `json:"href"`
}

// ParticipantsResponse ...
type ParticipantsResponse struct {
	Participants []Participant `json:"participants"`
}

// Participant ...
type Participant struct {
	Role     string `json:"role"`
	Approved bool   `json:"approved"`
	User     User   `json:"user"`
}

// User ...
type User struct {
	DisplayName string `json:"display_name"`
}

//
// Given a BB API, return JSON as a map interface
//
func getJSON(URL string, v interface{}) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return fmt.Errorf("request error: %s", err)
	}

	req.SetBasicAuth(bitbucketUserName, bitbucketPassword)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %s", err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("response code: %d", res.StatusCode)
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&v); err == io.EOF {
		return fmt.Errorf("decode error: %s", err)
	}

	return nil
}

func displayParticipantInfo(ID int, selfHrefLink string) error {
	participantsResponse := ParticipantsResponse{}

	// Get more details about the PR
	err := getJSON(selfHrefLink, &participantsResponse)
	if err != nil {
		return err
	}

	// For determining if the PR is ready to merge
	numApprovedReviewers := 0
	numReviewers := 0

	// For each participant in the PR, display state depending on role
	for _, value := range participantsResponse.Participants {
		role := value.Role
		approved := value.Approved
		displayName := value.User.DisplayName

		// TODO Rewrite with one line RegEx?
		var approvedS = " "
		if approved {
			approvedS = "X"
		}

		switch role {
		case "REVIEWER":
			fmt.Printf("        %s %s\n", approvedS, displayName)
			numReviewers++
			if approved {
				numApprovedReviewers++
			}
		case "PARTICIPANT":
			fmt.Printf("        %s (%s)\n", approvedS, displayName)
		default:
			fmt.Printf("        %s %s (%s)\n", approvedS, displayName, role)
		}
	}

	var isOrNot = "IS NOT"
	if numReviewers > 0 && numReviewers == numApprovedReviewers {
		isOrNot = "IS"
	}

	fmt.Printf("    #%d %s READY TO MERGE, %d of %d REVIEWERS APPROVED\n\n", ID, isOrNot, numApprovedReviewers, numReviewers)

	return nil
}

//
// Given a PR URL, iterate through state and print info
//
func listPR(pullRequestsLink string) error {
	var prAPI = pullRequestsLink

	// PR API has pagination, code for > 1 page
	for len(prAPI) > 0 {
		pRResponse := PRResponse{}

		err := getJSON(prAPI, &pRResponse)
		if err != nil {
			return err
		}

		// For each PR in the repo
		for _, value := range pRResponse.PRs {
			// Display base info about the PR
			fmt.Printf("    #%d %s (%s -> %s) by %s\n",
				value.ID,
				value.Title,
				value.Source.Branch.Name,
				value.Destination.Branch.Name,
				value.Author.DisplayName)
			// Prep the URL for more details about the PR
			links := value.PRLinks
			self := links.Self
			selfHref := self.Href

			// Display participant details about the PR
			err := displayParticipantInfo(value.ID, selfHref)
			if err != nil {
				return err
			}
		}

		// Determine if there's more results - if so, this assignment will loop control back
		prAPI = pRResponse.Next
	}

	return nil
}

func init() {
	flag.StringVar(&bitbucketOwnerName, "ownername", "", "Bitbucket repository owner account")
	flag.StringVar(&bitbucketUserName, "username", "", "Bitbucket account username")
	flag.StringVar(&bitbucketPassword, "password", "", "Bitbucket account password")
}

func rootRepos(bbOwnername, bbUsername, bbPassword string) error {
	var reposAPI = bbBaseURL + "/repositories/" + bitbucketOwnerName

	// Repo API has pagination, code for > 1 page
	for len(reposAPI) > 0 {
		repoResponse := RepoResponse{}

		// Get the list of repos for this user / group
		err := getJSON(reposAPI, &repoResponse)
		if err != nil {
			fmt.Println(err)
			return err
		}

		// For each repo, get the PR URL and process
		for _, value := range repoResponse.Repos {
			//fmt.Printf("value: %s\n\n", value)
			links := value.Links
			pullRequests := links.PullRequests
			pullRequestsHref := pullRequests.Href
			pullRequestsLink := fmt.Sprint(pullRequestsHref)
			fmt.Println("Repo:", pullRequestsLink)

			err := listPR(pullRequestsLink)
			if err != nil {
				fmt.Println(err) // OK to continue here if error, no need to exit the program
			}
		}

		// Determine if there's more results - if so, this assignment will loop control back
		reposAPI = repoResponse.Next
	}

	return nil
}

func main() {
	// Command line args
	flag.Parse()
	if len(os.Args) != 4 {
		flag.Usage()
		os.Exit(1)
	}

	err := rootRepos(bitbucketOwnerName, bitbucketUserName, bitbucketPassword)
	if err != nil {
		os.Exit(1)
	}
}
