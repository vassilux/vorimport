#!/usr/bin/python

import os
import sys
import MySQLdb
'''
Please create virtualenv revor 
mkvirtualenv --no-site-packages revor 
and install mysql-python with pip
pip install mysql-python 
'''

'''
incomming
INSERT INTO asteriskcdrdb.cdr (calldate, clid, src, dst, channel, dcontext, disposition,billsec,duration,
uniqueid,dstchannel, dnid, recordfile, import) VALUES(now(), 'Kebab <551234578>', 
'551234578', '4002', 'DAHDI/i5/551234578-1', 'incomming', 'ANSWERED', '5', 
'10', '139412787.59', '', '1157', NULL, 0)

'''

'''
Outgoing example 
 '1400147484', '"Cyril Hellstern" <6006>', '6006', '8181', 'SIP/6006-0000001b', 'to-g200be', 'ANSWERED', '0', '1', '1400147484.27', 'SIP/g200beprovider-0000001c', '8181', NULL

'''

incall_clid="Toto <551234578>"
incall_src="551234578"
incall_dst="6005"
incall_dnid="1157"

db = MySQLdb.connect(host="localhost", user="root", passwd="lepanos", db="asteriskcdrdb")

cursor = db.cursor()

def generate_incomming():
	query="INSERT INTO asteriskcdrdb.cdr (calldate, clid, src, dst, channel, dcontext, disposition,billsec,\
		duration,uniqueid,dstchannel, dnid, recordfile, import) VALUES(now(),'%s', '%s', '%s', 'DAHDI/i5/551234578-1', \
		'incomming', 'ANSWERED', '5','1', '139412787.58', 'SIP/%s', '%s', NULL, 0)" % \
		(incall_clid, incall_src, incall_dst, incall_dst, incall_dnid)

	print " ---- query --- : %s" % (query)	
	cursor.execute(query)

def generate_outgoing():
	pass

def main():
	print "main----"
	generate_outgoing()
	generate_incomming()

if __name__ == '__main__':
	main()