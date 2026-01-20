#!/usr/bin/env node

const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');

function main() {
  const binName = process.platform === 'win32' ? 'rlm.exe' : 'rlm';
  const binPath = path.resolve(__dirname, '..', '.rlm-bin', binName);

  if (!fs.existsSync(binPath)) {
    console.error(`rlm: native binary not found at ${binPath}`);
    console.error('This npm package builds the Go binary during install.');
    console.error('Try reinstalling with Go installed and available in PATH.');
    process.exit(2);
  }

  const args = process.argv.slice(2);
  const r = spawnSync(binPath, args, { stdio: 'inherit' });

  if (r.error) {
    console.error(`rlm: failed to execute native binary: ${r.error.message}`);
    process.exit(2);
  }

  process.exit(typeof r.status === 'number' ? r.status : 0);
}

main();
