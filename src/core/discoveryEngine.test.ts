import assert from 'node:assert/strict';
import { mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { resolveProjectRoot } from './discoveryEngine';

test('resolveProjectRoot prefers the current project directory over the git repository root', async () => {
  const tempRoot = await mkdtemp(path.join(os.tmpdir(), 'daemon-discovery-'));
  const projectRoot = path.join(tempRoot, 'python-project');
  const nestedDir = path.join(projectRoot, 'services', 'api');
  const gitRoot = path.join(tempRoot, 'repo-root');

  try {
    await mkdir(nestedDir, { recursive: true });
    await mkdir(gitRoot, { recursive: true });
    await writeFile(path.join(projectRoot, 'requirements.txt'), 'requests==2.31.0\n', 'utf8');

    const resolvedRoot = await resolveProjectRoot(nestedDir, gitRoot);
    assert.equal(resolvedRoot, projectRoot);
  } finally {
    await rm(tempRoot, { recursive: true, force: true });
  }
});
