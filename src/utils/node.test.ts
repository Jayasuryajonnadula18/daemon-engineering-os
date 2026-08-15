import assert from 'node:assert/strict';
import { mkdtemp, rm, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { detectPackageManager } from './node';

test('detectPackageManager identifies pnpm from lockfiles', async () => {
  const tempRoot = await mkdtemp(path.join(os.tmpdir(), 'daemon-node-'));

  try {
    await writeFile(path.join(tempRoot, 'pnpm-lock.yaml'), 'lockfileVersion: 9.0\n', 'utf8');
    assert.equal(await detectPackageManager(tempRoot), 'pnpm');
  } finally {
    await rm(tempRoot, { recursive: true, force: true });
  }
});
