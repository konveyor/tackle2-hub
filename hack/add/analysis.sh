#!/bin/bash

set -e

host="${HOST:-localhost:8080}"
appId="${1:-1}"
nRuleSet="${2:-10}"
nIssue="${3:-10}"
nIncident="${4:-25}"
tmp=/tmp/${self}-${pid}
file="/tmp/manifest.yaml"

echo " Application: ${appId}"
echo " RuleSets: ${nRuleSet}"
echo " Issues: ${nIssue}"
echo " Incidents: ${nIncident}"
echo " Manifest path: ${file}"

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
# Analysis
#
printf "\x1DBEGIN-MAIN\x1D\n" > ${file}
echo -n "---
commit: "1234"
" >> ${file}
printf "\x1DEND-MAIN\x1D\n" >> ${file}
#
# Issues
#
printf "\x1DBEGIN-ISSUES\x1D\n" >> ${file}
for r in $(seq 1 ${nRuleSet})
do
for i in $(seq 1 ${nIssue})
do
echo -n "---
ruleset: ruleSet-${r}
rule: rule-${i}
name: Rule-${r}.${i}-Violated
description: |
  This is a test ${r}/${i} violation.
    This is a **description** of the issue in markdown*.
    Here's how to fix the issue.
    
    For example:
    
        This is some bad code.
    
    Should become:
    
        This is some good code.
category: warning
effort: 10
labels:
- RULE-${i}
- RULESET-${r}
- ${sources[$((${i}%${#sources[@]}))]}
- ${targets[$((${i}%${#targets[@]}))]}
links:
- title: Document A
  url: http://ruleset/${r}/rule/${i}/documentA.html
- title: Document B
  url: http://ruleset/${r}/rule/${i}/documentB.html
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
printf "\x1DEND-ISSUES\x1D
\x1DBEGIN-DEPS\x1D\n" >> ${file}
#
# Deps
#
echo -n "---
name: github.com/jboss
version: 4.0
labels:
- konveyor.io/language=java
- konveyor.io/otherA=dog
" >> ${file}
echo -n "---
name: github.com/jboss
version: 5.0
labels:
- konveyor.io/language=java
- konveyor.io/otherA=cat
" >> ${file}
echo -n "---
name: github.com/hybernate
indirect: "true"
version: 4.6
" >> ${file}
echo -n "---
name: github.com/hybernate
indirect: "true"
version: 5.0
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
echo -n "---
name: github.com/java
version: 8
" >> ${file}
printf "\x1DEND-DEPS\x1D\n" >> ${file}

echo "Manifest (file) GENERATED: ${file}"

#
# Post manifest.
code=$(curl -kSs -o ${tmp} -w "%{http_code}" -F "file=@${file};type=application/x-yaml" http://${host}/files/manifest)
if [ ! $? -eq 0 ]
then
  exit $?
fi
case ${code} in
  201)
    manifestId=$(cat ${tmp}|jq .id)
    echo "manifest (file): ${name} posted. id=${manifestId}"
    ;;
  *)
    echo "manifest (file) post - FAILED: ${code}."
    cat ${tmp}
    exit 1
esac
#
# Post analysis.
d="
id: ${manifestId}
"
code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/applications/${appId}/analyses -H "Content-Type:application/x-yaml" -d "${d}")
if [ ! $? -eq 0 ]
then
  exit $?
fi
case ${code} in
  201)
    id=$(cat ${tmp}|jq .id)
    echo "analysis: ${name} posted. id=${id}"
    ;;
  *)
    echo "analysis post  - FAILED: ${code}."
    cat ${tmp}
    exit 1
esac

