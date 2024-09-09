# operator auto test
## Introduction
This document serves as a guide for executing automated tests designed for the OB-Operator

First modify env.sh and fill in sensitive variables

Running Tests
To run a single test case, follow this syntax:
./case_p0_ClusterAndTenant/check_P0_cluster01_create.sh

Or, utilize the run.sh script to target a specific category cases:
./run.sh case_p0_ClusterAndTenant/

For efficiency, you can execute multiple test categories concurrently using the && operator in Unix-based systems:
./run.sh case_p0_ClusterAndTenant/ && ./run.sh case_p1_BackupAndRestore


## Test Case Categorization
The test cases are organized into the following categories:

case_p0_ClusterAndTenant:

Ensures proper setup, configuration, and administration of OceanBase clusters and tenants.

case_p1_BackupAndRestore:

Validates backup creation, storage, and the recovery process to safeguard data integrity.

case_p2_Resource:

Resource related operations and validations

case_p3_AdvancedConfig:

Test with advanced configs

case_p4_ResourceModification:

Test resource modifications

case_p5_Operations:

Test operations resources
