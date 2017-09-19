# mquery
A simple utility and backend for querying Docker v2 API-supporting registry images
and reporting on "manifest list" multi-platform image support.

## About
This project uses [IBM Cloud Functions](https://console.bluemix.net/docs/openwhisk/index.html) (built on [OpenWhisk](https://openwhisk.incubator.apache.org/)) as a backend, in concert
with the [manifest-tool](https://github.com/estesp/manifest-tool) `inspect` capability
(packaged as a Docker function) to easily report on the status of whether an image is
a manifest list entry in the registry, and if so, what architecture/os pairs are supported
by the image.

## Usage
You can easily publish these functions yourself to [IBM Bluemix Cloud Functions](https://console.bluemix.net) using
the scripts in the two action directories. It will require having a bound Cloudant
database instance in your IBM Cloud account and your Cloudant credentials must be provided
to the function via the parameters.json file.

If you are interested in using the already published functions in my account, you can use
the "Web Action" URL with a query parameter to easily query any publicly accessible image
on any registry that supports the Docker v2 API without authentication (for public images).

The API endpoint is: *https://openwhisk.ng.bluemix.net/api/v1/web/estesp%40us.ibm.com_dev/default/archList.json*

You can use it with an `image` parameter like this:
```
$ curl 'https://openwhisk.ng.bluemix.net/api/v1/web/estesp%40us.ibm.com_dev/default/archList.json?image=estesp/busybox'
{
  "payload": {
     "manifestList": "Yes",
     "tag": "latest",
     "_id": "estesp/busybox",
     "cachetime": 1505832200347,
     "repoTags": ["aarch64", "amd64", "armfh", "latest", "ppc64le", "s390x"],
     "_rev": "3-0493c3315169c8ceac2d419463abb7e2",
     "archList": ["ppc64le/linux", "amd64/linux", "s390x/linux", "arm/linux (variant: armv7)", "arm64/linux (variant: armv8)"]
  }
}
```

Piping this output to `jq '.payload.archList'` would print just the list of architectures
supported by the image.

## References
More information about manifest lists and multi-platform image support is available in these blog posts:
 - [DockerHub Official Images Go Multi-platform!](https://integratedcode.us/2017/09/13/dockerhub-official-images-go-multi-platform/) - 13 Sep 2017
 - [[Docker Blog] Docker Official Images are now Multi-platform](https://blog.docker.com/2017/09/docker-official-images-now-multi-platform/) - 19 Sep 2017
  - [A big step towards multi-platform Docker images](https://integratedcode.us/2016/04/22/a-step-towards-multi-platform-docker-images/) - 22 April 2016

Also see the [manifest-tool project](https://github.com/estesp/manifest-tool) for how manifest lists are being created today while tooling is being completed with the Docker client.

## License
This project is licensed under the Apache Public License, v2.0.


