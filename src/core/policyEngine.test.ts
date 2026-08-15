import test from 'node:test';
import assert from 'node:assert/strict';
import { PolicyEngine } from './policyEngine';
import { RecoveryAction } from '../types';

test('PolicyEngine denies unsafe recovery actions', () => {
  const pe = new PolicyEngine();
  const unsafeAction: RecoveryAction = {
    id: 'test.unsafe',
    title: 'Unsafe command',
    description: 'Deletes root directory',
    safe: false,
    confidence: 50,
    executed: false,
    command: 'rm -rf /',
  };

  const res = pe.evaluateAction(unsafeAction);
  assert.equal(res.decision, 'deny');
});

test('PolicyEngine allows safe recovery actions', () => {
  const pe = new PolicyEngine();
  const safeAction: RecoveryAction = {
    id: 'test.safe',
    title: 'Install missing package',
    description: 'Runs npm install',
    safe: true,
    confidence: 100,
    executed: false,
    command: 'npm install',
  };

  const res = pe.evaluateAction(safeAction);
  assert.equal(res.decision, 'allow');
});
