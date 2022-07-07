from opensearchpy import OpenSearch
import json
import sys

def payload_constructor(data,action):
    action_string = json.dumps(action) + "\n"

    payload_string=""
    count=0
    for line in data:
        payload_string += action_string
        this_line = json.dumps(line) + "\n"
        payload_string += this_line
        count+=1
        if count == 100:
          response=client.bulk(body=payload_string,index="ocm-api-data")
          print(response)
          count=0
          payload_string=""
        
    return payload_string



filename= sys.argv[1]



# Create the client with SSL/TLS enabled, but hostname verification disabled.
client = OpenSearch(
    hosts = [{'host':'perf-results-elastic.apps.observability.perfscale.devcluster.openshift.com', 'port':443}],
    http_compress = True, # enables gzip compression for request bodies
    http_auth = ('<username>', '<password>'),
    use_ssl = True,
    verify_certs = False,
    timeout=60
    # max_retries=10, 
    # retry_on_timeout=True
)

# To check if connected to server
if not client.ping():
    raise ValueError("Connection failed")


with open(filename) as f:
    for line in f:
      data = json.loads(line)

# Created index
client.indices.create(index="ocm-api-data", ignore=400)

# Below document is appended to the json file, as this foramt is used for bulk uploading the files to the ES server.
action={
    "index": {
        "_index": "ocm-api-data"
    }
}

# For Bulk Upload
payload= payload_constructor(data, action)

#To check if all the data is uploaded
if not payload:
  print("Successful")
else:
  response=client.bulk(body=payload_constructor(data,action),index="ocm-api-data")


# Check all indexes present on the server
for index in client.indices.get('*'):
  print(index)
