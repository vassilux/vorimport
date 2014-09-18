#!/bin/bash


function dprint()
{
  tput setb 1
  level="$1"
  dtype="$2"
  value="$3"
  echo -e "\e[00;31m $dtype:$level $value \e[00m" >&2
}


function main()
{
	dprint 0 INFO "Do you want to delete all revor import data?[y/N]"
  	read resp
  	if [ "$resp" = "" -o "$resp" = "n" ]; then
    	dprint 0 INFO "See you next time. Bye"
    	exit 0
  	fi  

	mongo revor --eval "db.cdrs.drop()"
	mongo revor --eval "db.dailyanalytics_incomming.drop()"

	mongo revor --eval "db.dailyanalytics_outgoing.drop()"
	mongo revor --eval "db.dailydid_incomming.drop()"
	mongo revor --eval "db.monthlyanalytics__incomming.drop()"
	mongo revor --eval "db.monthlyanalytics_incomming.drop()"
	mongo revor --eval "db.monthlyanalytics_outgoing.drop()"
	mongo revor --eval "db.monthlydid_incomming.drop()"

}

main
















