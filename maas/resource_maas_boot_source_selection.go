package maas

import (
	"context"
	"encoding/json"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASBootSourceSelection() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a resource to manage a MAAS boot source selection.",
		ReadContext: resourceBootSourceSelectionRead,

		Schema: map[string]*schema.Schema{
			"arches": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The architecture list for this resource",
			},
			"boot_source_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The BootSource this resource is associated with",
			},
			"labels": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The label lists for this resource",
			},
			"os": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Operating system for this resource",
			},
			"release": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The specific release of the Operating system for this resource",
			},
			"subarches": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of subarches for this resource",
			},
		},
	}
}

func resourceBootSourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsourceselection, err := getBootSourceSelection(client, d.Get("boot_source_id").(int), d.Get("os").(string), d.Get("release").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]interface{}{
		"arches":         bootsourceselection.Arches,
		"boot_source_id": bootsourceselection.BootSourceID,
		"labels":         bootsourceselection.Labels,
		"os":             bootsourceselection.OS,
		"release":        bootsourceselection.Release,
		"subarches":      bootsourceselection.Subarches,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func fetchDefaultBootSourceSelection(client *client.Client) (*entity.BootSourceSelection, error) {
	// Fetch the default commissioning details
	var default_os string

	default_os_bytes, err := client.MAASServer.Get("default_osystem")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(default_os_bytes, &default_os)
	if err != nil {
		return nil, err
	}

	var default_release string

	default_release_bytes, err := client.MAASServer.Get("default_distro_Series")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(default_release_bytes, &default_release)
	if err != nil {
		return nil, err
	}

	// Only then fetch the default bootsource that refers to them

	bootsource, err := getBootSource(client)
	if err != nil {
		return nil, err
	}

	return getBootSourceSelection(client, bootsource.ID, default_os, default_release)
}
