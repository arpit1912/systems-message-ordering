#!/bin/bash

# Basic while loop
counter=1

ports=()

port=8081
while [ $counter -le $1 ]; do
    ports[$counter]=$port
    echo "${ports[$counter]}"  
    ((port++))
    ((counter++))
done


LEADER_PORT=8080
xterm -e "go run total-ordering-leader.go $LEADER_PORT" &

echo "printing"
for i in "${ports[@]}"; do
    other_ports="$LEADER_PORT"
    for j in "${ports[@]}"; do
        if [ "$j" -ne "$i" ]; then
            other_ports="$other_ports $j"
        fi
    done
    echo "$other_ports"  
    xterm -e "go run total-ordering.go $i $other_ports" &
done
