#!/bin/bash

set -e

host="${HOST:-localhost:8080}"
app="${1:-1}"
nRuleSet="${2:-10}"
nIssue="${3:-10}"
nIncident="${4:-25}"
iPath="/tmp/issues.yaml"
dPath="/tmp/deps.yaml"

echo " Application: ${app}"
echo " RuleSets: ${nRuleSet}"
echo " Issues: ${nIssue}"
echo " Incidents: ${nIncident}"
echo " Issues path: ${iPath}"
echo " Deps path: ${dPath}"

#
# Issues
#
file=${iPath}
echo "" > ${file}
for r in $(seq 1 ${nRuleSet})
do
for i in $(seq 1 ${nIssue})
do
echo -n "---
ruleset: ruleSet-${r}
rule: rule-${i}
name: Rule-${r}.${i}-Violated
description: This is a test ${r}/${i} violation.
category: warning
effort: 10
labels:
- konveyor.io/target=RULESET-${r}
- konveyor.io/source=RULE-${i}
incidents:
" >> ${file}
for n in $(seq 1 ${nIncident})
do
f=$(($n%3))
echo -n "- file: /thing.com/file/${i}${f}
  line: ${n}
  message: Thing happend line:${n}
  facts:
    factA: ${i}-${n}.A
    factB: ${i}-${n}.B
" >> ${file}
if ((${n} < 6)); then echo -n "  codesnip: |
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
file=${dPath}
echo -n "---
name: github.com/jboss
version: 5.0
" > ${file}
echo -n "---
name: github.com/hybernate
indirect: "true"
version: 4.6
" >> ${file}
echo -n "---
name: github.com/ejb
indirect: "true"
version: 4.3
" >> ${file}
echo -n "---
name: github.com/java
indirect: "true"
version: 8
" >> ${file}

echo "Report CREATED"

curl \
  -F "issues=@${iPath}" \
  -F "dependencies=@${dPath}" \
  ${host}/applications/${app}/analyses \
  -H 'Accept:application/x-yaml'
