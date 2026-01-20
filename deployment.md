# npm deployment (publish) guide

This repo ships `rlm` as a Go CLI, with an npm wrapper package.

The npm package:
- installs a small Node shim (`bin/rlm.js`) that runs the native binary
- builds the native Go binary during `postinstall` (`scripts/postinstall.js`)

**Important:** This means users must have **Go installed** when installing from npm.

## 0) Prerequisites

- Node.js + npm installed
- Go installed and available in `PATH` (because `postinstall` runs `go build`)
- An npm account

## 1) Pick a package name (avoid collisions)

Right now the package name in `package.json` is:

- `"name": "rlm"`

That name is very likely to be **already taken** on npm.

Recommended options:

- Use a **scoped** package name you control, e.g. `@brainqub3/rlm`
- Or pick a unique unscoped name, e.g. `rlm-cli` or `brainqub3-rlm`

To change the name, edit `package.json`:

```json
{
  "name": "@brainqub3/rlm",
  "version": "0.1.0"
}
```

If you use a scoped name and want it public, publish with `--access public`.

## 2) Verify the package content locally

From the repo root:

```bash
npm pack
```

This produces a `.tgz` tarball and prints its contents.

## 3) Smoke test install locally (recommended)

Install the tarball into an isolated prefix (does not modify your system-global npm):

```bash
PREFIX="$(pwd)/.npm-smoketest"
rm -rf "$PREFIX"
mkdir -p "$PREFIX"

npm install -g --prefix "$PREFIX" ./rlm-0.1.0.tgz

"$PREFIX/bin/rlm" --help
```

## 4) Login to npm

```bash
npm login
npm whoami
```

## 5) Publish

### Unscoped package

```bash
npm publish
```

### Scoped package (public)

```bash
npm publish --access public
```

If you see warnings like:

- `npm warn publish ... auto-corrected some errors in your package.json`

Run:

```bash
npm pkg fix
```

Then publish again.

## 5.1) Important: npm versions are immutable

If you try to publish the same version twice, npm will reject it:

- `You cannot publish over the previously published versions: X.Y.Z.`

In that case, bump the version and re-publish.

## 6) Verify install from npm

After publishing:

```bash
npm view <package-name>

npm i -g <package-name>
rlm --help
```

## 7) Versioning / updates

Update the version in `package.json` before publishing again:

```bash
npm version patch
npm publish
```

(Use `minor` / `major` as needed.)

## Notes / gotchas

- This package currently builds `rlm` during install. If you want installs to work **without Go**, switch to prebuilt binaries and download them during install.
- On Windows, the binary name is built as `rlm.exe`.
