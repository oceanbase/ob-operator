#!/bin/bash

kubectl config view --raw --minify --output="go-template-file=cluster.tfvars.gotemplate" > terraform.tfvars
