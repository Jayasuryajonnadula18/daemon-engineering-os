import { execa } from 'execa';
import { NodeSummary } from '../types';

async function runCommand(command: string, args: string[]): Promise<string> {
  const result = await execa(command, args, { stderr: 'ignore' });
  return result.stdout.trim();
}

export async function getNodeSummary(): Promise<NodeSummary> {
  const version = process.version;
  const npmVersion = await runCommand('npm', ['--version']).catch(() => 'Unavailable');
  const packageManager = 'npm';

  return {
    version,
    npmVersion,
    packageManager,
  };
}
