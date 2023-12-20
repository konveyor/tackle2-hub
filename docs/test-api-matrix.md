# Test API matrix

Endpoint | binding functions | basic CRUD test | deeper test| notes/component/status
-- | -- | -- | -- | --
**Application Inventory**||||
application|:white_check_mark: partially|:heavy_check_mark:||
bucket|:heavy_check_mark:|:heavy_check_mark:||partially within application
dependency|:heavy_check_mark:|:heavy_check_mark:|:heavy_check_mark:|
file|:heavy_check_mark:|:heavy_check_mark:||
import||:heavy_check_mark:||
review|:heavy_check_mark:|:heavy_check_mark:||
**Controls**||||
businessservice|:heavy_check_mark:|:heavy_check_mark:||
group|:heavy_check_mark:|:heavy_check_mark:||aka StakeholderGroup
identity|:heavy_check_mark:|:heavy_check_mark:||
jobfunction|:heavy_check_mark:|:heavy_check_mark:||
proxy|:heavy_check_mark:|:heavy_check_mark:||
stakeholder|:heavy_check_mark:|:heavy_check_mark:||
tag|:heavy_check_mark:|:heavy_check_mark:||
tagcategory|:heavy_check_mark:|:heavy_check_mark:||
**Dynamic Reports**||||
adoptionplan||||
analysis||||
ruleset|:heavy_check_mark:|:heavy_check_mark:||
**Migrationwaves and Jira**||||
batch||||
migrationwave|:heavy_check_mark:|:heavy_check_mark:||
ticket|:heavy_check_mark:|:heavy_check_mark:||
tracker|:heavy_check_mark:|:heavy_check_mark:||
**Assessments**||||
archetype|:heavy_check_mark:|:heavy_check_mark:||
assessment|:heavy_check_mark:|:heavy_check_mark:||
questionnaire|:heavy_check_mark:|:heavy_check_mark:||
**Other**||||
addon || | |
auth||||used in tests
cache||||
schema||:heavy_check_mark:||
setting|:heavy_check_mark:|:heavy_check_mark:||
task|:heavy_check_mark:|:heavy_check_mark:||
taskgroup||||

API tests are organized in https://github.com/konveyor/tackle2-hub/tree/main/test/api. One package/directory per endpoint.

This should be updated in PRs with API tests to keep track on API test coverage.
