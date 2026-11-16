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

## Log download and custom release ServiceAccount

On an authorized disposable K3s environment, deploy with a non-default Helm
release and use the same private configuration format described above. Run:

```sh
CHROMIUM_PATH=/path/to/chromium node tests/dashboard/log_download.cjs /private/evidence-directory
```

This drives Cluster Management -> Log Download -> last 30 minutes in the browser.
It verifies the created Job uses the actual Dashboard Pod ServiceAccount, waits
for successful completion and the visible download link, downloads the archive,
and rejects generated credential paths. The evidence directory also contains a
private log archive and the private login configuration; never publish either.
`kubectl` must access the same disposable cluster as the configured Dashboard.

For the missing-account negative case, explicitly configure the disposable
Dashboard's `SERVICE_ACCOUNT` to a nonexistent name, put that exact name in the
private config's `expectedServiceAccount`, and run with `EXPECT_MISSING_ACCOUNT=1`.
The scenario requires the browser to show View Details and the actual Kubernetes
FailedCreate/serviceaccount-not-found error. Restore the deployment environment
when finished. Do not run this fault injection on a user acceptance environment.
