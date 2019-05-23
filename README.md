

# godumper

***godumper*** is a multi-threaded MySQL backup and restore tool, and it is compatible with [maxbube/mydumper](https://github.com/maxbube/mydumper) in the layout.


### dumper

```
./bin/dumper --help
Usage: ./dumper -h [HOST] -P [PORT] -u [USER] -p [PASSWORD] -db [DATABASE] -o [OUTDIR]
  -F int
    	Split tables into chunks of this output file size. This value is in MB (default 128)
  -P int
    	TCP/IP port to connect to (default 3306)
  -db string
    	Database to dump
  -h string
    	The host to connect to
  -o string
    	Directory to output files to
  -p string
    	User password
  -s int
    	Attempted size of INSERT statement in bytes (default 1000000)
  -t int
    	Number of threads to use (default 16)
  -table string
    	Table to dump
  -u string
    	Username with privileges to run the dump
  -vars string
    	Session variables

eg:
$./dumper -h 192.168.0.1 -P 3306 -u root -p root -db HR  -o HR.sql


### loader

```
$ ./bin/loader --help
Usage: ./bin/myloader -h [HOST] -P [PORT] -u [USER] -p [PASSWORD] -d  [DIR] -db [database name]
  -P int
    	TCP/IP port to connect to (default 3306)
  -d string
    	Directory of the dump to import
  -h string
    	The host to connect to
  -p string
    	User password
  -t int
    	Number of threads to use (default 16)
  -u string
    	Username with privileges to run the loader
  -db string
      database name what you use

eg:
$./loader -h 192.168.0.2 -P 3306 -u root -p root -d db.sql

### streamer

Streaming mode, dumps datas from upstream to downstream in parallel instead of dumping to the out directory.
```
$./streamer
Usage: .streamer -h [HOST] -P [PORT] -u [USER] -p [PASSWORD] -db [DATABASE] -2h [DOWNSTREAM-HOST] -2P [DOWNSTREAM-PORT] -2u [DOWNSTREAM-USER] -2p [DOWNSTREAM-PASSWORD] [-2db DOWNSTREAM-DATABASE] [-o]
  -2P int
    	Downstream TCP/IP port to connect to (default 3306)
  -2db string
    	Downstream database, default is same as upstream db
  -2h string
    	The downstream host to connect to
  -2p string
    	Downstream user password
  -2u string
    	Downstream username with privileges to run the streamer
  -P int
    	Upstream TCP/IP port to connect to (default 3306)
  -db string
    	Database to stream
  -h string
    	The upstream host to connect to
  -o	Drop tables if they already exist
  -p string
    	Upstream user password
  -s int
    	Attempted size of INSERT statement in bytes (default 1000000)
  -t int
    	Number of threads to use (default 16)
  -table string
    	Table to stream
  -u string
    	Upstream username with privileges to run the streamer

eg:
$./streamer -h 192.168.0.2 -P 3306 -u mock -p mock -2h 192.168.0.3 -2P 3306 -2u mock -2p db2 -db sbtest

## License
godumper is released under the GPLv3. See LICENSE
