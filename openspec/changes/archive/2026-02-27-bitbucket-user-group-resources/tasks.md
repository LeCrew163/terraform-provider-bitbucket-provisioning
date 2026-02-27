## 1. Group Resource

- [ ] 1.1 Create `internal/provider/resource_group.go` with Create/Read/Delete/Import (name as ID, RequiresReplace)
- [ ] 1.2 Create `internal/provider/resource_group_test.go` with basic CRUD and import acceptance tests
- [ ] 1.3 Register `bitbucketdc_group` in `provider.go`

## 2. User Resource

- [ ] 2.1 Create `internal/provider/resource_user.go` with Create/Read/Update(rename)/Delete/Import (write-only password, sensitive)
- [ ] 2.2 Create `internal/provider/resource_user_test.go` with basic CRUD, rename, and import acceptance tests
- [ ] 2.3 Register `bitbucketdc_user` in `provider.go`

## 3. UserGroup Resource

- [ ] 3.1 Create `internal/provider/resource_user_group.go` with Create/Read/Delete/Import (ID = username/group, RequiresReplace all)
- [ ] 3.2 Create `internal/provider/resource_user_group_test.go` with add/remove and import acceptance tests
- [ ] 3.3 Register `bitbucketdc_user_group` in `provider.go`

## 4. Documentation and Validation

- [ ] 4.1 Add examples to `tests/terraform/main.tf` for all three resources
- [ ] 4.2 Run `make docs` to regenerate provider documentation
- [ ] 4.3 Run `make install` and validate with `terraform plan` (zero-diff)
- [ ] 4.4 Update `CHANGELOG.md` with new resources
