# json2csv

Convert JSON data to CSV data easily

# Usage

`json2csv` is a simple filter reading from standard input and sending to standard output.

```
$ json2csv <inputfile >outputfile
```
Usage of json2csv:

`-dedupe`
Remove duplicate rows. Usage: -dedupe=false (default true)

`-index` string
List index column name part. (default "#")

`-list` string
List column name part. (default "LIST")

`-separator` string
column name internal separator. (default ".")

# Description

`json2csv` works for JSON files which are already organized into simple record structures.

```JSON
[
  { "Title": "The Life of Brian",
    "date": { "year": "1908", "month": "January"},
    "cast" : [ "Joanna Lumley", "Terrance Trent Darby" ]
  },
  { "Title": "The Life of Pi",
    "date": { "year": "2012", "month": "January"},
    "cast" : ["Suraj Sharma", "Nota Tiger" ]
  }
]
```

The map keys are expected to become column names (not values), the map values are 
mapped to cells. Nested maps result in compound column names. Embedded lists have a 
row per each item in the list. The above is output as:

```
LIST,LIST.#,LIST.Title,LIST.date.year,LIST.cast.LIST,LIST.date.month,LIST.cast.LIST.#
,1,The Life of Brian,1908,Joanna Lumley,January,1
,1,The Life of Brian,1908,Terrance Trent Darby,January,2
,2,The Life of Pi,2012,Suraj Sharma,January,1
,2,The Life of Pi,2012,Nota Tiger,January,2
```

Here's the pretty version:

|LIST|LIST.#|LIST.Title       |LIST.date.year|LIST.cast.LIST      |LIST.date.month|LIST.cast.LIST.#|
|----|------|-----------------|--------------|--------------------|---------------|----------------|
|    |1     |The Life of Brian|1908          |Joanna Lumley       |January        |1               |
|    |1     |The Life of Brian|1908          |Terrance Trent Darby|January        |2               |
|    |2     |The Life of Pi   |2012          |Suraj Sharma        |January        |1               |
|    |2     |The Life of Pi   |2012          |Nota Tiger          |January        |2               |

# Column names

The column names format can be altered to avoid issues by using the command-line options. 
Example:
```
$ ./json2csv -list 'items' -index number -separator '_' <test/fixtures/movies.json
items,items_number,items_Title,items_date_year,items_cast_items,items_date_month,items_cast_items_number
,1,The Life of Brian,1908,Joanna Lumley,January,1
,2,The Life of Pi,2012,Suraj Sharma,January,1
,1,The Life of Brian,1908,Terrance Trent Darby,January,2
,2,The Life of Pi,2012,Nota Tiger,January,2
```

# Duplicate rows

The program would generate duplicate rows if a subtree lacks lists which other subtrees have. 
By default it remembers all rows previously output and skips duplicates. This may consume 
too much memory with very large input files, in which case the option `-dedupe=false` is available/

# Other Formats

To convert any arbitrary JSON file into useful CSV is a hard problem. Because the infinite
variety of formats possible make it hard to perform the right transformations. Forms which 
store data in map keys are hard. For example json2csv given this: 

```JSON
{
  "aliceblue": [240, 248, 255, 1],
  "antiquewhite": [250, 235, 215, 1],
  "aqua": [0, 255, 255, 1],
  "aquamarine": [127, 255, 212, 1],
. . .
```
generates something we cannot use:
```
aliceblue.LIST,antiquewhite.LIST,aqua.LIST, . . .
240,250,0
240,250,255
240,250,255
240,250,1
```
With tools like 'jp', convert the input to a conventional form with the required
meta-data (field names):
```JSON
{
  "colours": [
    {
      "blue": 255,
      "green": 248,
      "name": "aliceblue",
      "red": 240
    },
. . .
```

# Using SQL to query the CSV

json2csv denormalizes the JSON - you can query the output table with SQL, here I use `duckdb`

Example IP Addresses JSON
Query on any key in the JSON:
```
$ <test/fixtures/ip-addresses-style1.json ./json2csv -separator _ > foo.csv
$ duckdb -cmd '.mode box' -c "SELECT LIST_ip, LIST_country, LIST_city, LIST_connection_asn FROM read_csv('foo.csv') WHERE LIST_connection_isp = 'ITonCloud';"              
┌──────────────┬──────────────┬───────────┬─────────────────────┐
│   LIST_ip    │ LIST_country │ LIST_city │ LIST_connection_asn │
├──────────────┼──────────────┼───────────┼─────────────────────┤
│ 103.215.20.6 │ Australia    │ Sydney    │ 132015              │
└──────────────┴──────────────┴───────────┴─────────────────────┘
```

If the CSV file is large, save space with Parquet file, convert with duckdb and the query the parquet file:

```
$ duckdb -c "COPY (SELECT * FROM 'foo.csv') TO 'foo.parquet' (FORMAT PARQUET);"`
$ duckdb -cmd '.mode box' -c "SELECT LIST_ip, LIST_country, LIST_city, LIST_connection_asn FROM read_parquet('foo.parquet') WHERE LIST_connection_isp = 'ITonCloud';"
```
