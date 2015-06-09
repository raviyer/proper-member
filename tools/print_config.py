import pprint
import json
import sys
f = open (sys.argv[1])
s = f.read()
o = json.loads(s)

pprint.pprint(o, indent=2)
