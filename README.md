## Building

### Install dependencies

 - [GNU Make](https://gnu.org/s/make)
 - [Go](https://golang.org/) 1.15 or newer
 - Docker

### Run the build

 1. Run the `make` command.  This will create a local Docker image
    named `ko.local/example-rls-SOMEHASH:SOMEHASH`, and print that
    name on stdout.

### Push the image somewhere

 1. Copy the full name from the output of `make` above.
 2. Use `docker tag` to tag it with
    `docker.io/yourname/yourrepo:yourversion` or whichever name you
    like.
 3. Use `docker push` to push your tag to your registry.
