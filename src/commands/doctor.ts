import ora from 'ora';
import chalk from 'chalk';
import { Daemon } from '../core/daemon';
import { renderDoctorReport } from '../core/reporter';
import { getDockerSummary } from '../utils/docker';
import { getEnvSummary } from '../utils/env';
import { getGitSummary } from '../utils/git';
import { getNodeSummary } from '../utils/node';
import { getSystemSummary } from '../utils/system';

export async function runDoctor(): Promise<void> {
  const spinner = ora({ text: 'Discovering project context...', color: 'cyan' }).start();

  try {
    const daemon = new Daemon();
    const { context, healthReport } = await daemon.diagnose();
    const gitSummary = await getGitSummary();
    const dockerSummary = await getDockerSummary();
    const nodeSummary = await getNodeSummary();
    const envSummary = await getEnvSummary();
    const systemSummary = await getSystemSummary();

    spinner.succeed('Discovery complete');

    renderDoctorReport({
      git: {
        repository: gitSummary.repository,
        branch: gitSummary.branch,
        status: gitSummary.status,
        clean: gitSummary.clean,
        lastCommit: gitSummary.lastCommit,
      },
      docker: {
        installed: dockerSummary.installed,
        available: dockerSummary.available,
        version: dockerSummary.version,
      },
      node: {
        version: nodeSummary.version,
        npmVersion: nodeSummary.npmVersion,
        packageManager: nodeSummary.packageManager,
      },
      env: envSummary,
      system: systemSummary,
      overallHealthScore: healthReport.score,
    });
  } catch (error) {
    spinner.fail('Doctor scan failed');
    console.error(chalk.red(error instanceof Error ? error.message : String(error)));
    process.exit(1);
  }
}
