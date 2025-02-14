package maas

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
		Importer: &schema.ResourceImporter{
			State: resourceBootSourceSelectionImport,
		},

		Schema: map[string]*schema.Schema{
			"arches": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "The architecture list for this resource",
			},
			"boot_source": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The BootSource database ID this resource is associated with",
			},
			"labels": {
				Type:        schema.TypeSet,
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
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "The list of subarches for this resource",
			},
		},
	}
}

func resourceBootSourceSelectionImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected BOOT_SOURCE:BOOT_SOURCE_SELECTION_ID", d.Id())
	}

	d.Set("boot_source", idParts[0])
	d.SetId(idParts[1])

	return []*schema.ResourceData{d}, nil
}

func resourceBootSourceSelectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsourceselectionParams := entity.BootSourceSelectionParams{
		OS:        d.Get("os").(string),
		Release:   d.Get("release").(string),
		Arches:    convertToStringSlice(d.Get("arches").(*schema.Set).List()),
		Subarches: convertToStringSlice(d.Get("subarches").(*schema.Set).List()),
		Labels:    convertToStringSlice(d.Get("labels").(*schema.Set).List()),
	}

	bootsourceselection, err := client.BootSourceSelections.Create(d.Get("boot_source").(int), &bootsourceselectionParams)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId((fmt.Sprintf("%v", bootsourceselection.ID)))

	return resourceBootSourceSelectionRead(ctx, d, meta)
}

func resourceBootSourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceselection, err := getBootSourceSelection(client, d.Get("boot_source").(int), id)
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]interface{}{
		"arches":      bootsourceselection.Arches,
		"boot_source": bootsourceselection.BootSourceID,
		"labels":      bootsourceselection.Labels,
		"os":          bootsourceselection.OS,
		"release":     bootsourceselection.Release,
		"subarches":   bootsourceselection.Subarches,
	}
	d.SetId((fmt.Sprintf("%v", bootsourceselection.ID)))

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceBootSourceSelectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceselection, err := getBootSourceSelection(client, d.Get("boot_source").(int), id)
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceselectionParams := entity.BootSourceSelectionParams{
		OS:        d.Get("os").(string),
		Release:   d.Get("release").(string),
		Arches:    convertToStringSlice(d.Get("arches").(*schema.Set).List()),
		Subarches: convertToStringSlice(d.Get("subarches").(*schema.Set).List()),
		Labels:    convertToStringSlice(d.Get("labels").(*schema.Set).List()),
	}

	if _, err := client.BootSourceSelection.Update(bootsourceselection.BootSourceID, bootsourceselection.ID, &bootsourceselectionParams); err != nil {
		return diag.FromErr(err)
	}

	return resourceBootSourceSelectionRead(ctx, d, meta)
}

func resourceBootSourceSelectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceselection, err := getBootSourceSelection(client, d.Get("boot_source").(int), id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Fetch the default commissioning details
	var default_release string
	default_release_bytes, err := client.MAASServer.Get("commissioning_distro_series")
	if err != nil {
		return diag.FromErr(err)
	}
	err = json.Unmarshal(default_release_bytes, &default_release)
	if err != nil {
		return diag.FromErr(err)
	}

	// if the boot source selection is the default, we need to treat it differently
	if bootsourceselection.OS == "ubuntu" && bootsourceselection.Release == default_release {
		bootsourceselectionParams := entity.BootSourceSelectionParams{
			OS:        "ubuntu",
			Release:   default_release,
			Arches:    []string{"amd64"},
			Subarches: []string{"*"},
			Labels:    []string{"*"},
		}
		if _, err := client.BootSourceSelection.Update(bootsourceselection.BootSourceID, bootsourceselection.ID, &bootsourceselectionParams); err != nil {
			return diag.FromErr(err)
		}
	} else {
		return diag.FromErr(client.BootSourceSelection.Delete(bootsourceselection.BootSourceID, bootsourceselection.ID))
	}

	return nil
}

func getBootSourceSelection(client *client.Client, boot_source int, id int) (*entity.BootSourceSelection, error) {
	bootsourceselection, err := client.BootSourceSelection.Get(boot_source, id)
	if err != nil {
		return nil, err
	}
	if bootsourceselection == nil {
		return nil, fmt.Errorf("boot source selection (%v %v) was not found", boot_source, id)
	}
	return bootsourceselection, nil
}
