# API Specification

## Files

| File | Description |
|---|---|
| `bitbucket-openapi.json` | Official Atlassian Bitbucket DC OpenAPI spec (unmodified) |
| `overlay.jq` | Patches applied on top of the official spec before code generation |
| `bitbucket-openapi-patched.json` | Generated at build time — do not edit or commit |

## Why we use an overlay instead of editing the spec directly

The official `bitbucket-openapi.json` is sourced from Atlassian and should remain
unmodified. Editing it directly creates a maintenance problem: when a new version of
the spec is released, we would need to manually re-apply all patches to the new file,
with no record of what was changed or why.

Instead, patches are maintained in `overlay.jq` as a jq transformation. The patched
spec is produced at build time:

```bash
make generate-client
# internally runs:
#   jq -f specs/overlay.jq specs/bitbucket-openapi.json > specs/bitbucket-openapi-patched.json
#   openapi-generator generate -i specs/bitbucket-openapi-patched.json ...
```

When upgrading to a new spec version, replace `bitbucket-openapi.json` and re-run
`make generate-client`. The overlay is applied automatically and all patches are
preserved with their documented rationale.

## Generator quirks (see .openapi-generator-ignore for full detail)

Some files are excluded from regeneration in `.openapi-generator-ignore`:

- **`api_deprecated.go` / `client.go`** — the deprecated API file declares duplicate types that conflict with regular API files. The committed `client.go` has two references to `DeprecatedAPIService` removed.
- **`model_comment.go`, `model_comment_thread.go`, `model_pull_request_participant.go`** — these have mutually recursive types. The generator emits value types (`[]Comment`, `CommentThread`) which Go rejects as infinite-size; the committed files use pointer types. None are used by the provider.

## Current patches (see overlay.jq for full detail)

### RestWebhook.id
The Bitbucket DC API returns an `id` field on every webhook response, but the
Atlassian spec omits it. Without it the provider cannot track webhook IDs after
creation.

### AddSshKeyRequest
The official spec defines this schema as an empty object `{}`, losing all field
definitions. The actual API returns a full SSH key payload. `createdDate` is
omitted intentionally — it is typed `string/date-time` in the spec but the API
returns a Unix millisecond integer, which Go's `time.Time` JSON unmarshaler
rejects. The provider never reads `createdDate`, so omitting it lets the JSON
decoder silently skip the field without error.

### RestSshAccessKey.key → $ref AddSshKeyRequest
Without this patch the generator creates an anonymous inline type
`RestSshAccessKeyKey` instead of reusing `AddSshKeyRequest`. The provider calls
`.GetKey()` on `RestSshAccessKey` and expects `AddSshKeyRequest`.
