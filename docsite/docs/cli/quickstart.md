# obocli  

##  Quick Start

Obocli (ob-operator cli) is a command line tool that is compatible with [ob-operator](https://github.com/oceanbase/ob-operator). it has following features:

- cluster management
- tenant management
- backup policy management
- component installation and update. 
- interactive command for user to deploy a ob cluster and tenant easily.

## Install obocli 

Currently obocli is not released, you can build it from source code.

## Build obocli from source code

```bash
make cli-build
```

if build successfully, you can find the binary file in `./bin/obocli`, you can copy it to your PATH.

Assuming you have added it to your PATH, you can run the following command to check if obocli is installed successfully.

```bash
obocli -v
```

## Use obocli

Deploy an OceanBase cluster and correspond tenant easily using `demo` command, currently support single node cluster and 1-1-1 three node cluster.

```bash
obocli demo
```

Create an OceanBase cluster named `test` in default namespace with default config.

```bash
obocli cluster create test  
```

Check the status of cluster `test`, until the status is `Running`.

```bash
obocli cluster list
```

Create a tenant `t1` with resource name called `demo-tenant` in cluster `test`.

```bash
obocli tenant create demo-tenant --cluster=test --tenant-name=t1
```

For more command information, you can use the following command to check the help information of obocli.

```bash
obocli -h
```
