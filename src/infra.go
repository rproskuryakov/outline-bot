package src


import (
    "path/filepath"
    "context"
    "os"
    "log"

    "github.com/hashicorp/go-version"
    "github.com/hashicorp/hc-install/product"
    "github.com/hashicorp/hc-install/releases"
    jsii "github.com/aws/jsii-runtime-go"
    "github.com/hashicorp/terraform-exec/tfexec"
//     tfjson "github.com/hashicorp/terraform-json"
    "github.com/hashicorp/terraform-cdk-go/cdktf"
    digitalocean "github.com/cdktf/cdktf-provider-digitalocean-go/digitalocean/v13/provider"
    "github.com/aws/constructs-go/constructs/v10"
)


func NewStack(scope constructs.Construct, id string, apiToken string) cdktf.TerraformStack {
    stack := cdktf.NewTerraformStack(scope, &id)

    // config provider
    digitalocean.NewDigitaloceanProvider(
        scope,
        &id,
        &digitalocean.DigitaloceanProviderConfig{
            Token: jsii.String(apiToken),
        },
    )

//     droplet := digitalocean.NewDroplet(
//         scope,
//         &id,
//         &digitalocean.DropletConfig{
//             Image: "123",
//             Name: "123",
//             Size: "123",
//             SshKeys: []*string{},
//             Region: "AMS3",
//         },
//     )

    return stack
}


func ExecuteTerraform(ctx context.Context) (error){
    name := "outline-bot-infra"
    digitalocean_api_token := "12345"
    tempDir, err := os.MkdirTemp("", "outline-bot-infra-")
    if err != nil {
        return err
    }
    defer os.RemoveAll(tempDir)

    app := cdktf.NewApp(nil)
    NewStack(app, name, digitalocean_api_token)
    app.Synth()

    installer := &releases.ExactVersion{
        Product: product.Terraform,
        Version: version.Must(version.NewVersion("1.1.4")),
    }

    execPath, err := installer.Install(context.Background())
    if err != nil {
        log.Fatalf("error installing Terraform: %s", err)
        return err
    }

    workingDir := filepath.Join(tempDir, "stacks", name)
    tf, err := tfexec.NewTerraform(workingDir, execPath)
    if err != nil {
        log.Fatalf("error running NewTerraform: %s", err)
        return err
    }

    err = tf.Init(context.Background(), tfexec.Upgrade(true))
    if err != nil {
        log.Fatalf("error running Init: %s", err)
        return err
    }

    err = tf.Apply(context.Background())
    if err != nil {
        log.Fatalf("error running Apply: %s", err)
        return err
    }
    return nil
}