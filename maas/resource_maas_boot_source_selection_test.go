package maas_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASBootSourceSelection_basic(t *testing.T) {

	var bootsourceselection entity.BootSourceSelection
	os := "ubuntu"
	release := "oracular"
	arches := []string{"amd64"}

	checks := []resource.TestCheckFunc{
		testAccMAASBootSourceSelectionCheckExists("maas_boot_source_selection.test", &bootsourceselection),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "os", os),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "release", release),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "arches.#", "1"),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "arches.0", arches[0]),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASBootSourceSelectionDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASBootSourceSelection(os, release, arches),
				Check:  resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	})
}

func testAccMAASBootSourceSelectionCheckExists(rn string, bootSourceSelection *entity.BootSourceSelection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}
		boot_source_id, err := strconv.Atoi(rs.Primary.Attributes["boot_source"])
		if err != nil {
			return err
		}
		gotBootSourceSelection, err := conn.BootSourceSelection.Get(boot_source_id, id)
		if err != nil {
			return fmt.Errorf("error getting boot source selection: %s", err)
		}

		*bootSourceSelection = *gotBootSourceSelection

		return fmt.Errorf("Boot selection created: %v", bootSourceSelection.ID)

		return nil
	}
}

func testAccMAASBootSourceSelection(os string, release string, arches []string) string {
	return fmt.Sprintf(`
resource "maas_boot_source" "test" {
	url 			 = "http://images.maas.io/ephemeral-v3/stable/"
	keyring_filename = "/snap/maas/current/usr/share/keyrings/ubuntu-cloudimage-keyring.gpg"
}

resource "maas_boot_source_selection" "test" {
	boot_source = maas_boot_source.test.id

	os      = "%s"
	release = "%s"
	arches  = ["%s"]
}
`, os, release, arches[0])
}

func testAccCheckMAASBootSourceSelectionDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// Fetch the default commissioning details
	var default_release string
	default_release_bytes, err := conn.MAASServer.Get("default_distro_series")
	if err != nil {
		return err
	}
	err = json.Unmarshal(default_release_bytes, &default_release)
	if err != nil {
		return err
	}

	// loop through the resources in state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_boot_source_selection" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err

		}
		boot_source_id, err := strconv.Atoi(rs.Primary.Attributes["boot_source"])
		if err != nil {
			return err
		}

		response, err := conn.BootSourceSelection.Get(boot_source_id, id)
		if err != nil {
			return err
		}
		if err == nil {
			// default boot source selection
			if response.OS == "ubuntu" && response.Release == default_release {
				if len(response.Arches) != 1 || response.Arches[0] != "amd64" {
					return fmt.Errorf("MAAS Boot Source Selection (%s) Arches not reset to default. Returned value: %s", rs.Primary.ID, response.Arches)
				}
				if len(response.Subarches) != 1 || response.Subarches[0] != "*" {
					return fmt.Errorf("MAAS Boot Source Selection (%s) Subarches not reset to default. Returned value: %s", rs.Primary.ID, response.Subarches)
				}
				if len(response.Labels) != 1 || response.Labels[0] != "*" {
					return fmt.Errorf("MAAS Boot Source Selection (%s) Labels not reset to default. Returned value: %s", rs.Primary.ID, response.Labels)
				}
			} else if response != nil && response.ID == id {
				return fmt.Errorf("MAAS Boot Source Selection (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		return err
	}

	return nil
}
