package xlog

import (
	"reflect"
	"testing"
)

func TestXlogParseEmpty(t *testing.T) {
	res, err := Parse("   ")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if len(res) != 0 {
		t.Errorf("expected empty map, got %v", res)
	}
}

func TestXlogParseSingleField(t *testing.T) {
	res, err := Parse("  cow =moo ")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if res["cow"] != "moo" {
		t.Errorf("unexpected parse: %v, expected cow=moo", res)
	}
}

func TestXlogParseFull(t *testing.T) {
	line := "v=0.15-a0:vlong=0.15-a0-1506-g1b030bc:lv=0.1:tiles=1:name=Atomikkrab:race=Demonspawn:cls=Fighter:char=DsFi:xl=2:sk=Fighting:sklev=3:title=Skirmisher:place=D::$:br=D:lvl=0:absdepth=1:hp=19:mhp=24:mmhp=24:str=16:int=9:dex=12:ac=6:ev=6:sh=8:start=20140514224835S:dur=66:turn=69:aut=803:kills=3:gold=8:goldfound=8:goldspent=0:sc=8:ktyp=leaving:dam=-9999:sdam=0:tdam=0:end=20140514224942S:map=eino_arrival_water_star:tmsg=got out of the dungeon alive:vmsg=got out of the dungeon alive."
	expectedMap := map[string]string{
		"v":         "0.15-a0",
		"vlong":     "0.15-a0-1506-g1b030bc",
		"lv":        "0.1",
		"tiles":     "1",
		"name":      "Atomikkrab",
		"race":      "Demonspawn",
		"cls":       "Fighter",
		"char":      "DsFi",
		"xl":        "2",
		"sk":        "Fighting",
		"sklev":     "3",
		"title":     "Skirmisher",
		"place":     "D:$",
		"br":        "D",
		"lvl":       "0",
		"absdepth":  "1",
		"hp":        "19",
		"mhp":       "24",
		"mmhp":      "24",
		"str":       "16",
		"int":       "9",
		"dex":       "12",
		"ac":        "6",
		"ev":        "6",
		"sh":        "8",
		"start":     "20140514224835S",
		"dur":       "66",
		"turn":      "69",
		"aut":       "803",
		"kills":     "3",
		"gold":      "8",
		"goldfound": "8",
		"goldspent": "0",
		"sc":        "8",
		"ktyp":      "leaving",
		"dam":       "-9999",
		"sdam":      "0",
		"tdam":      "0",
		"end":       "20140514224942S",
		"map":       "eino_arrival_water_star",
		"tmsg":      "got out of the dungeon alive",
		"vmsg":      "got out of the dungeon alive.",
	}
	res, err := Parse(line)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(res, expectedMap) {
		t.Errorf("parsed map (%d) %v doesn't match expected (%d) %v",
			len(res), res, len(expectedMap), expectedMap)
	}
}

func TestXlogParseBroken(t *testing.T) {
	line := "cow:boy"
	_, err := Parse(line)
	if err == nil {
		t.Errorf("expected error parsing %s, didn't get it", line)
	}

	line = "moo=foo:cow"
	_, err = Parse(line)
	if err == nil {
		t.Errorf("expected error parsing %s, didn't get it", line)
	}
}
