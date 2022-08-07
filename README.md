# json2csv

Convert JSON data to CSV data easily

# Usage

`json2csv` is a simple filter reading from standard input and sending to standard output.

```
$ json2csv <inputfile >outputfile
```

# Description

`json2csv` works for JSON files which are already organized into simple record structures.

```
[
  { "Title": "The Life of Brian",
    "date": { "year": "1908", "month": "January"},
	"cast" : [
	    "Joanna Lumley",
		"Terrance Trent Darby"
		]
	"
},
 etc ...
]
```

The map keys are expected to become column names (not values), the map values are 
mapped to cells. Nested maps result in compound column names. Embedded lists have a 
row per each item in the list. The above is output as:

```
$,$.Title,$.cast.$,$.date.month,$.date.year
1,The Life of Brian,Joanna Lumley,January,1908
1,The Life of Brian,Terrance Trent Darby,January,1908
```

# Other Formats

To convert any arbitrary JSON file into useful CSV is a hard problem. Because the infinite
variety of formats possible make it hard to perform the right transformations. Forms which 
store data in map keys are hard. For example json2csv given this: 

```
{
  "aliceblue": [240, 248, 255, 1],
  "antiquewhite": [250, 235, 215, 1],
  "aqua": [0, 255, 255, 1],
  "aquamarine": [127, 255, 212, 1],
. . .
```
generates something we cannot use:
```
aliceblue.$,antiquewhite.$,aqua.$, . . .
240,250,0
240,250,255
240,250,255
240,250,1
```
With tools like 'jp', convert the input to a conventional form with the required
meta-data (field names):
```
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