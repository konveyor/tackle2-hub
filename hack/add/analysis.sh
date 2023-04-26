#!/bin/bash

host="${HOST:-localhost:8080}"
app="${1:-1}"
nRuleSet="${2:-10}"
nIssue="${3:-10}"
file="/tmp/analysis.yaml"

echo "Writing report: ${file}"
echo " Application: ${app}"
echo " RuleSets: ${nRuleSet}"
echo " Issues: ${nIssue}"

#
# RuleSets
#
echo -n "---
rulesets:
" > ${file}
for r in $(seq 1 ${nRuleSet})
do 
echo -n "- name: Main${r}
  description: Testing.
  technologies:
  - name: TechA
  - name: TechB
  - name: TechC
  - name: TechS
    source: true
  issues:
" >> ${file}
#
# Issues
#
for i in $(seq 1 ${nIssue})
do
echo -n "  - ruleid: Main${r}.00${i}
    description: This is a test ${r}/${i}.
    category: warning
    effort: 10
    labels:
    - RULESET-${r}
    - RULE-${i}
    incidents:
    - uri: http://thing.com/file:1
      message: Thing happend line:1
      facts:
        factA: 1.A
        factB: 1.B
    - uri: http://thing.com/file:2
      message: Thing happend line:2
      facts:
        factA: 1.C
        factB: 1.D
    - uri: http://thing.com/file:3
      message: Thing happend line:3
      facts:
        factA: 1.E
        factB: 1.F
" >> ${file}
done
done
#
# Deps
#
echo -n "dependencies:
- name: github.com/jboss
  version: 5.0
- name: github.com/hybernate
  indirect: "true"
  version: 4.6 
- name: github.com/ejb
  indirect: "true"
  version: 4.3 
- name: github.com/java
  indirect: "true"
  version: 8
" >> ${file}

curl -i -X POST ${host}/applications/${app}/analyses \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
  --data-binary @$file
