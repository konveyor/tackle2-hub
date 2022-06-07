# Tackle CLI tool

A tool for Konveyor Tackle application maintenance written in Python.

For more details about the Tackle project, see [Tackle2-Hub README](https://github.com/konveyor/tackle2-hub) or https://github.com/konveyor/tackle-documentation.

## Usage

Use ```tackle-config.yml.example``` file as a template to set your Tackle endpoints and credentials and save it as ```tackle-config.yml``` before running the ```tackle``` command.

### Supported actions
- ```export-tackle1``` exports Tackle 1.2 API objects into local JSON files
- ```import``` creates objects in Tackle 2 from local JSON files
- ```clean``` deletes objects uploaded to Tackle 2 from local JSON files
- ```clean-all``` deletes ALL data from Tackle 2 (including seeds)

### Scenarios

#### Migrate data from running Tackle 1.2 to running Tackle 2 instance

With tags, tag-types and job functions automatic re-mapping.

- ```tackle export-tackle1```
- ```tackle import```

#### Export data from Tackle 1.2 to be later imported to some Tackle 2 instance

With full export and full cleanup of the Tackle 2 before running the import.

- ```tackle --skip-destination-check export-tackle1``` (ensures that all seeds objects are exported too)
- ```tackle clean-all```
- ```tackle import```

### Export Tackle 1.2

Run ```tackle export-tackle1``` to get dump of Tackle 1.2 objects into JSON files in local directory ```./tackle-data```.

The ```export-tackle1``` command looks into Tackle2 to grab seed resources first, then downloads all resources from Tackle 1.2 API, transforms it to format expected by the Tackle 2 Hub and re-map resources to seeds already existing in destination Tackle2 Hub API. The result is serialized and stored into local JSON files.

### Import to Tackle 2 Hub

Check local JSON dump files in ```./tackle-data``` directory (if needed) and create objects in Tackle 2 Hub running ```tackle import```.

The import command connects to Tackle2 Hub, check existing objects for possible collisions (by IDs) and uploads resources from local JSON files.

### Delete uploaded objects
To delete objects previously created by the ```import``` command, run ```tackle clean```. This can address also existing Tackle 2 objects which are in collision with local JSON dump files.

### Command line options

Config file ```-c / --config``` path specifies a YAML file with configuration options including endpoints and credentials for Tackle APIs (```tackle-config.yml``` by default).

Verbose output ```-v / --verbose``` option logs all API requests and responses providing more information for possible debugging (```off``` by default).

Data directory ```-d / --data-dir``` specifies path to local directory with Tackle database records in JSON format (```./tackle-data``` by default).

SSL warnings ```-w / --disable-ssl-warnings``` optional suppress ssl warning for api requests.

Import errors could be skipped with ``` -i / --ignore-import-errors ``` -  not recommended - use with high attention to avoid data inconsistency. If the import has failed, it is recommended use ```tackle clean``` command to remove only imported resources.

## Example

```
$ tackle --help
usage: tackle [-h] [-c [CONFIG]] [-d [DATA_DIR]] [-v] [-s] [action ...]

Konveyor Tackle maintenance tool.

positional arguments:
  action                One or more Tackle commands that should be executed, options: export-tackle1 import clean

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

Unverified HTTPS warnings from Python could be hidden by ```export PYTHONWARNINGS="ignore:Unverified HTTPS request"```.
