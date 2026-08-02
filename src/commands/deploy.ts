import chalk from 'chalk';
import { Daemon } from '../core/daemon';
import { renderDoctorReport } from '../core/reporter';
import { getDockerSummary } from '../utils/docker';
import { getEnvSummary } from '../utils/env';
import { getGitSummary } from '../utils/git';
import { getNodeSummary } from '../utils/node';
import { getSystemSummary } from '../utils/system';

export async function runDeploy(): Promise<void> {
  console.log(chalk.cyan.bold('Coordinating deployment pipeline...'));

  const daemon = new Daemon();
  const { healthReport } = await daemon.inspect();
  const gitSummary = await getGitSummary();
  const dockerSummary = await getDockerSummary();
  const nodeSummary = await getNodeSummary();
  const envSummary = await getEnvSummary();
  const systemSummary = await getSystemSummary();

  if (healthReport.score < 70) {
    console.log(chalk.yellow('Deployment readiness is limited.'));
    console.log(chalk.gray('Review the following report and address the recommended issues before deploying.'));
  } else {
    console.log(chalk.green('Deployment readiness checks passed.'));
  }

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

  console.log(chalk.yellow('Deployment orchestration is in progress as a core workflow.'));
  console.log(chalk.cyan('Next step: integrate Kubernetes and cloud deployment orchestration flows.'));
}
