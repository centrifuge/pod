const swaggermerge = require('swagger-merge')
const fs = require('fs')
const path = require('path')
const util = require('util');
const exec = util.promisify(require('child_process').exec);

const swaggerFilesPath = path.resolve(__dirname, '../centrifuge/protobufs/gen/swagger')
const swaggerJsonPath = path.resolve(__dirname, '../centrifuge/protobufs/gen/swagger.json')
const swaggerHtmlPath = path.resolve(__dirname, '../centrifuge/protobufs/gen/swagger/html')
const swaggerConfig = require(path.resolve(__dirname, '../swagger_config'));
const pathPrefix = ""
/* # build_swagger.js
 *
 * This script recursively searches the swaggerFilesPath for any file ending in .swagger.json
 * loading all matching files and merging them into one for easier portability. In the second
 * step it then creates an html version of the documentation using spectacles.
 *
 * The defaults are defined in ../swagger_config.js
 */


// Find matching swagger files in path
// From: https://gist.github.com/kethinov/6658166
var getSwaggerFiles = function(dir, filelist) {
    var files = fs.readdirSync(dir);
    filelist = filelist || [];
    files.forEach(function(file) {
        if (fs.statSync(dir + '/' + file).isDirectory()) {
            getSwaggerFiles(dir + '/' + file, filelist);
        }
        else {
            if (file.indexOf(".swagger.json") > 0) {
                filelist.push(path.join(dir, file));
            }
        }
    });
    return filelist;
};

var files = getSwaggerFiles(swaggerFilesPath)
// There is a default swagger definition in swaggerConfig.defaultSwagger which we add first
swaggerModules = [swaggerConfig.defaultSwagger,]
files.forEach(function (f) {
    swaggerModules.push(require(f))
})

swaggermerge.on('warn', function (msg) {
    console.log(msg)
})
swaggermerge.on('err', function (msg) {
    console.error(msg)
    process.exit(1)
})

var merged = swaggermerge.merge(swaggerModules, swaggerConfig.info, swaggerConfig.pathPrefix, swaggerConfig.host, swaggerConfig.schemes)
var json = JSON.stringify(merged);
console.log("Merged swagger.json, writing to:", swaggerJsonPath);
fs.writeFileSync(swaggerJsonPath, json);

var spectacles = exec('spectacle '+swaggerJsonPath+' -t '+swaggerHtmlPath).then(function (msg) {
    console.log(msg['stdout'])
    console.log("Wrote html files to: ", swaggerHtmlPath)
    process.exit(0)
}).catch(function (err) {
    console.error(err)
    process.exit(1)
});

