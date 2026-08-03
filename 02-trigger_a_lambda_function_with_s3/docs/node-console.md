# Node.js with the AWS Console

Repeat the S3 trigger tutorial using Node.js and record what differs from Python.

## What I did

- Reused the same bucket, and this time explicitly picked **Use an existing role** so the function ran under the role I'd already fixed with S3 permissions — no repeat of the Python IAM issues.
- Deployed the sample Node handler, which uses `HeadObjectCommand` (metadata only, no body download) instead of Python's `GetObject` (full object) to read `ContentType`.
- Worked without errors on the first try — the permissions and key-prefix lessons from the Python pass carried over.
