import chalk from 'chalk';
import { Daemon } from '../core/daemon';
import { renderHealthReport } from '../core/reporter';

export async function runVerify(): Promise<void> {
  console.log(chalk.cyan.bold('Running targeted verification tasks...'));

  const daemon = new Daemon();
  const { healthReport } = await daemon.inspect();

  if (healthReport.score >= 80) {
    console.log(chalk.green.bold('Verification result: PASS'));
  } else if (healthReport.score >= 60) {
    console.log(chalk.yellow.bold('Verification result: PARTIAL'));
  } else {
    console.log(chalk.red.bold('Verification result: FAIL'));
  }

  renderHealthReport(healthReport);
}
