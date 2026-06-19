// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"e2b": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("E2B_API_KEY") == "" {
		loadDotEnv(t)
	}

	if os.Getenv("E2B_API_KEY") == "" {
		t.Fatal("E2B_API_KEY must be set for acceptance tests")
	}
}

func testAccProviderConfig() string {
	return `provider "e2b" {}`
}

func loadDotEnv(t *testing.T) {
	t.Helper()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("read home directory: %s", err)
	}

	content, err := os.ReadFile(filepath.Join(home, ".env"))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("read ~/.env: %s", err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "E2B_API_KEY" && os.Getenv("E2B_API_KEY") == "" {
			t.Setenv("E2B_API_KEY", value)
		}
	}
}
