import { execa } from 'execa';
import { DockerSummary } from '../types';

export async function getDockerSummary(): Promise<DockerSummary> {
  try {
    const versionResult = await execa('docker', ['version', '--format', '{{.Server.Version}}']);
    return {
      installed: true,
      available: true,
      version: versionResult.stdout.trim(),
    };
  } catch {
    try {
      await execa('docker', ['--version']);
      return { installed: true, available: false };
    } catch {
      return { installed: false, available: false };
    }
  }
}
