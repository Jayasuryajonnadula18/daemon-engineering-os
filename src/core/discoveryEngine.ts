import fs from 'fs/promises';
import path from 'path';
import { ProjectContext } from '../types';
import { getGitSummary } from '../utils/git';
import { getNodeSummary } from '../utils/node';
import { getEnvSummary } from '../utils/env';
import { ConfidenceEngine } from './confidenceEngine';

const frameworkKeywords: Record<string, string> = {
  next: 'Next.js',
  react: 'React',
  express: 'Express',
};

const databaseKeywords: Record<string, string> = {
  pg: 'PostgreSQL',
  postgres: 'PostgreSQL',
  mongoose: 'MongoDB',
  redis: 'Redis',
};

const envFiles = ['.env', '.env.local', '.env.development', '.env.production'];
const lockFiles = ['pnpm-lock.yaml', 'package-lock.json', 'yarn.lock', 'bun.lockb'];
const cloudFiles = [
  'azure-pipelines.yml',
  'cloudbuild.yaml',
  'serverless.yml',
  'pulumi.yaml',
  'pulumi.yml',
  'terraform.tf',
  'docker-compose.yml',
  'docker-compose.yaml',
  'kustomization.yaml',
  'helm-chart.yaml',
  'Chart.yaml',
  'deployment.yaml',
  'service.yaml',
];

const dockerServiceDependencies = ['PostgreSQL', 'Redis', 'Prisma', 'MongoDB'];

const serviceRelationshipMap: Record<string, string[]> = {
  'Next.js': ['React'],
  React: [],
  Express: [],
  Prisma: ['PostgreSQL', 'MySQL', 'SQLite', 'MariaDB'],
  PostgreSQL: [],
  Redis: [],
  MongoDB: [],
  'Developer runtime': [],
  'Start script': [],
  'Unknown service': [],
};

function normalizePackageDependencies(pkg: any): string[] {
  return [
    ...Object.keys(pkg.dependencies || {}),
    ...Object.keys(pkg.devDependencies || {}),
  ].map((name) => name.toLowerCase());
}

async function exists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

export async function detectProjectLanguages(root: string, packageJson: Record<string, unknown> = {}): Promise<string[]> {
  const languages = new Set<string>();

  if (packageJson && Object.keys(packageJson).length > 0) {
    languages.add('JavaScript');
  }

  if (await exists(path.join(root, 'tsconfig.json'))) {
    languages.add('TypeScript');
  } else if (
    packageJson &&
    typeof packageJson === 'object' &&
    ((packageJson as Record<string, unknown>).dependencies && typeof (packageJson as Record<string, unknown>).dependencies === 'object' && 'typescript' in ((packageJson as Record<string, unknown>).dependencies as Record<string, unknown>)) ||
    ((packageJson as Record<string, unknown>).devDependencies && typeof (packageJson as Record<string, unknown>).devDependencies === 'object' && 'typescript' in ((packageJson as Record<string, unknown>).devDependencies as Record<string, unknown>))
  ) {
    languages.add('TypeScript');
  }

  if (await exists(path.join(root, 'go.mod'))) {
    languages.add('Go');
  }

  if (
    await exists(path.join(root, 'pyproject.toml')) ||
    await exists(path.join(root, 'requirements.txt')) ||
    await exists(path.join(root, 'setup.py')) ||
    await exists(path.join(root, 'Pipfile'))
  ) {
    languages.add('Python');
  }

  return Array.from(languages);
}

export async function resolveProjectRoot(startPath: string, gitRoot?: string): Promise<string> {
  const candidates = [startPath];

  if (gitRoot) {
    candidates.push(gitRoot);
  }

  for (const candidate of candidates) {
    const current = path.resolve(candidate);
    const markers = [
      path.join(current, 'package.json'),
      path.join(current, 'tsconfig.json'),
      path.join(current, 'go.mod'),
      path.join(current, 'pyproject.toml'),
      path.join(current, 'requirements.txt'),
      path.join(current, 'setup.py'),
      path.join(current, 'Pipfile'),
    ];

    const hasMarkers = await Promise.all(markers.map((marker) => exists(marker)));
    if (hasMarkers.some(Boolean)) {
      return current;
    }
  }

  let current = path.resolve(startPath);
  let parent = path.dirname(current);
  while (parent !== current) {
    const markers = [
      path.join(current, 'package.json'),
      path.join(current, 'tsconfig.json'),
      path.join(current, 'go.mod'),
      path.join(current, 'pyproject.toml'),
      path.join(current, 'requirements.txt'),
      path.join(current, 'setup.py'),
      path.join(current, 'Pipfile'),
    ];

    const hasMarkers = await Promise.all(markers.map((marker) => exists(marker)));
    if (hasMarkers.some(Boolean)) {
      return current;
    }

    current = parent;
    parent = path.dirname(current);
  }

  return gitRoot ? path.resolve(gitRoot) : path.resolve(startPath);
}

