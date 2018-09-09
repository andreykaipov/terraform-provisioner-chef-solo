package main

import (
	"github.com/andreykaipov/terraform-provisioner-chef-solo/chef-solo"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: chefsolo.Provisioner,
	})
}
