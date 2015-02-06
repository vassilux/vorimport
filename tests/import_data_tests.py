#!/usr/bin/python

import os
import sys
import uuid
import random
import string
import time
import MySQLdb
import getopt
import datetime

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

incall_clid="551234578"
incall_src="551234578"
incall_dst="6005"
incall_dnid="1157"

outcall_clid="Tata <6006>"
outcall_src="6006"
outcall_dst="0493948118"
outcall_dnid="0493948118"

incalls_default_clids = """\
0493941157
Boris<007898989898989>
Loup<0493948400>""".split()

incall_default_peer = """\
6005
6006
6000
""".split()

incall_default_dnid = """\
1157
1158
""".split()

outcall_default_clids = """\
Isabelle<6005>
Marina<6006>
Natasha<6000>
6666
""".split()

outcall_default_dst = """\
0493941157
007898989898989
0493948400
""".split()



hangupcause_causes = """\
ANSWERED
BUSY
NOANSWER
ANSWERED
ANSWERED
ANSWERED
""".split()


MAX_INCALLS_NUMBER = 0

MAX_OUTCALLS_NUMBER = 0

seconds_in_a_day = 60 * 60 * 24

MONTH_AGO = 0

#
char_set_uniqueid = string.digits

db = MySQLdb.connect(host="192.168.3.20", user="root", passwd="lepanos", db="asteriskcdrdb")

cursor = db.cursor()

def generate_incomming():
	#test = ClidNameGenerator()
	for clid in incalls_default_clids:
		uniqueid = ''.join(random.sample(char_set_uniqueid*12, 12))
		incall_clid = random.choice(incalls_default_clids)
		incall_dst = random.choice(incall_default_peer)
		incall_dnid = random.choice(incall_default_dnid)
		incall_hangup_cause = random.choice(hangupcause_causes)
		#
		incall_src = ""

		if incall_clid == "none":
			incall_clid = ""
		else:
			tokens = incall_clid.split('<')
			if(len(tokens) == 2):
				incall_src = tokens[1].replace(">", "")
			else:
				incall_src = tokens[0]
			

		channel = "DAHDI/i5/%s-1" %(incall_src)

		duration = random.randint(1, 300)
		billsec = duration - random.randint(1, 10)
		#callername = "%s%d" % (it_name, count_contact)
		#hangupcause_id = random.randint(15, 17)
    	timestamp = int(str(time.time()).split('.')[0])
    	timestamp = timestamp - (seconds_in_a_day * 30 * MONTH_AGO)
    	timestamp = timestamp - random.randint(1, 86400) 


    	if incall_hangup_cause != "ANSWERED":
    		billsec = 0

    	query="INSERT INTO asteriskcdrdb.cdr (calldate, clid, src, dst, channel, dcontext, disposition, \
billsec, duration,uniqueid,dstchannel, dnid, recordfile, import) \
VALUES(from_unixtime(%s),'%s', '%s', '%s', '%s', 'incomming', '%s', '%s','%s', '%s', 'SIP/%s', '%s', NULL, 0)" % \
(timestamp,incall_clid, incall_src, incall_dst, channel, incall_hangup_cause, billsec, duration, uniqueid, incall_dst, incall_dnid)

    	cursor.execute(query)

def generate_outgoing():
	#
	for clid in incalls_default_clids:
		uniqueid = ''.join(random.sample(char_set_uniqueid*12, 12))
		outcall_clid = random.choice(outcall_default_clids)
		outcall_dst = random.choice(outcall_default_dst)
		outcall_dnid = outcall_dst
		outcall_hangup_cause = random.choice(hangupcause_causes)
		#
		outcall_src = ""

		if outcall_src == "none":
			outcall_src = ""
		else:
			tokens = outcall_clid.split('<')
			if(len(tokens) == 2):
				outcall_src = tokens[1].replace(">", "")
			else:
				outcall_src = tokens[0]
			

		channel = "SIP/%s-1" %(outcall_src)
		dchannel = "DAHDI/i5/%s-1" %(outcall_dst)

		duration = random.randint(1, 300)
		billsec = duration - random.randint(1, 10)
		#callername = "%s%d" % (it_name, count_contact)
		#hangupcause_id = random.randint(15, 17)
    	timestamp = int(str(time.time()).split('.')[0])
    	timestamp = timestamp - (seconds_in_a_day * 30 * MONTH_AGO)
    	timestamp = timestamp - random.randint(1, 86400)

    	if outcall_hangup_cause != "ANSWERED":
    		billsec = 0

    	query="INSERT INTO asteriskcdrdb.cdr (calldate, clid, src, dst, channel, dcontext, disposition, \
    	billsec, duration,uniqueid,dstchannel, dnid, recordfile, import) \
    	VALUES(from_unixtime(%s),'%s', '%s', '%s', '%s', 'outgoing', '%s', '%s','%s', '%s', '%s', '%s', NULL, 0)" % \
    	(timestamp,outcall_clid, outcall_src, outcall_dst, channel, outcall_hangup_cause, billsec, duration, uniqueid, dchannel, outcall_dnid)

    	cursor.execute(query)

def invoke_incall_generator():
	go = True
	i=0
	while go:
		#print " i %d" %(i)
		generate_incomming()
		i = i + 1
		if i>= MAX_INCALLS_NUMBER:
			go = False

def invoke_outcall_generator():
	go = True
	i=0
	while go:
		#print " i %d" %(i)
		generate_outgoing()
		i = i + 1
		if i>= MAX_OUTCALLS_NUMBER:
			go = False

def main():
	print "main----"
	global MAX_INCALLS_NUMBER
	global MAX_OUTCALLS_NUMBER
	global MONTH_AGO
	try:
	    myopts, args = getopt.getopt(sys.argv[1:],"i:o:m")
	except getopt.GetoptError as e:
	    print (str(e))
	    print("Usage: %s -i number of incomming calls -o numbers of outgoing calls " % sys.argv[0])
	    sys.exit(2)
	 
	for o, a in myopts:
	    if o == '-i':
	        MAX_INCALLS_NUMBER = string.atoi(a)
	    elif o == '-o':
	        MAX_OUTCALLS_NUMBER = string.atoi(a)
	    elif o == '-m':
	    	MONTH_AGO = string.atoi(a)

	#generate_outgoing()
	if MAX_INCALLS_NUMBER > 0:
		print " MAX_INCALLS_NUMBER %s"%(MAX_INCALLS_NUMBER)
		invoke_incall_generator()

	if MAX_OUTCALLS_NUMBER > 0:
		print "MAX_OUTCALLS_NUMBER %s"%(MAX_OUTCALLS_NUMBER)
		invoke_outcall_generator()


	print "Live good buddy"
if __name__ == '__main__':
	main()