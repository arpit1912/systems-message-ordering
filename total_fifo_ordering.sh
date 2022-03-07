#!/bin/bash

# Basic while loop
counter=1

ports=()

port=8081
while [ $counter -le $1 ]; do
    ports[$counter]=$port 
    ((port++))
    ((counter++))
done


LEADER_PORT=8080
gnome-terminal -e "go run total-fifo-ordering-leader.go $LEADER_PORT $1" &

for i in "${ports[@]}"; do
    other_ports="$LEADER_PORT"
    for j in "${ports[@]}"; do
        if [ "$j" -ne "$i" ]; then
            other_ports="$other_ports $j"
        fi
    done  
    gnome-terminal -e "go run total-ordering.go $i $other_ports" &
done
