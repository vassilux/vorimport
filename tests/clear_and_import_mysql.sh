#!/bin/bash
#########################################################################################
#											
#											
#											
#											
#											
# Description : Clear mysql data and import dump for test				
# Author : vassilux
# Last modified : 2014-09-18 14:00:00 
########################################################################################## 

SCRIPTPATH=$(cd "$(dirname "$0")"; pwd)


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
	
	echo "delete from cdr;" | mysql -u root -plepanos -h127.0.0.1 asteriskcdrdb
	
	echo "delete from cel;" | mysql -u root -plepanos -h127.0.0.1 asteriskcdrdb
	
	mysql -uroot -plepanos -h127.0.0.1 asteriskcdrdb < ${SCRIPTPATH}/asteriskcdrdb.sql 
}

main
















