import chalk from 'chalk';
import { Daemon } from '../core/daemon';
import { renderStatusReport } from '../core/reporter';

export async function runAnalyze(): Promise<void> {
  console.log(chalk.cyan.bold('Running engineering intelligence analysis...'));

  const daemon = new Daemon();
  const { context, healthReport, intelligence } = await daemon.analyze();

  renderStatusReport(context, healthReport, intelligence);

  console.log(chalk.cyan.bold('\nArchitecture analysis'));
  intelligence.architecture.issues.forEach((issue) => console.log(`- ${issue}`));
  console.log(chalk.cyan.bold('\nRisk summary'));
  console.log(`- ${intelligence.risk.summary}`);
  console.log(chalk.cyan.bold('\nBlast radius'));
  console.log(`- ${intelligence.blastRadius.summary}`);
  console.log(chalk.cyan.bold('\nBreaking changes'));
  console.log(`- ${intelligence.breakingChanges.summary}`);
  console.log(chalk.cyan.bold('\nInsights'));
  intelligence.insights.forEach((insight) => console.log(`- ${insight.title}: ${insight.detail}`));
}
