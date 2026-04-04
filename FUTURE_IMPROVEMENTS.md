# Future Improvements

This file tracks worthwhile next improvements for the starter kit. The items below are intentionally scoped as future work, not current guarantees.

## Platform

- [ ] Add Redis-backed rate limiting for higher-throughput production deployments.
- [ ] Add background jobs and an outbox pattern for durable async event delivery.
- [ ] Add deployability assets such as Helm charts, Terraform examples, or production reverse-proxy configs.

## Auth and Security

- [ ] Add password reset flow.
- [ ] Add email verification flow.
- [ ] Add admin controls for revoking another user's sessions.
- [ ] Add stricter token claim validation and secret rotation guidance.
- [ ] Add CORS configuration and security headers middleware.
- [ ] Add account lockout or anomaly-detection protections for suspicious auth activity.

## Admin and Audit

- [ ] Add audit trail query endpoints for administrators.
- [ ] Add moderation capabilities such as soft delete, review status, or content flags.
- [ ] Add admin content-moderation APIs beyond user-role management.

## API and Product Features

- [ ] Add search support for posts.
- [ ] Add filtering and sorting for posts and comments.
- [ ] Add request and response examples throughout the OpenAPI spec.
- [ ] Split or generate the OpenAPI spec as it grows to keep it maintainable.

## Testing and Quality

- [ ] Add dedicated tests for rate limiting behavior.
- [ ] Add dedicated tests for session revocation and refresh-token reuse edge cases.
- [ ] Add dedicated tests for admin role changes and moderation behavior.
- [ ] Add dedicated tests for audit persistence.
- [ ] Add migration idempotency and migration-order tests.
- [ ] Enable automatic CI triggers in derived repos and optionally add coverage reporting.
