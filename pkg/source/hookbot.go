package source

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"regexp"
)

var (
	hookbotGithostRe = regexp.MustCompile("^/sub/([^/]+)/repo/([^/]+)/([^/]+)" +
		"/branch/([^/#]+)(?:#(.*))?$")
	hookbotDockerPullSub = regexp.MustCompile("^/sub/docker-pull/(.*)/tag/([^/]+)$")
	hookbotCwdRe         = regexp.MustCompile("^/sub/hanoverd/cwd$")
)

func GetSourceFromHookbot(hookbotURLStr string) (string, ImageSource, error) {

	hookbotURL, err := url.Parse(hookbotURLStr)
	if err != nil {
		return "", nil, fmt.Errorf("Hookbot URL %q does not parse: %v",
			hookbotURLStr, err)
	}

	switch {
	case hookbotGithostRe.MatchString(PathWithFragment(hookbotURL)):
		return NewGitHostSource(hookbotURL)

	case hookbotDockerPullSub.MatchString(hookbotURL.Path):
		return NewDockerPullSource(hookbotURL)

	case hookbotCwdRe.MatchString(hookbotURL.Path):
		return "cwd", &CwdSource{}, nil
	}

	return "", nil, fmt.Errorf("Unrecogized hookbot URL %q", hookbotURL.Path)
}

// Represent the path as /foo or /foo#bar if #bar is specified.
func PathWithFragment(u *url.URL) string {
	pathWithFragment := u.Path
	if u.Fragment != "" {
		pathWithFragment += "#" + u.Fragment
	}
	return pathWithFragment
}

func NewGitHostSource(hookbotURL *url.URL) (string, ImageSource, error) {

	groups := hookbotGithostRe.FindStringSubmatch(PathWithFragment(hookbotURL))
	host, user, repository, branch := groups[1], groups[2], groups[3], groups[4]
	imageRoot := groups[5]

	imageSource := &GitHostSource{
		Host:          host,
		User:          user,
		Repository:    repository,
		InitialBranch: branch,
		ImageRoot:     imageRoot,
	}

	log.Printf("Hookbot monitoring %v@%v via %v (imageroot %q)",
		repository, branch, hookbotURL.Host, imageRoot)

	return repository, imageSource, nil
}

func NewDockerPullSource(hookbotURL *url.URL) (string, ImageSource, error) {

	groups := hookbotDockerPullSub.FindStringSubmatch(hookbotURL.Path)
	repository, tag := groups[1], groups[2]

	imageSource := &DockerPullSource{
		Repository: repository,
		Tag:        tag,
	}

	log.Printf("Hookbot monitoring %v@%v via %v",
		repository, tag, hookbotURL.Host)

	containerName := path.Base(repository)
	return containerName, imageSource, nil
}
