/*
 * go-mydumper
 * xelabs.org
 *
 * Copyright (c) XeLabs
 * GPL License
 *
 */

package main

import (
	"common"
	"flag"
	"fmt"
	"os"

	"github.com/xelabs/go-mysqlstack/xlog"
)

var (
	flagOverwriteTables                                                                                                         bool
	flagThreads, flagPort, flag2port, flagStmtSize                                                                              int
	flagUser, flagPasswd, flagHost, flag2user, flag2passwd, flag2host, flagDb, flag2Db, flag2Engine, flagTable, flagSessionVars string

	log = xlog.NewStdLog(xlog.Level(xlog.INFO))
)

func init() {
	flag.StringVar(&flagUser, "u", "veevaapp", "Upstream username with privileges to run the streamer")
	flag.StringVar(&flagPasswd, "p", "ECJC5f4gnwK6Q6Wr", "Upstream user password")
	flag.StringVar(&flagHost, "h", "stgdb.veevasfa.cn", "The upstream host to connect to")
	flag.IntVar(&flagPort, "P", 3306, "Upstream TCP/IP port to connect to")
	flag.StringVar(&flag2user, "2u", "root", "Downstream username with privileges to run the streamer")
	flag.StringVar(&flag2passwd, "2p", "root123", "Downstream user password")
	flag.StringVar(&flag2host, "2h", "localhost", "The downstream host to connect to")
	flag.IntVar(&flag2port, "2P", 3306, "Downstream TCP/IP port to connect to")
	flag.StringVar(&flag2Db, "2db", "SteamTest", "Downstream database, default is same as upstream db")
	flag.StringVar(&flag2Engine, "2engine", "", "Downstream table engine")
	flag.StringVar(&flagDb, "db", "MSBeigene", "Database to stream")
	flag.StringVar(&flagTable, "table", "", "Table to stream")
	flag.IntVar(&flagThreads, "t", 16, "Number of threads to use")
	flag.IntVar(&flagStmtSize, "s", 1000000, "Attempted size of INSERT statement in bytes")
	flag.BoolVar(&flagOverwriteTables, "o", true, "Drop tables if they already exist")
	flag.StringVar(&flagSessionVars, "vars", "", "Session variables")
}

func usage() {
	fmt.Println("Usage: " + os.Args[0] + " -h [HOST] -P [PORT] -u [USER] -p [PASSWORD] -db [DATABASE] -2h [DOWNSTREAM-HOST] -2P [DOWNSTREAM-PORT] -2u [DOWNSTREAM-USER] -2p [DOWNSTREAM-PASSWORD] [-2db DOWNSTREAM-DATABASE] [-o]")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = func() { usage() }
	flag.Parse()

	if flagHost == "" || flagUser == "" || flagDb == "" || flag2host == "" || flag2user == "" {
		usage()
		os.Exit(0)
	}

	args := &common.Args{
		User:            flagUser,
		Password:        flagPasswd,
		Address:         fmt.Sprintf("%s:%d", flagHost, flagPort),
		ToUser:          flag2user,
		ToPassword:      flag2passwd,
		ToAddress:       fmt.Sprintf("%s:%d", flag2host, flag2port),
		Database:        flagDb,
		ToDatabase:      flag2Db,
		ToEngine:        flag2Engine,
		Table:           flagTable,
		Threads:         flagThreads,
		StmtSize:        flagStmtSize,
		IntervalMs:      10 * 1000,
		OverwriteTables: flagOverwriteTables,
		SessionVars:     flagSessionVars,
	}
	common.Streamer(log, args)
}
