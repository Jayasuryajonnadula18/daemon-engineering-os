import os from 'os';
import { SystemSummary } from '../types';

export async function getSystemSummary(): Promise<SystemSummary> {
  const platform = `${os.type()} ${os.release()}`;
  const cpuCount = os.cpus().length;
  const memoryGb = Number((os.totalmem() / 1024 / 1024 / 1024).toFixed(2));
  const freeMemoryGb = Number((os.freemem() / 1024 / 1024 / 1024).toFixed(2));

  return {
    platform,
    cpuCount,
    memoryGb,
    freeMemoryGb,
  };
}
