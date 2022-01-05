/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package shell

import (
	"context"
	"fmt"
	"time"
)

type Program string

const (
	Sh        = "sh"
	RootUser  = "root"
	AdminUser = "admin"
)

const DefaultProgram = Sh
const DefaultTimeout = 10 * time.Second

type Command interface {
	Cmd() string
	User() string
	Program() Program
	Timeout() time.Duration
	WithContext(ctx context.Context) Command
	WithUser(user string) Command
	WithProgram(program Program) Command
	WithTimeout(timeout time.Duration) Command
	Execute() (*ExecuteResult, error)
	ExecuteWithDebug() (*ExecuteResult, error)
	ExecuteAllowFailure() (*ExecuteResult, error)
}

func NewCommand(cmd string) Command {
	return &command{
		program: DefaultProgram,
		cmd:     cmd,
		timeout: DefaultTimeout,
	}
}

type command struct {
	user    string  // Run command as this user, if not provided, run command as current process's user
	program Program // Shell program to execute command, e.g. sh, bash
	cmd     string
	timeout time.Duration
	context context.Context
}

func (c *command) Cmd() string {
	return c.cmd
}

func (c *command) User() string {
	return c.user
}

func (c *command) Program() Program {
	return c.program
}

func (c *command) Timeout() time.Duration {
	return c.timeout
}

func (c *command) WithContext(ctx context.Context) Command {
	c.context = ctx
	return c
}

func (c *command) WithUser(user string) Command {
	c.user = user
	return c
}

func (c *command) WithProgram(program Program) Command {
	c.program = program
	return c
}

func (c *command) WithTimeout(timeout time.Duration) Command {
	c.timeout = timeout
	return c
}

func (c *command) String() string {
	return fmt.Sprintf("Command{user=%s, program=%s, cmd=%s, timeout=%s}", c.user, c.program, c.cmd, c.timeout)
}
