package src

import (
    "testing"
	"context"
	"fmt"
	"os"

	getter "github.com/hashicorp/go-getter"
	"github.com/testcontainers/testcontainers-go/modules/compose"
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

func createDockerComposeStack(dockerComposeContent string) (*stack compose.ComposeStack) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    stack, err := compose.NewDockerComposeWith(compose.WithStackReaders(strings.NewReader(dockerComposeContent)))
    if err != nil {
        log.Printf("Failed to create stack: %v", err)
        return
    }

    err = stack.
        WithEnv(map[string]string{
            "bar": "BAR",
        }).
        WaitForService("nginx", wait.NewHTTPStrategy("/").WithPort("80/tcp").WithStartupTimeout(10*time.Second)).
        Up(ctx, compose.Wait(true))
    if err != nil {
        log.Printf("Failed to start stack: %v", err)
        return
    }
    defer func() {
        err = stack.Down(
            context.Background(),
            compose.RemoveOrphans(true),
            compose.RemoveVolumes(true),
            compose.RemoveImagesLocal,
        )
        if err != nil {
            log.Printf("Failed to stop stack: %v", err)
        }
    }()
    return &stack
//     serviceNames := stack.Services()
//
//     fmt.Println(serviceNames)
}


func main() {
    destDir := "tmp/gogetter"
    downloadGithubRepository("github.com/Jigsaw-Code/outline-server/", destDir)
}