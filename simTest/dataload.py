#!/usr/bin/env python
import json
import psycopg2
from pprint import pprint

NODES = [ # list of nodes to load
    'fnode0',
    'fnode01',
    'fnode02'
]

LOGS = [ # list of logfiles to load for each node
    'simtest',
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
        with open("%s_%s.txt" % (node, log), "r") as _file:
            for d in _file:
                load(d, l, n)

    # NOTE: you may need to change this connection string to match your database setup
    #conn = psycopg2.connect("postgres://load:load@localhost:5432")
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

    print("Loading log files into db as json")
    for n in NODES:
        #x('DROP SCHEMA IF EXISTS %s CASCADE' % n)
        x('CREATE SCHEMA IF NOT EXISTS %s' % n)

    for l in LOGS:
        for n in NODES:
            print("%s_%s.txt\n" % (n, l))
            try:
                x('CREATE TABLE IF NOT EXISTS %s.%s () INHERITS (public.logs)' % (n, l))
                extract(l, n)
            except Excetion as ex:
                print(ex)
                #print('skipped %s_%s' %(n,l))

    conn.close()

if __name__ == '__main__':
    main()
