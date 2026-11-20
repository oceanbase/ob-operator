/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND.
*/

package database

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type rotationSource struct{ driver, password string }

func (s *rotationSource) ID() string             { return s.driver }
func (s *rotationSource) DriverName() string     { return s.driver }
func (s *rotationSource) DataSourceName() string { return s.password }
func (s *rotationSource) GetAddress() string     { return "localhost" }
func (s *rotationSource) GetPort() int64         { return 2881 }
func (s *rotationSource) GetUser() string        { return "fixture" }
func (s *rotationSource) GetPassword() string    { return s.password }
func (s *rotationSource) GetDatabase() string    { return "fixture" }
func (s *rotationSource) String() string         { return "fixture" }

type rotationDriver struct {
	valid atomic.Value
}

func (d *rotationDriver) Open(password string) (driver.Conn, error) {
	if password != d.valid.Load().(string) {
		return nil, fmt.Errorf("driver rejected %s", password)
	}
	return rotationConn{}, nil
}

type rotationConn struct{}

func (rotationConn) Close() error                        { return nil }
func (rotationConn) Begin() (driver.Tx, error)           { return nil, errors.New("unused") }
func (rotationConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unused") }

var rotationDriverID atomic.Uint64

func TestPoolCredentialRotation(t *testing.T) {
	d := &rotationDriver{}
	initialPassword := rand.Text()
	s := &rotationSource{driver: fmt.Sprintf("rotation-%d", rotationDriverID.Add(1)), password: initialPassword}
	sql.Register(s.driver, d)
	d.valid.Store(s.password)
	t.Cleanup(func() { p.Cache.Remove(s.ID()) })
	first, err := GetConnector(s)
	if err != nil {
		t.Fatal(err)
	}
	if same, err := GetConnector(s); err != nil || same != first {
		t.Fatal("expected the authenticated pool to be reused")
	}

	s.password = rand.Text()
	got, err := GetConnector(s)
	if got != nil || err == nil {
		t.Fatal("changed credential reused the existing pool without authenticating")
	}
	if strings.Contains(err.Error(), s.password) {
		t.Fatal("connection error included the supplied credential")
	}
	s.password = initialPassword
	if same, err := GetConnector(s); err != nil || same != first || !first.IsAlive() {
		t.Fatal("unsuccessful authentication replaced the valid pool")
	}

	s.password = rand.Text()
	d.valid.Store(s.password)
	const callers = 16
	var wg sync.WaitGroup
	pools := make(chan *Connector, callers)
	for i := range callers {
		request := *s
		valid := i%2 == 0
		if !valid {
			request.password = rand.Text()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := GetConnector(&request)
			if !valid {
				if err == nil || c != nil {
					t.Error("a concurrent invalid credential received a pool")
				}
				return
			}
			if err != nil {
				t.Errorf("new credential: %v", err)
				return
			}
			pools <- c
		}()
	}
	wg.Wait()
	close(pools)
	current, err := GetConnector(s)
	if err != nil || current == first || !current.IsAlive() {
		t.Fatal("rotation did not install a new usable pool")
	}
	for c := range pools {
		if c != current {
			t.Fatal("concurrent callers received different pools for the new credential")
		}
	}
	s.password = initialPassword
	if previous, err := GetConnector(s); err == nil || previous != nil {
		t.Fatal("retired credential reused the replacement pool")
	}
	if first.IsAlive() {
		t.Fatal("the replaced pool is still open")
	}
}
