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

func TestAccTemplateResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	name := "tf-acc-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("e2b_template.test", "id"),
					resource.TestCheckResourceAttr("e2b_template.test", "name", name),
					resource.TestCheckResourceAttr("e2b_template.test", "from_image", "e2bdev/base:latest"),
					resource.TestCheckResourceAttr("e2b_template.test", "build_status", "ready"),
					resource.TestCheckResourceAttr("e2b_template.test", "cpu_count", "2"),
					resource.TestCheckResourceAttr("e2b_template.test", "memory_mb", "512"),
					resource.TestCheckResourceAttrPair("data.e2b_template.by_id", "id", "e2b_template.test", "id"),
					resource.TestCheckResourceAttrSet("e2b_sandbox.from_template", "id"),
					resource.TestCheckResourceAttrPair("e2b_sandbox.from_template", "template_id", "e2b_template.test", "id"),
				),
			},
			{
				ResourceName:            "e2b_template.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"from_image", "start_cmd", "ready_cmd", "cpu_count", "memory_mb", "force", "build_status"},
			},
		},
	})
}

func testAccTemplateResourceConfig(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "e2b_template" "test" {
  name       = %[1]q
  from_image = "e2bdev/base:latest"
  start_cmd  = "sh -c \"sleep 3600\""
  ready_cmd  = "true"
  cpu_count  = 2
  memory_mb  = 512
  force      = true
}

data "e2b_template" "by_id" {
  id = e2b_template.test.id
}

resource "e2b_sandbox" "from_template" {
  template_id = e2b_template.test.id
  timeout     = 60

  metadata = {
    managed_by = "terraform-provider-e2b"
    test_name  = %[1]q
  }
}
`, name)
}
