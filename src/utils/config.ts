import { readFile } from 'node:fs/promises';
import path from 'node:path';

export interface DaemonConfig {
  healthScoreWarningThreshold: number;
  healthScoreCriticalThreshold: number;
  startupPortTimeoutMs: number;
}

const DEFAULT_CONFIG: DaemonConfig = {
  healthScoreWarningThreshold: 70,
  healthScoreCriticalThreshold: 40,
  startupPortTimeoutMs: 15000,
};

function getEnvNumber(name: string, fallback: number): number {
  const rawValue = process.env[name];
  if (!rawValue) {
    return fallback;
  }

  const parsed = Number(rawValue);
  return Number.isFinite(parsed) ? parsed : fallback;
}

export async function loadDaemonConfig(projectRoot: string = process.cwd()): Promise<DaemonConfig> {
  const configPath = path.join(projectRoot, '.daemon', 'config.json');

  try {
    const raw = await readFile(configPath, 'utf8');
    const parsed = JSON.parse(raw) as Partial<DaemonConfig>;

    return {
      healthScoreWarningThreshold: parsed.healthScoreWarningThreshold ?? getEnvNumber('DAEMON_HEALTH_SCORE_WARNING_THRESHOLD', DEFAULT_CONFIG.healthScoreWarningThreshold),
      healthScoreCriticalThreshold: parsed.healthScoreCriticalThreshold ?? getEnvNumber('DAEMON_HEALTH_SCORE_CRITICAL_THRESHOLD', DEFAULT_CONFIG.healthScoreCriticalThreshold),
      startupPortTimeoutMs: parsed.startupPortTimeoutMs ?? getEnvNumber('DAEMON_STARTUP_PORT_TIMEOUT_MS', DEFAULT_CONFIG.startupPortTimeoutMs),
    };
  } catch {
    return {
      healthScoreWarningThreshold: getEnvNumber('DAEMON_HEALTH_SCORE_WARNING_THRESHOLD', DEFAULT_CONFIG.healthScoreWarningThreshold),
      healthScoreCriticalThreshold: getEnvNumber('DAEMON_HEALTH_SCORE_CRITICAL_THRESHOLD', DEFAULT_CONFIG.healthScoreCriticalThreshold),
      startupPortTimeoutMs: getEnvNumber('DAEMON_STARTUP_PORT_TIMEOUT_MS', DEFAULT_CONFIG.startupPortTimeoutMs),
    };
  }
}
