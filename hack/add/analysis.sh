#!/bin/bash

host="${HOST:-localhost:8080}"
app="${1:-1}"
nRuleSet="${2:-10}"
nIssue="${3:-10}"
nIncident="${4:-25}"
file="${HOME}/analysis.yaml"

echo "Writing report: ${file}"
echo " Application: ${app}"
echo " RuleSets: ${nRuleSet}"
echo " Issues: ${nIssue}"
echo " Incidents: ${nIncident}"

#
# Issues
#
echo -n "---
issues:
" > ${file}
for r in $(seq 1 ${nRuleSet})
do
for i in $(seq 1 ${nIssue})
do
echo -n "- ruleset: ruleSet-${r}
  rule: rule-${i}
  name: Rule-${i}-Violated
  description: This is a test ${r}/${i}.
  category: warning
  effort: 10
  labels:
  - konveyor.io/target=RULESET-${r}
  - konveyor.io/source=RULE-${i}
  incidents:
" >> ${file}
for n in $(seq 1 ${nIncident})
do
echo -n "  - uri: http://thing.com/file/${i}:${n}
    message: Thing happend line:${n}
    facts:
      factA: ${i}-${n}.A
      factB: ${i}-${n}.B
" >> ${file}
if ((${n} < 6)); then echo -n "    codesnip: |
      public class SwapNumbers {
          public static void main(String[] args) {
              float first = 1.20f, second = 2.45f;
      
              System.out.println(\"--Before swap--\");
              System.out.println(\"First number = \" + first);
              System.out.println(\"Second number = \" + second);
      
              // Value of first is assigned to temporary
              float temporary = first;
      
              // Value of second is assigned to first
              first = second;
      
              // Value of temporary (which contains the initial value of first) is assigned to second
              second = temporary;
      
              System.out.println(\"--After swap--\");
              System.out.println(\"First number = \" + first);
              System.out.println(\"Second number = \" + second);
          }
      }
" >> ${file}
fi
done
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

echo "Report CREATED"

curl -i -X POST ${host}/applications/${app}/analyses \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
  --data-binary @$file
