/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND.
*/

package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var credentialNames = []string{"minio_access_id", "minio_access_key", "ob_root_password", "ob_monitor_password"}

const credentialFormatVersion = 2

// initCredentials runs in an init container before any database or object-store
// process. The named volume survives container recreation and compose down.
func initCredentials(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("credential directory must be a real directory")
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return err
	}
	marker := filepath.Join(dir, ".initialized")
	version := 0
	markerValue, err := os.ReadFile(marker)
	if err == nil {
		if _, err := fmt.Sscanf(string(markerValue), "v%d", &version); err != nil || version < 1 || version > credentialFormatVersion {
			return errors.New("credential marker has an unsupported version")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, name := range credentialNames {
		path := filepath.Join(dir, name)
		info, err := os.Lstat(path)
		if err == nil {
			if !info.Mode().IsRegular() {
				return fmt.Errorf("%s must be a regular credential file", name)
			}
			if _, err := readCredential(path); err != nil {
				return fmt.Errorf("%s is invalid; restore the credentials volume from backup", name)
			}
			if err := os.Chmod(path, 0600); err != nil {
				return err
			}
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		// Version 2 adds a least-privilege OBAgent password. It is the only
		// credential that may be added to an existing version-1 volume.
		if version > 0 && !(version == 1 && name == "ob_monitor_password") {
			return fmt.Errorf("%s is missing from initialized credentials; restore the credentials volume from backup", name)
		}
		var random [24]byte
		if _, err := rand.Read(random[:]); err != nil {
			return fmt.Errorf("generate %s: %w", name, err)
		}
		value := hex.EncodeToString(random[:])
		if name == "minio_access_id" {
			value = value[:20]
		}
		if err := writePrivateFile(path, []byte(value+"\n")); err != nil {
			return fmt.Errorf("save %s: %w", name, err)
		}
	}
	if version != credentialFormatVersion {
		return replacePrivateFile(marker, []byte(fmt.Sprintf("v%d\n", credentialFormatVersion)))
	}
	return nil
}

func replacePrivateFile(path string, contents []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".credential-marker-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(contents); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func writePrivateFile(path string, contents []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".credential-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(contents); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// Link publishes a complete private file without replacing an existing
	// credential if two initializers run concurrently.
	if err := os.Link(tmp.Name(), path); err != nil {
		return fmt.Errorf("publish credential file: %w", err)
	}
	return nil
}
