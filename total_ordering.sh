#!/bin/bash
# Basic while loop
# counter=1

# ports=()

# port=81
# while [ $counter -le $1 ]; do
#     ports[$i]=$port
#     ((port++))
#     ((counter++))
#     echo "${ports[$i]}"  
# done


gnome-terminal -e "go run total-ordering-leader.go 80"
gnome-terminal -e "go run total-ordering.go 81 80"
