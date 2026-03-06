# specs/overlay.jq
#
# Patches applied on top of the official Atlassian Bitbucket DC OpenAPI spec
# before running openapi-generator. The official spec is never modified.
#
# Apply with:
#   jq -f specs/overlay.jq specs/bitbucket-openapi.json > specs/bitbucket-openapi-patched.json
#
# Each patch is documented with the reason it is needed.

# ── Patch 1: RestWebhook.id ──────────────────────────────────────────────────
# The Bitbucket DC API returns an "id" field on every webhook response object,
# but the Atlassian spec omits it entirely. Without this field the provider
# cannot read the webhook ID after creation and all webhook CRUD operations fail.
.components.schemas.RestWebhook.properties.id = {
  "type": "integer",
  "format": "int64",
  "readOnly": true,
  "description": "Unique webhook ID. Returned by the API but absent from the official spec."
}

# ── Patch 2: AddSshKeyRequest ────────────────────────────────────────────────
# The official spec defines AddSshKeyRequest as an empty object {}, losing all
# field definitions. The actual API response for SSH key operations returns the
# full key payload.
#
# createdDate is intentionally omitted from the properties below.
# The Atlassian spec types the equivalent field as "string / date-time", but the
# API returns a Unix millisecond integer. Go's time.Time JSON unmarshaler rejects
# integers, causing all SSH key reads to fail. The provider never uses createdDate,
# so omitting it means the JSON decoder silently skips the field — no unmarshal
# error and no hand-edit to a generated file required.
#
# The same field also appears in the inline request body schema on the
# POST /ssh/latest/keys endpoint. The generator resolves that inline schema
# into AddSshKeyRequest, so createdDate must be removed there too.
| del(.paths."/ssh/latest/keys".post.requestBody.content."application/json".schema.properties.createdDate)
# ── Patch 3: RestSshAccessKey.key → $ref AddSshKeyRequest ───────────────────
# The spec defines RestSshAccessKey.key as an inline schema identical to
# RestSshKey. Without a $ref, the generator creates a new anonymous type
# (RestSshAccessKeyKey) instead of reusing AddSshKeyRequest. The provider
# calls .GetKey() and expects AddSshKeyRequest, so we point the field at the
# named schema we define in Patch 2.
| .components.schemas.RestSshAccessKey.properties.key = {
  "$ref": "#/components/schemas/AddSshKeyRequest"
}

| .components.schemas.AddSshKeyRequest = {
  "type": "object",
  "properties": {
    "algorithmType":     { "type": "string" },
    "bitLength":         { "type": "integer", "format": "int32" },
    "expiryDays":        { "type": "integer", "format": "int32" },
    "fingerprint":       { "type": "string", "readOnly": true },
    "id":                { "type": "integer", "format": "int32", "readOnly": true },
    "label":             { "type": "string" },
    "lastAuthenticated": { "type": "string", "readOnly": true },
    "text":              { "type": "string" },
    "warning":           {
      "type": "string",
      "readOnly": true,
      "description": "Contains a warning about the key, for example that it's deprecated."
    }
  }
}
