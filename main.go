package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/urfave/cli/v2"
)

func StringPrompt(prompt string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s: ", prompt)
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func getRelease(ctx context.Context, client *github.Client, owner, repo, tag string) (*github.RepositoryRelease, error) {
	release, resp, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("release %s not found", tag)
		}

		return nil, fmt.Errorf("failed to get release %s: %w", tag, err)
	}

	return release, nil
}

func getAllReleases(ctx context.Context, client *github.Client, owner, repo string) ([]*github.RepositoryRelease, error) {
	var allReleases []*github.RepositoryRelease
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		releases, resp, err := client.Repositories.ListReleases(ctx, owner, repo, opts)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				return nil, fmt.Errorf("repository %s/%s not found", owner, repo)
			}

			return nil, fmt.Errorf("failed to list releases: %w", err)
		}
		allReleases = append(allReleases, releases...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allReleases, nil
}

func updateReleaseBody(ctx context.Context, client *github.Client, release *github.RepositoryRelease, owner, repo, value, replacement string) error {
	body := strings.ReplaceAll(release.GetBody(), value, replacement)
	if strings.Compare(body, release.GetBody()) == 0 {
		fmt.Printf("%s: No changes to release body found, skipping\n", release.GetTagName())
		return nil
	}

	release.Body = &body
	fmt.Printf("%s: Updating release body\n", release.GetTagName())
	_, _, err := client.Repositories.EditRelease(ctx, owner, repo, release.GetID(), release)
	if err != nil {
		return fmt.Errorf("%s: failed to update release body: %w", release.GetTagName(), err)
	}

	return nil
}

func run(ctx *cli.Context) error {
	owner := ctx.String("owner")
	repo := ctx.String("repo")
	value := ctx.String("value")
	replacement := ctx.String("replacement")
	token := ctx.String("token")
	all := ctx.Bool("all")
	tag := ctx.String("release")

	if all && tag != "" {
		return fmt.Errorf("cannot specify both --all and --release")
	}

	if !all && tag == "" {
		return fmt.Errorf("must specify either --all or --release")
	}

	if all {
		response := StringPrompt("Are you sure you want to update all releases? (y/n)")
		if response != "y" {
			fmt.Println("Invalid response, aborting")
			return nil
		}

		ctx := context.Background()
		client := github.NewClient(nil).WithAuthToken(token)
		fmt.Println("Fetching all releases")
		releases, err := getAllReleases(ctx, client, owner, repo)
		if err != nil {
			return err
		}

		for _, release := range releases {
			fmt.Printf("%s: Checking for updates to release body\n", release.GetTagName())
			err := updateReleaseBody(ctx, client, release, owner, repo, value, replacement)
			if err != nil {
				return err
			}
		}
	}

	if tag != "" {
		response := StringPrompt(fmt.Sprintf("Are you sure you want to update release %s? (y/n)", tag))
		if response != "y" {
			fmt.Println("Invalid response, aborting")
			return nil
		}

		ctx := context.Background()
		client := github.NewClient(nil).WithAuthToken(token)
		release, err := getRelease(ctx, client, owner, repo, tag)
		if err != nil {
			return err
		}

		return updateReleaseBody(ctx, client, release, owner, repo, value, replacement)
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:                 "update-release",
		Usage:                "Update strings in GitHub releases",
		Version:              "1.0.0",
		EnableBashCompletion: true,
		Suggest:              true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "owner",
				Aliases:  []string{"o"},
				Usage:    "GitHub repository organization or owner",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "repo",
				Aliases:  []string{"r"},
				Usage:    "GitHub repository name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "value",
				Aliases:  []string{"l"},
				Usage:    "Value to replace",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "replacement",
				Aliases:  []string{"p"},
				Usage:    "Replacement value",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "token",
				Aliases:  []string{"t"},
				Usage:    "GitHub personal access token",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Update all releases",
			},
			&cli.StringFlag{
				Name:    "release",
				Aliases: []string{"s"},
				Usage:   "Update a specific release",
			},
		},
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
