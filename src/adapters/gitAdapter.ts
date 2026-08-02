import { Adapter } from './adapter';
import { HealthReport, ProjectContext, RecoveryAction, RecoveryPlan } from '../types';
import { getGitSummary } from '../utils/git';
import { HealthEngine } from '../core/healthEngine';

export class GitAdapter implements Adapter {
  async detect(context: ProjectContext): Promise<boolean> {
    const gitSummary = await getGitSummary();
    return gitSummary.repository !== 'Not a git repository';
  }

  async verify(context: ProjectContext): Promise<HealthReport> {
    const healthEngine = new HealthEngine();
    return healthEngine.assess(context);
  }

  async health(context: ProjectContext): Promise<HealthReport> {
    return this.verify(context);
  }

  async recover(context: ProjectContext, actions: RecoveryAction[]): Promise<RecoveryPlan> {
    const planActions = actions.filter((action) => action.id === 'recover.git');
    return {
      summary: 'Git recovery actions',
      actions: planActions,
    };
  }
}
