# Dashboard browser regression

Run on an authorized disposable pooled test machine with Node.js, Playwright and
Chromium installed. Never run against an acceptance environment: this scenario
creates and deletes temporary admin accounts and changes their initial passwords.
It also completes the supplied admin account's initial password reset if needed.

Create an owner-only evidence directory containing `browser-default-config.json`
(mode 0600) with `url`, admin `password`, `namespace`, Helm `release`, and
`clusterName`. Supply the password locally; do not commit or publish this file.
The script updates the private file if the admin password is rotated.

```sh
CHROMIUM_PATH=/path/to/chromium node tests/dashboard/login_readiness.cjs /private/evidence-directory
```

The scenario delays account/permission responses by three seconds, completes
first-login password resets with new accounts, and navigates through topology,
cluster overview and global overview. It fails on uncaught JavaScript errors,
missing navigation/content, or failed static assets. The private configuration
must be excluded when collecting screenshots and JSON evidence.

Set `RESET_ITERATIONS=10` for repeated cold-context first-login coverage. An
optional `nodeName` in the private config makes the global overview assertion
wait for that specific Kubernetes node row. Topology and cluster overview also
wait for their LogService content before navigation continues.
