# Changelog

## [0.2.4](https://github.com/home-operations/echo/compare/0.2.3...0.2.4) (2026-08-27)


### Features

* **go:** update module github.com/kimmachinegun/automemlimit (v0.7.5 → v1.0.0) ([#91](https://github.com/home-operations/echo/issues/91)) ([95ccf2d](https://github.com/home-operations/echo/commit/95ccf2d3c0bcfd546f92b13b9ac35e43cf3ade57))


### Bug Fixes

* drop UPX compression, packed binaries segfault (upx/upx[#18902](https://github.com/home-operations/echo/issues/18902)) ([#92](https://github.com/home-operations/echo/issues/92)) ([d7fd6ef](https://github.com/home-operations/echo/commit/d7fd6ef6318bf3d5cf999998bdcf2d667dee07ae))


### Miscellaneous Chores

* **github-action:** update action docker/github-builder (v1.16.0 → v1.17.0) ([#90](https://github.com/home-operations/echo/issues/90)) ([5fd0eaf](https://github.com/home-operations/echo/commit/5fd0eaf50b16af86a995793f107b043556a78e76))
* **github-action:** update action docker/setup-buildx-action (v4.2.0 → v4.3.0) ([#87](https://github.com/home-operations/echo/issues/87)) ([bc39b6d](https://github.com/home-operations/echo/commit/bc39b6dc5a938ac59b4c545d0e12d47134070bdd))
* **mise:** update mise tools ([#89](https://github.com/home-operations/echo/issues/89)) ([9861ad2](https://github.com/home-operations/echo/commit/9861ad210bb9d8a1dd456c8787d4d7ac9dbc466d))
* **mise:** update tool oxfmt (0.63.0 → 0.64.0) ([#85](https://github.com/home-operations/echo/issues/85)) ([818fed8](https://github.com/home-operations/echo/commit/818fed87625acae1669ce9c57ad6cfa0d834d047))
* **mise:** update tool yq (4.53.3 → 4.53.4) ([#86](https://github.com/home-operations/echo/issues/86)) ([0f1eb0a](https://github.com/home-operations/echo/commit/0f1eb0adc80dcc457da92002fdc568dbf75fe571))

## [0.2.3](https://github.com/home-operations/echo/compare/0.2.2...0.2.3) (2026-08-21)


### Bug Fixes

* **chart:** keep example comments out of unrelated schema descriptions ([#83](https://github.com/home-operations/echo/issues/83)) ([6ce9ab2](https://github.com/home-operations/echo/commit/6ce9ab2dd22f867a671a58244d599c17e51ecc1a))
* **ci:** fail the merge gate on cancelled jobs, and key the lint cache on the toolchain ([#58](https://github.com/home-operations/echo/issues/58)) ([b4afcdd](https://github.com/home-operations/echo/commit/b4afcdd065ab0eb101a068f22e8ee9b9e205b6ca))
* **go:** update module go (1.26.4 → 1.26.5) ([#70](https://github.com/home-operations/echo/issues/70)) ([8661ad5](https://github.com/home-operations/echo/commit/8661ad5dd3a64d173de75140c674056aa53808e4))
* **go:** update to go 1.27.0 ([#84](https://github.com/home-operations/echo/issues/84)) ([f9ba0a6](https://github.com/home-operations/echo/commit/f9ba0a6e5cf9b39a491e12071a32bc4ca819d7ad))


### Documentation

* add AGENTS.md with Go conventions ([#62](https://github.com/home-operations/echo/issues/62)) ([b218df4](https://github.com/home-operations/echo/commit/b218df464d345f43130bb1520016a5fc90b51026))
* make AGENTS.md a generic, drift-resistant template ([#64](https://github.com/home-operations/echo/issues/64)) ([28d45d9](https://github.com/home-operations/echo/commit/28d45d9e2846088dfb4ab8505451733b3706aa56))


### Build System

* **mise:** add actionlint and refresh the lockfile ([#50](https://github.com/home-operations/echo/issues/50)) ([beda1ff](https://github.com/home-operations/echo/commit/beda1fff5978cde643a56ae1e3fefa84a84a1f6f))


### Continuous Integration

* gate pull requests on a single Build Success check ([#49](https://github.com/home-operations/echo/issues/49)) ([8c394eb](https://github.com/home-operations/echo/commit/8c394eb7e9fa8e6c242900d08ecc3f6355976e5e))
* **github-action:** Update action actions/stale (v10.4.0 → v11.0.0) ([#59](https://github.com/home-operations/echo/issues/59)) ([6255ec0](https://github.com/home-operations/echo/commit/6255ec06b1634f6023c623a876bdb818b57b7505))
* **github-action:** Update action docker/github-builder (v1.14.0 → v1.15.0) ([#57](https://github.com/home-operations/echo/issues/57)) ([25c7b5b](https://github.com/home-operations/echo/commit/25c7b5b0792a2e58c88937555ab586074b55787f))
* **github-action:** Update action docker/github-builder (v1.15.0 → v1.16.0) ([#73](https://github.com/home-operations/echo/issues/73)) ([6fc568c](https://github.com/home-operations/echo/commit/6fc568cbbcc68075cb1b8c0387174d5ba21347fe))
* **github-action:** Update action docker/login-action (v4.5.1 → v4.5.2) ([#60](https://github.com/home-operations/echo/issues/60)) ([85c0d4d](https://github.com/home-operations/echo/commit/85c0d4d6fab32cea1c10f3e79ae48f7799934918))
* **github-action:** Update action docker/login-action (v4.5.2 → v4.6.0) ([#63](https://github.com/home-operations/echo/issues/63)) ([5cf5338](https://github.com/home-operations/echo/commit/5cf53380d775342afef7cd903240ee85513f0ecf))
* **github-action:** Update action home-operations/.github/actions/workflow-lint (v1.0.2 → v1.0.3) ([#69](https://github.com/home-operations/echo/issues/69)) ([42bac11](https://github.com/home-operations/echo/commit/42bac118c36f2f9080e0b65329c73456d65d5d5b))
* **github-action:** Update action jdx/mise-action (v4.2.1 → v4.2.2) ([#52](https://github.com/home-operations/echo/issues/52)) ([c9298e0](https://github.com/home-operations/echo/commit/c9298e094cc35a86b339951349818d652461f011))
* **github-action:** Update action jdx/mise-action (v4.2.3 → v4.2.4) ([#71](https://github.com/home-operations/echo/issues/71)) ([5283840](https://github.com/home-operations/echo/commit/5283840d16ab0125233f00f65a946cc8652a7109))
* **github-action:** Update github-actions ([#53](https://github.com/home-operations/echo/issues/53)) ([d6334ce](https://github.com/home-operations/echo/commit/d6334ceea17efb1c7f31a619ad5be6dcb6902a22))
* **github-action:** update workflow-lint action (1.0.0 → v1.0.2) ([#66](https://github.com/home-operations/echo/issues/66)) ([967c73a](https://github.com/home-operations/echo/commit/967c73a08954fc32cc46fe03620efea0c598660f))
* lint workflows with the shared composite action ([#51](https://github.com/home-operations/echo/issues/51)) ([c668830](https://github.com/home-operations/echo/commit/c6688306f1603cce30a4783d9db6f3f5fdaf8a6a))
* skip release-please version-bump PRs in checks ([#48](https://github.com/home-operations/echo/issues/48)) ([2de44cc](https://github.com/home-operations/echo/commit/2de44ccad75994d8c94ce69a55979913ca469b32))
* wire govulncheck into mise and CI ([#68](https://github.com/home-operations/echo/issues/68)) ([2eb630c](https://github.com/home-operations/echo/commit/2eb630cd2586bebec5694c61ea3f88426c7413e0))


### Miscellaneous Chores

* **github-action:** update action jdx/mise-action (v4.2.4 → v4.2.5) ([#78](https://github.com/home-operations/echo/issues/78)) ([d4e845d](https://github.com/home-operations/echo/commit/d4e845d6d5c6f06bca14ef24f1eaf16a47526f1a))
* **go:** pin go directive to 1.26.0 ([#79](https://github.com/home-operations/echo/issues/79)) ([88e2148](https://github.com/home-operations/echo/commit/88e2148d7ee85fc7eebac07672055a07ea38ea46))
* **mise:** prune lockfile to used platforms ([#67](https://github.com/home-operations/echo/issues/67)) ([32591e5](https://github.com/home-operations/echo/commit/32591e54149ec5a53b5a52217beae7fdac925e7a))
* **mise:** Update tool cosign (3.1.2 → 3.1.3) ([#74](https://github.com/home-operations/echo/issues/74)) ([5edab8a](https://github.com/home-operations/echo/commit/5edab8a17fc4f8ff8b698154f642066052eb11b3))
* **mise:** update tool go (1.26.5 → 1.26.6) ([#82](https://github.com/home-operations/echo/issues/82)) ([9d4a187](https://github.com/home-operations/echo/commit/9d4a187dc6ca20350e9685f62032031552f2cdb3))
* **mise:** update tool go:golang.org/x/vuln/cmd/govulncheck (1.6.0 → v1.7.0) ([#77](https://github.com/home-operations/echo/issues/77)) ([731f12d](https://github.com/home-operations/echo/commit/731f12d00660943587dc70a6b39a98e2cc05fed3))
* **mise:** update tool helm (4.2.3 → 4.2.4) ([#81](https://github.com/home-operations/echo/issues/81)) ([32b85b4](https://github.com/home-operations/echo/commit/32b85b4b1cc43c2d59688e1e41709ef1e7432d4c))
* **mise:** Update tool oxfmt (0.60.0 → 0.61.0) ([#54](https://github.com/home-operations/echo/issues/54)) ([d40cd22](https://github.com/home-operations/echo/commit/d40cd22d1c959cdb0657aa1fcb0926f3c2279be1))
* **mise:** Update tool oxfmt (0.61.0 → 0.62.0) ([#72](https://github.com/home-operations/echo/issues/72)) ([5b7a778](https://github.com/home-operations/echo/commit/5b7a778707fdd6b9d262a0aeabd76f4cf7d927e4))
* **mise:** Update tool oxfmt (0.62.0 → 0.63.0) ([#75](https://github.com/home-operations/echo/issues/75)) ([7114dc7](https://github.com/home-operations/echo/commit/7114dc73d6ba787100d638ee144ad2df456a45a5))
* **mise:** Update tool zizmor (1.28.0 → 1.29.0) ([#65](https://github.com/home-operations/echo/issues/65)) ([4ad14aa](https://github.com/home-operations/echo/commit/4ad14aa759df077a7f22d5157002b7bc7dcc2d33))
* **release-please:** standardize the release pull request title pattern ([#61](https://github.com/home-operations/echo/issues/61)) ([d13101b](https://github.com/home-operations/echo/commit/d13101b3ea4b4123189da0e4bef6fca5e64b74c3))
* standardize release-please changelog sections ([#56](https://github.com/home-operations/echo/issues/56)) ([007ec54](https://github.com/home-operations/echo/commit/007ec54cc6ed20151d024c8af90d043f06cf2b9a))

## [0.2.2](https://github.com/home-operations/echo/compare/0.2.1...0.2.2) (2026-07-24)


### Features

* **deps:** update module github.com/prometheus/client_golang (v1.23.2 → v1.24.0) ([#41](https://github.com/home-operations/echo/issues/41)) ([a26d502](https://github.com/home-operations/echo/commit/a26d502719e30a218041f4fe6c1f90db81d7d868))
* **deps:** update module golang.org/x/sync (v0.21.0 → v0.22.0) ([#34](https://github.com/home-operations/echo/issues/34)) ([c04c282](https://github.com/home-operations/echo/commit/c04c28296cca3e86322f281a30cdd27470bb77b9))


### Bug Fixes

* **deps:** update module github.com/prometheus/client_golang (v1.24.0 → v1.24.1) ([#45](https://github.com/home-operations/echo/issues/45)) ([495cc58](https://github.com/home-operations/echo/commit/495cc58cfe6dd65a4c648fb129067f9ea91715e3))
* **helm:** stamp Chart.yaml version on release ([#47](https://github.com/home-operations/echo/issues/47)) ([bbb7458](https://github.com/home-operations/echo/commit/bbb745854a443f718a6b7d64a02c05a98bc835e4))


### Styles

* indent markdown at 2 to match embedded yaml ([#42](https://github.com/home-operations/echo/issues/42)) ([71318c2](https://github.com/home-operations/echo/commit/71318c27f92677821c96d80173c5ba7161dfedf0))


### Miscellaneous Chores

* **github-release:** Update release helm-unittest/helm-unittest (v1.1.1 → v1.1.2) ([#46](https://github.com/home-operations/echo/issues/46)) ([ce2033e](https://github.com/home-operations/echo/commit/ce2033e12406601170e8b1e0f45f1df6435f61a7))
* **mise:** Update tool cosign (3.1.1 → 3.1.2) ([#40](https://github.com/home-operations/echo/issues/40)) ([0f0c7a6](https://github.com/home-operations/echo/commit/0f0c7a6f0538e20e554f9f7647a63095012363b7))
* **mise:** Update tool go (1.26.4 → 1.26.5) ([#35](https://github.com/home-operations/echo/issues/35)) ([a31848d](https://github.com/home-operations/echo/commit/a31848d80523c7ea6a84197d4033dc659e223c40))
* **mise:** Update tool helm (4.2.2 → 4.2.3) ([#36](https://github.com/home-operations/echo/issues/36)) ([ff36c16](https://github.com/home-operations/echo/commit/ff36c162ac1f62c230f2509baa0f963747e2a25b))
* **mise:** Update tool lefthook (2.1.9 → 2.1.10) ([#33](https://github.com/home-operations/echo/issues/33)) ([a2b1dda](https://github.com/home-operations/echo/commit/a2b1dda9a217e09d54eaca6353d07eb7cb4affcc))
* **mise:** Update tool oxfmt (0.57.0 → 0.58.0) ([#31](https://github.com/home-operations/echo/issues/31)) ([193a734](https://github.com/home-operations/echo/commit/193a734c1e9fb1afaa0a65ee619d8bb50dcd7c5c))
* **mise:** Update tool oxfmt (0.58.0 → 0.59.0) ([#37](https://github.com/home-operations/echo/issues/37)) ([ce34281](https://github.com/home-operations/echo/commit/ce342816d23581d13aaddae55ce326ab55eb5cb7))
* **mise:** Update tool oxfmt (0.59.0 → 0.60.0) ([#44](https://github.com/home-operations/echo/issues/44)) ([15b9b71](https://github.com/home-operations/echo/commit/15b9b711b6ff490b691192a9eef8cb0cbfe5266a))
* **mise:** Update tool zizmor (1.26.1 → 1.27.0) ([#38](https://github.com/home-operations/echo/issues/38)) ([872f91b](https://github.com/home-operations/echo/commit/872f91b45fabbfacb739e4260c6150cecea9453f))
* **mise:** Update tool zizmor (1.27.0 → 1.28.0) ([#43](https://github.com/home-operations/echo/issues/43)) ([7ede577](https://github.com/home-operations/echo/commit/7ede577f6c9409915816148a07c18344869be7f3))

## [0.2.1](https://github.com/home-operations/echo/compare/0.2.0...0.2.1) (2026-07-04)


### Bug Fixes

* review findings — probe log spam, 1xx echo-code, Set-Cookie domain scoping, WS idle bound ([#29](https://github.com/home-operations/echo/issues/29)) ([699e4d1](https://github.com/home-operations/echo/commit/699e4d1f8c7f320b8ad248dbaddea7171c1b531c))

## [0.2.0](https://github.com/home-operations/echo/compare/0.1.3...0.2.0) (2026-07-04)


### ⚠ BREAKING CHANGES

* serve health on the main port; metrics port becomes fully optional ([#28](https://github.com/home-operations/echo/issues/28))

### Features

* serve health on the main port; metrics port becomes fully optional ([#28](https://github.com/home-operations/echo/issues/28)) ([d439355](https://github.com/home-operations/echo/commit/d43935597061bb2992814b6941bc0e224b6255b5))


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
