package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v52/github"

	"golang.org/x/oauth2"
)

var auth *ssh.PublicKeys

const (
	GITHUB_TOKEN    = "GITHUB_TOKEN"
	GITHUB_ORG      = "GITHUB_ORG"
	CLONE_BASE_PATH = "CLONE_BASE_PATH"
	GIT_PRIVATE_KEY = "GIT_PRIVATE_REPO"
)

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: GITHUB_TOKEN})
	oc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(oc)

	InitGitClientWithPrivateKey(GIT_PRIVATE_KEY)

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}

	var repositories []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, GITHUB_ORG, opt)
		if err != nil {
			log.Panicf("error getting repository list from github. %v", err)
		}
		repositories = append(repositories, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	for _, repo := range repositories {
		if err := Clone(repo, CLONE_BASE_PATH); err != nil {
			log.Printf("error cloning repository. %v", err)
		}
	}
}

func InitGitClientWithPrivateKey(privkey string) {
	var err error
	auth, _ = ssh.NewPublicKeysFromFile("git", privkey, "")
	if err != nil {
		log.Panicf("error setup git authentication using ssh private key. %v", err)
	}
}

func Clone(repository *github.Repository, basepath string) error {
	targetPath := fmt.Sprintf("%s/%s", basepath, repository.GetName())
	log.Printf("cloning %s/%s to %s", repository.GetOwner().GetLogin(), repository.GetName(), targetPath)
	_, err := git.PlainClone(targetPath, false, &git.CloneOptions{
		Auth:     auth,
		URL:      repository.GetSSHURL(),
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}
	log.Printf("finished cloning %s/%s", repository.GetOwner().GetLogin(), repository.GetName())
	return nil
}
