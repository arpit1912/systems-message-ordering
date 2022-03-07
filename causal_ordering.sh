#!/bin/bash

# Basic while loop
counter=1

ports=()

port=8080
while [ $counter -le $1 ]; do
    ports[$counter]=$port 
    ((port++))
    ((counter++))
done

for i in "${ports[@]}"; do
    other_ports=""
    for j in "${ports[@]}"; do
        if [ "$j" -ne "$i" ]; then
            other_ports="$other_ports $j"
        fi
    done  
    gnome-terminal -e "go run causal-ordering.go $i $other_ports" &
done
