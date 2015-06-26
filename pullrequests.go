//
// Brett Morrison, June 2015
//
// A simple program to display a list of open pull requests from BitBucket ✔
// 
package main

import (
    "os"
    "flag"
    "fmt"
    "io"
    "encoding/json"
    "net/http"
    "errors"
)

// Setup the Authentication info
const bb_base_url = "https://bitbucket.org/api/2.0"

var bitbucket_owner_name string
var bitbucket_username string
var bitbucket_password string

//
// Given a BB API, return JSON as a map interface
//
func getJSON(url string) (map[string]interface{}, error) {
    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
         return nil, errors.New(fmt.Sprintf("Request Error: %s", err))
    }

    req.SetBasicAuth(bitbucket_username, bitbucket_password)

    res, err := client.Do(req)
    if err != nil {
         return nil, errors.New(fmt.Sprintf("Request Error: %s", err))
    }
    
    defer res.Body.Close()
     
    if (res.StatusCode != 200) {
        return nil, errors.New(fmt.Sprintf("Response Code: %d", res.StatusCode))
    }
     
    var dat interface{}     

    decoder := json.NewDecoder(res.Body)
    if err := decoder.Decode(&dat); err == io.EOF {
         return nil, errors.New(fmt.Sprintf("Decode Error: %s", err))
    }
    
    json_response := dat.(map[string]interface{})
    
    return json_response, nil
}

//
// Given a PR url, iterate through state and print info
//
func listPR(pull_requests_link string) (error) {
    var pr_api = pull_requests_link
    
    // PR API has pagination, code for > 1 page
    for len(pr_api) > 0 {
        json_response, err := getJSON(pr_api)
        if err != nil {
            return err
        }
        
        prs := json_response["values"]
        prs_i := prs.([]interface{})
    
        // For each PR in the repo
        for _, value := range prs_i {
            value_map := value.(map[string]interface{})
            id := value_map["id"]
            
            // Display base info about the PR
            fmt.Printf("    #%.0f %s (%s -> %s) by %s\n",
                id,
                value_map["title"],
                value_map["source"].(map[string]interface{})["branch"].(map[string]interface{})["name"],
                value_map["destination"].(map[string]interface{})["branch"].(map[string]interface{})["name"],
                value_map["author"].(map[string]interface{})["display_name"])
    
            // Prep the URL for more details about the PR
            links := value_map["links"]
            self := links.(map[string]interface{})["self"]
            self_href := self.(map[string]interface{})["href"]
            self_href_link := fmt.Sprint(self_href)
    
            // Get more details about the PR
            json_response_det, err := getJSON(self_href_link)
            if err != nil {
                return err
            }
            
            // Get details about the participants
            prs_det := json_response_det["participants"]
            prs_det_i := prs_det.([]interface{})
    
            // For determining if the PR is ready to merge
            num_approved_reviewers := 0
            num_reviewers := 0
            
            // For each participant in the PR, display state depending on role
            for _, value := range prs_det_i {
                value_map := value.(map[string]interface{})
                role := value_map["role"]
                approved := value_map["approved"] == true
                display_name := value_map["user"].(map[string]interface{})["display_name"]
                
                // TODO Rewrite with one line RegEx?
                var approved_s = " "
                if approved {
                    approved_s = "X"
                }
    
                switch (role) {
                    case "REVIEWER":
                        fmt.Printf("        %s %s\n", approved_s, display_name)
                        num_reviewers++
                        if (approved) {
                            num_approved_reviewers++
                        }
                    case "PARTICIPANT":
                        fmt.Printf("        %s (%s)\n", approved_s, display_name)
                    default:
                        fmt.Printf("        %s %s (%s)\n", approved_s, display_name, role)
                }
            }
    
            var is_or_not = "IS NOT"
            if num_reviewers > 0 && num_reviewers == num_approved_reviewers {
                is_or_not = "IS"
            }
    
            fmt.Printf("    #%.0f %s READY TO MERGE, %d of %d REVIEWERS APPROVED\n\n", id, is_or_not, num_approved_reviewers, num_reviewers)
        }

        // Determine if there's more results - if so, loop control back
        next := json_response["next"]
        if next != nil {
            pr_api = fmt.Sprint(next)
        } else {
            pr_api = ""
        }
    }
    
    return nil
}

func init() {
    flag.StringVar(&bitbucket_owner_name, "ownername", "", "Bitbucket repository owner account")
    flag.StringVar(&bitbucket_username, "username", "", "Bitbucket account username")
    flag.StringVar(&bitbucket_password, "password", "", "Bitbucket account password")
}

func main() {
    // Command line args
    flag.Parse()
    if len(os.Args) != 4 {
        flag.Usage()
        os.Exit(1)
    }
    
    var repos_api = bb_base_url + "/repositories/" + bitbucket_owner_name
    
    // Repo API has pagination, code for > 1 page
    for len(repos_api) > 0 {
        // Get the list of repos for this user / group
        json_response, err := getJSON(repos_api)
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        
        repos := json_response["values"]
        repos_i := repos.([]interface{})
        
        // For each repo, get the PR URL and process
     	for _, value := range repos_i {
            links := value.(map[string]interface{})["links"]
            pullrequests := links.(map[string]interface{})["pullrequests"]
            pullrequests_href := pullrequests.(map[string]interface{})["href"]
    
            pull_requests_link := fmt.Sprint(pullrequests_href)
            fmt.Println("Repo:", pull_requests_link)
    
            err := listPR(pull_requests_link)
            if err != nil {
                 fmt.Println(err) // OK to continue here if error, no need to exit the program
            }
     	}
        
        // Determine if there's more results - if so, loop control back
        next := json_response["next"]
        if next != nil {
            repos_api = fmt.Sprint(next)
        } else {
            repos_api = ""
        }
    }
}
