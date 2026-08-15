import { execa } from 'execa';
import { ExecutionPlan, ProjectContext, StartupOptions, StartupResult } from '../types';
import { PlanningEngine } from '../core/planningEngine';
import { StorageService } from './storageService';
import { waitForPort } from '../utils/ports';
import { loadDaemonConfig } from '../utils/config';
import chalk from 'chalk';
import ora from 'ora';

export class StartupService {
  private readonly planningEngine = new PlanningEngine();
  private readonly storage = new StorageService();

  async start(context: ProjectContext, options: StartupOptions = {}): Promise<StartupResult> {
    const spinner = ora({ text: 'Building project startup plan...', color: 'cyan' }).start();
    const plan = await this.planningEngine.buildExecutionPlan(context, options);
    spinner.succeed('Startup plan ready');

    await this.storage.writeExecutionPlan(plan);

    if (options.skipInstall) {
      console.log(chalk.yellow('Previous startup memory indicates dependencies are already installed. Skipping install step.'));
    }

    let startupSuccess = true;
    let failedStep: string | undefined;

    for (const step of plan.steps) {
      const stepLabel = chalk.bold(step.title);
      if (!step.command) {
        console.log(`${chalk.cyan('›')} ${stepLabel} — ${step.description}`);
        continue;
      }

      const actionSpinner = ora({ text: `${step.title}` }).start();
      try {
        if (step.action === 'start' && step.expectedPorts && step.expectedPorts.length > 0) {
          const subprocess = execa(step.command, { shell: true, stdio: 'inherit' });
          const config = await loadDaemonConfig(context.root);
          const portsReady = await Promise.all(step.expectedPorts.map((port) => waitForPort(port, config.startupPortTimeoutMs)));

          if (!portsReady.every(Boolean)) {
            subprocess.kill('SIGTERM');
            throw new Error(`Expected ports did not become available: ${step.expectedPorts.join(', ')}`);
          }

          await subprocess;
        } else {
          await execa(step.command, { shell: true, stdio: 'inherit' });
        }

        actionSpinner.succeed(`${step.title}`);
      } catch (error) {
        startupSuccess = false;
        failedStep = step.title;
        actionSpinner.fail(`${step.title} failed`);
        console.error(chalk.red(`Error executing: ${step.command}`));
        console.error(error instanceof Error ? error.message : error);
        break;
      }
    }

    return {
      plan,
      success: startupSuccess,
      failedStep,
    };
  }
}
