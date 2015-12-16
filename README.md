# Couchlock


[![Travis](https://img.shields.io/travis/tomologic/couchlock.svg?style=flat-square)](https://travis-ci.org/tomologic/couchlock)
[![](https://badge.imagelayers.io/tomologic/couchlock:latest.svg)](https://imagelayers.io/?images=tomologic/couchlock:latest 'Get your own badge on imagelayers.io')

Couchlock those running pipelines!

## What is this?

This is currently the way we are trying to get global locks into our CD pipelines. Historically tried different Jenkins plugins for mutexes/semaphores and hacks by using jobs as locks. 

Lock that shared resource!

- Build machine
- Build only one image for specific githash of docker image (at the same time)
- Postgres server
- Staging system
- Single Arduino board hooked up to Jenkins

```
docker run -it --rm tomologic/couchlock \
    --couchdb "https://USER.cloudant.com/couchlock" \
    --lock shared_resource \
    --name $BUILD_TAG \ # Unique identifier, example Jenkins $BUILD_TAG
    lock

## Do crazy things with shared resource

docker run -it --rm tomologic/couchlock \
    --couchdb "https://USER.cloudant.com/couchlock" \
    --lock shared_resource \
    --name $BUILD_TAG \ # Unique identifier, example Jenkins $BUILD_TAG
    unlock
```

Couchlock is provided through our [homebrew tap](https://github.com/tomologic/homebrew-tap) for OSX and linux users.

```
brew install tomologic/tap/couchlock
```

## Build

[Travis CI Gimme](https://github.com/travis-ci/gimme) is used to cross-compile couchlock.

```
$ brew install gimme
```

```
$ make build
$ make build_darwin
$ make build_linux
```

## Package

[Effing Package Management - fpm](https://github.com/jordansissel/fpm) is used for packaging.

```
$ brew install gnu-tar
$ brew install rpm
$ gem install fpm
```

```
$ make package
$ make package_deb
$ make package_rpm
```

## CouchDB

If you don't have couchdb internally then totally try this out with [Cloudant](https://cloudant.com/) which has a great free tier.

## Development

### Generate bindata

All files in data directory is packed with [go-bindata](https://github.com/jteeuwen/go-bindata).

```
go get -u github.com/jteeuwen/go-bindata/...
go-bindata data/...
```
