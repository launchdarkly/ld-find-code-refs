#!/bin/sh
git init
git add .
git commit -m "init"
git checkout -b temp
git checkout master
echo `git rev-parse master`
echo `git rev-parse temp`