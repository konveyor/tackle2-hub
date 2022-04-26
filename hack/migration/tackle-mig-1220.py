import argparse
import copy
import json
import os
import requests

###############################################################################

parser = argparse.ArgumentParser(description='Migrate data from Tackle 1.2 to Tackle 2.')
parser.add_argument('steps', type=str, nargs='*',
                    help='One or more steps of migration that should be executed (dump and upload by default), options: dump  upload  clean')
parser.add_argument('-d','--debug', dest='debug', action='store_const', const=True, default=False,
                    help='Print debug output including all API requests')
args = parser.parse_args()

###############################################################################

def ensureDataDir(dataDir):
    if os.path.isdir(dataDir):
        debugPrint("Data directory already exists, using %s" % dataDir)
    else:
      debugPrint("Creating data directories at %s" % dataDir)
      os.mkdir(dataDir)

def checkConfig(expected_vars):
    for varKey in expected_vars:
        if os.environ.get(varKey) is None:
            print("ERROR: Missing required environment variable %s, define it first." % varKey)
            exit(1)

def debugPrint(str):
    if args.debug:
        print(str)

def getKeycloakToken(host, username, password, client_id='tackle-ui', realm='tackle'):
    url  = "%s/auth/realms/%s/protocol/openid-connect/token" % (host, realm)
    data = {'username': username, 'password': password, 'client_id': client_id, 'grant_type': 'password'}

    r = requests.post(url, data=data, verify=False)
    if r.ok:
        respData = json.loads(r.text)
        return respData['access_token']
    else:
        print("ERROR getting auth token from %s" % url)
        print(data, r)
        exit(1)

def apiJSON(url, token, data=None, method='GET', ignoreErrors=False):
    debugPrint("Querying: %s" % url)
    match method:
        case 'DELETE':
            r = requests.delete(url, headers={"Authorization": "Bearer %s" % token, "Content-Type": "text/json"}, verify=False)
        case 'POST':
            r = requests.post(url, data=json.dumps(data), headers={"Authorization": "Bearer %s" % token, "Content-Type": "text/json"}, verify=False)
        case _: # GET
            r = requests.get(url, headers={"Authorization": "Bearer %s" % token, "Content-Type": "text/json"}, verify=False)  # add pagination?

    if not r.ok:
        if ignoreErrors:
            debugPrint("Got status %d for %s, ignoring" % (r.status_code, url))
        else:
            if data:
                print("ERROR: POST data: %s" % data)
            print("ERROR: API request failed with status %d for %s" % (r.status_code, url))
            exit(1)

    if r.text is None or r.text ==  '':
        return

    debugPrint("Response: %s" % r.text)

    respData = json.loads(r.text)
    if '_embedded' in respData:
        debugPrint("Unwrapping Tackle1 JSON")
        return respData['_embedded'][url.rsplit('/')[-1]] # unwrap Tackle1 JSON response (e.g. _embedded -> application -> [{...}])
    else:
        return respData # raw return JSON (Tackle2)

def loadDump(path):
    data = open(path)
    return json.load(data)

def saveJSON(path, jsonData):
    dumpFile = open(path + ".json", 'w')
    dumpFile.write(json.dumps(jsonData, indent=4, default=vars))
    dumpFile.close()

def cmdWanted(args, step):
    if step in args.steps:
        return True
    else:
        return False

###############################################################################

