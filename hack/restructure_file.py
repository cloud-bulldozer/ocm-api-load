import json
import sys

filename= sys.argv[1]
savefile= sys.argv[2]
printfile= sys.argv[3]


def has_body(list1):
    message=""
    if list1["body"]:
        message="True"
    else:
        message= "False"
    return message

def has_error(list2):
    message1=""
    if 200 <= (int(list2["code"])) <=299 :
        message1="False"
    else:
        message1= "True"
    return message1



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
if printfile =='true':
    print(json.dumps(data, indent=4,sort_keys=True))
