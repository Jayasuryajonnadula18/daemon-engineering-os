import { DoctorReport } from '../types';

export class DiagnosisService {
  public evaluateHealth(report: DoctorReport): number {
    let score = 100;

    if (!report.git.repository || report.git.repository === 'Not a git repository') {
      score -= 30;
    }

    if (!report.docker.installed || !report.docker.available) {
      score -= 20;
    }

    if (report.env.missing.length > 0) {
      score -= report.env.missing.length * 10;
    }

    if (report.system.freeMemoryGb < 1) {
      score -= 10;
    }

    return Math.max(0, Math.min(score, 100));
  }
}
