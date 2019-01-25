#!/usr/bin/env python
import json

NODES = ['fnode0']

LOGS = [
    #'ackqueue',
    #'apilog',
    #'balancehash',
    #'commits',
    #'dbsig-eom',
    #'dbsig',
    #'dbstateprocess',
    #'duplicatesend',
    #'election',
    #'entrycredits',
    #'entrycredits_trans',
    #'entrysync',
    #'executemsg',
    #'factoids',
    #'factoids_trans',
    #'faulting',
    'holding',
    #'inmsgqueue',
    #'inmsgqueue2',
    #'missing_messages',
    #'msgqueue',
    #'networkinputs',
    #'networkoutputs',
    #'pendingchainheads',
    #'process',
    #'processlist',
    #'processstatus',
    #'simtest',
    #'graphdata',
    #'marshalsizes',
]


def main(*params):
    fo = file("%s_%s.txt" % params, "rw+")

    for l in fo.readlines():
        load(l, *params)

    fo.close()


def load(data, log, node):
    try:
        d = json.loads(data)
        # TODO: do someting w/ data
    except Exception as x:
        print(log, node, data)

if __name__ == '__main__':

# load all log data

    for l in LOGS:
        for n in NODES:
            main(n, l)
