#!/bin/bash
# get all filename in folder
source env.sh

set +x
 
path=$1
files=$(ls $path)
 
# iterate over all test files and execute
for filename in $files
do
  result=$(echo $filename | grep "check_P")
  if [[ "$result" != "" ]];then
    date
    echo "==============================================case $filename===================================================="
    bash $path/$filename
  fi
done
