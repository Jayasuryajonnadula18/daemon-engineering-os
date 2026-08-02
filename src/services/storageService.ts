import fs from 'fs/promises';
import path from 'path';
import {
  ExecutionHistoryEntry,
  ExecutionPlan,
  HealthHistoryEntry,
  HealthReport,
  ProjectContext,
  RecoveryHistoryEntry,
  RecoveryPlan,
} from '../types';

export class StorageService {
  constructor(private root: string = process.cwd()) {}

  private get daemonDir(): string {
    return path.join(this.root, '.daemon');
  }

  async ensureDaemonDirectory(): Promise<void> {
    await fs.mkdir(this.daemonDir, { recursive: true });
  }

  async writeJson(fileName: string, data: unknown): Promise<void> {
    await this.ensureDaemonDirectory();
    const target = path.join(this.daemonDir, fileName);
    await fs.writeFile(target, JSON.stringify(data, null, 2), 'utf-8');
  }

  async readJson<T>(fileName: string): Promise<T | undefined> {
    try {
      const target = path.join(this.daemonDir, fileName);
      const content = await fs.readFile(target, 'utf-8');
      return JSON.parse(content) as T;
    } catch {
      return undefined;
    }
  }

  async writeContext(context: ProjectContext): Promise<void> {
    await this.writeJson('context.json', context);
  }

  async readContext(): Promise<ProjectContext | undefined> {
    return this.readJson<ProjectContext>('context.json');
  }

  async writeHealth(health: HealthReport): Promise<void> {
    await this.writeJson('health.json', health);
  }

  async readHealthHistory(): Promise<HealthHistoryEntry[]> {
    return (await this.readJson<HealthHistoryEntry[]>('health-history.json')) || [];
  }

  async appendExecutionHistory(entry: ExecutionHistoryEntry): Promise<void> {
    const existing = (await this.readJson<ExecutionHistoryEntry[]>('execution-history.json')) || [];
    existing.push(entry);
    await this.writeJson('execution-history.json', existing);
  }

  async readExecutionHistory(): Promise<ExecutionHistoryEntry[]> {
    return (await this.readJson<ExecutionHistoryEntry[]>('execution-history.json')) || [];
  }

  async appendRecoveryHistory(entry: RecoveryHistoryEntry): Promise<void> {
    const existing = (await this.readJson<RecoveryHistoryEntry[]>('recovery-history.json')) || [];
    existing.push(entry);
    await this.writeJson('recovery-history.json', existing);
  }

  async appendHealthHistory(entry: HealthHistoryEntry): Promise<void> {
    const existing = (await this.readJson<HealthHistoryEntry[]>('health-history.json')) || [];
    existing.push(entry);
    await this.writeJson('health-history.json', existing);
  }

  async writeExecutionPlan(plan: ExecutionPlan): Promise<void> {
    await this.writeJson('execution-plan.json', plan);
  }

  async writeRecoveryPlan(plan: RecoveryPlan): Promise<void> {
    await this.writeJson('recovery-plan.json', plan);
  }

  async readRecoveryHistory(): Promise<RecoveryHistoryEntry[]> {
    return (await this.readJson<RecoveryHistoryEntry[]>('recovery-history.json')) || [];
  }

  async writeIntelligenceReport(report: unknown): Promise<void> {
    await this.writeJson('intelligence-report.json', report);
  }

  async readIntelligenceReport<T>(): Promise<T | undefined> {
    return this.readJson<T>('intelligence-report.json');
  }
}
