# Changelog

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.3.0](https://github.com/schnurbus/go-kegelmaster/compare/v0.2.1...v0.3.0) (2026-02-23)


### Features

* **footer:** add Buy Me a Coffee link via public config endpoint ([#44](https://github.com/schnurbus/go-kegelmaster/issues/44)) ([23a6560](https://github.com/schnurbus/go-kegelmaster/commit/23a6560f58995888a63570523ef9bb233c2f8ae9))
* **frontend:** show money inputs with two decimal places ([#43](https://github.com/schnurbus/go-kegelmaster/issues/43)) ([73dd254](https://github.com/schnurbus/go-kegelmaster/commit/73dd254fc415e2e429e13947c34e1724e0814d68))


### Bug Fixes

* **frontend:** fixed double v character ([#41](https://github.com/schnurbus/go-kegelmaster/issues/41)) ([0014929](https://github.com/schnurbus/go-kegelmaster/commit/0014929aba5794cbe8155d1bf57b78b1924793cb))
* **frontend:** Seitenmenü-Flackern durch persistentes Dashboard-Layout beheben ([#45](https://github.com/schnurbus/go-kegelmaster/issues/45)) ([ef0baf0](https://github.com/schnurbus/go-kegelmaster/commit/ef0baf08958a877c01cced7aef2747d4102272db))

## [0.2.1](https://github.com/schnurbus/go-kegelmaster/compare/v0.2.0...v0.2.1) (2026-02-21)


### Bug Fixes

* **ci:** use PAT for Release Please so docker-release triggers on release ([#39](https://github.com/schnurbus/go-kegelmaster/issues/39)) ([2a69691](https://github.com/schnurbus/go-kegelmaster/commit/2a696919b44f38aa8f364e91ec7b031cee7a7e00))

## [0.2.0](https://github.com/schnurbus/go-kegelmaster/compare/v0.1.0...v0.2.0) (2026-02-21)


### Features

* **auth:** Login/Register UI, Fehlermeldungen, Passwort-Reset, Merken ([#13](https://github.com/schnurbus/go-kegelmaster/issues/13)) ([fde0d75](https://github.com/schnurbus/go-kegelmaster/commit/fde0d75924c9c66b996e605f7c043fa12043b7de))
* **balance:** club start balance, recalculation, and UI ([#14](https://github.com/schnurbus/go-kegelmaster/issues/14)) ([b5cdd2b](https://github.com/schnurbus/go-kegelmaster/commit/b5cdd2b400bc35acb7f7e8230ff57220986a5136))
* **chatbot:** add help chatbot with Gemini ([#30](https://github.com/schnurbus/go-kegelmaster/issues/30)) ([5d530cb](https://github.com/schnurbus/go-kegelmaster/commit/5d530cb8976dcd24e5ffd7ec1b0ea4f736b72760))
* CLI-Import für Spieltag-CSV ([#8](https://github.com/schnurbus/go-kegelmaster/issues/8)) ([c43de35](https://github.com/schnurbus/go-kegelmaster/commit/c43de356ae3c483052be1fc064e9832e036cb742))
* **club:** Paar-Modus für Einzahlungen und Spielerauswahl sortiert ([#18](https://github.com/schnurbus/go-kegelmaster/issues/18)) ([ac86be0](https://github.com/schnurbus/go-kegelmaster/commit/ac86be0d3be281b59854b37a9674b255237e4b06))
* **club:** transfer owner and delete club with password confirmation ([#22](https://github.com/schnurbus/go-kegelmaster/issues/22)) ([b4cb7c7](https://github.com/schnurbus/go-kegelmaster/commit/b4cb7c7269e24968bee895786a4b1d16e0b458a9))
* competitions ([#6](https://github.com/schnurbus/go-kegelmaster/issues/6)) ([659dd82](https://github.com/schnurbus/go-kegelmaster/commit/659dd827f5b3d2a953492be55ec6dd7a3dad482c))
* dashboard improvements ([#5](https://github.com/schnurbus/go-kegelmaster/issues/5)) ([f383892](https://github.com/schnurbus/go-kegelmaster/commit/f383892cc7c912c5bcac093299532771ff9438d5))
* **dashboard:** competition charts, penalty history, and my transactions table ([#20](https://github.com/schnurbus/go-kegelmaster/issues/20)) ([329794c](https://github.com/schnurbus/go-kegelmaster/commit/329794ce5fe07c952abc3999329a340637366825))
* **db:** auto-migration on startup with AUTO_MIGRATE flag ([#33](https://github.com/schnurbus/go-kegelmaster/issues/33)) ([dc1a39f](https://github.com/schnurbus/go-kegelmaster/commit/dc1a39f09bcadeb2a607fd1fd76906ef7ef6aacc))
* **docs:** add end-user help page and OpenAPI/Swagger API docs ([#32](https://github.com/schnurbus/go-kegelmaster/issues/32)) ([643f330](https://github.com/schnurbus/go-kegelmaster/commit/643f33042cffaaad680b7a3a8314280965bac193))
* game days and fees added ([5239250](https://github.com/schnurbus/go-kegelmaster/commit/5239250bfb6ba492f2f016831a0fb724f24a25f4))
* game days and fees added ([1945730](https://github.com/schnurbus/go-kegelmaster/commit/19457307299ed67afc19ef78813c043a67d16862))
* **gameday:** add draft status for game days ([#36](https://github.com/schnurbus/go-kegelmaster/issues/36)) ([305422e](https://github.com/schnurbus/go-kegelmaster/commit/305422e6137affca874535adc188818d56fecb4a))
* Gleitkomma-Anzahl bei Fee Types (Fixed-Point) ([#10](https://github.com/schnurbus/go-kegelmaster/issues/10)) ([8277ffb](https://github.com/schnurbus/go-kegelmaster/commit/8277ffbd9ab2f7bf4517fed98582deaad7195655))
* invitation added ([#4](https://github.com/schnurbus/go-kegelmaster/issues/4)) ([4c5b55e](https://github.com/schnurbus/go-kegelmaster/commit/4c5b55e7e3eaa5f0cfa50bcdc6f033187bd3683a))
* old data migration ([#17](https://github.com/schnurbus/go-kegelmaster/issues/17)) ([71a4e27](https://github.com/schnurbus/go-kegelmaster/commit/71a4e2794084afcb7fd927f77a2fc23ad1016d3a))
* **permissions:** add transactions to role permission matrix ([#27](https://github.com/schnurbus/go-kegelmaster/issues/27)) ([4565f9b](https://github.com/schnurbus/go-kegelmaster/commit/4565f9b67b5d1dcc383a3345916d4a681eb0cd7f))
* **permissions:** tie roles to permissions and improve UI ([#26](https://github.com/schnurbus/go-kegelmaster/issues/26)) ([92adfa7](https://github.com/schnurbus/go-kegelmaster/commit/92adfa7bfb412900e664438b6c10b41f5607badc))
* **players:** add inactive flag (no base fee, UI display) ([#24](https://github.com/schnurbus/go-kegelmaster/issues/24)) ([465d3ca](https://github.com/schnurbus/go-kegelmaster/commit/465d3cac022da3b456a6ece86ee3424b1d3396e7))
* **players:** optional partner and pair balance in Paar-Modus ([#34](https://github.com/schnurbus/go-kegelmaster/issues/34)) ([2860a2d](https://github.com/schnurbus/go-kegelmaster/commit/2860a2dc68fa9a093b967c78d989fd4a66e8ad66))
* **releases:** add Release Please, Docker release tags, footer version, Impressum and Datenschutz ([#37](https://github.com/schnurbus/go-kegelmaster/issues/37)) ([8e43fd9](https://github.com/schnurbus/go-kegelmaster/commit/8e43fd962dc52181ddff57348df5134cdc6e3d07))
* single binary ([#7](https://github.com/schnurbus/go-kegelmaster/issues/7)) ([11b2704](https://github.com/schnurbus/go-kegelmaster/commit/11b270472005ce53d747a6b2c08796a96ae16b33))
* transactions added ([#2](https://github.com/schnurbus/go-kegelmaster/issues/2)) ([2b6daaa](https://github.com/schnurbus/go-kegelmaster/commit/2b6daaa8f080f3c38e0a83ead003c441d45ec285))
* **transactions:** transaction date for manual tx, player required for tip ([#11](https://github.com/schnurbus/go-kegelmaster/issues/11)) ([71d9326](https://github.com/schnurbus/go-kegelmaster/commit/71d9326dd0823b859a0cf0de4b46b4a079426336))
* UI-Verbesserungen, Club-Auswahl, Spieler-Sortierung ([#12](https://github.com/schnurbus/go-kegelmaster/issues/12)) ([1ded381](https://github.com/schnurbus/go-kegelmaster/commit/1ded3812d89d9bf3b4604392654c6c738d59579e))
* **ui:** hide inactive players filter, game day sort and delete redirect ([#28](https://github.com/schnurbus/go-kegelmaster/issues/28)) ([8caeeb5](https://github.com/schnurbus/go-kegelmaster/commit/8caeeb5a31c4b3811d9f9c543c7a71f4b68c9033))


### Bug Fixes

* **club-switcher:** clubs nach Login laden ([#21](https://github.com/schnurbus/go-kegelmaster/issues/21)) ([9ad6677](https://github.com/schnurbus/go-kegelmaster/commit/9ad66774079dc3481fe4b66ab7a26afce6cfa15a))
* **frontend:** hide game day edit UI when user lacks update permission ([#31](https://github.com/schnurbus/go-kegelmaster/issues/31)) ([2df249c](https://github.com/schnurbus/go-kegelmaster/commit/2df249cb92222e44864a95090135e10b879b5b74))
* **frontend:** transaction dialog usable on mobile and footer spacing ([#35](https://github.com/schnurbus/go-kegelmaster/issues/35)) ([20ffb12](https://github.com/schnurbus/go-kegelmaster/commit/20ffb12a91e14ba0ecb985826905b7820d9b68b3))
* increase db connection ([#29](https://github.com/schnurbus/go-kegelmaster/issues/29)) ([10657d4](https://github.com/schnurbus/go-kegelmaster/commit/10657d43ae962f9ad360407e6381dfdb9fb2734f))
* **migrate-old:** Strafen-Count, Vorzeichen und Transaktionen korrigieren ([#23](https://github.com/schnurbus/go-kegelmaster/issues/23)) ([e973e1f](https://github.com/schnurbus/go-kegelmaster/commit/e973e1ffe42dab9b0657c6ba571484ede4646fda))
* Rollen-/User-Zuweisung bei Spielern erhalten ([#9](https://github.com/schnurbus/go-kegelmaster/issues/9)) ([e5f9c3f](https://github.com/schnurbus/go-kegelmaster/commit/e5f9c3fd13c83d4d7c769c5d2c135a756b7abce1))
* **transactions:** apply type filter server-side for correct pagination ([#15](https://github.com/schnurbus/go-kegelmaster/issues/15)) ([8e181ed](https://github.com/schnurbus/go-kegelmaster/commit/8e181ed01a650dd78393ff1e084cce3c8d977b18))

## [0.1.0] - Initial release

- Initial application release (Kegelmaster: clubs, players, game days, transactions).

[0.1.0]: https://github.com/schnurbus/go-kegelmaster/releases/tag/v0.1.0
