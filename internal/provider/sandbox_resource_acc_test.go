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

func TestAccSandboxResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	name := "tf-acc-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSandboxResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("e2b_sandbox.test", "id"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "template_id", "base"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "alias", "base"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "metadata.managed_by", "terraform-provider-e2b"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "metadata.test_name", name),
					resource.TestCheckResourceAttrSet("data.e2b_sandbox.by_id", "id"),
					resource.TestCheckResourceAttrSet("data.e2b_sandboxes.filtered", "sandboxes.0.id"),
				),
			},
			{
				ResourceName:            "e2b_sandbox.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_id", "timeout", "metadata", "env_vars"},
			},
		},
	})
}

func testAccSandboxResourceConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "e2b_sandbox" "test" {
  template_id = "base"
  timeout     = 60

  metadata = {
    managed_by = "terraform-provider-e2b"
    test_name  = %[2]q
  }
}

data "e2b_sandbox" "by_id" {
  id = e2b_sandbox.test.id
}

data "e2b_sandboxes" "filtered" {
  depends_on = [e2b_sandbox.test]

  metadata = {
    test_name = %[2]q
  }

  states = ["running"]
  limit  = 10
}
`, testAccProviderConfig(), name)
}
