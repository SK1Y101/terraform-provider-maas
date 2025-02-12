package maas_test

import (
	"fmt"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMaasBootSourceSelection_basic(t *testing.T) {

	boot_source_id := "1"
	os := "ubuntu"
	release := "noble"

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("data.maas_boot_source_selection.test", "arches"),
		resource.TestCheckResourceAttr("data.maas_boot_source_selection.test", "boot_source_id", boot_source_id),
		resource.TestCheckResourceAttrSet("data.maas_boot_source_selection.test", "labels"),
		resource.TestCheckResourceAttr("data.maas_boot_source_selection.test", "os", os),
		resource.TestCheckResourceAttr("data.maas_boot_source_selection.test", "release", release),
		resource.TestCheckResourceAttrSet("data.maas_boot_source_selection.test", "subarches"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testutils.PreCheck(t, nil) },
		Providers:  testutils.TestAccProviders,
		ErrorCheck: func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMaasBootSourceSelection(boot_source_id, os, release),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func testAccDataSourceMaasBootSourceSelection(boot_source_id string, os string, release string) string {
	return fmt.Sprintf(`
data "maas_boot_source_selection" "test" {
	boot_source_id = "%v"
	
	os      = "%s"
	release = "%s"
}
`, boot_source_id, os, release)
}
