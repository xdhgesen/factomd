#!/bin/bash
now=$(date +"%Y-%m-%d-%H-%M")
mkdir -p logs
ls fnode0_*.txt | xargs -I filename -n 1 sh -c "echo filename;tail -n 4000 filename > logs/filename"
tar czvf logs_$now.tgz logs/*
ls -lh logs_$now.tgz
