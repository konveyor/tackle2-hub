# tackle-tool-12-to-20

A tool migrating Konveyor Tackle 1.2 data into Tackle 2.0 written as a Python script. For more details about the Tackle project, see [Tackle2-Hub README](https://github.com/konveyor/tackle2-hub) or https://github.com/konveyor/tackle-documentation.

## Usage

Use ```config-vars.example``` file as a template to set your Tackle 1.2 and 2 endpoints and credentials before running the data migration.

### Dump Tackle 1.2

Run ```. config-vars && python tackle-mig-1220.py dump``` to get dump of Tackle 1.2 objects into JSON files in local directory ```mig-data```.

The dump command looks into Tackle2 to grab seed resources first, then downloads all resources from Tackle 1.2 API, transforms it to format expected by the Tackle 2 Hub and re-map resources to seeds already existing in destination Tackle2 Hub API. The result is serialized and stored into local JSON files.

### Upload to Tackle 2 Hub

Check local JSON dump files in ```mig-data``` directory (if needed) and upload it to Tackle 2 Hub running ```. config-vars && python tackle-mig-1220.py upload```.

The upload command connects to Tackle2 Hub, check existing objects for possible collisions (by IDs) and uploads resources from local JSON files.

### Supported actions
- ```dump``` exports Tackle 1.2 API objects into local JSON files
- ```upload``` creates objects in Tackle 2 from local JSON files
- ```clean``` deletes objects uploaded to Tackle 2 from local JSON files

There is a ```-d/--debug``` option which logs all API requests and responses providing more information for possible debugging.

## Example

```
$ python tackle-mig-1220.py --help
usage: tackle-mig-1220.py [-h] [steps ...]

Migrate data from Tackle 1.2 to Tackle 2.

positional arguments:
  steps       One or more steps of migration that should be executed (dump and upload by default), options: dump upload clean

options:
  -h, --help  show this help message and exit
  -d, --debug  Print debug output including all API requests
```

API endpoints and credentials should be set in the environment or source an env file.

```
export TACKLE1_URL=https://tackle-tackle.apps.mta01.cluster.local
export TACKLE1_USERNAME=tackle
export TACKLE1_PASSWORD=...
export TACKLE2_URL=https://tackle-konveyor-tackle.apps.cluster.local
export TACKLE2_USERNAME=admin
export TACKLE2_PASSWORD=...
```

### Sample command output

```
$ . config-vars && python tackle-mig-1220.py dump
Starting Tackle 1.2 -> 2 data migration tool
Dumping Tackle1.2 objects..
Writing JSON data files into ./mig-data..
Done.
```

Sample migration JSON dump files are available in [mig-data directory](mig-data).

Unverified HTTPS warnings from Python could be hidden by ```export PYTHONWARNINGS="ignore:Unverified HTTPS request"```.
