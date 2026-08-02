export interface GitSummary {
  repository: string;
  branch: string;
  status: string;
  clean: boolean;
  lastCommit?: string;
}

export interface DockerSummary {
  installed: boolean;
  available: boolean;
  version?: string;
}

export interface NodeSummary {
  version: string;
  npmVersion: string;
  packageManager?: string;
}

export interface EnvSummary {
  required: string[];
  present: string[];
  missing: string[];
  envFiles: string[];
}

export interface SystemSummary {
  platform: string;
  cpuCount: number;
  memoryGb: number;
  freeMemoryGb: number;
}

export interface DoctorReport {
  git: GitSummary;
  docker: DockerSummary;
  node: NodeSummary;
  env: EnvSummary;
  system: SystemSummary;
  overallHealthScore?: number;
}

export interface ProjectContext {
  name: string;
  root: string;
  framework?: string;
  languages: string[];
  runtimes: string[];
  packageManager?: string;
  packageManagerLock?: string;
  docker: boolean;
  git: boolean;
  database?: string;
  services: string[];
  envFiles: string[];
  healthHints: string[];
  confidence: number;
  packageJson?: Record<string, unknown>;
  monorepo: boolean;
  ports: number[];
  cloudConfig: string[];
  serviceGraph: Record<string, string[]>;
  runningServices: string[];
}

export interface ExecutionHistoryEntry {
  id: string;
  timestamp: string;
  command: string;
  outcome: 'success' | 'warning' | 'failure';
  summary: string;
  details?: string;
}

export interface RecoveryHistoryEntry {
  id: string;
  timestamp: string;
  actions: RecoveryAction[];
  outcome: 'success' | 'warning' | 'failure';
  summary: string;
}

export interface HealthHistoryEntry {
  id: string;
  timestamp: string;
  score: number;
  issueCount: number;
  warningCount: number;
  recommendationCount: number;
}

export interface StartupOptions {
  skipInstall?: boolean;
  activePorts?: number[];
  previousStartupTimestamp?: string;
}

export interface ExecutionPlanStep {
  title: string;
  description: string;
  action: 'verify' | 'install' | 'start' | 'check' | 'recommend';
  command?: string;
  expectedPorts?: number[];
  confidence: number;
}

export interface ExecutionPlan {
  name: string;
  summary: string;
  steps: ExecutionPlanStep[];
}

export interface HealthIssue {
  id: string;
  title: string;
  severity: 'info' | 'warning' | 'critical';
  detail: string;
}

export interface HealthRuleContext {
  context: ProjectContext;
  gitSummary: GitSummary;
  dockerSummary: DockerSummary;
  nodeSummary: NodeSummary;
  envSummary: EnvSummary;
  systemSummary: SystemSummary;
  healthReport?: HealthReport;
}

export interface HealthRule {
  id: string;
  title: string;
  severity: 'info' | 'warning' | 'critical';
  detail: (ctx: HealthRuleContext) => string;
  condition: (ctx: HealthRuleContext) => boolean;
  recommendation?: (ctx: HealthRuleContext) => string;
}

export interface RecoveryRule {
  id: string;
  title: string;
  description: (ctx: HealthRuleContext) => string;
  safe: boolean;
  confidence: number;
  condition: (ctx: HealthRuleContext) => boolean;
  command?: (ctx: HealthRuleContext) => string;
  metadata?: (ctx: HealthRuleContext) => Record<string, unknown>;
}

export interface HealthReport {
  score: number;
  issues: HealthIssue[];
  warnings: string[];
  recommendations: string[];
  timestamp: string;
}

export interface RecoveryAction {
  id: string;
  title: string;
  description: string;
  confidence: number;
  safe: boolean;
  executed: boolean;
  command?: string;
  metadata?: Record<string, unknown>;
}

export interface RecoveryPlan {
  summary: string;
  actions: RecoveryAction[];
}

export interface StartupResult {
  plan: ExecutionPlan;
  success: boolean;
  failedStep?: string;
}

export interface DaemonStartResult {
  plan: ExecutionPlan;
  startupSuccess: boolean;
  failedStep?: string;
  health: HealthReport;
  context: ProjectContext;
}

export interface ArchitectureAnalysis {
  issues: string[];
  recommendations: string[];
}

export interface BreakingChangeAnalysis {
  detected: boolean;
  reasons: string[];
  summary: string;
}

export interface RiskAnalysis {
  level: 'Low' | 'Medium' | 'High' | 'Critical';
  reasons: string[];
  summary: string;
}

export interface BlastRadiusAnalysis {
  level: 'Low' | 'Medium' | 'High' | 'Critical';
  affectedAreas: string[];
  summary: string;
}

export interface AnalysisInsight {
  title: string;
  detail: string;
  severity: 'info' | 'warning' | 'critical';
}

export interface EngineeringIntelligenceReport {
  startupSequence: string[];
  serviceGraph: Record<string, string[]>;
  blastRadius: BlastRadiusAnalysis;
  architecture: ArchitectureAnalysis;
  risk: RiskAnalysis;
  breakingChanges: BreakingChangeAnalysis;
  recommendations: string[];
  insights: AnalysisInsight[];
}

export interface InspectionResult {
  context: ProjectContext;
  healthReport: HealthReport;
  intelligence?: EngineeringIntelligenceReport;
}
