package db

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/greensnark/go-sequell/pg"
)

func ExportTV(db pg.ConnSpec) error {
	c, err := db.Open()
	if err != nil {
		return err
	}
	sch := CrawlSchema()
	for _, table := range sch.PrimaryTableNames() {
		if err := writeTVData(c, table); err != nil {
			return err
		}
	}
	return nil
}

func writeTVData(c pg.DB, table string) error {
	q := "select g.game_key, t.ntv from " + table + " as t " +
		`inner join l_game_key g on t.game_key_id = g.id
              where t.ntv > 0`
	rows, err := c.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()
	var gameKey string
	var ntv int
	for rows.Next() {
		if err := rows.Scan(&gameKey, &ntv); err != nil {
			return err
		}
		fmt.Printf("%s\t%s\t%d\n", table, gameKey, ntv)
	}
	return rows.Err()
}

func ImportTV(db pg.ConnSpec) error {
	c, err := db.Open()
	if err != nil {
		return err
	}

	type keyTV map[string]string
	type tableKeyTV map[string]keyTV

	tableTV := tableKeyTV{}
	pendingCount := 0
	const flushAt = 50000
	reader := bufio.NewReader(os.Stdin)

	tvUpdateQuery := func(table string, nupdates int) string {
		buf := bytes.Buffer{}
		buf.WriteString(
			"update " + table + " as t set ntv = c.ntv " +
				"from (values ")
		nbind := 1
		for i := 0; i < nupdates; i++ {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("(")
			buf.WriteString("$")
			buf.WriteString(strconv.Itoa(nbind))
			nbind++
			buf.WriteString(",")
			buf.WriteString("$")
			buf.WriteString(strconv.Itoa(nbind))
			buf.WriteString("::int")
			nbind++
			buf.WriteString(")")
		}
		buf.WriteString(
			`) as c (game_key, ntv), l_game_key as k
             where t.game_key_id = k.id
               and c.game_key = k.game_key`)
		return buf.String()
	}

	updateTV := func() error {
		if len(tableTV) == 0 {
			return nil
		}
		for table, keyTVs := range tableTV {
			query := tvUpdateQuery(table, len(keyTVs))
			args := make([]interface{}, len(keyTVs)*2)
			i := 0
			for key, ntv := range keyTVs {
				args[i] = key
				args[i+1] = ntv
				i += 2
			}
			_, err := c.Exec(query, args...)
			if err != nil {
				return err
			}
		}
		tableTV = tableKeyTV{}
		pendingCount = 0
		return nil
	}

	addTableKeyTV := func(table, key, ntv string) {
		tableMap := tableTV[table]
		if tableMap == nil {
			tableMap = keyTV{}
			tableTV[table] = tableMap
		}
		tableMap[key] = ntv
		pendingCount++
		if pendingCount >= flushAt {
			updateTV()
		}
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		line = strings.TrimSpace(line)
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			fmt.Fprintf(os.Stderr, "Rejecting malformed line: %s\n", line)
			continue
		}

		table, key, ntv := parts[0], parts[1], parts[2]
		addTableKeyTV(table, key, ntv)
	}

	return updateTV()
}
