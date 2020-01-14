#!/bin/bash
/usr/bin/scl enable devtoolset-8
gcc -v
make
mv bridge /tmp/
