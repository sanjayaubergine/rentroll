"use strict";

/* global module */
module.exports = function gruntInit(grunt) {

    // ========== GRUNT INIT CONFIG ==========
    grunt.initConfig({
        distHTMLs: './html/*.html', // all html files
        distInput: './js/elems/*.js',   // input source files
        distConcatOutput: './js/bundle.js', // output bundle
        distMinifiedOutput: './js/bundle.min.js',   // output bundle in minified version
        banner: "/*! '<%= grunt.template.today(\"yyyy-mm-dd\") %> */",  // banner for output file
        qUnitInstrumentedFiles: "./temp",   // qunit instrumentedFiles temporary folder, will be deleted after it's done

        //configure concat plugins
        concat: {
            options:{
                separator: '\n',
                banner: '<%= banner %>'
            },
            dist:{
                src: ['<%= distInput %>'],
                dest: '<%= distConcatOutput %>'
            },
        },

        // validate files with JSHINT
        jshint: {
            // REF: https://glebbahmutov.com/blog/linting-js-inside-html/
            html: {
                options: {
                    extract: 'always',
                    browser: true,
                },
                files: {
                    src: ['<%= distHTMLs %>']
                }
            },
            js: {
                src: ['Gruntfile.js', '<%= distInput %>']
            },
            options:{
                jshintrc: '.jshintrc'
            },
            // beforeconcat: ['<%= distInput %>'],
            // afterconcat: ['<%= distConcatOutput %>'], // when concatOutput is created then only run lint on it
            // all: ['Grunfile.js', '<%= distInput %>']
        },

        // QUnit test cases in a headless phantomjs instance
        qunit: {
            options: {
                // timeout: 30000,
                "--web-security": "no",
                coverage: {
                    src: [ "<%= distInput %>" ],
                    instrumentedFiles: '<%= qUnitInstrumentedFiles %>',
                    htmlReport: "./coverage/html",
                    lcovReport: "./coverage/lcov",
                    coberturaReport: "./coverage/cobertura",
                    linesThresholdPct: 0
                }
            },
            all: ["./qunit/index.html"]
        },

        // clean the files, created on temporary based
        clean: {
            js: ['<%= distConcatOutput %>', '<%= distMinifiedOutput %>', '<%= qUnitInstrumentedFiles %>', './coverage/']
        },

        // minification options
        uglify: {
            options: {
                banner: '<%= banner %>',
            },
            build: {
                files: {
                    '<%= distMinifiedOutput %>': ['<%= distInput %>']
                }
            }
        },

        // run tasks whenever watched files changed
        watch: {
            scripts: {
                files: ['<%= distInput %>', '<%= distHTMLs %>'],
                tasks: ['jshint', 'concat', 'uglify', 'qunit']
            }
        },

    }); // initConfig::END

    // ========== AVAILABLE TASKS ==========
    grunt.loadNpmTasks('grunt-contrib-clean');
    grunt.loadNpmTasks('grunt-contrib-jshint');
    grunt.loadNpmTasks('grunt-contrib-concat');
    grunt.loadNpmTasks('grunt-contrib-uglify');
    grunt.loadNpmTasks("grunt-qunit-istanbul");
    grunt.loadNpmTasks('grunt-contrib-watch');

    // ========== REGISTERED TASKS ==========
    grunt.registerTask('default', ['jshint', 'concat', 'uglify', 'qunit']);
    grunt.registerTask('dev', ['watch']);
};