export class DiscoveryEngine {
  async discover(): Promise<ProjectContext> {
    const root = await this.resolveRoot();
    const packageJson = await this.loadPackageJson(root);
    const gitSummary = await getGitSummary();
    const nodeSummary = await getNodeSummary();
    const envSummary = await getEnvSummary();

    const dependencies = normalizePackageDependencies(packageJson);
    const languages = await detectProjectLanguages(root, packageJson);
    const framework = this.detectFramework(dependencies, packageJson);
    const packageManager = await this.detectPackageManager(root, packageJson);
    const packageManagerLock = await this.detectLockFile(root);
    const database = this.detectDatabase(dependencies, envSummary);
    const services = this.detectServices(dependencies, root, packageJson);
    const foundEnvFiles = await this.detectEnvFiles(root);
    const ports = this.detectPorts(packageJson);
    const runningServices = this.detectRunningServices(dependencies, packageJson);
    const cloudConfig = await this.detectCloudConfig(root);
    const monorepo = await this.detectMonorepo(packageJson, root);
    const requiresDocker = this.inferDockerRequirement(services, cloudConfig, packageJson);
    const serviceGraph = this.buildServiceGraph(services, database);
    const confidence = new ConfidenceEngine().computeContextConfidence({
      name: packageJson.name || path.basename(root),
      root,
      framework,
      languages,
      runtimes: this.detectRuntimes(languages, nodeSummary),
      packageManager,
      packageManagerLock,
      docker: requiresDocker,
      git: gitSummary.repository !== 'Not a git repository',
      database,
      services,
      envFiles: foundEnvFiles,
      healthHints: [],
      confidence: 0,
      packageJson,
      monorepo,
      ports,
      cloudConfig,
      runningServices,
      serviceGraph,
    });

    return {
      name: packageJson.name || path.basename(root),
      root,
      framework,
      languages,
      runtimes: this.detectRuntimes(languages, nodeSummary),
      packageManager,
      packageManagerLock,
      docker: requiresDocker,
      git: gitSummary.repository !== 'Not a git repository',
      database,
      services,
      envFiles: foundEnvFiles,
      healthHints: [],
      confidence,
      packageJson,
      monorepo,
      ports,
      cloudConfig,
      serviceGraph,
      runningServices,
    };
  }

  private async resolveRoot(): Promise<string> {
    try {
      const gitSummary = await getGitSummary();
      if (gitSummary.repository && gitSummary.repository !== 'Not a git repository') {
        return resolveProjectRoot(process.cwd(), gitSummary.repository);
      }
    } catch {
      // fallback to cwd
    }

    return resolveProjectRoot(process.cwd());
  }

  private async loadPackageJson(root: string): Promise<any> {
    const packagePath = path.join(root, 'package.json');
    try {
      const contents = await fs.readFile(packagePath, 'utf-8');
      return JSON.parse(contents);
    } catch {
      return {};
    }
  }

  private detectFramework(dependencies: string[], pkg: any): string | undefined {
    const packageManagerField = String(pkg.packageManager || '').toLowerCase();
    if (packageManagerField.includes('next')) {
      return 'Next.js';
    }

    for (const keyword of Object.keys(frameworkKeywords)) {
      if (dependencies.includes(keyword)) {
        return frameworkKeywords[keyword];
      }
    }

    return undefined;
  }

  private async detectPackageManager(root: string, pkg: any): Promise<string | undefined> {
    for (const lockFile of lockFiles) {
      if (await exists(path.join(root, lockFile))) {
        if (lockFile === 'pnpm-lock.yaml') return 'pnpm';
        if (lockFile === 'package-lock.json') return 'npm';
        if (lockFile === 'yarn.lock') return 'yarn';
        if (lockFile === 'bun.lockb') return 'bun';
      }
    }

    const packageManager = pkg.packageManager;
    if (typeof packageManager === 'string') {
      return packageManager;
    }

    return undefined;
  }

  private async detectLockFile(root: string): Promise<string | undefined> {
    for (const lockFile of lockFiles) {
      if (await exists(path.join(root, lockFile))) {
        return lockFile;
      }
    }
    return undefined;
  }

