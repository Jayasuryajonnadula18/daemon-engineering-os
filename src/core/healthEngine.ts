import { getGitSummary } from '../utils/git';
import { getDockerSummary } from '../utils/docker';
import { getNodeSummary } from '../utils/node';
import { getEnvSummary } from '../utils/env';
import { getSystemSummary } from '../utils/system';
import {
  HealthReport,
  ProjectContext,
  HealthIssue,
  GitSummary,
  DockerSummary,
  NodeSummary,
  EnvSummary,
  SystemSummary,
  HealthRule,
  HealthRuleContext,
} from '../types';

const healthRules: HealthRule[] = [
  {
    id: 'git.missing',
    title: 'Git repository not detected',
    severity: 'warning',
    condition: (ctx) => !ctx.context.git,
    detail: () => 'The current directory is not a Git repository. Git tracking improves history and workflow awareness.',
    recommendation: () => 'Initialize a Git repository or open a project under an existing repository.',
  },
  {
    id: 'git.dirty',
    title: 'Working tree is dirty',
    severity: 'warning',
    condition: (ctx) => !ctx.gitSummary.clean,
    detail: () => 'Uncommitted changes are present. This may impact recovery and deployment workflows.',
    recommendation: () => 'Commit or stash changes before running recovery or deployment workflows.',
  },
  {
    id: 'docker.unavailable',
    title: 'Docker not found',
    severity: 'warning',
    condition: (ctx) => ctx.context.docker && !ctx.dockerSummary.installed,
    detail: () => 'Docker is required by the project but is not installed or available in PATH.',
    recommendation: () => 'Install Docker Desktop or ensure Docker is available in your environment.',
  },
  {
    id: 'node.missing',
    title: 'Node.js not available',
    severity: 'critical',
    condition: (ctx) =>
      (ctx.context.languages.includes('JavaScript') || ctx.context.languages.includes('TypeScript')) && !ctx.nodeSummary.version,
    detail: () => 'Node.js is required for the detected JavaScript or TypeScript project.',
    recommendation: () => 'Install Node.js and make it available on PATH.',
  },
  {
    id: 'env.missing',
    title: 'Missing environment variables',
    severity: 'warning',
    condition: (ctx) => ctx.envSummary.missing.length > 0,
    detail: (ctx) => `The following environment variables are missing: ${ctx.envSummary.missing.join(', ')}`,
    recommendation: (ctx) => `Populate the missing environment variables in ${ctx.envSummary.envFiles.join(', ') || '.env'}.`, 
  },
  {
    id: 'package-manager.unknown',
    title: 'Package manager could not be detected',
    severity: 'warning',
    condition: (ctx) => !ctx.context.packageManager,
    detail: () => 'A package manager could not be inferred from lock files or package.json.',
    recommendation: () => 'Add a supported lock file or packageManager field to package.json.',
  },
  {
    id: 'low.memory',
    title: 'Low available memory',
    severity: 'warning',
    condition: (ctx) => ctx.systemSummary.freeMemoryGb < 1,
    detail: (ctx) => `Less than 1GB of free memory is available (${ctx.systemSummary.freeMemoryGb} GB).`, 
    recommendation: () => 'Close unused apps or increase available memory before running heavy workflows.',
  },
];

export class HealthEngine {
  async assess(context: ProjectContext): Promise<HealthReport> {
    const gitSummary = await getGitSummary();
    const dockerSummary = await getDockerSummary();
    const nodeSummary = await getNodeSummary();
    const envSummary = await getEnvSummary();
    const systemSummary = await getSystemSummary();

    const healthContext: HealthRuleContext = {
      context,
      gitSummary,
      dockerSummary,
      nodeSummary,
      envSummary,
      systemSummary,
    };

    const findings = healthRules.filter((rule) => rule.condition(healthContext));
    const issues: HealthIssue[] = findings.map((rule) => ({
      id: rule.id,
      title: rule.title,
      severity: rule.severity,
      detail: rule.detail(healthContext),
    }));

    const warnings = findings
      .filter((rule) => rule.severity === 'warning')
      .map((rule) => rule.recommendation?.(healthContext) || rule.title);

    const recommendations = findings
      .map((rule) => rule.recommendation?.(healthContext))
      .filter((recommendation): recommendation is string => Boolean(recommendation));

    const score = this.calculateScore(issues, warnings);

    if (score < 70 && !recommendations.some((text) => text.includes('daemon doctor'))) {
      recommendations.push('Run `daemon doctor` and `daemon recover` to improve your project health score.');
    }

    return {
      score,
      issues,
      warnings,
      recommendations,
      timestamp: new Date().toISOString(),
    };
  }

  private calculateScore(issues: HealthIssue[], warnings: string[]): number {
    let score = 100;
    for (const issue of issues) {
      score -= issue.severity === 'critical' ? 25 : 12;
    }

    // Keep the score tied to actual issues, not duplicate warning sentiment.
    return Math.max(0, Math.min(100, score));
  }
}
