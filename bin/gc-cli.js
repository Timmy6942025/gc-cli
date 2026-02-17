#!/usr/bin/env node

const { spawnSync } = require('node:child_process');
const { existsSync } = require('node:fs');
const { join } = require('node:path');

const targetMap = {
  darwin: {
    arm64: 'darwin-arm64/gc-cli',
    x64: 'darwin-amd64/gc-cli',
  },
  linux: {
    arm64: 'linux-arm64/gc-cli',
    x64: 'linux-amd64/gc-cli',
  },
  win32: {
    arm64: 'windows-arm64/gc-cli.exe',
    x64: 'windows-amd64/gc-cli.exe',
  },
};

const platform = process.platform;
const arch = process.arch;
const relPath = targetMap[platform] && targetMap[platform][arch];

if (!relPath) {
  console.error(`Unsupported platform/arch: ${platform}/${arch}`);
  console.error('This package currently supports darwin/linux/win32 with arm64/x64.');
  process.exit(1);
}

const binPath = join(__dirname, '..', 'dist', relPath);
if (!existsSync(binPath)) {
  console.error(`Missing bundled binary: ${binPath}`);
  console.error('Reinstall the package or file an issue.');
  process.exit(1);
}

const child = spawnSync(binPath, process.argv.slice(2), {
  stdio: 'inherit',
  env: process.env,
});

if (child.error) {
  console.error(child.error.message);
  process.exit(1);
}

if (typeof child.status === 'number') {
  process.exit(child.status);
}

process.exit(1);
