import { EnvSummary } from '../types';

const requiredVariables = ['NODE_ENV', 'DATABASE_URL', 'GIT_AUTHOR_NAME'];
const envFiles = ['.env', '.env.local', '.env.development', '.env.production'];

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
