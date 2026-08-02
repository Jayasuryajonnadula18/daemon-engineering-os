import { DiscoveryEngine } from './discoveryEngine';
import { HealthEngine } from './healthEngine';
import { RecoveryEngine } from './recoveryEngine';
import { IntelligenceEngine } from './intelligenceEngine';
import { StartupService } from '../services/startupService';
import { StorageService } from '../services/storageService';
import {
  ExecutionHistoryEntry,
  EngineeringIntelligenceReport,
  HealthReport,
  ProjectContext,
  RecoveryHistoryEntry,
  RecoveryPlan,
  DaemonStartResult,
  InspectionResult,
  StartupResult,
} from '../types';

export class Daemon {
  private readonly discoveryEngine = new DiscoveryEngine();
  private readonly healthEngine = new HealthEngine();
  private readonly recoveryEngine = new RecoveryEngine();
  private readonly intelligenceEngine = new IntelligenceEngine();
  private readonly startupService = new StartupService();
  private readonly storageService = new StorageService();

  async inspect(): Promise<InspectionResult> {
    const context = await this.discoveryEngine.discover();
    const healthReport = await this.healthEngine.assess(context);
    const intelligenceReport = await this.intelligenceEngine.analyze(context, healthReport);

    await this.storageService.writeContext(context);
    await this.storageService.writeHealth(healthReport);
    await this.storageService.writeIntelligenceReport(intelligenceReport);
    await this.storageService.appendHealthHistory({
      id: `health-${Date.now()}`,
      timestamp: new Date().toISOString(),
      score: healthReport.score,
      issueCount: healthReport.issues.length,
      warningCount: healthReport.warnings.length,
      recommendationCount: healthReport.recommendations.length,
    });

    return { context, healthReport, intelligence: intelligenceReport };
  }

  async diagnose(): Promise<{ context: ProjectContext; healthReport: HealthReport; recoveryPlan: RecoveryPlan }> {
    const context = await this.discoveryEngine.discover();
    const healthReport = await this.healthEngine.assess(context);
    const recoveryPlan = await this.recoveryEngine.buildRecoveryPlan(context, healthReport);

    await this.storageService.writeRecoveryPlan(recoveryPlan);
    await this.storageService.appendRecoveryHistory({
      id: `recover-${Date.now()}`,
      timestamp: new Date().toISOString(),
      actions: recoveryPlan.actions,
      outcome: recoveryPlan.actions.length === 0 ? 'success' : 'warning',
      summary: 'Recovery plan generated during diagnosis.',
    });

    return { context, healthReport, recoveryPlan };
  }

  async analyze(): Promise<{ context: ProjectContext; healthReport: HealthReport; intelligence: EngineeringIntelligenceReport }> {
    const context = await this.discoveryEngine.discover();
    const healthReport = await this.healthEngine.assess(context);
    const intelligenceReport = await this.intelligenceEngine.analyze(context, healthReport);

    await this.storageService.writeIntelligenceReport(intelligenceReport);

    return { context, healthReport, intelligence: intelligenceReport };
  }

  async start(): Promise<DaemonStartResult> {
    const context = await this.discoveryEngine.discover();
    const healthReport = await this.healthEngine.assess(context);
    const previousContext = await this.storageService.readContext();
    const existingExecutionHistory = await this.storageService.readExecutionHistory();
    const previousStartup = existingExecutionHistory
      .slice()
      .reverse()
      .find((entry) => entry.command === 'daemon start' && entry.outcome === 'success');

    const samePackageJson =
      previousContext?.packageJson && context.packageJson
        ? JSON.stringify(previousContext.packageJson) === JSON.stringify(context.packageJson)
        : false;

    const sameLockFile =
      previousContext?.packageManagerLock && context.packageManagerLock
        ? previousContext.packageManagerLock === context.packageManagerLock
        : previousContext?.packageManagerLock === context.packageManagerLock;

    const safeToReuseStartup =
      Boolean(previousStartup && previousContext?.root === context.root && samePackageJson && sameLockFile);

    if (safeToReuseStartup) {
      console.log('Reusing execution memory from previous startup.');
    }

    const startupResult = await this.startupService.start(context, {
      skipInstall: safeToReuseStartup,
      activePorts: context.ports,
      previousStartupTimestamp: previousStartup?.timestamp,
    });
    const startupSuccess = startupResult.success;

    const historyEntry: ExecutionHistoryEntry = {
      id: `startup-${Date.now()}`,
      timestamp: new Date().toISOString(),
      command: 'daemon start',
      outcome: startupSuccess ? 'success' : 'failure',
      summary: startupSuccess ? 'Daemon startup completed' : `Startup failed at ${startupResult.failedStep ?? 'unknown step'}`,
      details: startupResult.failedStep ? `Failed step: ${startupResult.failedStep}` : 'All startup checks passed.',
    };

    await this.storageService.appendExecutionHistory(historyEntry);
    await this.storageService.writeContext(context);
    await this.storageService.writeHealth(healthReport);

    return {
      plan: startupResult.plan,
      startupSuccess,
      failedStep: startupResult.failedStep,
      health: healthReport,
      context,
    } as DaemonStartResult;
  }

  async recover(): Promise<{ context: ProjectContext; healthReport: HealthReport; recoveryPlan: RecoveryPlan }> {
    const context = await this.discoveryEngine.discover();
    const healthReport = await this.healthEngine.assess(context);
    const recoveryPlan = await this.recoveryEngine.buildRecoveryPlan(context, healthReport);

    await this.storageService.writeRecoveryPlan(recoveryPlan);
    await this.storageService.appendRecoveryHistory({
      id: `recover-${Date.now()}`,
      timestamp: new Date().toISOString(),
      actions: recoveryPlan.actions,
      outcome: recoveryPlan.actions.length === 0 ? 'success' : 'warning',
      summary: 'Recovery plan generated during recover flow.',
    });

    return { context, healthReport, recoveryPlan };
  }
}
