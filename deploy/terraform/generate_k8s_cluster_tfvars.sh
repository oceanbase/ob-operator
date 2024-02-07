#!/bin/bash

kubectl config view --raw --output="go-template-file=cluster.tfvars.gotemplate" > terraform.tfvars
