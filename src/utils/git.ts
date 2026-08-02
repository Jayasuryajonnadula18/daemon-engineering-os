import { execa } from 'execa';
import { GitSummary } from '../types';

async function runGitCommand(args: string[]): Promise<string> {
  const result = await execa('git', args, { stderr: 'ignore' });
  return result.stdout.trim();
}

export async function getGitSummary(): Promise<GitSummary> {
  try {
    const repository = await runGitCommand(['rev-parse', '--show-toplevel']);
    const branch = await runGitCommand(['rev-parse', '--abbrev-ref', 'HEAD']);
    const statusOutput = await runGitCommand(['status', '--short']);
    const clean = statusOutput.length === 0;
    const lastCommit = await runGitCommand(['log', '-1', '--pretty=%B']).catch(() => 'Unavailable');

    return {
      repository,
      branch,
      status: clean ? 'Clean working tree' : statusOutput.split('\n')[0] || 'Uncommitted changes',
      clean,
      lastCommit,
    };
  } catch {
    return {
      repository: 'Not a git repository',
      branch: 'N/A',
      status: 'Unavailable',
      clean: false,
      lastCommit: 'Unavailable',
    };
  }
}
