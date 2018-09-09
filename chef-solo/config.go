package chefsolo

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

type provisioner struct {
	ChefEnvironment  string
	ConfigTemplate   string
	CookbookPaths    []string
	EnvironmentsPath string
	ExecuteCommand   string
	InstallCommand   string
	GuestOSType      string
	JSON             map[string]interface{}
	PreventSudo      bool
	RunList          []string
	SkipInstall      bool
	RolesPath        string
	StagingDirectory string
	Version          string

	createDirCommand string
}

// Provisioner returns a Chef Solo provisioner
func Provisioner() terraform.ResourceProvisioner {
	return &schema.Provisioner{
		Schema: map[string]*schema.Schema{
			"chef_environment": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"config_template": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"cookbook_paths": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"environments_path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"execute_command": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"guest_os_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"install_command": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"json": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"prevent_sudo": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"roles_path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"run_list": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"staging_directory": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"skip_install": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"version": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ApplyFunc:    applyFn,
		ValidateFunc: func(c *terraform.ResourceConfig) (ws []string, es []error) { return },
	}
}

// takes the data from the provisioner schema
func decodeConfig(d *schema.ResourceData) (*provisioner, error) {
	p := &provisioner{
		ChefEnvironment:  d.Get("chef_environment").(string),
		ConfigTemplate:   d.Get("config_template").(string),
		CookbookPaths:    getStringList(d.Get("cookbook_paths")),
		EnvironmentsPath: d.Get("environments_path").(string),
		ExecuteCommand:   d.Get("execute_command").(string),
		GuestOSType:      d.Get("guest_os_type").(string),
		InstallCommand:   d.Get("install_command").(string),
		PreventSudo:      d.Get("prevent_sudo").(bool),
		RunList:          getStringList(d.Get("run_list")),
		RolesPath:        d.Get("roles_path").(string),
		StagingDirectory: d.Get("staging_directory").(string),
		SkipInstall:      d.Get("skip_install").(bool),
		Version:          d.Get("version").(string),
	}
	if unparsed, ok := d.GetOk("json"); ok {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(unparsed.(string)), &parsed); err != nil {
			return nil, fmt.Errorf("Error parsing `json`: %#v", err)
		}
		p.JSON = parsed
	}
	return p, nil
}
