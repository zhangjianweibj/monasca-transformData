#!/usr/bin/env python

from __future__ import print_function

import os
import sys

from jinja2 import Template

def main():
    if len(sys.argv) != 3:
        print('Usage: {} [input] [output]'.format(sys.argv[0]))
        sys.exit(1)

    in_path = sys.argv[1]
    out_path = sys.argv[2]

    with open(in_path, 'r') as in_file, open(out_path, 'w') as out_file:
        t = Template(in_file.read(),
                     keep_trailing_newline=True,
                     lstrip_blocks=True,
                     trim_blocks=True)
        out_file.write(t.render(os.environ))

if __name__ == '__main__':
    main()