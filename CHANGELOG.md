# Changelog

## [0.1.4](https://github.com/home-operations/echo/compare/0.1.3...0.1.4) (2026-07-01)


### Miscellaneous Chores

* **mise:** Update tool oxfmt (0.56.0 → 0.57.0) ([#27](https://github.com/home-operations/echo/issues/27)) ([50211ba](https://github.com/home-operations/echo/commit/50211ba46f7c5a1c47e3dfb8601279f7b66806f2))
* **renovate:** inherit shared toolchain + chart-docs presets ([#24](https://github.com/home-operations/echo/issues/24)) ([4577ea0](https://github.com/home-operations/echo/commit/4577ea0a08a06899898e3b8df294c450495671b7))

## [0.1.3](https://github.com/home-operations/echo/compare/0.1.2...0.1.3) (2026-06-24)


### Features

* caller-controlled response shaping and pretty-print ([#23](https://github.com/home-operations/echo/issues/23)) ([8507eb7](https://github.com/home-operations/echo/commit/8507eb72290282fad423f19dc7f4d4654344fce3))
* **container:** update image mirror.gcr.io/curlimages/curl (8.20.0 → 8.21.0) ([#22](https://github.com/home-operations/echo/issues/22)) ([3bbc295](https://github.com/home-operations/echo/commit/3bbc295555ed38b7929ed6854b8edfe6ae718c73))


### Miscellaneous Chores

* **mise:** Update tool oxfmt (0.55.0 → 0.56.0) ([#20](https://github.com/home-operations/echo/issues/20)) ([231f094](https://github.com/home-operations/echo/commit/231f09430ce2f4c017c522522c0af459085260b6))
* **mise:** Update tool zizmor (1.25.2 → 1.26.1) ([#18](https://github.com/home-operations/echo/issues/18)) ([e4ebc1a](https://github.com/home-operations/echo/commit/e4ebc1a6ea6f86bc231dde4a49b921707bc18829))
* update Renovate configuration for Go toolchain ([71012e8](https://github.com/home-operations/echo/commit/71012e80787583e628fe25c2a01a97dd2dc70d54))

## [0.1.2](https://github.com/home-operations/echo/compare/0.1.1...0.1.2) (2026-06-18)


### Features

* serve metrics and health probes on 8081 ([#17](https://github.com/home-operations/echo/issues/17)) ([40712fd](https://github.com/home-operations/echo/commit/40712fde1866e5f15bdb753a934fff4315ad9eac))


### Miscellaneous Chores

* **mise:** update tool helm (4.2.1 → 4.2.2) ([#14](https://github.com/home-operations/echo/issues/14)) ([c4b8cbd](https://github.com/home-operations/echo/commit/c4b8cbdcf072c065e87df1ba188b96d3e0a3411b))

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
