# simTest

This folder contains scripts used to parse raw log data into a report.


### Usage

* run a script to generate logs ( should see .txt files in this folder )
* get a postgres db ( docker-compose file is included )
* load data
* run report

### Run ./report/notebook.sh

using pandas to load data frames from pandas and plotted w/ plotly python api

* https://dev.socrata.com/blog/2016/02/02/plotly-pandas.html

NOTE: usage of this library requires registering an account & setting up these files:

 ~/.plotly/.config
 ~/.plotly/.credentials
