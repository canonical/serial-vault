#!/bin/sh

# Build the project
npm run build

# Copy the main assets to the static directory
cp -R build/static/js/*.js ../static/js/app_user.js
cp -R build/static/css/*.css ../static/css/app_user.css
cp -R build/static/css/*.css.map ../static/css/app_user.css.map