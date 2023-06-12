#! /bin/sh

usage() { echo "Usage: $0 [-v <true/false>] [-f <true/false>] [-s <true/false>]" 1>&2; exit 1; }
_term(){
    echo "shutting down the application container"
    /usr/sbin/crond stop
    kill -TERM "$child" 2>/dev/null
}

trap _term SIGTERM #so as to pass it down
echo "starting cron deamon..."
/usr/sbin/crond -f -l 8&


# getting all the command line arguments 
verbose="false"
filelog="false"
seed="false"
while getopts ":v:f:s:" o; do
    case "${o}" in
        v)
            verbose=${OPTARG}
            ;;
        f)
            filelog=${OPTARG}
            ;;
        s) 
            seed=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done
echo $verbose
echo $filelog
echo $seed
echo "now booting the botmincock application.."
/usr/bin/botmincock -verbose $verbose -flog $filelog -seed $seed&

# waiting for seller pro application 
child=$!
wait "$child"
/usr/sbin/crond stop # if the go app container exits gracefully without any user interruption 