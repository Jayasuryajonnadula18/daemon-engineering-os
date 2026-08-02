import { ProjectContext, HealthReport, EngineeringIntelligenceReport, ArchitectureAnalysis, BlastRadiusAnalysis, BreakingChangeAnalysis, RiskAnalysis, AnalysisInsight } from '../types';
import { StorageService } from '../services/storageService';

export class IntelligenceEngine {
  private readonly storage = new StorageService();

  async analyze(context: ProjectContext, healthReport: HealthReport): Promise<EngineeringIntelligenceReport> {
    const previousContext = await this.storage.readContext();
    const startupSequence = this.buildStartupSequence(context);
    const breakingChanges = this.detectBreakingChanges(context, previousContext);
    const architecture = this.analyzeArchitecture(context, healthReport);
    const risk = this.assessRisk(context, healthReport, breakingChanges);
    const blastRadius = this.estimateBlastRadius(context, healthReport, breakingChanges);
    const recommendations = this.summarizeRecommendations(context, healthReport, architecture, breakingChanges);
    const insights = this.buildInsights(context, healthReport, breakingChanges);

    return {
      startupSequence,
      serviceGraph: context.serviceGraph,
      blastRadius,
      architecture,
      risk,
      breakingChanges,
      recommendations,
      insights,
    };
  }

  private buildStartupSequence(context: ProjectContext): string[] {
    const graph = context.serviceGraph || {};
    const visited = new Set<string>();
    const stack = new Set<string>();
    const order: string[] = [];

    const visit = (node: string) => {
      if (stack.has(node)) {
        return; // skip cycles gracefully
      }
      if (visited.has(node)) {
        return;
      }

      stack.add(node);
      const targets = graph[node] || [];
      for (const target of targets) {
        visit(target);
      }
      stack.delete(node);
      visited.add(node);
      order.push(node);
    };

    for (const node of Object.keys(graph)) {
      visit(node);
    }

    if (order.length === 0 && context.services.length > 0) {
      return [...context.services];
    }

    return order;
  }

  private detectBreakingChanges(
    context: ProjectContext,
    previousContext?: ProjectContext,
  ): BreakingChangeAnalysis {
    const reasons: string[] = [];
    let detected = false;

    if (!previousContext) {
      return { detected: false, reasons, summary: 'No prior project context available for comparison.' };
    }

    if (previousContext.packageManagerLock !== context.packageManagerLock) {
      detected = true;
      reasons.push('Package manager lock file changed.');
    }

    if (JSON.stringify(previousContext.packageJson || {}) !== JSON.stringify(context.packageJson || {})) {
      detected = true;
      reasons.push('package.json differs from the last known context.');
    }

    if (JSON.stringify(previousContext.serviceGraph || {}) !== JSON.stringify(context.serviceGraph || {})) {
      detected = true;
      reasons.push('Detected a change in service relationships or startup dependencies.');
    }

    return {
      detected,
      reasons,
      summary: detected ? 'Breaking-change indicators were found during analysis.' : 'No breaking-change indicators were detected.',
    };
  }

  private analyzeArchitecture(context: ProjectContext, healthReport: HealthReport): ArchitectureAnalysis {
    const issues: string[] = [];
    const recommendations: string[] = [];

    if (context.services.includes('Unknown service')) {
      issues.push('Project services could not be fully identified.');
      recommendations.push('Add clearer dependencies or startup scripts so Daemon can infer the system structure.');
    }

    if (context.packageManagerLock && !context.packageManager) {
      issues.push('Lock file is present but package manager is unknown.');
      recommendations.push('Verify package.json packageManager field or supported lock file mapping.');
    }

    if (healthReport.issues.some((issue) => issue.id === 'env.missing')) {
      issues.push('Required environment variables are missing.');
      recommendations.push('Declare required variables and provide a sample .env file for local developers.');
    }

    if (context.serviceGraph && Object.keys(context.serviceGraph).length > 0) {
      const duplicateServices = Object.entries(context.serviceGraph).filter(([service, relations]) => relations.includes(service));
      if (duplicateServices.length > 0) {
        issues.push('Service definition contains self-references.');
        recommendations.push('Review service graph relationships for duplicate or circular dependencies.');
      }
    }

    return { issues, recommendations };
  }

