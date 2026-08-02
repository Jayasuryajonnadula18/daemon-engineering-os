import chalk from 'chalk';
import { Daemon } from '../core/daemon';
import ora from 'ora';
import { execa } from 'execa';
import { writeFile } from 'fs/promises';

export async function runRecover(): Promise<void> {
  console.log(chalk.cyan.bold('Recovering project issues...'));

  const daemon = new Daemon();
  const { recoveryPlan } = await daemon.recover();
  let executedCount = 0;

  for (const action of recoveryPlan.actions) {
    console.log(chalk.cyan(`› ${action.title}`));
    console.log(chalk.gray(`  ${action.description}`));
    if (!action.safe) {
      console.log(chalk.yellow('  Skipped unsafe recovery action.'));
      continue;
    }

    const spinner = ora({ text: `Executing ${action.title}`, color: 'cyan' }).start();
    try {
      if (action.id === 'recover.env' && Array.isArray(action.metadata?.missing)) {
        const missing = action.metadata.missing as string[];
        const content = missing.map((name) => `${name}=YOUR_${name}`).join('\n') + '\n';
        await writeFile('.env.example', content, 'utf8');
        spinner.succeed(`Created .env.example`);
        executedCount += 1;
        continue;
      }

      if (!action.command) {
        spinner.info(`No automatic command available; review this action manually.`);
        continue;
      }

      await execa(action.command, { shell: true, stdio: 'inherit' });
      spinner.succeed(`Executed ${action.title}`);
      executedCount += 1;
    } catch (error) {
      spinner.fail(`Failed ${action.title}`);
      console.error(chalk.red(`  ${error instanceof Error ? error.message : String(error)}`));
    }
  }

  if (recoveryPlan.actions.length === 0) {
    console.log(chalk.green('No recovery actions needed.'));
  } else {
    console.log(chalk.green(`Recovery plan generated with ${recoveryPlan.actions.length} actions.`));
    console.log(chalk.green(`${executedCount} safe action(s) executed.`));
  }
}
