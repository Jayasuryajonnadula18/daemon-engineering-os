import { RecoveryAction } from '../types';

export type PolicyDecision = 'allow' | 'deny' | 'confirm';

export interface PolicyEvaluation {
  decision: PolicyDecision;
  reason: string;
}

export class PolicyEngine {
  evaluateAction(action: RecoveryAction): PolicyEvaluation {
    if (!action.safe) {
      return {
        decision: 'deny',
        reason: 'Action is flagged as unsafe by recovery engine.',
      };
    }

    if (action.command) {
      const lowerCmd = action.command.toLowerCase();
      if (lowerCmd.includes('rm -rf /') || lowerCmd.includes('push --force') || lowerCmd.includes('force')) {
        return {
          decision: 'deny',
          reason: 'Hard ceiling: destructive or force command blocked by Policy Engine.',
        };
      }
    }

    return {
      decision: 'allow',
      reason: 'Action approved by policy checks.',
    };
  }
}
