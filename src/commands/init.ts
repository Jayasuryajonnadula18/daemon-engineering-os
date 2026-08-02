import chalk from 'chalk';
import path from 'path';
import { StorageService } from '../services/storageService';

export async function runInit(): Promise<void> {
  const storage = new StorageService();
  await storage.ensureDaemonDirectory();

  const metadata = {
    initializedAt: new Date().toISOString(),
    project: {
      root: process.cwd(),
      name: path.basename(process.cwd()),
    },
    daemonVersion: '0.1.0',
  };

  await storage.writeJson('daemon.json', metadata);

  console.log(chalk.cyan.bold('Initializing Daemon in the current project...'));
  console.log(chalk.green('✓') + ' .daemon directory created');
  console.log(chalk.green('✓') + ' Daemon metadata saved to .daemon/daemon.json');
  console.log(chalk.green('✓') + ' Run `daemon doctor` to inspect your repository and complete setup');
}
