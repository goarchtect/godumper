/*
 * go-mydumper
 * xelabs.org
 *
 * Copyright (c) XeLabs
 * GPL License
 *
 */

package common

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xelabs/go-mysqlstack/common"
	"github.com/xelabs/go-mysqlstack/xlog"
)

// Files tuple.
type Files struct {
	databases []string
	schemas   []string
	tables    []string
	triggers    []string
}

var (
	dbSuffix     = "-schema-create.sql"
	schemaSuffix = "-schema.sql"
	tableSuffix  = ".sql"
	triggerSuffix = "-trigger.sql"
)

func loadFiles(log *xlog.Log, dir string) *Files {
	files := &Files{}
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Panicf("loader.file.walk.error:%+v", err)
		}

		if !info.IsDir() {
			switch {
			case strings.HasSuffix(path, dbSuffix):
				files.databases = append(files.databases, path)
			case strings.HasSuffix(path, schemaSuffix):
				files.schemas = append(files.schemas, path)
			case strings.HasSuffix(path, triggerSuffix):
				files.triggers = append(files.triggers, path)
			default:
				if strings.HasSuffix(path, tableSuffix) {
					files.tables = append(files.tables, path)
				}
			}
		}
		return nil
	}); err != nil {
		log.Panicf("loader.file.walk.error:%+v", err)
	}
	return files
}

func restoreDatabaseSchema(log *xlog.Log, dbs []string, conn *Connection) {
	for _, db := range dbs {
		base := filepath.Base(db)
		name := strings.TrimSuffix(base, dbSuffix)

		data, err := ReadFile(db)
		AssertNil(err)
		sql := common.BytesToString(data)

		err = conn.Execute(sql)
		AssertNil(err)
		log.Info("restoring.database[%s]", name)
	}
}

func restoreTrigger(log *xlog.Log, overwrite bool, triggers []string, conn *Connection,dbName string)  {
	for _, trigger := range triggers {
		// use
		var db string
		base := filepath.Base(trigger)
		name := strings.TrimSuffix(base, triggerSuffix)
		if dbName == "" {
			db = strings.Split(name, ".")[0]
		} else {
			db = dbName
		}
		tr := strings.Split(name, ".")[1]
		name = fmt.Sprintf("`%v`.`%v`", db, tr)

		log.Info("working.trigger[%s]", name)

		err := conn.Execute(fmt.Sprintf("USE `%s`", db))
		AssertNil(err)

		data, err := ReadFile(trigger)
		AssertNil(err)
		query := common.BytesToString(data)
		//querys := strings.Split(query1, ";\n")
		//fmt.Println(querys)

		if !strings.HasPrefix(query, "/*") && query != "" {
			if overwrite {
				log.Info("drop(overwrite.is.true).trigger[%s]", name)
				dropQuery := fmt.Sprintf("DROP TRIGGER IF EXISTS %s", name)
				err = conn.Execute(dropQuery)
				AssertNil(err)
			}
			err = conn.Execute(query)
			AssertNil(err)
		}
		log.Info("restoring.trigger[%s]", name)
	}
}

func restoreTableSchema(log *xlog.Log, overwrite bool, tables []string, conn *Connection,dbName string) {
	for _, table := range tables {
		// use
		var db string
		base := filepath.Base(table)
		name := strings.TrimSuffix(base, schemaSuffix)
		if dbName == ""{
			db = strings.Split(name, ".")[0]
		}else {
			db = dbName
		}
		tbl := strings.Split(name, ".")[1]
		name = fmt.Sprintf("`%v`.`%v`", db, tbl)

		log.Info("working.table[%s]", name)

		err := conn.Execute(fmt.Sprintf("USE `%s`", db))
		AssertNil(err)

		err = conn.Execute("SET FOREIGN_KEY_CHECKS=0")
		AssertNil(err)

		data, err := ReadFile(table)
		AssertNil(err)
		query1 := common.BytesToString(data)
		querys := strings.Split(query1, ";\n")
		for _, query := range querys {
			if !strings.HasPrefix(query, "/*") && query != "" {
				if overwrite {
					log.Info("drop(overwrite.is.true).table[%s]", name)
					dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s", name)
					err = conn.Execute(dropQuery)
					AssertNil(err)
				}
				err = conn.Execute(query)
				AssertNil(err)
			}
		}
		log.Info("restoring.schema[%s]", name)
	}
}

