#!/bin/bash

set -e

host="${HOST:-localhost:8080}"
app="${1:-1}"
nRuleSet="${2:-10}"
nIssue="${3:-10}"
nIncident="${4:-25}"
aPath="/tmp/analysis.yaml"
iPath="/tmp/issues.yaml"
dPath="/tmp/deps.yaml"

echo " Application: ${app}"
echo " RuleSets: ${nRuleSet}"
echo " Issues: ${nIssue}"
echo " Incidents: ${nIncident}"
echo " Issues path: ${iPath}"
echo " Deps path: ${dPath}"

sources=(
konveyor.io/source=oraclejdk
konveyor.io/source=oraclejdk
konveyor.io/source=oraclejdk
""
""
""
""
""
""
""
""
""
""
""
""
""
""
""
""
""
)
targets=(
konveyor.io/target=openjdk7
konveyor.io/target=openjdk11+
konveyor.io/target=openjdk17+
konveyor.io/target=cloud-readiness
konveyor.io/target=openliberty
konveyor.io/target=quarkus
konveyor.io/target=jakarta-ee9+
konveyor.io/target=rhr
konveyor.io/target=azure-aks
konveyor.io/target=azure-appservice
konveyor.io/target=azure-container-apps
konveyor.io/target=azure-spring-apps
konveyor.io/target=eap
konveyor.io/target=eap7
konveyor.io/target=eap8
konveyor.io/target=drools
konveyor.io/target=camel
konveyor.io/target=hibernate
konveyor.io/target=jbpm
)

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
- RULE-${i}
- RULESET-${r}
- ${sources[$((${i}%${#sources[@]}))]}
- ${targets[$((${i}%${#targets[@]}))]}
incidents:
" >> ${file}
for n in $(seq 1 ${nIncident})
do
f=$(($n%3))
echo -n "- file: /thing.com/file/${i}${f}.java
  message: |
    This is a **description** of the issue on line ${n} *in markdown*. Here's how to fix the issue.
    
    For example:
    
        This is some bad code.
    
    Should become:
    
        This is some good code.
    
    Some documentation links will go here.
  facts:
    factA: ${i}-${n}.A
    factB: ${i}-${n}.B
  line: 106
" >> ${file}
if ((${n} < 6)); then echo -n "  codesnip: |2
     97  public class SwapNumbers {
     98      public static void main(String[] args) {
     99          float first = 1.20f, second = 2.45f;
    100 
    101          System.out.println(\"--Before swap--\");
    102          System.out.println(\"First number = \" + first);
    103          System.out.println(\"Second number = \" + second);
    104 
    105          // Value of first is assigned to temporary
    106          float temporary = first;
    107 
    108          // Value of second is assigned to first
    109          first = second;
    110 
    111          // Value of temporary assigned to second
    112          second = temporary;
    113 
    114          System.out.println(\"--After swap--\");
    115          System.out.println(\"First number = \" + first);
    116          System.out.println(\"Second number = \" + second);
    117      }
    118  }
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
#
# Analysis
#
file=${aPath}
echo -n "---
issues:
dependencies:
" > ${file}

echo "Report CREATED"

mime="application/x-yaml"

curl \
  -F "file=@${aPath};type=${mime}" \
  -F "issues=@${iPath};type=${mime}" \
  -F "dependencies=@${dPath};type=${mime}" \
  ${host}/applications/${app}/analyses \
  -H "Accept:${mime}"
