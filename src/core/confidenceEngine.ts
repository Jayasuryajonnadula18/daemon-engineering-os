import { ProjectContext } from '../types';

const baseConfidence = 40;

export class ConfidenceEngine {
  computeContextConfidence(context: ProjectContext): number {
    let score = baseConfidence;

    if (context.framework) {
      score += 20;
    }

    if (context.packageManager) {
      score += 15;
    }

    if (context.docker) {
      score += 10;
    }

    if (context.git) {
      score += 10;
    }

    if (context.services.length > 0) {
      score += 10;
    }

    if (context.envFiles.length > 0) {
      score += 5;
    }

    return Math.min(100, Math.max(0, score));
  }

  format(score: number): string {
    if (score >= 95) {
      return 'Certain';
    }
    if (score >= 80) {
      return 'Likely';
    }
    if (score >= 60) {
      return 'Partial';
    }
    return 'Unknown';
  }
}
