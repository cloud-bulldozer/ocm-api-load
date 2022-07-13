import json
import sys
import argh

filename= sys.argv[1]
savefile= sys.argv[2]
printfile= sys.argv[3]


def has_body(item):
    message=""
    if item["body"]:
        hasbody="True"
    else:
        hasbody= "False"
    return hasbody

def has_error(item):
    message1=""
    if 200 <= (int(item["code"])) <=299 :
        haserror="False"
    else:
        haserror= "True"
    return haserror



# Loading json file
with open(filename) as f:
    data = [json.loads(line) for line in f]


# Adding the required fields
version="version"
version_value="b4d0329"
haserror="has_error"
hasbody="has_body"
uuid="uuid"
uuid_value="c750f388-3eb1-5baa-bfa3-13ec62b02ccd"


for i in data:
    body_value = has_body(i)
    error_val = has_error(i)
    i[version]=version_value
    i[uuid]=uuid_value
    i[hasbody]=body_value
    i[haserror]=error_val
        

# Dumping(saving) data into JSON file
with open(savefile,"w") as f:
    json.dump(data, f)

# Pretty print the loaded file
if printfile=='True':
    print(json.dumps(data, indent=4,sort_keys=True))
