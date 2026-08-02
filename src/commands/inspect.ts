import chalk from 'chalk';
import { Daemon } from '../core/daemon';
import { renderStatusReport } from '../core/reporter';

export async function runInspect(): Promise<void> {
  console.log(chalk.cyan.bold('Inspecting engineering environment...'));

  const daemon = new Daemon();
  const { context, healthReport } = await daemon.inspect();

  renderStatusReport(context, healthReport);
}
