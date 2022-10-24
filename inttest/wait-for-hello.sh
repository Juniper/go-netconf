#!/bin/sh

usage() {
    printf "Usage: %s [-p <port>] [-s <pass>] [-t <timeout>] [-w <wait>] [<username>@]<host>\n" "$0" >&2
    printf "\n" >&2
    printf "SSH_PASS enviroment variable can be used in place of -p as well\n\n" >&2
}

pass="${SSH_PASS}"
timeout=300 # 5 minutes
wait=10

while getopts "h:p:u:s:t:w:" arg; do
    case "${arg}" in
    p) port="${OPTARG}" ;;
    s) pass="${OPTARG}" ;;
    t) timeout="${OPTARG}" ;;
    w) wait="${OPTARG}" ;;
    h) 
        usage
        exit 0
        ;; 
    ?)
        echo "invalid option: -${OPTARG}."
        usage
        exit 1
        ;;
    esac
done
shift "$((OPTIND-1))"

if [ -z "$1" ]; then
    printf "wait-for-hello: missing host\n\n" >&2
    usage
    exit 1
fi 

host="${1}"

ssh_cmd="ssh ${host}" 

if [ -n "$port" ]; then
    ssh_cmd="${ssh_cmd} -p $port"
fi

ssh_cmd="${ssh_cmd} -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -s netconf"

orig_ssh_cmd=${ssh_cmd}

if [ -n "$pass" ]; then 
    ssh_cmd="sshpass -p ${pass} ${ssh_cmd}"
fi


printf "wait-for-hello: waiting for netconf subsystem on %s for %d seconds\n" "${host}" "${timeout}" >&2

tries=0
endtime=$(($(date +%s) + timeout))
while [ "$(date -u +%s)" -le "$endtime" ]; do
    echo "${orig_ssh_cmd}"
    out=$(echo | ${ssh_cmd} | sed 's/]]>]]>//')
    session_id=$(echo "$out" | xpath -q -e 'hello/session-id/text()')

    if [ -n "$session_id" ]; then
        printf "wait-for-hello: success! session-id: %d\n" "${session_id}" >&2
         exit 0
    fi

    tries=$((tries+1))
    printf "wait-for-hello: unsucessful waiting %d before trying again\n\n" "${wait}" >&2
    sleep "$wait"
done

printf "wait-for-hello: failed to connect to the netconf subsystem after %d seconds (%d tries)\n" "${timeout}" "${tries}" >&2
exit 1
