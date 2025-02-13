package maas

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASBootSourceSelection() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage a MAAS boot source selection.",
		CreateContext: resourceBootSourceSelectionCreate,
		ReadContext:   resourceBootSourceSelectionRead,
		UpdateContext: resourceBootSourceSelectionUpdate,
		DeleteContext: resourceBootSourceSelectionDelete,

		Schema: map[string]*schema.Schema{
			"arches": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "The architecture list for this resource",
			},
			"boot_source_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The BootSource this resource is associated with",
			},
			"labels": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
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
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "The list of subarches for this resource",
			},
		},
	}
}

func resourceBootSourceSelectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceselectionParams := entity.BootSourceSelectionParams{
		OS:        d.Get("os").(string),
		Release:   d.Get("release").(string),
		Arches:    sliceToString(d.Get("arches")),
		Subarches: sliceToString(d.Get("subarches")),
		Labels:    sliceToString(d.Get("labels")),
	}

	bootsourceselection, err := client.BootSourceSelections.Create(bootsource.ID, &bootsourceselectionParams)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId((fmt.Sprintf("%v", bootsourceselection.ID)))

	return resourceBootSourceSelectionRead(ctx, d, meta)
}

func resourceBootSourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceselection, err := getBootSourceSelection(client, bootsource.ID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]interface{}{
		"arches":         bootsourceselection.Arches,
		"boot_source_id": bootsource.ID,
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
func resourceBootSourceSelectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceselection, err := getBootSourceSelection(client, bootsource.ID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceselectionParams := entity.BootSourceSelectionParams{
		OS:        d.Get("os").(string),
		Release:   d.Get("release").(string),
		Arches:    sliceToString(d.Get("arches")),
		Subarches: sliceToString(d.Get("subarches")),
		Labels:    sliceToString(d.Get("labels")),
	}

	if _, err := client.BootSourceSelection.Update(bootsource.ID, bootsourceselection.ID, &bootsourceselectionParams); err != nil {
		return diag.FromErr(err)
	}

	return resourceBootSourceSelectionRead(ctx, d, meta)
}

func resourceBootSourceSelectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceselection, err := getBootSourceSelection(client, bootsource.ID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	// if the boot source selection is the default, we need to treat it differently
	defaultbootsourceselection, err := fetchDefaultBootSourceSelection(client)
	if err != nil {
		return diag.FromErr(err)
	}

	if defaultbootsourceselection.BootSourceID == bootsourceselection.BootSourceID && defaultbootsourceselection.ID == bootsourceselection.ID {
		bootsourceselectionParams := entity.BootSourceSelectionParams{
			OS:        defaultbootsourceselection.OS,
			Release:   defaultbootsourceselection.Release,
			Arches:    []string{"amd64"},
			Subarches: []string{"*"},
			Labels:    []string{},
		}
		if _, err := client.BootSourceSelection.Update(bootsource.ID, bootsourceselection.ID, &bootsourceselectionParams); err != nil {
			return diag.FromErr(err)
		}
	} else {
		return diag.FromErr(client.BootSourceSelection.Delete(bootsource.ID, bootsourceselection.ID))
	}

	return nil
}

func sliceToString(v interface{}) []string {
	if v == nil {
		return nil
	}
	list, ok := v.([]interface{})
	if !ok {
		return nil
	}
	output_list := make([]string, len(list))
	for i, item := range list {
		output_list[i], _ = item.(string)
	}
	return output_list
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

	default_release_bytes, err := client.MAASServer.Get("default_distro_series")
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

	return getBootSourceSelectionByRelease(client, bootsource.ID, default_os, default_release)
}