class Tackle12Import:
    # TYPES order matters for import/upload to Tackle2
    TYPES = ['applications', 'proxies', 'dependencies', 'reviews', 'identities', 'jobfunctions', 'stakeholdergroups', 'stakeholders', 'businessservices', 'tags', 'tagtypes']  # buckets
    TACKLE2_SEED_TYPES = ['tags', 'tagtypes', 'jobfunctions']

    def __init__(self, dataDir, tackle1Url, tackle1Token, tackle2Url, tackle2Token):
        self.dataDir      = dataDir
        self.tackle1Url   = tackle1Url
        self.tackle1Token = tackle1Token
        self.tackle2Url   = tackle2Url
        self.tackle2Token = tackle2Token
        # Dump data
        self.data         = dict()
        for t in self.TYPES:
            self.data[t] = []
        # Existing resources in destination
        self.destData        = dict()
        for t in self.TYPES:
            self.destData[t] = dict()

    # Gather existing seeded objects from Tackle2
    def loadTackle2Seeds(self):
        print("Checking Tackle 2 for seed objects..")

        # Tackle 2 TagTypes and Tags
        collection = apiJSON(tackle12import.tackle2Url + "/hub/tagtypes", tackle12import.tackle2Token)
        for tt2 in collection:
            tt  = Tackle2Object(tt2)
            tt.name = tt2['name']
            self.destData['tagtypes'][tt.name] = tt
            if tt2['tags']:
                for t2 in tt2['tags']:
                    tag             = Tackle2Object()
                    tag.id          = t2['id']
                    tag.name        = t2['name']
                    self.destData['tags'][tag.name] = tag

        # Tackle 2 JobFunctions
        collection = apiJSON(tackle12import.tackle2Url + "/hub/jobfunctions", tackle12import.tackle2Token)
        for jf2 in collection:
            jf              = Tackle2Object(jf2)
            jf.name         = jf2['name']
            self.destData['jobfunctions'][jf.name] = jf

    # Gather Tackle 1.2 API objects and map seeded Tackle2 API objects
    def dumpTackle1(self):
        # Iterate Tackle 1.2 objects
        print("Dumping Tackle 1.2 API objects")

        ### APPLICATION ###
        collection = apiJSON(self.tackle1Url + "/api/application-inventory/application", self.tackle1Token)
        for app1 in collection:
            # Temp holder for tags
            tags = []
            # Prepare Tags
            debugPrint(app1)
            if app1['tags']:
                for tag1 in app1['tags']:
                    # Check if Tag exists in Tackle2 destination
                    if tag1['name'] in self.destData['tags']:
                        # Re-map to existing Tackle2 Tag
                        tags.append(self.destData['tags'][tag1['name']])  # deepcopy?
                    else:
                        # Prepare new Tag
                        tag             = Tackle2Object(tag1)
                        tag.name        = tag1['name']
                        self.add('tags', tag)
                        tags.append(tag)
            # Prepare Application
            app                 = Tackle2Object(app1)
            app.name            = app1['name']
            app.description     = app1['description']
            app.businessService = app1['businessService']
            app.repository      = app1['repository']
            app.binary          = app1['binary']
            app.facts           = app1['facts']
            app.review          = app1['review']
            app.tags            = tags
            self.add('applications', app)

        ### PROXIES ###
        collection = apiJSON(self.tackle1Url + "/api/", self.tackle1Token)
        for proxy1 in collection:
            # Prepare Proxy
            proxy                 = Tackle2Object(proxy1)
            proxy.name            = proxy1['name']
            self.add('proxies', proxy)

        ### DEPENDENCIES ###
        collection = apiJSON(self.tackle1Url + "/api/", self.tackle1Token)
        for dep1 in collection:
            # Prepare Dependency
            dep                 = Tackle2Object(dep1)
            dep.name            = rev1['name']
            self.add('dependencies', dep)

        ### REVIEWS ###
        collection = apiJSON(self.tackle1Url + "/api/application-inventory/application", self.tackle1Token)
        for rev1 in collection:
            # Prepare Review
            rev                 = Tackle2Object(rev1)
            app.name            = rev1['name']
            self.add('reviews', rev)

        ### IDENTITIES ###
        collection = apiJSON(self.tackle1Url + "/api/", self.tackle1Token)
        for id1 in collection:
            # Prepare Review
            id                 = Tackle2Object(id1)
            id.name            = id1['name']
            self.add('identities', id)

        ### STAKEHOLDER ###
        collection = apiJSON(self.tackle1Url + "/api/controls/stakeholder", self.tackle1Token)
        for sh1 in collection:
            # Temp holder for stakeholder's groups
            shgs = []
            # Prepare StakeholderGroups
            for shg1 in sh1['stakeholderGroups']:
                shg             = Tackle2Object(shg)
                shg.name        = shg1['name']
                shg.description = shg1['description']
                self.add('stakeholdergroups', shg)
                shgs.append(shg)
            # Prepare StakeHolder
            sh            = Tackle2Object(sh1)
            sh.name       = sh1['displayName']
            sh.email      = sh1['email']
            sh.groups     = shgs
            if sh1['jobFunction']:
                if sh1['jobFunction']['name'] in self.destData['jobfunctions']:
                    # Re-map to JobFunction existing in Tackle2 destination
                    sh.jobFunction = self.destData['jobfunctions'][sh['jobFunction']['name']]
                else:
                    # Prepare new JobFunction
                    jf              = Tackle2Object(sh['jobFunction'])
                    jf.name         = sh['jobFunction']['role']
                    self.add('jobfunctions', jf)
                    sh.jobFunction = jf
            self.add('stakeholders', sh)
        
        ### STAKEHOLDER GROUPS ###
        collection = apiJSON(self.tackle1Url + "/api/controls/stakeholder-group", self.tackle1Token)
        for shg1 in collection:
            # Prepare StakeholderGroup
            shg             = Tackle2Object(shg1)
            shg.name        = shg1['name']
            shg.description = shg1['description']
            self.add('stakeholdergroups', shg)

        ### JOB FUNCTION ###
        collection = apiJSON(self.tackle1Url + "/api/controls/job-function", self.tackle1Token)
        for jf1 in collection:
            # Temp holder for stakeholders
            shs = []
            # Prepare JobFunction's Stakeholders
            for sh1 in jf1['stakeholders']:
                sh             = Tackle2Object(sh1)
                sh.name        = sh1['displayName']
                sh.email       = sh1['email']
                shs.append(sh)
            # Prepare JobFunction
            jf              = Tackle2Object(jf1)
            jf.name         = jf1['role']
            jf.stakeholders = shs
            # Store only if doesn't exist in Tackle2 destination already
            if jf.name not in self.destData['jobfunctions']:
                self.add('jobfunctions', jf)

        ### BUSINESS SERVICE ###
        collection = apiJSON(self.tackle1Url + "/api/controls/business-service", self.tackle1Token)
        for bs1 in collection:
            # Prepare JobFunction
            bs              = Tackle2Object(bs1)
            bs.name         = bs1['name']
            bs.description  = bs1['description']
            bs.owner        = bs1['owner']  # Stakeholder
            self.add('businessservices', bs)

        ### TAG TYPES & TAGS ###
        collection = apiJSON(self.tackle1Url + "/api/controls/tag-type", self.tackle1Token)
        for tt1 in collection:
            # Temp holder for tags
            tags = []
            # Prepare TagTypes's Tags
            for tag1 in tt1['tags']:
                tag             = Tackle2Object(tag1)
                tag.name        = tag1['name']
                # TagType is injected from tagType processing few lines below
                # Store Tag only if doesn't exist in Tackle2 destination already
                if tag.name not in self.destData['tags']:
                    self.add('tags', tag)
                tags.append(tag)
            # Prepare TagType
            tt            = Tackle2Object(tt1)
            tt.name       = tt1['name']
            tt.colour     = tt1['colour']
            tt.rank       = tt1['rank']
            tt.username   = tt1['createUser'] # Is there another relevant user?
            for tag in tags:
                tag.tagType = copy.deepcopy(tt) # Is this doule-nesting needed?
            tt.tags = tags
            # Store only if doesn't exist in Tackle2 destination already
            if tt.name not in self.destData['tagtypes']:
                self.add('tagtypes', tt)

    def add(self, type, item):
        for existingItem in self.data[type]:
            if item.id == existingItem.id:
                # The item is already present, skipping
                return
        self.data[type].append(item)

    def store(self):
        ensureDataDir(self.dataDir)
        for t in self.TYPES:
            saveJSON(os.path.join(self.dataDir, t), self.data[t])

    def uploadTackle2(self):
        for t in self.TYPES:
            print("Uploading %s.." % t)
            dictCollection = loadDump(os.path.join(self.dataDir, t + '.json'))
            for dictObj in dictCollection:
                debugPrint(dictObj)
                apiJSON(self.tackle2Url + "/hub/" + t, self.tackle2Token, dictObj, method='POST')

    def preImportCheck(self):
        for t in self.TYPES:
            print("Checking %s in destination Tackle2.." % t)
            destCollection = apiJSON(self.tackle2Url + "/hub/" + t, self.tackle2Token)
            localCollection = loadDump(os.path.join(self.dataDir, t + '.json'))
            for importObj in localCollection:
                for destObj in destCollection:
                    if importObj['id'] == destObj['id']:
                        print("ERROR: Resource %s/%d \"%s\" already exists in Tackle2 destination as \"%s\". Clean it before running import." % (t, importObj['id'], importObj['name'], destObj['name']))
                        exit(1)

    def cleanTackle2(self):
        self.TYPES.reverse()
        for t in self.TYPES:
            dictCollection = loadDump(os.path.join(self.dataDir, t + '.json'))
            for dictObj in dictCollection:
                print("Deleting %s/%s" % (t, dictObj['id']))
                apiJSON("%s/hub/%s/%d" % (self.tackle2Url, t, dictObj['id']), self.tackle2Token, method='DELETE')

