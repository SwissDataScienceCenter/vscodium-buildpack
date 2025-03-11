# VSCodium Buildpack

A buildpack that packages browser based vscodium into an image.

To build and run this buildpack, run

```bash
# build the buildpack
$ make build
# create a docker image using the buildpack and an application directory
$ pack build vscodium --path <application root dir> --buildpack . --builder paketobuildpacks/builder-jammy-base --verbose
# run the built image
$ docker run -u 1001 -p 8000:8000 -e HOME=/workspace -it --rm vscodium 
```

Note that since this buildpack does not actually inspect the application, the application root dir can be just an empty folder.

## Test

You can run the tests with `make test`.

## Packaging

To create a package (.cnb file) run `make package version=<version>` (with "<version>" being something like `1.2.3`).
