#!/usr/bin/python

import os
import sys
import uuid
import random
import string
import time
import MySQLdb
import getopt

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
8400
""".split()

hangupcause_causes = """\
ANSWERED
BUSY
ANSWERED
CONGESTION
ANSWERED
FAILED
ANSWERED
ANSWERED
ANSWERED
""".split()


MAX_INCALLS_NUMBER = 0

MAX_OUTCALLS_NUMBER = 0

#
'''
class ClidNameGenerator(object):
    def __init__(self, names=incalls_default_clids):
        self._names =  names #{i : name.strip() for i, name in enumerate(names)}
        self._total_names = len(self._names)
        self._used_indices = set()
    def __call__(self):
        index = random.randrange(self._total_names)
        name = self._names[index]
        return name
    def __iter__(self):
        while True:
            yield self()
'''
#
char_set_uniqueid = string.digits

db = MySQLdb.connect(host="localhost", user="root", passwd="lepanos", db="asteriskcdrdb")

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
    	timestamp = timestamp - random.randint(1, 86400)

    	if incall_hangup_cause != "ANSWERED":
    		billsec = 0

    	#print "Incall clid %s %s "%(incall_clid, timestamp)
    	query="INSERT INTO asteriskcdrdb.cdr (calldate, clid, src, dst, channel, dcontext, disposition, \
billsec, duration,uniqueid,dstchannel, dnid, recordfile, import) \
VALUES(from_unixtime(%s),'%s', '%s', '%s', '%s', 'incomming', '%s', '%s','%s', '%s', 'SIP/%s', '%s', NULL, 0)" % \
(timestamp,incall_clid, incall_src, incall_dst, channel, incall_hangup_cause, billsec, duration, uniqueid, incall_dst, incall_dnid)

    	cursor.execute(query)

def generate_outgoing():
	query="INSERT INTO asteriskcdrdb.cdr (calldate, clid, src, dst, channel, dcontext, disposition,billsec,\
		duration,uniqueid,dstchannel, dnid, recordfile, import) VALUES(now(),'%s', '%s', '%s', 'DAHDI/i5/551234578-1', \
		'outgoing', 'ANSWERED', '5','1', '139412787.59', '%s', '%s', NULL, 0)" % \
		(outcall_clid, outcall_src, outcall_dst, outcall_dst, outcall_dnid)
	
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

def main():
	print "main----"
	global MAX_INCALLS_NUMBER
	global MAX_OUTCALLS_NUMBER
	try:
	    myopts, args = getopt.getopt(sys.argv[1:],"i:o:")
	except getopt.GetoptError as e:
	    print (str(e))
	    print("Usage: %s -i number of incomming calls -o numbers of outgoing calls " % sys.argv[0])
	    sys.exit(2)
	 
	for o, a in myopts:
	    if o == '-i':
	        MAX_INCALLS_NUMBER = string.atoi(a)
	    elif o == '-o':
	        MAX_OUTCALLS_NUMBER = string.atoi(a)

	#generate_outgoing()
	if MAX_INCALLS_NUMBER > 0:
		print " MAX_INCALLS_NUMBER %s"%(MAX_INCALLS_NUMBER)
		invoke_incall_generator()

if __name__ == '__main__':
	main()