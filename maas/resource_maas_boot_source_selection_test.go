package maas_test

import (
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
	release := "focal"

	checks := []resource.TestCheckFunc{
		testAccMAASBootSourceSelectionCheckExists("maas_boot_source_selection.test", &bootsourceselection),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "os", os),
		resource.TestCheckResourceAttr("maas_boot_source_selection.test", "release", release),
		resource.TestCheckResourceAttrSet("maas_boot_source_selection.test", "arches"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASBootSourceSelection(os, release),
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
		boot_source_id, err := strconv.Atoi(rs.Primary.Attributes["boot_source_id"])
		if err != nil {
			return err
		}
		gotBootSourceSelection, err := conn.BootSourceSelection.Get(boot_source_id, id)
		if err != nil {
			return fmt.Errorf("error getting boot source selection: %s", err)
		}

		*bootSourceSelection = *gotBootSourceSelection

		return nil
	}
}

func testAccMAASBootSourceSelection(os string, release string) string {
	return fmt.Sprintf(`
data "maas_boot_source" "test" {
	url = "http://images.maas.io/ephemeral-v3/stable/"
}
data "maas_boot_source_selection" "test" {
	boot_source_id = maas_boot_source.test.id

	os      = "%s"
	release = "%s"
}
`, os, release)
}
