import chalk from 'chalk';
import { Daemon } from '../core/daemon';
import { renderHealthReport } from '../core/reporter';

export async function runHealth(): Promise<void> {
  console.log(chalk.cyan.bold('Collecting engineering health data...'));

  const daemon = new Daemon();
  const { healthReport } = await daemon.inspect();

  renderHealthReport(healthReport);
}
