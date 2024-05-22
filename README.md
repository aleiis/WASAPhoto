# WASAPhoto

[Web and Software Architecture](http://gamificationlab.uniroma1.it/en/wasa/) homework project.

## Project description

Each user will be presented with a stream of photos (images) in reverse chronological order, with information about when each photo was uploaded (date and time) and how many likes and comments it has. The stream is composed by photos from “following” (other users that the user follows). Users can place (and later remove) a “like” to photos from other users. Also, users can add comments to any image (even those uploaded by themself). Only authors can remove their comments.

Users can ban other users. If user Alice bans user Eve, Eve won’t be able to see anyinformation about Alice. Alice can decide to remove the ban at anymoment.

Users will have their profiles. The personal profile page for the user shows: the user’s photos (in reverse chronological order), how many photos have been uploaded, and the user’s followers and following. Users can change their usernames, upload photos, remove photos, and follow/unfollow other users. Removal of an image will also remove likes and comments. A user can search other user profiles via username. 

A user can login just by specifying the username.

"Fantastic coffee (decaffeinated)" is a simplified version for the WASA course, not suitable for a production environment.
The full version can be found in the "Fantastic Coffee" repository.

## Project structure

The project follows the "Fantastic coffee (decaffeinated)" pattern, a simplified version for the WASA course, not suitable for a production environment.
The full version can be found in the "Fantastic Coffee" repository.

* `cmd/` contains all executables; Go programs here should only do "executable-stuff", like reading options from the CLI/env, etc.
	* `cmd/healthcheck` is an example of a daemon for checking the health of servers daemons; useful when the hypervisor is not providing HTTP readiness/liveness probes (e.g., Docker engine)
	* `cmd/webapi` contains an example of a web API server daemon
* `demo/` contains a demo config file
* `doc/` contains the documentation (usually, for APIs, this means an OpenAPI file)
* `service/` has all packages for implementing project-specific functionalities
	* `service/api` contains an example of an API server
	* `service/globaltime` contains a wrapper package for `time.Time` (useful in unit testing)
* `vendor/` is managed by Go, and contains a copy of all dependencies
* `webui/` is an example of a web frontend in Vue.js; it includes:
	* Bootstrap JavaScript framework
	* a customized version of "Bootstrap dashboard" template
	* feather icons as SVG
	* Go code for release embedding

Other project files include:
* `open-npm.sh` starts a new (temporary) container using `node:lts` image for safe web frontend development (you don't want to use `npm` in your system, do you?)

## Go vendoring

This project uses [Go Vendoring](https://go.dev/ref/mod#vendoring). You must use `go mod vendor` after changing some dependency (`go get` or `go mod tidy`) and add all files under `vendor/` directory in your commit.

For more information about vendoring:

* https://go.dev/ref/mod#vendoring
* https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html

## Node/NPM vendoring

This repository contains the `webui/node_modules` directory with all dependencies for Vue.JS. You should commit the content of that directory and both `package.json` and `package-lock.json`.

## How to build

If you're not using the WebUI, or if you don't want to embed the WebUI into the final executable, then:

```shell
go build ./cmd/webapi/
```

If you're using the WebUI and you want to embed it into the final executable:

```shell
./open-npm.sh
# (here you're inside the NPM container)
npm run build-embed
exit
# (outside the NPM container)
go build -tags webui ./cmd/webapi/
```

## How to run (in development mode)

You can launch the backend only using:

```shell
go run ./cmd/webapi/
```

If you want to launch the WebUI, open a new tab and launch:

```shell
./open-npm.sh
# (here you're inside the NPM container)
npm run dev
```

## How to build for production / homework delivery

```shell
./open-npm.sh
# (here you're inside the NPM container)
npm run build-prod
```

## License

See [LICENSE](LICENSE).
