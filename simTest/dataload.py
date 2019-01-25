#!/usr/bin/env python
import json
import psycopg2
from pprint import pprint

NODES = ['fnode0']

LOGS = [
    'ackqueue',
    'apilog',
    #'balancehash',
    'commits',
    #'dbsig-eom',
    #'dbsig',
    #'dbstateprocess',
    #'duplicatesend',
    #'election',
    #'entrycredits',
    'entrycredits_trans',
    #'entrysync',
    #'executemsg',
    #'factoids',
    #'factoids_trans',
    'faulting',
    'holding',
    'inmsgqueue',
    'inmsgqueue2',
    'missing_messages',
    'msgqueue',
    'networkinputs',
    'networkoutputs',
    #'pendingchainheads',
    #'process',
    #'processlist',
    #'processstatus',
    'simtest',
]

#LOGS = [ 'missing']

def main():
    """ initalize the db """

    def x(sql, fetch=True):
        conn = psycopg2.connect("postgres://load:load@localdb:5432")
        cursor = conn.cursor()  
        cursor.execute(sql)
        rows = []

        if fetch:
            rows = cursor.fetchall()  

        conn.commit()
        conn.close()
        return rows

    def perform(sql):
        return x(sql, fetch=False)

    def load(data, log, node):
        """ load data """
        try:
            _ = json.loads(data)
            perform("INSERT INTO %s.%s (e) values('%s')" % (node, log, data))
        except Exception as ex:
            #print(ex)
            print(log, node, data)

    def extract(log, node):
        fo = open("%s_%s.txt" % (node, log), "r")

        for d in fo.readlines():
            load(d, l, n)

        fo.close()

    perform('DROP TABLE IF EXISTS logs CASCADE')
    perform('CREATE TABLE public.logs (e jsonb)')

    for n in NODES:
        perform('DROP SCHEMA IF EXISTS %s CASCADE' % n)
        perform('CREATE SCHEMA %s' % n)

    for l in LOGS:
        for n in NODES:
            try:
                perform('CREATE TABLE %s.%s () INHERITS (public.logs)' % (n, l))
                extract(l, n)
            except Excetion as ex:
                print(ex)
                print('skipped %s_%s' %(n,l))

if __name__ == '__main__':
    main()
