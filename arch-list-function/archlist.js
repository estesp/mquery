var Cloudant = require('cloudant'),
    Openwhisk = require('openwhisk');

function main(params) {
    var imageData = {};
    return new Promise(function(resolve, reject) {
        var image = params.image;
        if (image === undefined || image === "") {
            reject({
                'error': 'Please provide an image name to query.'
            });
            return;
        }

        if (!params.cloudantUser || !params.cloudantHost || !params.cloudantPassword) {
            console.error('CloudantDB parameters not set.');
            reject({
                'error': 'CloudantDB parameters not set.'
            });
            return;
        }
        // Configure database connection
        var cloudant = new Cloudant({
            account: params.cloudantUser,
            password: params.cloudantPassword,
            plugin: 'promises'
        });

        var db = cloudant.db.use('registryimages');
        var ow = Openwhisk();
        var errorStr = "";

        console.log("image lookup: " + image);
        getImageData(db, image).then(function(data) {
            imageData = data;
            var now = Date.now();
            if ((now - imageData.cachetime) > 60 * 60 * 1000) {
                // data older than one hour; don't use cache
                console.log("Expiring cached data for: " + image);
                imageData.expired = true;
            }
        }).catch(function(err) {
            console.log("[getImageData] Cloudant lookup error/empty: " + err);
        }).then(function() {
            if (isEmpty(imageData) || imageData.expired) {
                // need to query data from manifest tool inspect action
                return ow.actions.invoke({
                    actionName: "/estesp@us.ibm.com_dev/mplatformQuery",
                    blocking: true,
                    params: {
                        "image": image,
                    }
                }).catch(function(err) {
                    // our function invocation returned an error:
                    console.log("Error from mplatformQuery action: " + err);
                    errorStr = err.error.response.result.error;
                }).then(function(res) {
                    // check for error on action invoke:
                    if (errorStr !== "") {
                        reject({
                            error: errorStr,
                        });
                        return;
                    }
                    if (res.response.result.payload) {
                        var newData = res.response.result.payload;
                        if (!isEmpty(imageData)) {
                            newData._rev = imageData._rev;
                        }
                        newData._id = image;
                        newData.cachetime = Date.now();
                        return processImageData(db, newData);
                    }
                });
            } else {
                // give back the cached data
                console.log("Using cached data for: " + image);
                return imageData;
            }
        }).then(function(data) {
            resolve({ payload: data });
        });
    });
}

function processImageData(db, manifestData) {
    var filteredData = {};
    filteredData.cachetime = manifestData.cachetime;
    filteredData._id = manifestData._id;
    if (manifestData._rev !== null) {
        filteredData._rev = manifestData._rev;
    }
    // walk manifest entries for this name:tag
    var archList = [];
    var manifestList = false;
    for (var i = 0; i < manifestData.length; i++) {
        if (manifestData[i].MediaType == "application/vnd.docker.distribution.manifest.list.v2+json") {
            filteredData.manifestList = "Yes";
            manifestList = true;
            continue;
        }
        filteredData.repoTags = manifestData[i].RepoTags;
        filteredData.tag = manifestData[i].Tag;
        if (manifestList) {
            var arch = manifestData[i].Platform.architecture;
            var os = manifestData[i].Platform.os;
            var variant = "";
            if (arch == "arm" || arch == "arm64") {
                variant = manifestData[i].Platform.variant;
                archList.push("" + arch + "/" + os + " (variant: " + variant + ")");
            } else {
                archList.push("" + arch + "/" + os);
            }
        } else {
            filteredData.manifestList = "No";
            filteredData.Platform = manifestData[i].Architecture + "/" + manifestData[i].Os;
        }
    }
    if (manifestList) {
        filteredData.archList = archList;
    }
    putImageData(db, filteredData);
    console.log("cached image data for: " + filteredData._id);
    return filteredData;
}

function getImageData(db, image) {
    // query cloudant to see if we have cached any image data
    return db.get(image);
}

function putImageData(db, imageData) {
    db.insert(imageData, function(err, data) {
        if (err) {
            console.log("Error on registry data DB insert: " + err);
        }
    });
}

function isEmpty(obj) {
    if (obj === undefined) {
        return true;
    }
    return Object.keys(obj).length === 0;
}