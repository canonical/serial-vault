#!/bin/sh

# Build the project
npm run build

# Copy the main assets to the static directory
cp -R build/static/js ../static
cp -R build/static/css ../static
cp build/index.html ../static/app.html
# cleanup
rm -rf ./build
