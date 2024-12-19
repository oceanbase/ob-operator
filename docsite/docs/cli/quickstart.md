# okctl

```bash
=============================================
          _             _     _ 
   ___   | | __   ___  | |_  | |
  / _ \  | |/ /  / __| | __| | |
 | (_) | |   <  | (__  | |_  | |
  \___/  |_|\_\  \___|  \__| |_|

=============================================
A Command Line Tool compatible with OceanBase Operator
```

okctl is a powerful command line interface (CLI) tool compatible with [ob-operator](https://github.com/oceanbase/ob-operator/blob/master/README.md)

okctl has the following features:

- Cluster management, including creating, deleting, scaling, upgrading, etc.
- Tenant management, including creating, updating, activating standby tenant, switchover of tenants, etc.
- Backup policy management, including creating, resuming, pausing, etc.
- Component installation and update, support ob-operator, ob-dashboard, cert-manager, local-path-provisioner.
- Interactive command for user to deploy a ob cluster and tenant easily.

## Install okctl

Currently okctl is not released, you can build it from source code.

### Build okctl from source code

```bash
make cli-build
```

If build successfully, you can find the binary file in `./bin/okctl`, you can copy it to your `PATH` by the following command.

```bash
# using bash as example
echo "export PATH=$PATH:/path/to/okctl" >> ~/.bashrc
source ~/.bashrc
```

Assuming you have added it to your `PATH`, you can run the following command to check if okctl is installed successfully.

```bash
okctl -v
```

If install successfully, you will see the output like this:

```bash
OceanBase Operator Cli:
 Version:    0.1.0
 OS/Arch:    linux/amd64
 Go Version: go1.22.5
 Git Commit: 7074c066
 Build:      2024-12-18 16:22:11

```

## Use okctl with examples

### Interactive command

Deploy an OceanBase cluster and correspond tenant easily using `demo` command, currently support single node cluster and 1-1-1 three node cluster.

```bash
okctl demo
```

### Cluster management

Create an OceanBase cluster named `test` in default namespace with 1 zone.

```bash
okctl cluster create test --zones=z1=1
```

List all clusters in default namespace.

```bash
okctl cluster list
```

Add two zones `z2`, `z3` with replica 1 to the cluster `test` 

```bash
okctl cluster scale test --zones=z2=1,z3=1
```

Update the cluster `test` with 2 cpu and 12G memory.

```bash
okctl cluster update test --cpu=2 --memory=12
```

### Tenant management

Create a tenant `t1` with same resource name in cluster `test`.

```bash
okctl tenant create t1 --cluster=test --priority=z1=1
```

Create a empty standby tenant `t1s` with same resource name in cluster `test`.

```bash
okctl tenant create t1s --cluster=test --from=t1
```

Switch over roles of primary tenant `t1` and standby tenant `t1s`.

```bash
okctl tenant switchover t1 t1s
``` 

### Backup policy management

Create a backup policy by NFS for tenant `t1` in default namespace.

```bash
okctl backup create t1 --archive-path="t1/archive" --bak-data-path="t1/backup" --bak-encryption-password="xxxx" --inc="0 0 * * 1,2,3," --full="0 0 * * 4,5"
```

After creating the backup policy, you can restore `t1` by creating a new tenant `t1r` with the same resource name.

```bash
okctl tenant create t1r --cluster=test --from=t1 --restore --restore-type=nfs --archive-path="t1/archive" --bak-data-source="t1/backup" --restore-password="xxxx"  
```


Pause the backup policy of tenant `t1`.

```bash
okctl backup pause t1
```

### Component Installation and update

Use okctl to install `ob-operator` with version `2.2.0` in cluster.

```bash
okctl install ob-operator --version=2.2.0
```

Use okctl to update `ob-operator` to the latest version.

```bash
okctl update ob-operator
```

### Others

For better use experience, you can use autocomplete feature of okctl. You can run the following command to check the command for your shell, here is an example for bash.

```bash
# After install the bash-completion package, you can run the following command to enable autocomplete feature in the current shell.
source <(okctl completion bash)
```

For more command information, you can use `okctl -h` command to check the help information of okctl.
