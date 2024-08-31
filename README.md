# 　music-timer-api
[ English | [日本語]() ]

## Overview
WEB API for [music-timer](https://github.com/pp-develop/music-timer).  
create spotify playlist by specifying time  

## Setup
create env file

## Usage
1. Create and start containers
```
$ cd build
$ docker-compose up -d
```

2. Initialize CockroachDB
```
$ docker exec -it cockroachdb bash
$ cockroach sql --insecure --host=localhost:26257 < /cockroach/init.sql
```