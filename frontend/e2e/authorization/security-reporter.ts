import fs from 'node:fs'
import path from 'node:path'
import type { FullResult, Reporter, TestCase, TestResult } from '@playwright/test/reporter'

type Finding = {
  browser: string
  endpoint: string
  account: string
  expected: string
  actual: string
  risk: string
  rootCause: string
  artifact: string
}

export default class SecurityReporter implements Reporter {
  private findings: Finding[] = []

  onTestEnd(test: TestCase, result: TestResult) {
    if (result.status === test.expectedStatus) return
    const annotations = new Map(result.annotations.map(item => [item.type, item.description || '']))
    this.findings.push({
      browser: test.parent.project()?.name || 'unknown',
      endpoint: annotations.get('endpoint') || test.title,
      account: annotations.get('account') || 'multiple/see test title',
      expected: annotations.get('expected') || test.expectedStatus,
      actual: annotations.get('actual') || `${result.status}: ${result.error?.message?.replace(/\r?\n/g, ' ') || 'no error message'}`,
      risk: annotations.get('risk') || 'medium',
      rootCause: annotations.get('root_cause_hint') || 'Inspect the failing trace and authorization middleware.',
      artifact: result.attachments.map(item => item.path).filter(Boolean).join(', ') || 'See Playwright HTML report',
    })
  }

  onEnd(result: FullResult) {
    const dir = path.resolve('test-results/authz')
    fs.mkdirSync(dir, { recursive: true })
    const lines = [
      '# Automated Authorization Audit',
      '',
      `Overall status: **${result.status}**`,
      `Generated: ${new Date().toISOString()}`,
      '',
      '## Role mapping',
      '',
      '| Requested persona | Application role | Meaning in this audit |',
      '|---|---|---|',
      '| superadmin | admin | Full administrative policy |',
      '| manager | viewer | Read-only policy; the application has no manager role |',
      '| cashier | cashier | Operational write policy |',
      '| unauthenticated | none | No session or access token |',
      '',
      '## Failures',
      '',
    ]
    if (!this.findings.length) {
      lines.push('No authorization expectation failures were detected.')
    } else {
      lines.push('| Browser | Endpoint/UI route | Account | Expected | Actual / status code | Risk | Likely root cause | Artifacts |', '|---|---|---|---|---|---|---|---|')
      for (const finding of this.findings) {
        lines.push(`| ${escape(finding.browser)} | ${escape(finding.endpoint)} | ${escape(finding.account)} | ${escape(finding.expected)} | ${escape(finding.actual)} | ${escape(finding.risk)} | ${escape(finding.rootCause)} | ${escape(finding.artifact)} |`)
      }
    }
    lines.push(
      '',
      '## Scope notes',
      '',
      '- API tests cover anonymous, every disallowed role, every allowed role, ID replacement, logout replay, expiry, and direct requests.',
      '- UI tests cover anonymous redirects and direct route navigation for every persona.',
      '- Mutation probes use empty payloads and non-existent sentinel IDs; they verify authorization ordering without changing business data.',
      '- Screenshots, videos, traces, JSON, and the HTML report are stored below `test-results/authz/`.',
      '- Object ownership/tenant isolation cannot be asserted because the current data model has no resource owner or tenant boundary.',
      '',
    )
    fs.writeFileSync(path.join(dir, 'authorization-audit.md'), lines.join('\n'))
  }
}

function escape(value: string) {
  return value.replace(/\|/g, '\\|').replace(/\r?\n/g, ' ')
}
