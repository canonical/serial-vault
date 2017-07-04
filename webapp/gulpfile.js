var gulp = require('gulp');
var sass = require('gulp-sass');
var uglifycss = require('gulp-uglifycss');
var babel = require("gulp-babel");
var browserify = require('gulp-browserify');
var uglify = require('gulp-uglify');
var rename = require('gulp-rename');
var del = require('del');

var path = {
  BASE: './',
  SRC: './js/',
  BUILD: './build/',
  DIST: '../static/'
};

// Use production mode to omit debug code from modules
gulp.task('set-prod-node-env', function() {
    return process.env.NODE_ENV = 'production';
});

// Compile the Vanilla SASS files into the CSS folder
gulp.task('sass', function () {
    gulp.src(path.BASE + 'sass/*.scss')
        .pipe(sass().on('error', sass.logError))
        .pipe(uglifycss())
        .pipe(gulp.dest(path.DIST + 'css'));
});

// Compile the JSX/ES6 files to Javascript in the build directory
gulp.task('compile_components', function(){
    return gulp.src(path.SRC + 'components/*.js')
        .pipe(babel())
        .pipe(gulp.dest(path.BUILD + 'components/'));
});
gulp.task('compile_models', function(){
    return gulp.src(path.SRC + 'models/*.js')
        .pipe(babel())
        .pipe(gulp.dest(path.BUILD + 'models/'));
});
gulp.task('compile_app', ['compile_components', 'compile_models'], function(){
    return gulp.src(path.SRC + '*.js')
        .pipe(babel())
        .pipe(gulp.dest(path.BUILD));
});

gulp.task('build', ['compile_app'], function(){
  return gulp.src([path.BUILD + 'index.js'])
    .pipe(browserify({}))
    .on('prebundle', function(bundler) {
        // Make React available externally for dev tools
        bundler.require('react');
    })
    .pipe(rename('bundle.js'))
    .pipe(uglify())
    .pipe(gulp.dest(path.DIST + 'js/'));
});

// Clean the build files
gulp.task('clean', ['build', 'sass'], function() {
    del([path.BUILD + '**/*']);
});

// Default: remember that these tasks get run asynchronously
gulp.task('default', ['set-prod-node-env', 'build', 'sass', 'clean']);
