# Admin Web App

## Overview
The Admin Service webapp is built on React.js and uses the current LTS version of Node.js.

## Pre-requisites
- Install NVM
Install the [Node Version Manager](https://github.com/creationix/nvm) that will allow a specific
version of Node.js to be installed. Follow the installation instructions.

- Install the latest stable Node.js and npm
The latest stable (LTS) version of Node can be found on the [Node website](nodejs.org).
```bash
# Overview of available commands
nvm help

# Install the latest stable version
nvm install lts/*

# Select the version to use
nvm ls
nvm use lts/*
```

- Install the nodejs dependencies
```bash
cd webapp-admin # Make sure you are in the webapp directory
npm install
```

## Testing
To run the tests for an app:
```bash
cd webapp-admin # Make sure you are in the webapp directory
npm test
```

## Building
To run a full build:
```bash
cd webapp-admin  # Make sure you are in the webapp directory
./build.sh
```

This runs the build, which creates the files in the ./build directory and then copies the
relevant files to the static directories.

Create React App also has a handy development mode that loads the view in the browser and
refreshes it on every save - no extra build step required:
```bash
cd webapp-admin  # Make sure you are in the webapp-user directory
npm start
```
The dev mode launches ./public/index.html on port :3000.
