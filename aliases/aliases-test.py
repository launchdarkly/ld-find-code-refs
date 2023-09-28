#! /usr/bin/env python
import sys,json
for flag_key in sys.stdin:
    flag_key = flag_key.strip()
    alias = '_'.join([word.capitalize() for word in flag_key.split('_')])
    aliases = json.dumps([alias])
    print(aliases)
    break
