import json

def hasbody(list1):
    message=""
    if list1["body"]:
        message="True"
    else:
        message= "False"
    return message

def haserror(list2):
    message1=""
    if 200 <= (int(list2["code"])) <=299 :
        message1="False"
    else:
        message1= "True"
    return message1



# Loading json file
with open('/home/sahshah/Downloads/All_data/2022-04-20/requests/c750f388-3eb1-5baa-bfa3-13ec62b02ccd_self-terms-review.json') as f:
    data = [json.loads(line) for line in f]


# Adding the required fields
version="version"
version_value="b4d0329"
has_error="has_error"
has_body="has_body"
uuid="uuid"
uuid_value="c750f388-3eb1-5baa-bfa3-13ec62b02ccd"


for i in data:
    body_value = hasbody(i)
    error_val = haserror(i)
    i[version]=version_value
    i[uuid]=uuid_value
    i[has_body]=body_value
    i[has_error]=error_val
        

# Dumping(saving) data into JSON file
with open('/home/sahshah/Downloads/All_data/2022-04-20/requests/c750f388-3eb1-5baa-bfa3-13ec62b02ccd_self-terms-review.json',"w") as f:
    json.dump(data, f)

# Pretty print the loaded file
print(json.dumps(data, indent=4,sort_keys=True))
