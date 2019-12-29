#!/bin/bash
 
while true
do 
    procnum=`ps -ef|grep "linux-miner"|grep -v grep|wc -l`
   if [ $procnum -eq 0 ]; then
        echo "./linux-miner -C solo.conf >>/tmp/miner.log 2>&1 &"
       ./linux-miner -C solo.conf >>/tmp/miner.log 2>&1 &
   fi
   sleep 30
done