# Homepage — agent instructions

**Mandatory first read:** [current merged Homepage suite policy](https://github.com/richard-mauri/homepage/blob/main/SUITE_DEVELOPMENT_POLICY.md). A working-branch policy proposal is not authoritative. Follow its BEGIN, NEXT, FINISH, Issue bootstrap and checkpoint rules. Read [README.md](README.md), [ARCHITECTURE.md](ARCHITECTURE.md), [DEVELOPMENT.md](DEVELOPMENT.md) and [CURRENT_WORK.md](CURRENT_WORK.md) from the active project branch; reconcile actual GitHub state.

Homepage-specific architecture and build/acceptance procedures are in ARCHITECTURE and DEVELOPMENT. Preserve loopback listener, embedded assets, same-tab destinations, private credentials, data and tunnel isolation. Do not expose tokens, logs, private configuration or build artifacts. No runtime restart without separate owner authorization.
