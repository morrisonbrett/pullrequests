//
// Brett Morrison, June 2015
//
// A simple program to display a list of open pull requests from BitBucket ✔
//
package main

import (
	"encoding/json"
	"errors"
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

//
// Given a BB API, return JSON as a map interface
//
func getJSON(URL string) (map[string]interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("request error: %s", err))
	}

	req.SetBasicAuth(bitbucketUserName, bitbucketPassword)

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("request error: %s", err))
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("response code: %d", res.StatusCode))
	}

	var dat interface{}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&dat); err == io.EOF {
		return nil, errors.New(fmt.Sprintf("decode error: %s", err))
	}

	jsonResponse := dat.(map[string]interface{})

	return jsonResponse, nil
}

func displayParticipantInfo(ID int, selfHrefLink string) error {
	// Get more details about the PR
	jsonResponseDet, err := getJSON(selfHrefLink)
	if err != nil {
		return err
	}

	// Get details about the participants
	prsDet := jsonResponseDet["participants"]
	prsDetI := prsDet.([]interface{})

	// For determining if the PR is ready to merge
	numApprovedReviewers := 0
	numReviewers := 0

	// For each participant in the PR, display state depending on role
	for _, value := range prsDetI {
		valueMap := value.(map[string]interface{})
		role := valueMap["role"]
		approved := valueMap["approved"] == true
		displayName := valueMap["user"].(map[string]interface{})["display_name"]

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
		jsonResponse, err := getJSON(prAPI)
		if err != nil {
			return err
		}

		prs := jsonResponse["values"]
		prsI := prs.([]interface{})

		// For each PR in the repo
		for _, value := range prsI {
			valueMap := value.(map[string]interface{})
			ID := int(valueMap["id"].(float64))

			// Display base info about the PR
			fmt.Printf("    #%d %s (%s -> %s) by %s\n",
				ID,
				valueMap["title"],
				valueMap["source"].(map[string]interface{})["branch"].(map[string]interface{})["name"],
				valueMap["destination"].(map[string]interface{})["branch"].(map[string]interface{})["name"],
				valueMap["author"].(map[string]interface{})["display_name"])

			// Prep the URL for more details about the PR
			links := valueMap["links"]
			self := links.(map[string]interface{})["self"]
			selfHref := self.(map[string]interface{})["href"]
			selfHrefLink := fmt.Sprint(selfHref)

			// Display participant details about the PR
			err := displayParticipantInfo(ID, selfHrefLink)
			if err != nil {
				return err
			}
		}

		// Determine if there's more results - if so, loop control back
		next := jsonResponse["next"]
		if next != nil {
			prAPI = fmt.Sprint(next)
		} else {
			prAPI = ""
		}
	}

	return nil
}

func init() {
	flag.StringVar(&bitbucketOwnerName, "ownername", "", "Bitbucket repository owner account")
	flag.StringVar(&bitbucketUserName, "username", "", "Bitbucket account username")
	flag.StringVar(&bitbucketPassword, "password", "", "Bitbucket account password")
}

func main() {
	// Command line args
	flag.Parse()
	if len(os.Args) != 4 {
		flag.Usage()
		os.Exit(1)
	}

	var reposAPI = bbBaseURL + "/repositories/" + bitbucketOwnerName

	// Repo API has pagination, code for > 1 page
	for len(reposAPI) > 0 {
		// Get the list of repos for this user / group
		jsonResponse, err := getJSON(reposAPI)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		repos := jsonResponse["values"]
		reposI := repos.([]interface{})

		// For each repo, get the PR URL and process
		for _, value := range reposI {
			links := value.(map[string]interface{})["links"]
			pullRequests := links.(map[string]interface{})["pullrequests"]
			pullRequestsHref := pullRequests.(map[string]interface{})["href"]

			pullRequestsLink := fmt.Sprint(pullRequestsHref)
			fmt.Println("Repo:", pullRequestsLink)

			err := listPR(pullRequestsLink)
			if err != nil {
				fmt.Println(err) // OK to continue here if error, no need to exit the program
			}
		}

		// Determine if there's more results - if so, loop control back
		next := jsonResponse["next"]
		if next != nil {
			reposAPI = fmt.Sprint(next)
		} else {
			reposAPI = ""
		}
	}
}
