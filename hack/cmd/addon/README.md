Test addon.

The addon will list files in /etc and create artifacts for each file.

To run locally:

Be logged into a kubernetes/openshift cluster.

Environment variables: 
- **ADDON_SECRET_PATH** - The addon secret path. Recommend: `hack/cmd/addon/hub.json`
- **HUB_BASE_URL** - The hub API base URL. Default: `localhost:8080`.