func restoreTable(log *xlog.Log, table string, conn *Connection,dbName string) int {
	bytes := 0
	part := "0"
	var db string
	base := filepath.Base(table)
	name := strings.TrimSuffix(base, tableSuffix)
	splits := strings.Split(name, ".")

	if dbName == ""{
		db = splits[0]
	}else{
		db = dbName
	}
	tbl := splits[1]
	if len(splits) > 2 {
		part = splits[2]
	}

	log.Info("restoring.tables[%s].parts[%s].thread[%d]", tbl, part, conn.ID)
	err := conn.Execute(fmt.Sprintf("USE `%s`", db))
	AssertNil(err)
	err = conn.Execute("SET FOREIGN_KEY_CHECKS=0")
	AssertNil(err)
	err = conn.Execute("SET UNIQUE_CHECKS=0")
	AssertNil(err)
	err = conn.Execute("SET AUTOCOMMIT=0")
	AssertNil(err)

	data, err := ReadFile(table)
	AssertNil(err)
	query1 := common.BytesToString(data)
	querys := strings.Split(query1, ";\n")
	bytes = len(query1)
	for _, query := range querys {
		if !strings.HasPrefix(query, "/*") && query != "" {
			err = conn.Execute(query)
			AssertNil(err)
		}
	}
	err = conn.Execute("SET FOREIGN_KEY_CHECKS=1")
	AssertNil(err)
	err = conn.Execute("SET UNIQUE_CHECKS=1")
	AssertNil(err)
	err = conn.Execute("SET AUTOCOMMIT=1")
	AssertNil(err)
	log.Info("restoring.tables[%s].parts[%s].thread[%d].done...", tbl, part, conn.ID)
	return bytes
}

// Loader used to start the loader worker.
func Loader(log *xlog.Log, args *Args) {
	pool, err := NewPool(log, args.Threads, args.Address, args.User, args.Password, args.SessionVars)
	AssertNil(err)
	defer pool.Close()

	files := loadFiles(log, args.Outdir)


	if args.Database == ""{
		conn := pool.Get()
		restoreDatabaseSchema(log, files.databases, conn)
		pool.Put(conn)
	}

	conn := pool.Get()
	restoreTableSchema(log, args.OverwriteTables, files.schemas, conn,args.Database)
	pool.Put(conn)

	// Shuffle the tables
	for i := range files.tables {
		j := rand.Intn(i + 1)
		files.tables[i], files.tables[j] = files.tables[j], files.tables[i]
	}

	var wg sync.WaitGroup
	var bytes uint64
	t := time.Now()
	for _, table := range files.tables {
		conn := pool.Get()
		wg.Add(1)
		go func(conn *Connection, table string) {
			defer func() {
				wg.Done()
				pool.Put(conn)
			}()
			r := restoreTable(log, table, conn,args.Database)
			atomic.AddUint64(&bytes, uint64(r))
		}(conn, table)
	}

	connTrigger := pool.Get()
	restoreTrigger(log, args.OverwriteTables, files.triggers, connTrigger,args.Database)
	pool.Put(connTrigger)

	tick := time.NewTicker(time.Millisecond * time.Duration(args.IntervalMs))
	defer tick.Stop()
	go func() {
		for range tick.C {
			diff := time.Since(t).Seconds()
			bytes := float64(atomic.LoadUint64(&bytes) / 1024 / 1024)
			rates := bytes / diff
			log.Info("restoring.allbytes[%vMB].time[%.2fsec].rates[%.2fMB/sec]...", bytes, diff, rates)
		}
	}()

	wg.Wait()
	elapsed := time.Since(t).Seconds()
	log.Info("restoring.all.done.cost[%.2fsec].allbytes[%.2fMB].rate[%.2fMB/s]", elapsed, float64(bytes/1024/1024), (float64(bytes/1024/1024) / elapsed))
}
