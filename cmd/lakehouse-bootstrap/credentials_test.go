/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND.
*/

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialsCreatedOnceAndRetained(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "secrets")
	if err := initCredentials(dir); err != nil {
		t.Fatal(err)
	}
	before := map[string]string{}
	for _, name := range credentialNames {
		value, err := readCredential(filepath.Join(dir, name))
		if err != nil || len(value) < 20 {
			t.Fatalf("%s: insufficient generated credential: %v", name, err)
		}
		before[name] = value
		info, _ := os.Stat(filepath.Join(dir, name))
		if info.Mode().Perm() != 0600 {
			t.Fatalf("%s: credential permissions %o", name, info.Mode().Perm())
		}
	}
	if err := initCredentials(dir); err != nil {
		t.Fatal(err)
	}
	for _, name := range credentialNames {
		value, _ := readCredential(filepath.Join(dir, name))
		if value != before[name] {
			t.Fatal("credential changed on repeat initialization")
		}
	}
	if before["minio_access_key"] == before["ob_root_password"] {
		t.Fatal("database and storage share a password")
	}
}

func TestInitializedCredentialCorruptionDoesNotRotatePasswords(t *testing.T) {
	for _, corrupt := range []string{"missing", "empty", "symlink"} {
		t.Run(corrupt, func(t *testing.T) {
			dir := t.TempDir()
			if err := initCredentials(dir); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "ob_root_password")
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			switch corrupt {
			case "empty":
				if err := os.WriteFile(path, nil, 0600); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Symlink(filepath.Join(dir, "minio_access_key"), path); err != nil {
					t.Fatal(err)
				}
			}
			if err := initCredentials(dir); err == nil {
				t.Fatal("corrupted credential silently regenerated")
			}
		})
	}
}

func TestPartialInitializationPreservesCompletedCredential(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "minio_access_id")
	if err := os.WriteFile(path, []byte("previously-created-id\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := initCredentials(dir); err != nil {
		t.Fatal(err)
	}
	value, _ := readCredential(path)
	if value != "previously-created-id" {
		t.Fatal("partial credential changed")
	}
}
