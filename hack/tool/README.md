# Tackle CLI tool

A tool for Konveyor Tackle application maintenance written in Python.

For more details about the Tackle project, see [Tackle2-Hub README](https://github.com/konveyor/tackle2-hub) or https://github.com/konveyor/tackle-documentation.

## Scenarios

### Migrate data from running Tackle 1.2 to running Tackle 2 instance

Migrate data updating refs to the seed objects matching to the destination Tackle 2 (tags, tag-types and job functions seeded).

- ```tackle export-tackle1```
- ```tackle import```

### Export data from Tackle 1.2 to be later imported to some Tackle 2 instance

Exports full data dump including seeds to be later imported to currently not available Tackle 2 instance, which needs a cleanup before running the import.

- ```tackle --skip-destination-check export-tackle1```
- ```tackle clean-all```
- ```tackle import```

### If the import failed

The ```tackle import``` command could fail in a pre-import check phase which ensures there are no resources of given type with the same ID in the destination Tackle 2 (error after ```Checking tagtypes in destination Tackle2..``` etc.). In this case, run ```tackle clean``` command, which will remove such objects from the destination Tackle 2 API or remove it manually either from the destination Tackle 2 or from the JSON data files.

If the import failed in the upload phase (error after ```Uploading tagtypes..``` etc.), try  ```tackle clean``` to remove already imported objects followed by  ```tackle clean-all``` which lists all resources of all known data types in the destination Tackle 2 API and deletes it (without looking to local data files).

Note on ```clean-all``` command, it deletes all resources from Tackle 2 Hub API, however Pathfinder API doesn't support listing assessments without providing an applicationID. The applications could not be present in Hub, so an "orphaned" assessments could stay in Pathfinder. In order to resolve potential collision with imported data, run  ```tackle clean``` together with the ```clean-all``` command.

Caution: all clean actions might delete objects already present in the Tackle 2 and unrelated to the import data.

## Requirements

The tool requires Python3 to be installed, PyYAML module and git to get the source code, install it with  ```dnf install python39 python3-pyyaml git``` (for Red Hat-like Linux distros).

Since git and python3 should be present on most systems, it might be enought just install PyYAML with Python PIP tool ```python3 -m pip install pyyaml``` without need to use your OS package manager.

## Usage

Clone Github repository:
```git clone https://github.com/konveyor/tackle2-hub.git```

Change to the tool directory:
```cd hack/tool```

Use ```tackle-config.yml.example``` file as a template to set your Tackle endpoints and credentials and save it as ```tackle-config.yml```.

Run the tackle tool:
```./tackle```

### Supported actions
- ```export-tackle1``` exports Tackle 1.2 API objects into local JSON files
- ```import``` creates objects in Tackle 2 from local JSON files
- ```clean``` deletes objects uploaded to Tackle 2 from local JSON files
- ```clean-all``` deletes ALL data returned by Tackle 2 (including seeds, additional to ```clean```), skips some pathfinder stuff without index action

### Export Tackle 1.2

Run ```tackle export-tackle1``` to get dump of Tackle 1.2 objects into JSON files in local directory ```./tackle-data```.

The ```export-tackle1``` command looks into Tackle2 to grab seed resources first, then downloads all resources from Tackle 1.2 API, transforms it to format expected by the Tackle 2 Hub and re-map resources to seeds already existing in destination Tackle2 Hub API. The result is serialized and stored into local JSON files.

### Import to Tackle 2 Hub

Check local JSON dump files in ```./tackle-data``` directory (if needed) and create objects in Tackle 2 Hub running ```tackle import```.

The import command connects to Tackle2 Hub, check existing objects for possible collisions (by IDs) and uploads resources from local JSON files.

### Delete uploaded objects

To delete objects previously created by the ```import``` command, run ```tackle clean```. This can address also existing Tackle 2 objects which are in collision with local JSON dump files.

### Delete all objects

The Tackle2 instance could be cleaned-up with ```tackle clean``` command. It lists objects from all data types and deletes such resources.

There is a exception with Pathfinder API which doesn't support listing assessments without knowledge of applicationIDs, so this might stay in Pathfinder database.

### Command line options

Config file ```-c / --config``` path specifies a YAML file with configuration options including endpoints and credentials for Tackle APIs (```tackle-config.yml``` by default).

Verbose output ```-v / --verbose``` option logs all API requests and responses providing more information for possible debugging (```off``` by default).

Data directory ```-d / --data-dir``` specifies path to local directory with Tackle database records in JSON format (```./tackle-data``` by default).

A full export without having access to the destination Tackle 2 and including all seed objects can be executed with ```-s / --skip-destination-check``` option. When importing such data, the destination Tackle 2 needs to be empty (run ```clean-all``` first).

SSL warnings ```-w / --disable-ssl-warnings``` optional suppress ssl warning for api requests.

Import errors could be skipped with ``` -i / --ignore-import-errors ``` -  not recommended - use with high attention to avoid data inconsistency. If the import has failed, it is recommended use ```tackle clean``` command to remove only imported resources.

## Example

```
$ tackle --help
usage: tackle [-h] [-c [CONFIG]] [-d [DATA_DIR]] [-v] [-s] [action ...]

Konveyor Tackle maintenance tool.

positional arguments:
  action                One or more Tackle commands that should be executed, options: export-tackle1 import clean clean-all

options:
  -h, --help            show this help message and exit
  -c [CONFIG], --config [CONFIG]
                        A config file path (tackle-config.yml by default).
  -d [DATA_DIR], --data-dir [DATA_DIR]
                        Local Tackle data directory path (tackle-data by default).
  -v, --verbose         Print verbose output (including all API requests).
  -s, --skip-destination-check
                        Skip connection and data check of Tackle 2 destination.
  -w, --disable-ssl-warnings
                        Do not display warnings during ssl check for api requests.
  -i, --ignore-import-errors
                        Skip to next item if an item fails load.


```

API endpoints and credentials should be set in a config file (```tackle-config.yml``` by default).

```
---
# Main Tackle 2 endpoint and credentials
url: https://tackle-konveyor-tackle.apps.cluster.local
username: admin
password:

# Tackle 1.2 endpoint and credentials, e.g. to dump data and migrate to Tackle2
tackle1:
  url: https://tackle-tackle.apps.mta01.cluster.local
  username: tackle
  password:

```

Unverified HTTPS warnings from Python could be hidden by ```export PYTHONWARNINGS="ignore:Unverified HTTPS request"``` or with ```-w``` command option.
