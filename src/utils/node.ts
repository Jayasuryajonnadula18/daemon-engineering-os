import { access, readFile } from 'node:fs/promises';
import path from 'node:path';
import { execa } from 'execa';
import { NodeSummary } from '../types';

async function runCommand(command: string, args: string[]): Promise<string> {
  const result = await execa(command, args, { stderr: 'ignore' });
  return result.stdout.trim();
}

async function exists(filePath: string): Promise<boolean> {
  try {
    await access(filePath);
    return true;
  } catch {
    return false;
  }
}

function normalizePackageManager(value: string): string | undefined {
  const normalized = value.trim().toLowerCase();
  if (normalized.startsWith('pnpm')) return 'pnpm';
  if (normalized.startsWith('yarn')) return 'yarn';
  if (normalized.startsWith('bun')) return 'bun';
  if (normalized.startsWith('npm')) return 'npm';
  return undefined;
}

export async function detectPackageManager(projectRoot: string = process.cwd()): Promise<string | undefined> {
  const lockFiles = [
    { manager: 'pnpm', file: 'pnpm-lock.yaml' },
    { manager: 'npm', file: 'package-lock.json' },
    { manager: 'yarn', file: 'yarn.lock' },
    { manager: 'bun', file: 'bun.lockb' },
  ] as const;

  for (const lockFile of lockFiles) {
    if (await exists(path.join(projectRoot, lockFile.file))) {
      return lockFile.manager;
    }
  }

  try {
    const packageJsonPath = path.join(projectRoot, 'package.json');
    const packageJson = JSON.parse(await readFile(packageJsonPath, 'utf8'));
    const packageManagerField = typeof packageJson.packageManager === 'string' ? packageJson.packageManager : '';
    return normalizePackageManager(packageManagerField);
  } catch {
    return undefined;
  }
}

export async function getNodeSummary(projectRoot: string = process.cwd()): Promise<NodeSummary> {
  const version = process.version;
  const packageManager = await detectPackageManager(projectRoot);
  const packageManagerVersion = packageManager
    ? await runCommand(packageManager, ['--version']).catch(() => 'Unavailable')
    : 'Unavailable';

  return {
    version,
    npmVersion: packageManagerVersion,
    packageManager,
    packageManagerVersion,
  };
}
