#!/usr/bin/env bash

proc_name="linux-miner"

proc_num()
{
    num=`ps -ef | grep $proc_name | grep -v grep | wc -l`
    return $num
}

proc_num
number=$?
if [ $number -eq 0 ]
then
    ./linux-miner -C solo.conf
fi