  private assessRisk(
    context: ProjectContext,
    healthReport: HealthReport,
    breakingChanges: BreakingChangeAnalysis,
  ): RiskAnalysis {
    const reasons: string[] = [];
    let level: RiskAnalysis['level'] = 'Low';

    if (healthReport.score < 60) {
      level = 'High';
      reasons.push('The project health score is low.');
    } else if (healthReport.score < 80) {
      level = 'Medium';
      reasons.push('The project health score is moderate.');
    }

    if (context.docker && !context.packageManager) {
      reasons.push('Docker is required but package manager detection is incomplete.');
    }

    if (breakingChanges.detected) {
      level = 'Critical';
      reasons.push('Breaking changes were detected in package or service metadata.');
    }

    if (healthReport.warnings.length > 0 && level !== 'Critical') {
      level = level === 'Low' ? 'Medium' : level;
      reasons.push('There are active health warnings.');
    }

    const summary = reasons.length > 0 ? reasons.join(' ') : 'No heightened risk detected for the current project state.';

    return { level, reasons, summary };
  }

  private estimateBlastRadius(
    context: ProjectContext,
    healthReport: HealthReport,
    breakingChanges: BreakingChangeAnalysis,
  ): BlastRadiusAnalysis {
    const affectedAreas = new Set<string>(context.services);

    if (context.database) {
      affectedAreas.add('Database');
    }
    if (context.docker) {
      affectedAreas.add('Containers');
    }
    if (healthReport.issues.some((issue) => issue.id.startsWith('git'))) {
      affectedAreas.add('Repository');
    }
    if (healthReport.issues.some((issue) => issue.id.startsWith('env'))) {
      affectedAreas.add('Environment variables');
    }
    if (breakingChanges.detected) {
      affectedAreas.add('Startup workflow');
    }

    const level = breakingChanges.detected || healthReport.score < 60 ? 'High' : healthReport.score < 80 ? 'Medium' : 'Low';
    const summary = `The current change set impacts ${Array.from(affectedAreas).join(', ') || 'project startup and configuration'}.`;

    return {
      level,
      affectedAreas: Array.from(affectedAreas),
      summary,
    };
  }

  private summarizeRecommendations(
    context: ProjectContext,
    healthReport: HealthReport,
    architecture: ArchitectureAnalysis,
    breakingChanges: BreakingChangeAnalysis,
  ): string[] {
    const recommendations = [...healthReport.recommendations, ...architecture.recommendations];

    if (breakingChanges.detected) {
      recommendations.push('Review the changed package metadata and service relationships before running startup or recovery.');
    }

    if (context.services.includes('Unknown service')) {
      recommendations.push('Add or improve package.json scripts so Daemon can infer the application startup flow.');
    }

    if (recommendations.length === 0) {
      recommendations.push('Continue with startup; no additional recommendations were detected.');
    }

    return recommendations;
  }

  private buildInsights(
    context: ProjectContext,
    healthReport: HealthReport,
    breakingChanges: BreakingChangeAnalysis,
  ): AnalysisInsight[] {
    const insights: AnalysisInsight[] = [];

    if (breakingChanges.detected) {
      insights.push({
        title: 'Breaking change detected',
        detail: breakingChanges.reasons.join(' '),
        severity: 'critical',
      });
    }

    if (context.ports.length > 0) {
      insights.push({
        title: 'Active service ports discovered',
        detail: `Detected ${context.ports.length} port(s) configured in startup scripts: ${context.ports.join(', ')}.`,
        severity: 'info',
      });
    }

    if (healthReport.score < 70) {
      insights.push({
        title: 'Startup risk is elevated',
        detail: `Health score is ${healthReport.score}%. Review warnings and recover before continuing.`,
        severity: 'warning',
      });
    }

    if (context.services.includes('Unknown service')) {
      insights.push({
        title: 'Service discovery is incomplete',
        detail: 'Daemon could not fully infer which services are present from package metadata.',
        severity: 'warning',
      });
    }

    return insights;
  }
}
