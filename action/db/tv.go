package db

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/greensnark/go-sequell/ectx"
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

func extraIdentField(table string) string {
	if strings.Index(table, "milestone") != -1 {
		return "rtime"
	}
	return "rend"
}

func writeTVData(c pg.DB, table string) error {
	extraField := extraIdentField(table)
	q := "select g.game_key, t.ntv, t." + extraField + " from " + table +
		" as t " +
		`inner join l_game_key g on t.game_key_id = g.id
              where t.ntv > 0`
	rows, err := c.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()
	var gameKey, rowTime string
	var ntv int
	for rows.Next() {
		if err := rows.Scan(&gameKey, &ntv, &rowTime); err != nil {
			return err
		}
		fmt.Printf("%s\t%s\t%s\t%d\n", table, gameKey, rowTime, ntv)
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

	uniqKey := func(gameKey, time string) string {
		return gameKey + "/" + time
	}

	splitUniqKey := func(key string) (string, string) {
		split := strings.SplitN(key, "/", 2)
		return split[0], split[1]
	}

	tableTV := tableKeyTV{}
	pendingCount := 0
	const flushAt = 1000
	reader := bufio.NewReader(os.Stdin)

	tvUpdateQuery := func(table string, nupdates int) string {
		buf := bytes.Buffer{}
		buf.WriteString(
			"update " + table + " as t set ntv = c.ntv " +
				"from (values ")
		nbind := 1
		bindStr := func() string {
			s := "$" + strconv.Itoa(nbind)
			nbind++
			return s
		}
		for i := 0; i < nupdates; i++ {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("(")
			buf.WriteString(bindStr())
			buf.WriteString(",")
			buf.WriteString(bindStr() + "::int")
			buf.WriteString(",")
			buf.WriteString(bindStr())
			buf.WriteString(")")
		}
		buf.WriteString(
			`) as c (game_key, ntv, ttime), l_game_key as k
             where t.game_key_id = k.id
               and c.game_key = k.game_key 
			   and c.ttime = t.` + extraIdentField(table))
		return buf.String()
	}

	tx, err := c.Begin()
	if err != nil {
		return err
	}

	total := 0
	updateTV := func() error {
		if len(tableTV) == 0 {
			pendingCount = 0
			return nil
		}
		for table, keyTVs := range tableTV {
			query := tvUpdateQuery(table, len(keyTVs))
			args := make([]interface{}, len(keyTVs)*3)
			i := 0
			for key, ntv := range keyTVs {
				gameKey, ttime := splitUniqKey(key)
				args[i] = gameKey
				args[i+1] = ntv
				args[i+2] = ttime
				i += 3
			}
			total += len(keyTVs)
			log.Printf("%s: updating %d (total: %d) ntv rows\n", table,
				len(keyTVs), total)
			_, err := tx.Exec(query, args...)
			if err != nil {
				return ectx.Err(
					fmt.Sprintf("Query (%d binds): %s", len(args), query), err)
			}
		}
		tableTV = tableKeyTV{}
		pendingCount = 0
		return nil
	}

	addTableKeyTV := func(table, key, ntv, ttime string) error {
		tableMap := tableTV[table]
		if tableMap == nil {
			tableMap = keyTV{}
			tableTV[table] = tableMap
		}
		tableMap[uniqKey(key, ttime)] = ntv
		pendingCount++
		if pendingCount >= flushAt {
			return updateTV()
		}
		return nil
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			tx.Rollback()
			return err
		}
		line = strings.TrimSpace(line)
		parts := strings.Split(line, "\t")
		if len(parts) != 4 {
			fmt.Fprintf(os.Stderr, "Rejecting malformed line: %s\n", line)
			continue
		}

		table, key, ttime, ntv := parts[0], parts[1], parts[2], parts[3]
		if err := addTableKeyTV(table, key, ntv, ttime); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err = updateTV(); err != nil {
		return err
	}

	return tx.Commit()
}
