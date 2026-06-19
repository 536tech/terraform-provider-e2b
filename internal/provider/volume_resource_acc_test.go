// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVolumeResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}
	if os.Getenv("E2B_ACC_VOLUMES") == "" {
		t.Skip("set E2B_ACC_VOLUMES=1 to run volume acceptance tests; this E2B account must have volumes enabled")
	}

	name := "tf-acc-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("e2b_volume.test", "id"),
					resource.TestCheckResourceAttr("e2b_volume.test", "name", name),
					resource.TestCheckResourceAttrSet("data.e2b_volume.by_id", "id"),
				),
			},
			{
				ResourceName:      "e2b_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVolumeResourceConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "e2b_volume" "test" {
  name = %[2]q
}

data "e2b_volume" "by_id" {
  id = e2b_volume.test.id
}

data "e2b_volumes" "all" {}
`, testAccProviderConfig(), name)
}
