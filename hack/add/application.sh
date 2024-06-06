#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/applications \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
'
---
name: Dog
description: Dog application.
businessService: 
  id: 1
repository:
    kind: git
    url: https://github.com/WASdev/sample.daytrader7.git
identities:
  - id: 1
  - id: 2
tags:
  - id: 1
  - id: 16
'

curl -X POST ${host}/applications \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
'
---
name: Cat
description: Cat application.
identities:
  - id: 1
tags:
  - id: 1
  - id: 2
'

curl -X POST ${host}/applications \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
'
---
name: Lion
description: Lion application.
identities:
  - id: 1
tags:
  - id: 3
  - id: 4
'

curl -X POST ${host}/applications \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
'
---
name: Tiger
description: Tiger application.
identities:
  - id: 1
tags:
  - id: 1
  - id: 2
  - id: 3
  - id: 4
'

curl -X POST ${host}/applications \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
'
---
name: Bear
description: Bear application, oh my!.
identities:
  - id: 1
tags:
  - id: 5
  - id: 6
'

