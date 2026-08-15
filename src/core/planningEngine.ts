import { ExecutionPlan, ExecutionPlanStep, ProjectContext, StartupOptions } from '../types';
import { loadDaemonConfig } from '../utils/config';

export class PlanningEngine {
  async buildExecutionPlan(context: ProjectContext, options: StartupOptions = {}): Promise<ExecutionPlan> {
    const config = await loadDaemonConfig(context.root);
    const steps: ExecutionPlanStep[] = [];

    steps.push({
      title: 'Verify runtime',
      description: `Ensure the detected runtime is available for ${context.framework || 'the current project'}`,
      action: 'verify',
      command: 'node --version',
      confidence: 95,
    });

    if (context.packageManager) {
      steps.push({
        title: 'Verify package manager',
        description: `Confirm ${context.packageManager} is installed and available`,
        action: 'verify',
        command: `${context.packageManager} --version`,
        confidence: 95,
      });
    }

    if (context.packageJson?.scripts && typeof context.packageJson.scripts === 'object' && Object.keys(context.packageJson.scripts).length > 0) {
      if (options.skipInstall) {
        steps.push({
          title: 'Reuse previous installation state',
          description: 'Daemon detected a matching prior startup and is skipping dependency installation.',
          action: 'verify',
          confidence: 85,
        });
      } else {
        steps.push({
          title: 'Install dependencies',
          description: 'Ensure project dependencies are installed before starting the app',
          action: 'install',
          command: `${context.packageManager || 'install'} install`,
          confidence: 90,
        });
      }
    } else {
      steps.push({
        title: 'Review dependency installation',
        description: 'Dependencies may need to be installed manually if they are missing',
        action: 'recommend',
        confidence: 80,
      });
    }

    if (context.docker) {
      steps.push({
        title: 'Verify Docker',
        description: 'Check Docker daemon availability for container-backed services',
        action: 'verify',
        command: 'docker version --format "{{.Server.Version}}"',
        confidence: 90,
      });
    }

    if (context.database) {
      steps.push({
        title: 'Check database requirements',
        description: `Validate ${context.database} configuration and connection hints`,
        action: 'check',
        confidence: 85,
      });
    }

    const startCommand = this.detectStartCommand(context);
    if (startCommand) {
      steps.push({
        title: 'Start application runtime',
        description: `Run ${startCommand} to launch the application`,
        action: 'start',
        command: startCommand,
        expectedPorts: options.activePorts,
        confidence: 90,
      });
    } else {
      steps.push({
        title: 'Review startup scripts',
        description: 'No safe application startup script was detected; manual startup may be required',
        action: 'recommend',
        confidence: 70,
      });
    }

    if (options.activePorts && options.activePorts.length > 0) {
      steps.push({
        title: 'Monitor expected ports',
        description: `Verify service ports are available: ${options.activePorts.join(', ')}`,
        action: 'check',
        expectedPorts: options.activePorts,
        confidence: 85,
      });
    }

    steps.push({
      title: 'Run health checks',
      description: 'Verify the project is healthy after environment preparation',
      action: 'check',
      confidence: 95,
    });

    return {
      name: 'Daemon startup plan',
      summary: 'A deterministic startup plan for the current project',
      steps,
    };
  }

  private isSelfBootstrappingScript(script: string): boolean {
    return /(?:src\/index\.ts|dist\/index\.js|daemon(?:\s|-)cli|daemon\s*start|run\s*dev)/i.test(script);
  }

  private detectStartCommand(context: ProjectContext): string | undefined {
    const scripts = context.packageJson?.scripts as Record<string, unknown> | undefined;
    if (!scripts) {
      return undefined;
    }

    if (typeof scripts.start === 'string' && !this.isSelfBootstrappingScript(scripts.start)) {
      return `${context.packageManager || 'npm'} start`;
    }

    if (typeof scripts.dev === 'string' && !this.isSelfBootstrappingScript(scripts.dev)) {
      return `${context.packageManager ? `${context.packageManager} run dev` : 'npm run dev'}`;
    }

    return undefined;

    return undefined;
  }
}
