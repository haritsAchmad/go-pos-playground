import { defineConfig, devices } from '@playwright/test'

const baseURL = process.env.E2E_BASE_URL || 'http://127.0.0.1:3000'

export default defineConfig({
  testDir: './e2e/authorization',
  fullyParallel: false,
  workers: 1,
  timeout: 30_000,
  expect: { timeout: 8_000 },
  outputDir: 'test-results/authz/artifacts',
  reporter: [
    ['list'],
    ['./e2e/authorization/security-reporter.ts'],
    ['html', { outputFolder: 'test-results/authz/html', open: 'never' }],
    ['json', { outputFile: 'test-results/authz/results.json' }],
  ],
  use: {
    baseURL,
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
    video: 'retain-on-failure',
    ...devices['Desktop Chrome'],
  },
  projects: [
    { name: 'setup', testMatch: /global\.setup\.ts/ },
    {
      name: 'chromium',
      testIgnore: /global\.setup\.ts/,
      dependencies: ['setup'],
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'google-chrome',
      testIgnore: /global\.setup\.ts/,
      dependencies: ['setup'],
      use: { ...devices['Desktop Chrome'], channel: 'chrome' },
    },
    {
      name: 'microsoft-edge',
      testIgnore: /global\.setup\.ts/,
      dependencies: ['setup'],
      use: { ...devices['Desktop Edge'], channel: 'msedge' },
    },
    {
      name: 'firefox',
      testIgnore: /global\.setup\.ts/,
      dependencies: ['setup'],
      use: { ...devices['Desktop Firefox'] },
    },
  ],
})
