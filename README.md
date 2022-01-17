# mquery
A simple utility and serverless-based backend for querying Docker v2 & OCI v1 container images
and reporting on "manifest list"/OCI "index" multi-platform image support.

## About
This project uses [AWS Lambda](https://aws.amazon.com/lambda/) as a backend, in concert
with the [manifest-tool](https://github.com/estesp/manifest-tool) inspect API capability
to easily report on the status of whether an image is a manifest list/OCI index entry in a
registry, and if so, what platforms are supported by the image.

## Usage
You can use the public endpoint with `curl` and JSON formatting tools to query images directly.
See the next section for a tool which performs this for you and provides a simple text output.
This tool is published as a multi-platform image on DockerHub as `mplatform/mquery`; for example
you can look up the `ubuntu:latest` image as follows:

```
$ docker run --rm mplatform/mquery ubuntu:latest
Image: ubuntu:latest (digest: sha256:b5a61709a9a44284d88fb12e5c48db0409cfad5b69d4ff8224077c57302df9cf)
 * Manifest List: Yes (Image type: application/vnd.docker.distribution.manifest.list.v2+json)
 * Supported platforms:
   - linux/amd64
   - linux/arm/v7
   - linux/arm64/v8
   - linux/ppc64le
   - linux/riscv64
   - linux/s390x
```

#### Using the `mquery` tool

This project also includes a tool for querying the Lambda API Gateway-fronted endpoint with
a simple/readable output format for showing the list of platforms supported by a specific
image. You can build the tool yourself using the `Makefile`, or you can use a pre-packaged
multi-platform image on DockerHub as shown in the section above.

This Go program requires the [github.com/dghubble/sling](https://github.com/dghubble/sling) and
[github.com/opencontainers/image-spec/specs-go/v1](https://github.com/opencontainers/image-spec) packages.
You can add these to your Go development environment with:
```
$ go get -u github.com/dghubble/sling
$ go get -u github.com/opencontainers/image-spec/specs-go/v1
```

## References
More information about manifest lists and multi-platform image support is available in these blog posts:
 - [DockerHub Official Images Go Multi-platform!](https://integratedcode.us/2017/09/13/dockerhub-official-images-go-multi-platform/) - 13 Sep 2017
 - [[Docker Blog] Docker Official Images are now Multi-platform](https://blog.docker.com/2017/09/docker-official-images-now-multi-platform/) - 19 Sep 2017
  - [A big step towards multi-platform Docker images](https://integratedcode.us/2016/04/22/a-step-towards-multi-platform-docker-images/) - 22 April 2016

Also see the [manifest-tool project](https://github.com/estesp/manifest-tool) for an easy to use tool
for assembling and pushing manifest lists and OCI index images.

## License
This project is licensed under the Apache Public License, v2.0.


