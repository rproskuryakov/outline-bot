package src

import (
    "testing"
	"context"
	"fmt"
	"os"

	getter "github.com/hashicorp/go-getter"
)

func downloadGithubRepository(url string, path string) {
    client := &getter.Client{
		Ctx: context.Background(),
		//define the destination to where the directory will be stored. This will create the directory if it doesnt exist
// 	    Dst: "/tmp/gogetter",
		Dst: path,
		Dir: true,
		//the repository with a subdirectory I would like to clone only
		Src:  url,
// 		Src:  "github.com/hashicorp/terraform/examples/cross-provider",
		Mode: getter.ClientModeDir,
		//define the type of detectors go getter should use, in this case only github is needed
		Detectors: []getter.Detector{
			&getter.GitHubDetector{},
		},
		//provide the getter needed to download the files
		Getters: map[string]getter.Getter{
			"git": &getter.GitGetter{},
		},
	}
	//download the files
	if err := client.Get(); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting path %s: %v", client.Src, err)
		os.Exit(1)
	}
}