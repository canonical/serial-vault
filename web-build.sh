#!/bin/bash

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
