import { EnvSummary } from '../types';

const envFiles = ['.env', '.env.local', '.env.development', '.env.production'];

function inferRequiredVariables(): string[] {
  const variables = new Set<string>();

  if (process.env.NODE_ENV) {
    variables.add('NODE_ENV');
  }

  if (process.env.DATABASE_URL) {
    variables.add('DATABASE_URL');
  }

  if (process.env.GIT_AUTHOR_NAME) {
    variables.add('GIT_AUTHOR_NAME');
  }

  if (process.env.GITHUB_TOKEN || process.env.GITHUB_PAT) {
    variables.add('GITHUB_TOKEN');
  }

  if (variables.size === 0) {
    return ['NODE_ENV', 'DATABASE_URL', 'GIT_AUTHOR_NAME'];
  }

  return Array.from(variables);
}

async function detectEnvFiles(): Promise<string[]> {
  const found: string[] = [];
  const cwd = process.cwd();

  for (const file of envFiles) {
    try {
      await import('fs/promises').then(({ access }) => access(`${cwd}/${file}`));
      found.push(file);
    } catch {
      // ignore missing files
    }
  }

  return found;
}

export async function getEnvSummary(): Promise<EnvSummary> {
  const requiredVariables = inferRequiredVariables();
  const present = requiredVariables.filter((name) => Boolean(process.env[name]));
  const missing = requiredVariables.filter((name) => !Boolean(process.env[name]));
  const foundEnvFiles = await detectEnvFiles();

  return {
    required: requiredVariables,
    present,
    missing,
    envFiles: foundEnvFiles,
  };
}
