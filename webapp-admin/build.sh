#!/bin/sh

# Build the project
npm run build

# Copy the main assets to the static directory
cp -R build/static/* ../static
cp build/index.html ../static

# cleanup
rm -rf ./build
