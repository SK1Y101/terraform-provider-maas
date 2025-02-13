package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMaasBootSourceSelection_basic(t *testing.T) {
	os := "ubuntu"
	release := "noble"

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("data.maas_boot_source_selection.test", "arches.#"),
		resource.TestCheckResourceAttrSet("data.maas_boot_source_selection.test", "boot_source_id"),
		resource.TestCheckResourceAttrSet("data.maas_boot_source_selection.test", "labels.#"),
		resource.TestCheckResourceAttr("data.maas_boot_source_selection.test", "os", os),
		// the returned release depends on tested MAAS version
		resource.TestCheckResourceAttrSet("data.maas_boot_source_selection.test", "release"),
		resource.TestCheckResourceAttrSet("data.maas_boot_source_selection.test", "subarches.#"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasBootSourceSelection(os, release),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasBootSourceSelection(os string, release string) string {
	return fmt.Sprintf(`
data "maas_boot_source" "test" {
}

data "maas_boot_source_selection" "test" {
	os      = "%s"
	release = "%s"
}
`, os, release)
}
