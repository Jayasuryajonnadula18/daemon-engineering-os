import chalk from 'chalk';
import { StorageService } from '../services/storageService';
import { renderStatusReport } from '../core/reporter';
import { HealthReport } from '../types';

export async function runWorkspace(): Promise<void> {
  console.log(chalk.cyan.bold('Showing workspace dashboard...'));

  const storage = new StorageService();
  const context = await storage.readContext();
  const healthReport = await storage.readJson<HealthReport>('health.json');
  const executionHistory = await storage.readExecutionHistory();
  const recoveryHistory = await storage.readRecoveryHistory();
  const intelligence = await storage.readIntelligenceReport();

  if (!context || !healthReport) {
    console.log(chalk.yellow('No workspace state found. Run `daemon inspect` or `daemon start` first to populate workspace history.'));
    return;
  }

  renderStatusReport(context, healthReport, intelligence);

  console.log(chalk.cyan.bold('\nRecent activity'));
  if (executionHistory.length === 0) {
    console.log(chalk.gray('No execution history available.'));
  } else {
    executionHistory.slice(-5).reverse().forEach((entry) => {
      console.log(`- [${entry.outcome}] ${entry.command} at ${entry.timestamp}`);
      if (entry.details) {
        console.log(`  ${chalk.gray(entry.details)}`);
      }
    });
  }

  console.log(chalk.cyan.bold('\nRecovery history'));
  if (recoveryHistory.length === 0) {
    console.log(chalk.gray('No recovery history available.'));
  } else {
    recoveryHistory.slice(-5).reverse().forEach((entry) => {
      console.log(`- [${entry.outcome}] ${entry.summary} at ${entry.timestamp}`);
    });
  }
}
