## ADDED Requirements

### Requirement: Generate provider documentation with tfplugindocs
The system SHALL generate Terraform provider documentation in `docs/` using `tfplugindocs generate`.

#### Scenario: Documentation generated successfully
- **WHEN** `make docs` is run
- **THEN** `docs/index.md`, `docs/resources/*.md`, and `docs/data-sources/*.md` are created or updated

#### Scenario: Documentation reflects schema descriptions
- **WHEN** a resource schema has a `Description` field set
- **THEN** the generated doc for that resource contains the description text

### Requirement: Makefile target for documentation generation
The system SHALL expose a `make docs` target that runs tfplugindocs.

#### Scenario: Make docs target exists
- **WHEN** user runs `make docs`
- **THEN** tfplugindocs generates docs without error
