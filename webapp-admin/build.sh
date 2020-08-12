#!/bin/sh

# Build the project
npm run build

# Copy the main assets to the static directory
rm ../static/js/*
rm ../static/css/*
cp -R build/static/js ../static
cp -R build/static/css ../static
cp -R build/index.html ../static/app.html

# cleanup
rm -rf ./build
