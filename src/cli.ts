import { Command } from 'commander';
import { runDoctor } from './commands/doctor';
import { runInit } from './commands/init';
import { runInspect } from './commands/inspect';
import { runHealth } from './commands/health';
import { runPredict } from './commands/predict';
import { runVerify } from './commands/verify';
import { runDeploy } from './commands/deploy';
import { runWorkspace } from './commands/status';
import { runGraph } from './commands/graph';
import { runStart } from './commands/start';
import { runRecover } from './commands/recover';
import { runAnalyze } from './commands/analyze';

const program = new Command();

export async function runCLI(argv: string[]) {
  program
    .name('daemon')
    .description('Daemon CLI — engineering intelligence for modern development systems.')
    .version('0.1.0');

  program
    .command('init')
    .description('Initialize Daemon in the current project')
    .action(async () => {
      await runInit();
    });

  program
    .command('doctor')
    .description('Run a health check across the detected project runtime, dependencies, and system status')
    .action(async () => {
      await runDoctor();
    });

  program
    .command('health')
    .description('Display engineering health metrics')
    .action(async () => {
      await runHealth();
    });

  program
    .command('inspect')
    .description('Inspect engineering environment and project context')
    .action(async () => {
      await runInspect();
    });

  program
    .command('start')
    .description('Prepare the local project environment and start the development workflow')
    .action(async () => {
      await runStart();
    });

  program
    .command('recover')
    .description('Generate and execute safe recovery actions for the current project')
    .action(async () => {
      await runRecover();
    });

  program
    .command('predict')
    .description('Analyze recent code changes and predict risk')
    .action(async () => {
      await runPredict();
    });

  program
    .command('verify')
    .description('Run only the required verification tasks for the current change set')
    .action(async () => {
      await runVerify();
    });

  program
    .command('deploy')
    .description('Coordinate Docker, GitHub Actions, Kubernetes, and cloud deployment')
    .action(async () => {
      await runDeploy();
    });

  program
    .command('workspace')
    .description('Show workspace dashboard and recent engineering activity')
    .action(async () => {
      await runWorkspace();
    });

  program
    .command('status')
    .description('Alias for workspace dashboard')
    .action(async () => {
      await runWorkspace();
    });

  program
    .command('analyze')
    .description('Run engineering intelligence analysis for impact, risk, and architecture')
    .action(async () => {
      await runAnalyze();
    });

  program
    .command('graph')
    .description('Render an engineering dependency graph')
    .action(async () => {
      await runGraph();
    });

  await program.parseAsync(argv);
}
