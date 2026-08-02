import { HealthReport, RecoveryAction, RecoveryPlan, ProjectContext, RecoveryRule, HealthRuleContext } from '../types';
import { getGitSummary } from '../utils/git';
import { getDockerSummary } from '../utils/docker';
import { getNodeSummary } from '../utils/node';
import { getEnvSummary } from '../utils/env';
import { getSystemSummary } from '../utils/system';

function createEnvExampleCommand(missing: string[]): string {
  const content = missing.map((name) => `${name}=YOUR_${name}`).join('\n');
  const escaped = JSON.stringify(content);
  return `node -e "require('fs').writeFileSync('.env.example', ${escaped})"`;
}

const recoveryRules: RecoveryRule[] = [
  {
    id: 'recover.env',
    title: 'Create example environment file',
    safe: true,
    confidence: 90,
    condition: (ctx) => ctx.envSummary.missing.length > 0,
    description: (ctx) => `Create a sample .env.example file for missing variables: ${ctx.envSummary.missing.join(', ')}.`,
    metadata: (ctx) => ({ missing: ctx.envSummary.missing }),
  },
  {
    id: 'recover.docker',
    title: 'Install or enable Docker',
    safe: true,
    confidence: 85,
    condition: (ctx) => ctx.context.docker && !ctx.dockerSummary.installed,
    description: () => 'Install Docker Desktop or enable the Docker daemon to provide container-based databases and services.',
  },
  {
    id: 'recover.dependencies',
    title: 'Install project dependencies',
    safe: true,
    confidence: 95,
    condition: (ctx) => Boolean(ctx.context.packageManager && ctx.context.packageJson?.dependencies && Object.keys(ctx.context.packageJson.dependencies).length > 0),
    description: (ctx) => `Run ${ctx.context.packageManager || 'npm'} install to ensure dependencies are available for startup and verification tasks.`,
    command: (ctx) => `${ctx.context.packageManager || 'npm'} install`,
  },
  {
    id: 'recover.git',
    title: 'Initialize Git repository',
    safe: true,
    confidence: 80,
    condition: (ctx) => !ctx.context.git,
    description: () => 'Initialize a Git repository to enable Daemon project discovery and workflow tracking.',
    command: () => 'git init',
  },
];

export class RecoveryEngine {
  async buildRecoveryPlan(context: ProjectContext, healthReport: HealthReport): Promise<RecoveryPlan> {
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
      healthReport,
    };

    const actions: RecoveryAction[] = recoveryRules
      .filter((rule) => rule.condition(healthContext))
      .map((rule) => ({
        id: rule.id,
        title: rule.title,
        description: rule.description(healthContext),
        confidence: rule.confidence,
        safe: rule.safe,
        command: rule.command ? rule.command(healthContext) : undefined,
        metadata: rule.metadata ? rule.metadata(healthContext) : undefined,
        executed: false,
      }));

    return {
      summary: 'Recovery plan generated from the latest health assessment.',
      actions,
    };
  }
}
