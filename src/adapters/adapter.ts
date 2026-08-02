import { HealthReport, ProjectContext, RecoveryAction, RecoveryPlan } from '../types';

export interface Adapter {
  detect(context: ProjectContext): Promise<boolean>;
  verify(context: ProjectContext): Promise<HealthReport>;
  health(context: ProjectContext): Promise<HealthReport>;
  recover(context: ProjectContext, actions: RecoveryAction[]): Promise<RecoveryPlan>;
}
