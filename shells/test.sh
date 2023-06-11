#! /bin/sh

usage() { echo "Usage: $0 [-v <true/false>] [-f <true/false>] [-s <true/false>]" 1>&2; exit 1; }

verbose=false
filelog=false
seed=false
while getopts ":v:f:s:" o; do
    case "${o}" in
        v)
            verbose=true
            ;;
        f)
            filelog=true
            ;;
        s) 
            seed=true
            ;;
        *)
            usage
            ;;
    esac
done
echo $verbose
echo $filelog
echo $seed