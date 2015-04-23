package db

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/crawl/go-sequell/crawl/version"
	"github.com/crawl/go-sequell/ectx"
	"github.com/crawl/go-sequell/pg"
)

func RenumberVersions(dbc pg.ConnSpec) error {
	c, err := dbc.Open()
	if err != nil {
		return err
	}

	var tables = []struct {
		table           string
		verCol, vnumCol string
	}{
		{"l_version", "v", "vnum"},
		{"l_cversion", "cv", "cvnum"},
		{"l_vlong", "vlong", "vlongnum"},
	}
	for _, tab := range tables {
		err = renumberVersionTable(c, tab.table, tab.verCol, tab.vnumCol)
		if err != nil {
			return err
		}
	}
	return nil
}

func renumberVersionTable(c pg.DB, table, verCol, vnumCol string) error {
	fmt.Printf("Renumbering version table %s: %s => %s\n",
		table, verCol, vnumCol)
	const queueLen = 1000
	versionQueue := make([]string, 0, queueLen)

	versionUpdateStatement := func() string {
		buf := bytes.Buffer{}
		buf.WriteString("update " + table + " t set " + vnumCol +
			" = c.vnum from (values ")
		nvals := len(versionQueue)
		bind := 1
		for i := 0; i < nvals; i++ {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("(")
			buf.WriteString("$" + strconv.Itoa(bind) + ",")
			bind++
			buf.WriteString("$" + strconv.Itoa(bind) + "::bigint")
			bind++
			buf.WriteString(")")
		}
		buf.WriteString(") as c (ver, vnum) where t." + verCol +
			" = c.ver")
		return buf.String()
	}

	versionUpdateBinds := func() []interface{} {
		res := make([]interface{}, len(versionQueue)*2)
		for i, ver := range versionQueue {
			res[i*2] = ver
			res[i*2+1] = version.CachingVersionNumericId(ver)
		}
		return res
	}

	commitVersions := func() error {
		if len(versionQueue) == 0 {
			return nil
		}
		updateStatement := versionUpdateStatement()
		binds := versionUpdateBinds()
		res, err := c.Exec(updateStatement, binds...)
		if err != nil {
			return ectx.Err(fmt.Sprintf("Query: %s (%#v)", updateStatement, binds), err)
		}
		if rowsAffected, err := res.RowsAffected(); err == nil {
			fmt.Println("Updated", rowsAffected, "version rows")
		}
		return nil
	}

	queueVersion := func(ver string) error {
		versionQueue = append(versionQueue, ver)
		if len(versionQueue) >= queueLen {
			if err := commitVersions(); err != nil {
				return err
			}
			versionQueue = versionQueue[0:0]
		}
		return nil
	}

	versionsQuery := "select " + verCol + " from " + table
	verRows, err := c.Query(versionsQuery)
	if err != nil {
		return err
	}
	defer verRows.Close()
	var ver string
	for verRows.Next() {
		if err = verRows.Scan(&ver); err != nil {
			return err
		}
		if err = queueVersion(ver); err != nil {
			return err
		}
	}
	if err = commitVersions(); err != nil {
		return err
	}
	return verRows.Err()
}
