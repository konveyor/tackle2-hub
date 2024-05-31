#!/bin/bash

pid=$$
self=$(basename $0)
tmp=/tmp/${self}-${pid}

usage() {
  echo "Usage: ${self}"
  echo "  -u konveyor URL"
  echo "  -h help"
}

while getopts "u:h" arg; do
  case $arg in
    h)
      usage
      exit 1
      ;;
    u)
      host=$OPTARG
      ;;
  esac
done

if [ -z "${host}"  ]
then
  echo "-u required."
  usage
  exit 0
fi

code=$(curl -kSs -o ${tmp} -w "%{http_code}" ${host}/tasks)
if [ ! $? -eq 0 ]
then
  exit $?
fi
case ${code} in
  200)
    echo ${tmp}
    echo "ID  | Kind      | State         | Pty | Application"
    echo "--- | ----------|---------------|-----|---------------"
    readarray report <<< $(jq -c '.[]|"\(.id) \(.kind) \(.state) \(.priority) \(.application.id) \(.application.name)"' ${tmp})
    for r in "${report[@]}"
    do
      r=${r//\"/}
      t=($r)
      id=${t[0]}
      kind=${t[1]}
      state=${t[2]}
      pty=${t[3]}
      appId=${t[4]}
      appName=${t[5]}
      if [ "${pty}" = "null" ]
      then
        pty=0
      fi
      printf "%-6s%-12s%-16s%-4s%4s|%-10s\n" ${id} ${kind} ${state} ${pty} ${appId} ${appName}
    done
    ;;
  *)
    echo "FAILED: ${code}."
    cat ${tmp}
    exit 1
esac

rm -f ${tmp}
