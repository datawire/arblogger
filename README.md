# ARBLogger

ARBLogger is a demo REST service for [ARB](https://github.com/datawire/arb). It is
useful mostly in the context of ARB, so you are strongly encourage to check out the
[ARB README](https://github.com/datawire/arb/blob/master/README.md)!

## Building ARBLogger

### Install dependencies

 - [GNU Make](https://gnu.org/s/make)
 - [Go](https://golang.org/) 1.15 or newer
 - Docker

### Set up

Edit the `Makefile` to set `DOCKER_REGISTRY` to a registry to which 
you can push. You may also prefer to set `IMAGE_TAG` to give your image
a separate version number.

After that, `make tools` to set up [`ko`](https://github.com/google/ko)
in `tools/bin/ko`.

### Using `ko` for development

`make apply` will use `ko` to build ARB and apply it, using `arb.yaml`,
to your cluster.

### Pushing an image to your Docker registry

`make push` will build ARB with `ko`, then push it to `$DOCKER_REGISTRY/arb:$IMAGE_TAG`,
where the variables have their values from the `Makefile`.
