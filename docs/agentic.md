## Agentic ##
The agentic API is a thin REST facade over the custom resources owned by the
[agentic-controller](https://github.com/konveyor/agentic-controller): Agent, AgentRun,
AgentWorkflow, AgentWorkflowRun, SkillCard, SkillCollection and Gateway.  The hub does not
store these resources; each route reads and writes the CR in the hub namespace. As a result,
a resource may equally be created through the hub API or applied directly with (for example)
`kubectl apply`, a kustomize overlay or an operator.

### The konveyor.io/managed label ###
`POST /agentic/agents` and `POST /agentic/workflows` label the resource they create:

```yaml
metadata:
  labels:
    konveyor.io/managed: "true"
```

`GET /agentic/agents` and `GET /agentic/workflows` select on that label. **An Agent or
AgentWorkflow created without it is not listed.** This is the behavior to know about when
authoring these resources by hand: the CR applies cleanly, the controller reconciles it, and
it is still absent from the hub's lists - and so from the catalog and the _Create run_ dialogs
drawn from them. Nothing reports this, because nothing has failed.

Two ways to resolve it:

- Label the resource, either in the manifest or after the fact:

  ```bash
  kubectl label agent my-agent konveyor.io/managed=true
  ```

- Or ask for an unfiltered list with the `all` parameter:

  ```bash
  curl ${HUB}/agentic/agents?all=true
  ```

  `all=true` lists every resource of that kind in the namespace, labelled or not. Omitted or
  `all=false` keeps the default, label-filtered list. A value that is not a boolean is
  rejected with 400 rather than quietly ignored.

The other kinds - SkillCard, SkillCollection, Gateway, AgentRun and AgentWorkflowRun - are
neither labelled when created through the API nor filtered when listed. The whole namespace is
listed, so hand-authored resources of those kinds appear without anything further. The `all`
parameter is accepted only by the two list routes that filter.

### Addressing a resource by name ###
The single-resource routes do not filter on the label:

```bash
curl ${HUB}/agentic/agents/my-agent
```

returns an unlabelled Agent, and `PUT` and `DELETE` on it behave the same way. The label
records how a resource was created, not whether it may be addressed. So an unlabelled Agent
is reachable by name while absent from the list - fetching it successfully is not evidence
that it will appear in the catalog.
