package chefsolo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var osDefaults = map[string]provisioner{
	"unix": {
		PreventSudo:      false,
		StagingDirectory: "/tmp/terraform-chef-solo",
		InstallCommand:   "sh -c 'command -v chef-solo || (curl -LO https://omnitruck.chef.io/install.sh && sh install.sh{{if .Version}} -v {{.Version}}{{end}})'",
		ExecuteCommand:   "chef-solo --no-color -c {{.StagingDirectory}}/solo.rb -j {{.StagingDirectory}}/attributes.json",
		createDirCommand: "sh -c 'mkdir -p %q; chmod 777 %q'",
	},
	"windows": {
		PreventSudo:      true,
		StagingDirectory: "C:/Windows/Temp/packer-chef-solo",
		InstallCommand:   "powershell.exe -Command \". { iwr -useb https://omnitruck.chef.io/install.ps1 } | iex; Install-Project{{if .Version}} -version {{.Version}}{{end}}\"",
		ExecuteCommand:   "C:/opscode/chef/bin/chef-solo.bat --no-color -c {{.StagingDirectory}}/solo.rb -j {{.StagingDirectory}}/attributes.json",
		createDirCommand: "cmd /c if not exist %q mkdir %q",
	},
}

type soloRb struct {
	CookbookPaths string
	RolesPath     string
}

// defaultConfigTemplate
var defaultSoloRbTemplate = `
cookbook_path     [{{.CookbookPaths}}]
{{ if (not (eq .RolesPath "")) }}
role_path         "{{.RolesPath}}"
{{end}}
`

func applyFn(ctx context.Context) error {
	o := ctx.Value(schema.ProvOutputKey).(terraform.UIOutput)
	s := ctx.Value(schema.ProvRawStateKey).(*terraform.InstanceState)
	data := ctx.Value(schema.ProvConfigDataKey).(*schema.ResourceData)

	comm, err := getCommunicator(ctx, o, s)
	if err != nil {
		return err
	}

	// Decode the provisioner config
	p, err := decodeConfig(data)
	if err != nil {
		return err
	}

	// Find the OS based on the instance state
	p.GuestOSType, err = getGuestOSType(s)
	if err != nil {
		return err
	}

	// Setup based on OS
	setIfEmpty(&p.PreventSudo, osDefaults[p.GuestOSType].PreventSudo)
	setIfEmpty(&p.StagingDirectory, osDefaults[p.GuestOSType].StagingDirectory)
	setIfEmpty(&p.InstallCommand, osDefaults[p.GuestOSType].InstallCommand)
	setIfEmpty(&p.ExecuteCommand, osDefaults[p.GuestOSType].ExecuteCommand)

	p.InstallCommand = renderTemplate(p.InstallCommand, p)
	p.ExecuteCommand = renderTemplate(p.ExecuteCommand, p)

	o.Output("Creating configuration files...")
	p.createDir(o, comm, p.StagingDirectory)
	p.uploadCookbooks(o, comm)
	p.createAndUploadJSONAttributes(o, comm)
	p.createAndUploadSoloRb(o, comm)

	o.Output("Installing Chef-Solo...")
	if !p.SkipInstall {
		if err := p.runCommand(o, comm, p.InstallCommand); err != nil {
			return fmt.Errorf("Error installing Chef: %v", err)
		}
	}

	o.Output("Starting Chef-Solo...")
	if err := p.runCommand(o, comm, p.ExecuteCommand); err != nil {
		return fmt.Errorf("Error executing Chef: %v", err)
	}

	return nil
}

// maps the local cookbook paths to remote cookbook paths
func (p *provisioner) getRemoteCookbookPaths() []string {
	remoteCookbookPaths := make([]string, 0, len(p.CookbookPaths))
	for i := range p.CookbookPaths {
		remotePath := fmt.Sprintf("%s/cookbooks-%d", p.StagingDirectory, i)
		remoteCookbookPaths = append(remoteCookbookPaths, remotePath)
	}
	return remoteCookbookPaths
}

// uploads the cookbooks from the local cookbook paths to remote cookbook paths
func (p *provisioner) uploadCookbooks(o terraform.UIOutput, comm communicator.Communicator) error {
	for i, remotePath := range p.getRemoteCookbookPaths() {
		localPath := p.CookbookPaths[i]
		if err := p.uploadDir(o, comm, remotePath, localPath); err != nil {
			return fmt.Errorf("Error uploading cookbooks: %v", err)
		}
	}
	return nil
}

// get the node attributes, add the `run_list` if it's specified, and upload it
func (p *provisioner) createAndUploadJSONAttributes(o terraform.UIOutput, comm communicator.Communicator) error {
	o.Output("Creating Chef JSON attributes file...")

	jsonData := make(map[string]interface{})
	for k, v := range p.JSON {
		jsonData[k] = v
	}
	if len(p.RunList) > 0 {
		jsonData["run_list"] = p.RunList
	}
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return err
	}
	remotePath := filepath.ToSlash(filepath.Join(p.StagingDirectory, "attributes.json"))
	if err := comm.Upload(remotePath, bytes.NewReader(jsonBytes)); err != nil {
		return fmt.Errorf("Error creating the Chef JSON attributes file: %v", err)
	}
	return nil
}

// creates and uploads the solo.rb config file for Chef Solo to use
func (p *provisioner) createAndUploadSoloRb(o terraform.UIOutput, comm communicator.Communicator) error {
	o.Output("Creating solo.rb config filed...")

	quotedRemotePaths := make([]string, 0, len(p.CookbookPaths))
	for _, remotePath := range p.getRemoteCookbookPaths() {
		quotedRemotePath := fmt.Sprintf(`"%s"`, remotePath)
		quotedRemotePaths = append(quotedRemotePaths, quotedRemotePath)
	}

	soloRbConfig := renderTemplate(defaultSoloRbTemplate, &soloRb{
		CookbookPaths: strings.Join(quotedRemotePaths, ","),
		RolesPath:     p.RolesPath,
	})

	remoteSoloRbPath := filepath.ToSlash(filepath.Join(p.StagingDirectory, "solo.rb"))
	if err := comm.Upload(remoteSoloRbPath, strings.NewReader(soloRbConfig)); err != nil {
		return fmt.Errorf("Error creating the solo.rb file: %v", err)
	}
	return nil
}
