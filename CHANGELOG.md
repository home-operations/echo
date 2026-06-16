# Changelog

## [0.1.1](https://github.com/home-operations/echo/compare/0.1.0...0.1.1) (2026-06-16)


### Features

* **server:** log healthz/ping/metrics requests at debug level ([#12](https://github.com/home-operations/echo/issues/12)) ([b5e862a](https://github.com/home-operations/echo/commit/b5e862a4b700d228e2e1e335fd82bf471a65c1da))


### Bug Fixes

* **deps:** update module github.com/coder/websocket (v1.8.14 → v1.8.15) ([#7](https://github.com/home-operations/echo/issues/7)) ([e87d80c](https://github.com/home-operations/echo/commit/e87d80c46b2965fb603354e5dfaa4aac81293374))


### Miscellaneous Chores

* **mise:** update tool oxfmt (0.54.0 → 0.55.0) ([#11](https://github.com/home-operations/echo/issues/11)) ([242512e](https://github.com/home-operations/echo/commit/242512ed8e6bffadbb32280efcec0d5497e63675))


### Code Refactoring

* **echo:** use strings.Cut in isJSON ([#13](https://github.com/home-operations/echo/issues/13)) ([8ef7c2d](https://github.com/home-operations/echo/commit/8ef7c2d3c6c6d9d6889e9e03d86959015f7f14e8))

## 0.1.0 (2026-06-15)


### Features

* **container:** update image mirror.gcr.io/curlimages/curl (8.11.1 → 8.20.0) ([#6](https://github.com/home-operations/echo/issues/6)) ([3cb0647](https://github.com/home-operations/echo/commit/3cb06472c22029f89242923ede1705e132888502))
* echo request server with an OCI Helm chart ([b943f86](https://github.com/home-operations/echo/commit/b943f86f28ab134a2b44a57f270737a3d02642ea))


### Bug Fixes

* **chart:** render ECHO_MAX_BODY_BYTES as an integer, not scientific notation ([5199c00](https://github.com/home-operations/echo/commit/5199c001d75ddd23c95ec3ab4ef6dc51c5c9992d))


### Miscellaneous Chores

* Remove failure diagnostics step from e2e ([a416a50](https://github.com/home-operations/echo/commit/a416a50d0575b57e3eb22551642ec172b9de6401))
