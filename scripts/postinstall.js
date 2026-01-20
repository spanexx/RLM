const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');

function run(cmd, args, opts) {
  const r = spawnSync(cmd, args, { stdio: 'inherit', ...opts });
  if (r.error) {
    throw r.error;
  }
  if (r.status !== 0) {
    const err = new Error(`${cmd} exited with status ${r.status}`);
    err.code = r.status;
    throw err;
  }
}

function main() {
  const pkgRoot = path.resolve(__dirname, '..');
  const outDir = path.join(pkgRoot, '.rlm-bin');
  const outName = process.platform === 'win32' ? 'rlm.exe' : 'rlm';
  const outPath = path.join(outDir, outName);

  fs.mkdirSync(outDir, { recursive: true });

  // Build the Go binary from the package root.
  // Requires 'go' to be available on PATH.
  try {
    run('go', ['build', '-o', outPath, './cmd/rlm'], { cwd: pkgRoot });
  } catch (e) {
    console.error('\nERROR: failed to build the rlm Go binary during npm install.');
    console.error('Make sure Go is installed and available in PATH, then reinstall.');
    console.error(`Details: ${e && e.message ? e.message : String(e)}`);
    process.exit(1);
  }
}

main();
