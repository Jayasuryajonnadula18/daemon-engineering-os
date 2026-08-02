import boxen, { Options as BoxenOptions } from 'boxen';
import chalk from 'chalk';
import Table from 'cli-table3';
import { DoctorReport, HealthReport, ProjectContext } from '../types';

const boxOptions: BoxenOptions = {
  borderColor: 'cyan',
  borderStyle: 'round',
  padding: 1,
  margin: 1,
};

function renderSection(title: string, rows: Array<{ label: string; value: string; status?: string }>): void {
  const table = new Table({
    colWidths: [28, 52],
    style: { head: [], border: [] },
    wordWrap: true,
  });

  rows.forEach((row) => {
    const status = row.status ? ` ${chalk.gray('|')} ${row.status}` : '';
    table.push([chalk.bold(row.label), `${row.value}${status}`]);
  });

  console.log(chalk.cyan.bold(`
${title}
`));
  console.log(table.toString());
}

function renderListSection(title: string, items: string[]): void {
  console.log(chalk.cyan.bold(`
${title}
`));
  if (items.length === 0) {
    console.log(chalk.gray('None')); 
    return;
  }
  items.forEach((item) => console.log(`- ${item}`));
}

export function renderDoctorReport(report: DoctorReport): void {
  console.log(boxen(chalk.bold.cyan('Daemon Doctor Report'), boxOptions));

  renderSection('Git', [
    { label: 'Repository', value: report.git.repository || 'Unknown' },
    { label: 'Branch', value: report.git.branch || 'Unknown' },
    { label: 'Status', value: report.git.status, status: report.git.clean ? chalk.green('clean') : chalk.yellow('dirty') },
    { label: 'Last commit', value: report.git.lastCommit || 'Unavailable' },
  ]);

  renderSection('Docker', [
    { label: 'Installed', value: report.docker.installed ? chalk.green('Yes') : chalk.red('No') },
    { label: 'Daemon', value: report.docker.available ? chalk.green('Running') : chalk.red('Unavailable') },
    { label: 'Version', value: report.docker.version || 'N/A' },
  ]);

  renderSection('Node & npm', [
    { label: 'Node.js version', value: report.node.version },
    { label: 'npm version', value: report.node.npmVersion },
    { label: 'Package manager', value: report.node.packageManager || 'npm' },
  ]);

  renderSection('Environment', [
    { label: 'Required variables', value: report.env.required.join(', ') || 'None' },
    { label: 'Missing variables', value: report.env.missing.length > 0 ? report.env.missing.join(', ') : 'None', status: report.env.missing.length > 0 ? chalk.yellow('warning') : chalk.green('ok') },
  ]);

  renderSection('System', [
    { label: 'Platform', value: report.system.platform },
    { label: 'CPU cores', value: `${report.system.cpuCount}` },
    { label: 'Memory', value: `${report.system.memoryGb} GB` },
    { label: 'Free memory', value: `${report.system.freeMemoryGb} GB` },
  ]);

  const score = report.overallHealthScore || 0;
  const status = score >= 90 ? chalk.green('Healthy') : score >= 70 ? chalk.yellow('Moderate') : chalk.red('At risk');

  console.log(boxen(`Engineering Health: ${chalk.bold(score.toString())}%  ${status}`, {
    borderColor: score >= 90 ? 'green' : score >= 70 ? 'yellow' : 'red',
    borderStyle: 'round',
    padding: 1,
    margin: 1,
  }));
}

export function renderHealthReport(report: HealthReport): void {
  console.log(boxen(chalk.bold.cyan('Daemon Health Report'), boxOptions));

  renderSection('Summary', [
    { label: 'Health score', value: `${report.score}%`, status: report.score >= 80 ? chalk.green('good') : report.score >= 60 ? chalk.yellow('fair') : chalk.red('poor') },
    { label: 'Timestamp', value: report.timestamp },
  ]);

  renderListSection('Issues', report.issues.map((issue) => `${issue.title} (${issue.severity}) - ${issue.detail}`));
  renderListSection('Warnings', report.warnings);
  renderListSection('Recommendations', report.recommendations);
}

export function renderStatusReport(context: ProjectContext, report: HealthReport, intelligence?: unknown): void {
  console.log(boxen(chalk.bold.cyan('Daemon Status Report'), boxOptions));

  renderSection('Project', [
    { label: 'Name', value: context.name },
    { label: 'Root', value: context.root },
    { label: 'Framework', value: context.framework || 'Unknown' },
    { label: 'Package manager', value: context.packageManager || 'Unknown' },
    { label: 'Docker support', value: context.docker ? chalk.green('Detected') : chalk.red('Missing') },
  ]);

  renderSection('Services', [
    { label: 'Detected services', value: context.services.join(', ') || 'None' },
    { label: 'Detected env files', value: context.envFiles.join(', ') || 'None' },
  ]);

  renderHealthReport(report);
}
