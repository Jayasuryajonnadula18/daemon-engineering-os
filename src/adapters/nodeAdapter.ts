import { Adapter } from './adapter';
import { HealthReport, ProjectContext, RecoveryAction, RecoveryPlan } from '../types';
import { getNodeSummary } from '../utils/node';
import { HealthEngine } from '../core/healthEngine';

export class NodeAdapter implements Adapter {
  async detect(context: ProjectContext): Promise<boolean> {
    const nodeSummary = await getNodeSummary();
    return Boolean(nodeSummary.version);
  }

  async verify(context: ProjectContext): Promise<HealthReport> {
    const healthEngine = new HealthEngine();
    return healthEngine.assess(context);
  }

  async health(context: ProjectContext): Promise<HealthReport> {
    return this.verify(context);
  }

  async recover(context: ProjectContext, actions: RecoveryAction[]): Promise<RecoveryPlan> {
    return {
      summary: 'Node recovery plan',
      actions: actions.filter((action) => action.id === 'recover.dependencies'),
    };
  }
}
