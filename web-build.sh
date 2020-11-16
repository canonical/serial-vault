#!/bin/bash
#
# This script is used from the Makefile/TravisCI in order to find out if
# some changes where done in webapp-admin/ directory or not. 
# In the positive case Makefile will run frontend tests from webapp-admin/ directory.

set -e

CHANGED_FILES=`git diff --name-only master`
IS_WEB_BUILD=False
WEB_PATH="webapp-admin/*"

for CHANGED_FILE in $CHANGED_FILES; do
  if [[ $CHANGED_FILE =~ $WEB_PATH ]]; then
    IS_WEB_BUILD=True
    break
  fi
done

if [[ $IS_WEB_BUILD == True ]]; then
  echo "yes"
fi
