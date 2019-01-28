#!/usr/bin/env python
import json
import psycopg2
from pprint import pprint

NODES = ['fnode0']

LOGS = [
    'simtest',
    #'processstatus',
    #'processlist',
    #'process',
    #'pendingchainheads',
    #'networkoutputs', # 500k+ load by itself
    #'networkinputs', # 100k+
#    'msgqueue',
#    'missing_messages',
#    'inmsgqueue2',
#    'inmsgqueue',
    'holding',
#    'faulting',
    #'factoids_trans',
    #'factoids',
    #'executemsg',
    #'entrysync',
#    'entrycredits_trans',
    #'entrycredits',
    #'election',
    #'duplicatesend',
    #'dbstateprocess',
    #'dbsig',
    #'dbsig-eom',
#    'commits',
    #'balancehash',
#    'apilog',
#    'ackqueue',
]

def main():
    """ initalize the db """

    def load(data, log, node):
        """ load data """
        try:
            #_ = json.loads(data)

            if data[0] in ['[', '{']:
                x("""
                INSERT INTO %s.%s (e, run)
                values('%s', (select max(id) from public.log_runs))
                """ % (node, log, data))

        except Exception as ex:
            print(ex)
            print(log, node, data)

    def extract(log, node):
        fo = open("%s_%s.txt" % (node, log), "r")

        for d in fo.readlines():
            load(d, l, n)

        fo.close()

    conn = psycopg2.connect("postgres://load:load@localdb:5432")

    def x(sql, fetch=True):
        cursor = conn.cursor()  
        cursor.execute(sql)
        conn.commit()

    #x('DROP TABLE IF EXISTS logs CASCADE')
    #x('DROP TABLE IF EXISTS log_runs CASCADE')
    x('CREATE TABLE IF NOT EXISTS public.logs (e jsonb, run int)')
    x('CREATE TABLE IF NOT EXISTS public.log_runs (id serial, ts timestamp)')
    x('INSERT INTO public.log_runs(ts) values(now())')

    for n in NODES:
        #x('DROP SCHEMA IF EXISTS %s CASCADE' % n)
        x('CREATE SCHEMA IF NOT EXISTS %s' % n)

    for l in LOGS:
        for n in NODES:
            try:
                x('CREATE TABLE IF NOT EXISTS %s.%s () INHERITS (public.logs)' % (n, l))
                extract(l, n)
            except Excetion as ex:
                print(ex)
                #print('skipped %s_%s' %(n,l))

    conn.close()

if __name__ == '__main__':
    main()
