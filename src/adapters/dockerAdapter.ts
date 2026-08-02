import { Adapter } from './adapter';
import { HealthReport, ProjectContext, RecoveryAction, RecoveryPlan } from '../types';
import { getDockerSummary } from '../utils/docker';
import { HealthEngine } from '../core/healthEngine';

export class DockerAdapter implements Adapter {
  async detect(context: ProjectContext): Promise<boolean> {
    const dockerSummary = await getDockerSummary();
    return dockerSummary.installed;
  }

  async verify(context: ProjectContext): Promise<HealthReport> {
    const healthEngine = new HealthEngine();
    return healthEngine.assess(context);
  }

  async health(context: ProjectContext): Promise<HealthReport> {
    return this.verify(context);
  }

  async recover(context: ProjectContext, actions: RecoveryAction[]): Promise<RecoveryPlan> {
    const planActions = actions.filter((action) => action.id === 'recover.docker');
    return {
      summary: 'Docker recovery actions',
      actions: planActions,
    };
  }
}
