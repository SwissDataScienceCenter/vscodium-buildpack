# VSCodium Buildpack

A buildpack that packages browser based vscodium into an image.

To test this buildpack, run

```bash
$ ./scripts/build.sh
$ pack build vscodium --path <application root dir> --buildpack . --builder paketobuildpacks/builder-jammy-base --verbose
$ docker run -u 1001 -p 8000:8000 -e HOME=/workspace -it --rm vscodium 
```

Note that since this buildpack does not actually inspect the application, the application root dir can be just an empty folder.
