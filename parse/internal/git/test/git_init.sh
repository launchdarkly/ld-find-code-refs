#!/bin/sh
cp -r ./test-flags-repo /tmp
cd /tmp/test-flags-repo
git config --global user.email "you@example.com"
git config --global user.name "Your Name"
git init
git add .
git commit -m "init"
git checkout -b temp
git checkout master
echo `git rev-parse master`
echo `git rev-parse temp`