class Tackle2Object:
    def __init__(self, initAttrs = {}):
        if initAttrs:
            self.id         = initAttrs['id']
            self.createUser = initAttrs['createUser']
            self.updateUser = initAttrs['updateUser']

###############################################################################

dataDir = "./mig-data"

print("Tackle 1.2 -> 2 data migration tool")

# Gather Keycloak access tokens for Tackle1&2
token1 = getKeycloakToken(os.environ.get('TACKLE1_URL'), os.environ.get('TACKLE1_USERNAME'), os.environ.get('TACKLE1_PASSWORD'))
token2 = getKeycloakToken(os.environ.get('TACKLE2_URL'), os.environ.get('TACKLE2_USERNAME'), os.environ.get('TACKLE2_PASSWORD'))

# Tackle 2.0 objects to be imported
tackle12import = Tackle12Import(dataDir, os.environ.get('TACKLE1_URL'), token1, os.environ.get('TACKLE2_URL'), token2)

# Dump steps
if cmdWanted(args, "dump"):
    print("Dump Tackle objects..")
    tackle12import.loadTackle2Seeds()
    tackle12import.dumpTackle1()
    print("Writing JSON data files into %s.." % dataDir)
    tackle12import.store()

# Upload steps
if cmdWanted(args, "upload"):
    print("Uploading data to Tackle2..")
    tackle12import.preImportCheck()
    tackle12import.uploadTackle2()

# Clean uploaded objects
if cmdWanted(args, "clean"):
    print("Cleaning data uploaded to Tackle2..")
    tackle12import.cleanTackle2()

###############################################################################
