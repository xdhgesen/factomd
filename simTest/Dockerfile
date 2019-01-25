FROM postgres:11.1

RUN apt-get update 
RUN apt-get -y install git python3-dev python3-pip python3 postgresql-contrib-11 postgresql-plpython3-11

RUN apt-get clean && \
rm -rf /var/cache/apt/* /var/lib/apt/lists/*

# python package dependencies and db extension
WORKDIR /opt

RUN pip3 install psycopg2 numpy pprint

# install python db extension
ENV PYTHON_PATH /usr/local/lib/python3.5/dist-packages

WORKDIR /docker-entrypoint-initdb.d
RUN echo 'CREATE EXTENSION plpython3u;' > 000_extension.sql

# generate testing data
COPY report.sql 001_report.sql

ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 5432
CMD ["postgres"]
