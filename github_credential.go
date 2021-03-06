package main

import(
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/term"
)

type GitHubCredential struct {
	User string
	Token string
}

var HomeDir, _ = os.UserHomeDir()

func LoadGitHubCredential() *GitHubCredential {
	credentialFilePath := filepath.Join(HomeDir, ".toybox_github_credentials")

	if ((os.Getenv("GITHUB_USER") != "") && (os.Getenv("GITHUB_TOKEN") != "")) {
		return &GitHubCredential{os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_TOKEN")}
	} else if _, err := os.Stat(credentialFilePath); err == nil {
		jsonString, err := os.ReadFile(credentialFilePath)
    
	    if err != nil {
			ReportError("Error locating GitHub credentials", err, true)
	    }

		parsedCredential := GitHubCredential{}
    	err = json.Unmarshal([]byte(jsonString), &parsedCredential)

    	if err != nil {
			ReportError("Error parsing GitHub credentials", err, true)
    	}

    	return &parsedCredential
	}

	return nil
}

func RequestGitHubCredential() {
	Print("This process will store a Personal Access Token locally for making GitHub API calls.\n")
	Print("Follow the instructions here:")
	Print("https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token\n")
	Print("When creating the token, set a sensible expiration (60 days?)")
	Print("and it ONLY needs the repo:public_repo scope.\n")

	fmt.Printf("        GitHub username: ")
	username := ""
	fmt.Scanln(&username)

	fmt.Printf("        Personal Access Token: ")
	token, _ := term.ReadPassword(0)
	
	credential := GitHubCredential{username, string(token)}
	
	credentialFilePath := filepath.Join(HomeDir, ".toybox_github_credentials")
	file, _ := json.MarshalIndent(credential, "", " ")
	_ = ioutil.WriteFile(credentialFilePath, file, 0644)

	fmt.Println("\n")
	ReportProgress(fmt.Sprintf("👍 Credential stored in %s.", credentialFilePath))
}

func LoginOrRenewGitHubCredential() {
	if credential := LoadGitHubCredential(); credential != nil {
		Print(fmt.Sprintf("Logged in to GitHub as %s.", credential.User))
		fmt.Printf("        Update stored credentials? (y/n) ")
		
		answer := "n"
		fmt.Scanln(&answer)

		if answer == "y" {
			fmt.Println()
			RequestGitHubCredential()	
		}
	} else {
		RequestGitHubCredential()
	}
}