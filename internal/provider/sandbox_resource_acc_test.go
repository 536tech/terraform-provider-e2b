// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSandboxResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	name := "tf-acc-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	var sandboxID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSandboxResourceConfig(name, 60, "8.8.8.8/32", "203.0.113.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCaptureResourceID("e2b_sandbox.test", &sandboxID),
					resource.TestCheckResourceAttrSet("e2b_sandbox.test", "id"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "template_id", "base"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "timeout", "60"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "alias", "base"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "metadata.managed_by", "terraform-provider-e2b"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "metadata.test_name", name),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "network_allow_out.#", "1"),
					resource.TestCheckTypeSetElemAttr("e2b_sandbox.test", "network_allow_out.*", "8.8.8.8/32"),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "network_deny_out.#", "1"),
					resource.TestCheckTypeSetElemAttr("e2b_sandbox.test", "network_deny_out.*", "203.0.113.0/24"),
					resource.TestCheckResourceAttrSet("data.e2b_sandbox.by_id", "id"),
					resource.TestCheckTypeSetElemAttr("data.e2b_sandbox.by_id", "network_allow_out.*", "8.8.8.8/32"),
					resource.TestCheckTypeSetElemAttr("data.e2b_sandbox.by_id", "network_deny_out.*", "203.0.113.0/24"),
					resource.TestCheckResourceAttrSet("data.e2b_sandboxes.filtered", "sandboxes.0.id"),
				),
			},
			{
				Config: testAccSandboxResourceConfig(name, 120, "1.1.1.1/32", "198.51.100.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceIDUnchanged("e2b_sandbox.test", &sandboxID),
					resource.TestCheckResourceAttr("e2b_sandbox.test", "timeout", "120"),
					resource.TestCheckTypeSetElemAttr("e2b_sandbox.test", "network_allow_out.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr("e2b_sandbox.test", "network_deny_out.*", "198.51.100.0/24"),
					resource.TestCheckTypeSetElemAttr("data.e2b_sandbox.by_id", "network_allow_out.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr("data.e2b_sandbox.by_id", "network_deny_out.*", "198.51.100.0/24"),
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

func testAccSandboxResourceConfig(name string, timeout int, allowOut string, denyOut string) string {
	return fmt.Sprintf(`
%s

resource "e2b_sandbox" "test" {
  template_id = "base"
  timeout     = %[3]d

  network_allow_out = [%[4]q]
  network_deny_out  = [%[5]q]

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
`, testAccProviderConfig(), name, timeout, allowOut, denyOut)
}

func testAccCaptureResourceID(resourceName string, target *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instance, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		if instance.Primary == nil || instance.Primary.ID == "" {
			return fmt.Errorf("resource %s has no primary ID", resourceName)
		}

		*target = instance.Primary.ID
		return nil
	}
}

func testAccCheckResourceIDUnchanged(resourceName string, target *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		instance, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		if instance.Primary == nil || instance.Primary.ID == "" {
			return fmt.Errorf("resource %s has no primary ID", resourceName)
		}
		if *target == "" {
			return fmt.Errorf("no prior resource ID captured for %s", resourceName)
		}
		if instance.Primary.ID != *target {
			return fmt.Errorf("expected %s ID to remain %q, got %q", resourceName, *target, instance.Primary.ID)
		}

		return nil
	}
}
