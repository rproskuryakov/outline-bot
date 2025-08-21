package infra


import (
    "github.com/hashicorp/terraform-cdk-go/cdktf"
    "github.com/hashicorp/cdktf-provider-google-go/google"
    "github.com/hashicorp/cdktf-provider-digitalocean-go/digitalocean"
    "github.com/hashicorp/cdktf-provider-aws-go/aws"
    "github.com/aws/constructs-go/constructs/v10"
)

const (
    digitalocean_api_token := "12345"
)

func NewStack(scope constructs.Construct, id string) cdktf.TerraformStack {
    stack := cdktf.NewTerraformStack(scope, &id)

    // config provider
    digitalocean.NewDigitalOceanProvider(
        scope,
        &id,
        &digitalocean.DigitaloceanProviderConfig{
            Token: digitalocean_api_token,
        }
    )

    droplet := digitalocean.NewDroplet(
        scope,
        &id,
        &digitalocean.DropletConfig{
            Image: "123",
            Name: "123",
            Size: "123",
            SshKeys: [],
            Region: "AMS3"
        }
    )
}