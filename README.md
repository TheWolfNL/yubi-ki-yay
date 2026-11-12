# Yubi Ki-Yay

A macOS menu-bar app that tells you when your YubiKey wants a touch. YubiKey's only cue is a small blinking LED, which is easy to miss.

Website: https://thewolfnl.github.io/yubi-ki-yay/ (source in `docs/`, deployed by `.github/workflows/pages.yml`).

## How it works

Nothing on macOS reports that a YubiKey (OpenPGP applet behind `gpg-agent`) is waiting for a touch, so hooks fire *before* the operation that needs one:

- **SSH**: a `Match exec` line in `~/.ssh/config` runs a shim before connecting to the hosts whose key lives on the YubiKey.
- **git signing**: `gpg.program` points at a wrapper that alerts before signing (never on `--verify`).

Both call `yubikiyay notify <label>`. The CLI prints the terminal line itself (only it owns your tty) and hands the rest to the running app over `~/Library/Application Support/yubi-ki-yay/notify.sock`. When the app isn't running, the CLI starts it in the background (`open -g`) and covers that one alert with `osascript` + `afplay`. Turn on **Start at login** in Settings to keep it running.

The CLI walks its parent processes to say what triggered the touch (e.g. `git push · my-repo · via Claude Code`). The app slides a banner onto the screen you are working on (a non-activating panel; top right by default, configurable to top center/left or bottom right/left), plays a system sound and flashes **Touch!** in the menu bar. Native macOS notifications are not used: macOS delivers them for ad-hoc signed apps but suppresses their banners. Click the banner or press Escape to dismiss it; it hides by itself after 10 s. Settings live in `~/.config/yubi-ki-yay/config.json`. Use **＋** next to the sound picker to add your own audio file (aiff, caf, m4a, mp3 or wav); it is copied to `~/.config/yubi-ki-yay/sounds/`.

The menu bar icon can be shown always, only while a touch is needed (it also shows for 5 s at startup), or never. With the icon hidden, reach Settings by opening the app again (Spotlight/Finder) or running `yubikiyay settings`.

## Install

```bash
brew install --cask thewolfnl/tap/yubi-ki-yay
```

The cask installs the app, links the CLI as `yubikiyay` and clears the quarantine flag (the app is ad-hoc signed, not notarized). The CLI then lives at `$(brew --prefix)/bin/yubikiyay` instead of `~/.local/bin/yubikiyay`; adjust the hook paths below.

## Build from source

Needs Go, Node 20+ and the Wails v3 CLI (`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`).

```bash
wails3 task install
```

This builds the app, installs it to `~/Applications/Yubi Ki-Yay.app` and links the CLI to `~/.local/bin/yubikiyay`.

## Hooks

`~/.ssh/config`:

```
Match originalhost github-coconut exec "~/.local/bin/yk-touch-notify %n"
```

`~/.local/bin/yk-touch-notify` (must exit non-zero so the Match block never applies):

```bash
#!/usr/bin/env bash
"$HOME/.local/bin/yubikiyay" notify "ssh ${1:-?}"
exit 1
```

git (per directory via `includeIf`): `gpg.program = ~/.local/bin/yk-gpg`, a wrapper that calls `yubikiyay notify "git signing"` unless the arguments contain `--verify`, then execs `gpg`.

## Notes

- Vite is pinned to 6 because Vite 8 requires Node ≥ 20.19.
- The icon is a YubiKey in a grubby white tank top. Source SVGs are in `build/icon-src/`; regenerate with `inkscape` and `wails3 task common:generate:icons`.

## Releasing

Run `wails3 task release` on an up-to-date, clean `main`. It shows the commits since the last tag, suggests a major/minor/patch bump from Conventional Commit prefixes (`feat:` minor, `!` or `BREAKING CHANGE:` major, otherwise patch; the first release is v0.1.0), creates a signed `vX.Y.Z` tag and asks before pushing it. `.github/workflows/release.yml` builds a universal (arm64 + amd64) app, ad-hoc signs it and attaches `Yubi-Ki-Yay-X.Y.Z-universal.zip` plus its `.sha256` to a GitHub release. It then sends a `bump-cask` dispatch to [TheWolfNL/homebrew-tap](https://github.com/TheWolfNL/homebrew-tap), which updates the cask. This needs a `TAP_DISPATCH_TOKEN` secret: a fine-grained token for the tap repo with Contents read and write. Without it the step is skipped and you run the tap's **Bump cask** workflow by hand.

## License

[PolyForm Shield 1.0.0](LICENSE.md). Required Notice: Copyright Tjeerd Ritsma (https://github.com/TheWolfNL)
