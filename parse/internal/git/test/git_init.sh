#!/bin/sh
cp -r ./test-flags-repo /tmp
cd /tmp/test-flags-repo
git init
git add .
git commit -m "init"
git checkout -b temp
git checkout master
echo `git rev-parse master`
echo `git rev-parse temp`