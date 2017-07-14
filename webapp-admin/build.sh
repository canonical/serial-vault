#!/bin/sh

# Build the project
npm run build

# Copy the main assets to the static directory
cp -R build/static/js/*.js ../static/js/bundle.js
cp -R build/static/css/*.css ../static/css/application.css
cp -R build/static/css/*.css.map ../static/css/application.css.map

# cleanup
rm -rf ./build
