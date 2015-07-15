package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
	"strings"
)

type GithubContext struct {
	orgName       string
	accessToken   string
	gitUser       string
	exclusionFile string
}

type Config struct {
	DbPath                string           `json:"dbpath"`
	Repos                 map[string]*Repo `json:"repos"`
	MaxConcurrentIndexers int              `json:"max-concurrent-indexers"`
}

type Repo struct {
	Url    string `json:"url"`
	Branch string `json:"branch"`
}

func loadExclusions(file string) map[string]bool {
	var excluded map[string]bool = make(map[string]bool)

	if file == "" {
		// not an error, no need to warn - just skip
		return excluded
	}

	inFile, err := os.Open(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading exclusion file; skipping:", err)
		return excluded
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		// skip empty lines and comments
		if len(line) != 0 && line[0] != '#' {
			excluded[line] = true
		}
	}

	return excluded
}

func getFriendlyName(in string) string {
	// <in> has the form https://github.com/YikYakApp/tools.git
	return in[strings.LastIndex(in, "/")+1 : strings.LastIndex(in, ".git")]
}

func getAuthURL(url string, ctx *GithubContext) string {
	// insert github credentials into the access url
	// <url> has the form https://github.com/YikYakApp/tools.git
	return "https://" + ctx.gitUser + ":" + ctx.accessToken + "@" + url[8:]
}

func buildContext(ctx *GithubContext) error {
	flag.StringVar(&ctx.orgName, "org", "", "Github org name")
	flag.StringVar(&ctx.gitUser, "user", "", "Github user name")
	flag.StringVar(&ctx.accessToken, "token", "", "AccessToken")
	flag.StringVar(&ctx.exclusionFile, "excl", "", "File with repos to exclude")
	flag.Parse()

	if len(ctx.orgName)*len(ctx.gitUser)*len(ctx.accessToken) == 0 {
		return errors.New("Must specify -org, -user, and -token")
	}
	return nil

}

func main() {

	var ghCtx GithubContext
	err := buildContext(&ghCtx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	excludeList := loadExclusions(ghCtx.exclusionFile)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghCtx.accessToken},
	)

	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	// go-github can only fetch 100 repos at a time, so we need to paginate
	opt := &github.RepositoryListByOrgOptions{
		Type:        "private",
		ListOptions: github.ListOptions{PerPage: 100, Page: 1},
	}

	// fetch & build the repo list
	allRepos := make(map[string]*Repo)

	for {
		repos, resp, err := client.Repositories.ListByOrg(ghCtx.orgName, opt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching repo info:", err)
			os.Exit(1)
		}

		for _, r := range repos {
			// unless the repo is in our exclusion list, add it to the config
			repoName := getFriendlyName(*r.CloneURL)
			if !excludeList[repoName] {
				allRepos[repoName] = &Repo{Url: getAuthURL(*r.CloneURL, &ghCtx), Branch: "master"}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	// pretty print the config
	conf := Config{DbPath: "data", MaxConcurrentIndexers: 2, Repos: allRepos}
	out, err := json.MarshalIndent(conf, "  ", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error marshaling json:", err)
		os.Exit(1)
	}
	os.Stdout.Write(out)
}