  private detectDatabase(dependencies: string[], envSummary: any): string | undefined {
    for (const keyword of Object.keys(databaseKeywords)) {
      if (dependencies.includes(keyword)) {
        return databaseKeywords[keyword];
      }
    }

    if (envSummary.present.includes('DATABASE_URL')) {
      return 'PostgreSQL';
    }

    return undefined;
  }

  private detectServices(dependencies: string[], root: string, pkg: any): string[] {
    const services: string[] = [];
    if (dependencies.includes('next')) {
      services.push('Next.js');
    }
    if (dependencies.includes('react')) {
      services.push('React');
    }
    if (dependencies.includes('express')) {
      services.push('Express');
    }
    if (dependencies.includes('prisma')) {
      services.push('Prisma');
    }
    if (dependencies.includes('pg') || dependencies.includes('postgres')) {
      services.push('PostgreSQL');
    }
    if (dependencies.includes('redis')) {
      services.push('Redis');
    }
    if (pkg.scripts && typeof pkg.scripts === 'object') {
      if (pkg.scripts.dev) {
        services.push('Developer runtime');
      }
      if (pkg.scripts.start) {
        services.push('Start script');
      }
    }

    if (services.length === 0) {
      services.push('Unknown service');
    }

    return services;
  }

  private async detectEnvFiles(root: string): Promise<string[]> {
    const found: string[] = [];
    for (const file of envFiles) {
      if (await exists(path.join(root, file))) {
        found.push(file);
      }
    }
    return found;
  }

  private async detectCloudConfig(root: string): Promise<string[]> {
    const found: string[] = [];
    for (const file of cloudFiles) {
      if (await exists(path.join(root, file))) {
        found.push(file);
      }
    }
    return found;
  }

  private async detectMonorepo(pkg: any, root: string): Promise<boolean> {
    if (Array.isArray(pkg.workspaces) || typeof pkg.workspaces === 'object') {
      return true;
    }

    const packagesDir = path.join(root, 'packages');
    return await exists(packagesDir);
  }

  private detectRuntimes(languages: string[], nodeSummary: { version?: string }): string[] {
    const runtimes: string[] = [];

    if (languages.includes('JavaScript') || languages.includes('TypeScript')) {
      if (nodeSummary.version) {
        runtimes.push(`Node.js ${nodeSummary.version}`);
      }
    }

    if (languages.includes('Go')) {
      runtimes.push('Go');
    }

    if (languages.includes('Python')) {
      runtimes.push('Python');
    }

    return runtimes;
  }

  private detectPorts(pkg: any): number[] {
    const candidates = new Set<number>();
    const scriptRegex = /(?:--port|PORT=|port\s*[:=])\s*([0-9]{2,5})/gi;

    const scripts = pkg.scripts || {};
    for (const value of Object.values(scripts)) {
      if (typeof value !== 'string') {
        continue;
      }
      let match: RegExpExecArray | null;
      while ((match = scriptRegex.exec(value)) !== null) {
        candidates.add(Number(match[1]));
      }
    }

    return Array.from(candidates).sort((a, b) => a - b);
  }

  private inferDockerRequirement(services: string[], cloudConfig: string[], pkg: any): boolean {
    const hasDatabaseService = services.some((service) => dockerServiceDependencies.includes(service));
    const hasDockerCompose = cloudConfig.some((file) => file.startsWith('docker-compose'));
    const hasDockerScript = Object.values(pkg.scripts || {}).some((script) => typeof script === 'string' && /docker/.test(script));
    return hasDatabaseService || hasDockerCompose || hasDockerScript;
  }

  private buildServiceGraph(services: string[], database?: string): Record<string, string[]> {
    const graph: Record<string, string[]> = {};

    for (const service of services) {
      const relationships = serviceRelationshipMap[service] ?? [];
      const targets = [...relationships];

      if (service === 'Prisma' && database) {
        targets.push(database);
      }

      graph[service] = Array.from(new Set(targets));
    }

    if (database && !services.includes(database)) {
      graph[database] = [];
    }

    return graph;
  }

  private detectRunningServices(dependencies: string[], pkg: any): string[] {
    const services: string[] = [];

    if (dependencies.includes('next')) {
      services.push('Next.js');
    }
    if (dependencies.includes('react')) {
      services.push('React');
    }
    if (dependencies.includes('express')) {
      services.push('Express');
    }
    if (pkg.scripts && typeof pkg.scripts === 'object') {
      if (pkg.scripts.dev) {
        services.push('Developer runtime');
      }
      if (pkg.scripts.start) {
        services.push('Start script');
      }
    }

    return services.length > 0 ? services : ['Unknown service'];
  }
}
