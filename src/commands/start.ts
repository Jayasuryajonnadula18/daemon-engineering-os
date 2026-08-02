import chalk from 'chalk';
import { Daemon } from '../core/daemon';
import { renderHealthReport } from '../core/reporter';

export async function runStart(): Promise<void> {
  console.log(chalk.cyan.bold('Starting Daemon startup flow...'));

  const daemon = new Daemon();
  const result = await daemon.start();

  if (result.startupSuccess) {
    console.log(chalk.green('Daemon startup completed successfully.'));
  } else {
    console.log(chalk.yellow('Daemon startup completed with issues. Review the reported health and retry.'));
    if (result.failedStep) {
      console.log(chalk.red(`Failed on step: ${result.failedStep}`));
    }
  }

  renderHealthReport(result.health);
}
