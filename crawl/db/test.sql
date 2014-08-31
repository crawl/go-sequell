drop table if exists milestone;
drop table if exists logrecord;
drop table if exists nostalgia_milestone;
drop table if exists nostalgia_logrecord;
drop table if exists zot_milestone;
drop table if exists zot_logrecord;
drop table if exists spr_milestone;
drop table if exists spr_logrecord;
drop table if exists l_oplace;
drop table if exists l_milestone;
drop table if exists l_noun;
drop table if exists l_verb;
drop table if exists l_status;
drop table if exists l_maxskills;
drop table if exists l_mapdesc;
drop table if exists l_god;
drop table if exists l_ltyp;
drop table if exists l_br;
drop table if exists l_place;
drop table if exists l_kmod;
drop table if exists l_kpath;
drop table if exists l_ktyp;
drop table if exists l_title;
drop table if exists l_sk;
drop table if exists l_char;
drop table if exists l_cls;
drop table if exists l_crace;
drop table if exists l_race;
drop table if exists l_name;
drop table if exists l_lv;
drop table if exists l_src;
drop table if exists l_file;
drop table if exists l_game_key;
drop table if exists l_version;
drop table if exists l_msg;
drop table if exists l_kaux;
drop table if exists l_map;
drop table if exists l_killer;
drop table if exists l_cversion;
create table l_cversion (
  id serial unique,
  cv citext unique,
  cvnum numeric(18) unique,
  primary key (id)
);
create table l_killer (
  id serial unique,
  killer citext unique,
  ckiller citext unique,
  ikiller citext unique,
  primary key (id)
);
create table l_map (
  id serial unique,
  mapname citext unique,
  killermap citext unique,
  primary key (id)
);
create table l_kaux (
  id serial unique,
  kaux citext unique,
  ckaux citext unique,
  primary key (id)
);
create table l_msg (
  id serial unique,
  tmsg citext unique,
  vmsg citext unique,
  primary key (id)
);
create table l_version (
  id serial unique,
  v citext unique,
  vnum numeric(18) unique,
  primary key (id)
);
create table l_game_key (
  id serial unique,
  game_key citext unique,
  primary key (id)
);
create table l_file (
  id serial unique,
  file citext unique,
  primary key (id)
);
create table l_src (
  id serial unique,
  src citext unique,
  primary key (id)
);
create table l_lv (
  id serial unique,
  lv citext unique,
  primary key (id)
);
create table l_name (
  id serial unique,
  pname text unique,
  primary key (id)
);
create index ind_l_name_pname on l_name (cast(pname as citext));
create table l_race (
  id serial unique,
  race citext unique,
  primary key (id)
);
create table l_crace (
  id serial unique,
  crace citext unique,
  primary key (id)
);
create table l_cls (
  id serial unique,
  cls citext unique,
  primary key (id)
);
create table l_char (
  id serial unique,
  charabbrev citext unique,
  primary key (id)
);
create table l_sk (
  id serial unique,
  sk citext unique,
  primary key (id)
);
create table l_title (
  id serial unique,
  title citext unique,
  primary key (id)
);
create table l_ktyp (
  id serial unique,
  ktyp citext unique,
  primary key (id)
);
create table l_kpath (
  id serial unique,
  kpath citext unique,
  primary key (id)
);
create table l_kmod (
  id serial unique,
  kmod citext unique,
  primary key (id)
);
create table l_place (
  id serial unique,
  place citext unique,
  primary key (id)
);
create table l_br (
  id serial unique,
  br citext unique,
  primary key (id)
);
create table l_ltyp (
  id serial unique,
  ltyp citext unique,
  primary key (id)
);
create table l_god (
  id serial unique,
  god citext unique,
  primary key (id)
);
create table l_mapdesc (
  id serial unique,
  mapdesc citext unique,
  primary key (id)
);
create table l_maxskills (
  id serial unique,
  maxskills citext unique,
  primary key (id)
);
create table l_status (
  id serial unique,
  status citext unique,
  primary key (id)
);
create table l_verb (
  id serial unique,
  verb citext unique,
  primary key (id)
);
create table l_noun (
  id serial unique,
  noun text unique,
  primary key (id)
);
create index ind_l_noun_noun on l_noun (cast(noun as citext));
create table l_milestone (
  id serial unique,
  milestone citext unique,
  primary key (id)
);
create table l_oplace (
  id serial unique,
  oplace citext unique,
  primary key (id)
);
create table spr_logrecord (
  id serial,
  file_offset bigint default 0,
  game_key_id int,
  file_id int,
  alpha boolean,
  src_id int,
  v_id int,
  cv_id int,
  lv_id int,
  sc bigint default 0,
  pname_id int,
  race_id int,
  crace_id int,
  cls_id int,
  charabbrev_id int,
  xl int default 0,
  sk_id int,
  sklev int default 0,
  title_id int,
  ktyp_id int,
  killer_id int,
  ckiller_id int,
  ikiller_id int,
  kpath_id int,
  kmod_id int,
  kaux_id int,
  ckaux_id int,
  place_id int,
  br_id int,
  lvl int default 0,
  absdepth int default 0,
  ltyp_id int,
  hp int default 0,
  mhp int default 0,
  mmhp int default 0,
  dam int default 0,
  sdam int default 0,
  tdam int default 0,
  sstr int default 0,
  sint int default 0,
  sdex int default 0,
  god_id int,
  piety int default 0,
  pen int default 0,
  wiz int default 0,
  tstart timestamp,
  tend timestamp,
  dur bigint default 0,
  turn int default 0,
  urune int default 0,
  nrune int default 0,
  tmsg_id int,
  vmsg_id int,
  splat boolean,
  rstart citext,
  rend citext,
  ntv int default 0,
  mapname_id int,
  killermap_id int,
  mapdesc_id int,
  tiles boolean,
  gold int default 0,
  goldfound int default 0,
  goldspent int default 0,
  kills int default 0,
  ac int default 0,
  ev int default 0,
  sh int default 0,
  aut int default 0,
  maxskills_id int,
  status_id int,
  primary key (id),
  foreign key (game_key_id) references l_game_key (id),
  foreign key (file_id) references l_file (id),
  foreign key (src_id) references l_src (id),
  foreign key (v_id) references l_version (id),
  foreign key (cv_id) references l_cversion (id),
  foreign key (lv_id) references l_lv (id),
  foreign key (pname_id) references l_name (id),
  foreign key (race_id) references l_race (id),
  foreign key (crace_id) references l_crace (id),
  foreign key (cls_id) references l_cls (id),
  foreign key (charabbrev_id) references l_char (id),
  foreign key (sk_id) references l_sk (id),
  foreign key (title_id) references l_title (id),
  foreign key (ktyp_id) references l_ktyp (id),
  foreign key (killer_id) references l_killer (id),
  foreign key (ckiller_id) references l_killer (id),
  foreign key (ikiller_id) references l_killer (id),
  foreign key (kpath_id) references l_kpath (id),
  foreign key (kmod_id) references l_kmod (id),
  foreign key (kaux_id) references l_kaux (id),
  foreign key (ckaux_id) references l_kaux (id),
  foreign key (place_id) references l_place (id),
  foreign key (br_id) references l_br (id),
  foreign key (ltyp_id) references l_ltyp (id),
  foreign key (god_id) references l_god (id),
  foreign key (tmsg_id) references l_msg (id),
  foreign key (vmsg_id) references l_msg (id),
  foreign key (mapname_id) references l_map (id),
  foreign key (killermap_id) references l_map (id),
  foreign key (mapdesc_id) references l_mapdesc (id),
  foreign key (maxskills_id) references l_maxskills (id),
  foreign key (status_id) references l_status (id)
);
create index ind_spr_logrecord_file_file_offset on spr_logrecord (file, file_offset);
create index ind_spr_logrecord_file_offset on spr_logrecord (file_offset);
create index ind_spr_logrecord_game_key_id on spr_logrecord (game_key_id);
create index ind_spr_logrecord_file_id on spr_logrecord (file_id);
create index ind_spr_logrecord_src_id on spr_logrecord (src_id);
create index ind_spr_logrecord_v_id on spr_logrecord (v_id);
create index ind_spr_logrecord_cv_id on spr_logrecord (cv_id);
create index ind_spr_logrecord_lv_id on spr_logrecord (lv_id);
create index ind_spr_logrecord_sc on spr_logrecord (sc);
create index ind_spr_logrecord_pname_id on spr_logrecord (pname_id);
create index ind_spr_logrecord_race_id on spr_logrecord (race_id);
create index ind_spr_logrecord_crace_id on spr_logrecord (crace_id);
create index ind_spr_logrecord_cls_id on spr_logrecord (cls_id);
create index ind_spr_logrecord_charabbrev_id on spr_logrecord (charabbrev_id);
create index ind_spr_logrecord_xl on spr_logrecord (xl);
create index ind_spr_logrecord_sk_id on spr_logrecord (sk_id);
create index ind_spr_logrecord_sklev on spr_logrecord (sklev);
create index ind_spr_logrecord_title_id on spr_logrecord (title_id);
create index ind_spr_logrecord_ktyp_id on spr_logrecord (ktyp_id);
create index ind_spr_logrecord_killer_id on spr_logrecord (killer_id);
create index ind_spr_logrecord_ckiller_id on spr_logrecord (ckiller_id);
create index ind_spr_logrecord_ikiller_id on spr_logrecord (ikiller_id);
create index ind_spr_logrecord_kpath_id on spr_logrecord (kpath_id);
create index ind_spr_logrecord_kmod_id on spr_logrecord (kmod_id);
create index ind_spr_logrecord_kaux_id on spr_logrecord (kaux_id);
create index ind_spr_logrecord_ckaux_id on spr_logrecord (ckaux_id);
create index ind_spr_logrecord_place_id on spr_logrecord (place_id);
create index ind_spr_logrecord_br_id on spr_logrecord (br_id);
create index ind_spr_logrecord_ltyp_id on spr_logrecord (ltyp_id);
create index ind_spr_logrecord_hp on spr_logrecord (hp);
create index ind_spr_logrecord_mhp on spr_logrecord (mhp);
create index ind_spr_logrecord_god_id on spr_logrecord (god_id);
create index ind_spr_logrecord_tstart on spr_logrecord (tstart);
create index ind_spr_logrecord_tend on spr_logrecord (tend);
create index ind_spr_logrecord_dur on spr_logrecord (dur);
create index ind_spr_logrecord_turn on spr_logrecord (turn);
create index ind_spr_logrecord_urune on spr_logrecord (urune);
create index ind_spr_logrecord_nrune on spr_logrecord (nrune);
create index ind_spr_logrecord_tmsg_id on spr_logrecord (tmsg_id);
create index ind_spr_logrecord_vmsg_id on spr_logrecord (vmsg_id);
create index ind_spr_logrecord_rstart on spr_logrecord (rstart);
create index ind_spr_logrecord_rend on spr_logrecord (rend);
create index ind_spr_logrecord_ntv on spr_logrecord (ntv);
create index ind_spr_logrecord_mapname_id on spr_logrecord (mapname_id);
create index ind_spr_logrecord_killermap_id on spr_logrecord (killermap_id);
create index ind_spr_logrecord_mapdesc_id on spr_logrecord (mapdesc_id);
create index ind_spr_logrecord_maxskills_id on spr_logrecord (maxskills_id);
create index ind_spr_logrecord_status_id on spr_logrecord (status_id);
create table spr_milestone (
  id serial,
  game_key_id int,
  file_offset int default 0,
  file_id int,
  alpha boolean,
  src_id int,
  v_id int,
  cv_id int,
  pname_id int,
  race_id int,
  crace_id int,
  cls_id int,
  charabbrev_id int,
  xl int default 0,
  sk_id int,
  sklev int default 0,
  title_id int,
  place_id int,
  br_id int,
  lvl int default 0,
  absdepth int default 0,
  ltyp_id int,
  hp int default 0,
  mhp int default 0,
  mmhp int default 0,
  sstr int default 0,
  sint int default 0,
  sdex int default 0,
  god_id int,
  dur bigint default 0,
  turn int default 0,
  urune int default 0,
  nrune int default 0,
  ttime timestamp,
  rtime citext,
  tstart timestamp,
  rstart citext,
  verb_id int,
  noun_id int,
  milestone_id int,
  ntv int default 0,
  oplace_id int,
  tiles boolean,
  gold int default 0,
  goldfound int default 0,
  goldspent int default 0,
  kills int default 0,
  ac int default 0,
  ev int default 0,
  sh int default 0,
  aut int default 0,
  maxskills_id int,
  status_id int,
  primary key (id),
  foreign key (game_key_id) references l_game_key (id),
  foreign key (file_id) references l_file (id),
  foreign key (src_id) references l_src (id),
  foreign key (v_id) references l_version (id),
  foreign key (cv_id) references l_cversion (id),
  foreign key (pname_id) references l_name (id),
  foreign key (race_id) references l_race (id),
  foreign key (crace_id) references l_crace (id),
  foreign key (cls_id) references l_cls (id),
  foreign key (charabbrev_id) references l_char (id),
  foreign key (sk_id) references l_sk (id),
  foreign key (title_id) references l_title (id),
  foreign key (place_id) references l_place (id),
  foreign key (br_id) references l_br (id),
  foreign key (ltyp_id) references l_ltyp (id),
  foreign key (god_id) references l_god (id),
  foreign key (verb_id) references l_verb (id),
  foreign key (noun_id) references l_noun (id),
  foreign key (milestone_id) references l_milestone (id),
  foreign key (oplace_id) references l_oplace (id),
  foreign key (maxskills_id) references l_maxskills (id),
  foreign key (status_id) references l_status (id)
);
create index ind_spr_milestone_file_file_offset on spr_milestone (file, file_offset);
create index ind_spr_milestone_verb_noun on spr_milestone (verb, noun);
create index ind_spr_milestone_game_key_id on spr_milestone (game_key_id);
create index ind_spr_milestone_file_offset on spr_milestone (file_offset);
create index ind_spr_milestone_file_id on spr_milestone (file_id);
create index ind_spr_milestone_src_id on spr_milestone (src_id);
create index ind_spr_milestone_v_id on spr_milestone (v_id);
create index ind_spr_milestone_cv_id on spr_milestone (cv_id);
create index ind_spr_milestone_pname_id on spr_milestone (pname_id);
create index ind_spr_milestone_race_id on spr_milestone (race_id);
create index ind_spr_milestone_crace_id on spr_milestone (crace_id);
create index ind_spr_milestone_cls_id on spr_milestone (cls_id);
create index ind_spr_milestone_charabbrev_id on spr_milestone (charabbrev_id);
create index ind_spr_milestone_xl on spr_milestone (xl);
create index ind_spr_milestone_sk_id on spr_milestone (sk_id);
create index ind_spr_milestone_sklev on spr_milestone (sklev);
create index ind_spr_milestone_title_id on spr_milestone (title_id);
create index ind_spr_milestone_place_id on spr_milestone (place_id);
create index ind_spr_milestone_br_id on spr_milestone (br_id);
create index ind_spr_milestone_ltyp_id on spr_milestone (ltyp_id);
create index ind_spr_milestone_hp on spr_milestone (hp);
create index ind_spr_milestone_mhp on spr_milestone (mhp);
create index ind_spr_milestone_god_id on spr_milestone (god_id);
create index ind_spr_milestone_turn on spr_milestone (turn);
create index ind_spr_milestone_urune on spr_milestone (urune);
create index ind_spr_milestone_nrune on spr_milestone (nrune);
create index ind_spr_milestone_ttime on spr_milestone (ttime);
create index ind_spr_milestone_rtime on spr_milestone (rtime);
create index ind_spr_milestone_tstart on spr_milestone (tstart);
create index ind_spr_milestone_rstart on spr_milestone (rstart);
create index ind_spr_milestone_verb_id on spr_milestone (verb_id);
create index ind_spr_milestone_noun_id on spr_milestone (noun_id);
create index ind_spr_milestone_milestone_id on spr_milestone (milestone_id);
create index ind_spr_milestone_ntv on spr_milestone (ntv);
create index ind_spr_milestone_oplace_id on spr_milestone (oplace_id);
create index ind_spr_milestone_maxskills_id on spr_milestone (maxskills_id);
create index ind_spr_milestone_status_id on spr_milestone (status_id);
create table zot_logrecord (
  id serial,
  file_offset bigint default 0,
  game_key_id int,
  file_id int,
  alpha boolean,
  src_id int,
  v_id int,
  cv_id int,
  lv_id int,
  sc bigint default 0,
  pname_id int,
  race_id int,
  crace_id int,
  cls_id int,
  charabbrev_id int,
  xl int default 0,
  sk_id int,
  sklev int default 0,
  title_id int,
  ktyp_id int,
  killer_id int,
  ckiller_id int,
  ikiller_id int,
  kpath_id int,
  kmod_id int,
  kaux_id int,
  ckaux_id int,
  place_id int,
  br_id int,
  lvl int default 0,
  absdepth int default 0,
  ltyp_id int,
  hp int default 0,
  mhp int default 0,
  mmhp int default 0,
  dam int default 0,
  sdam int default 0,
  tdam int default 0,
  sstr int default 0,
  sint int default 0,
  sdex int default 0,
  god_id int,
  piety int default 0,
  pen int default 0,
  wiz int default 0,
  tstart timestamp,
  tend timestamp,
  dur bigint default 0,
  turn int default 0,
  urune int default 0,
  nrune int default 0,
  tmsg_id int,
  vmsg_id int,
  splat boolean,
  rstart citext,
  rend citext,
  ntv int default 0,
  mapname_id int,
  killermap_id int,
  mapdesc_id int,
  tiles boolean,
  gold int default 0,
  goldfound int default 0,
  goldspent int default 0,
  kills int default 0,
  ac int default 0,
  ev int default 0,
  sh int default 0,
  aut int default 0,
  maxskills_id int,
  status_id int,
  primary key (id),
  foreign key (game_key_id) references l_game_key (id),
  foreign key (file_id) references l_file (id),
  foreign key (src_id) references l_src (id),
  foreign key (v_id) references l_version (id),
  foreign key (cv_id) references l_cversion (id),
  foreign key (lv_id) references l_lv (id),
  foreign key (pname_id) references l_name (id),
  foreign key (race_id) references l_race (id),
  foreign key (crace_id) references l_crace (id),
  foreign key (cls_id) references l_cls (id),
  foreign key (charabbrev_id) references l_char (id),
  foreign key (sk_id) references l_sk (id),
  foreign key (title_id) references l_title (id),
  foreign key (ktyp_id) references l_ktyp (id),
  foreign key (killer_id) references l_killer (id),
  foreign key (ckiller_id) references l_killer (id),
  foreign key (ikiller_id) references l_killer (id),
  foreign key (kpath_id) references l_kpath (id),
  foreign key (kmod_id) references l_kmod (id),
  foreign key (kaux_id) references l_kaux (id),
  foreign key (ckaux_id) references l_kaux (id),
  foreign key (place_id) references l_place (id),
  foreign key (br_id) references l_br (id),
  foreign key (ltyp_id) references l_ltyp (id),
  foreign key (god_id) references l_god (id),
  foreign key (tmsg_id) references l_msg (id),
  foreign key (vmsg_id) references l_msg (id),
  foreign key (mapname_id) references l_map (id),
  foreign key (killermap_id) references l_map (id),
  foreign key (mapdesc_id) references l_mapdesc (id),
  foreign key (maxskills_id) references l_maxskills (id),
  foreign key (status_id) references l_status (id)
);
create index ind_zot_logrecord_file_file_offset on zot_logrecord (file, file_offset);
create index ind_zot_logrecord_file_offset on zot_logrecord (file_offset);
create index ind_zot_logrecord_game_key_id on zot_logrecord (game_key_id);
create index ind_zot_logrecord_file_id on zot_logrecord (file_id);
create index ind_zot_logrecord_src_id on zot_logrecord (src_id);
create index ind_zot_logrecord_v_id on zot_logrecord (v_id);
create index ind_zot_logrecord_cv_id on zot_logrecord (cv_id);
create index ind_zot_logrecord_lv_id on zot_logrecord (lv_id);
create index ind_zot_logrecord_sc on zot_logrecord (sc);
create index ind_zot_logrecord_pname_id on zot_logrecord (pname_id);
create index ind_zot_logrecord_race_id on zot_logrecord (race_id);
create index ind_zot_logrecord_crace_id on zot_logrecord (crace_id);
create index ind_zot_logrecord_cls_id on zot_logrecord (cls_id);
create index ind_zot_logrecord_charabbrev_id on zot_logrecord (charabbrev_id);
create index ind_zot_logrecord_xl on zot_logrecord (xl);
create index ind_zot_logrecord_sk_id on zot_logrecord (sk_id);
create index ind_zot_logrecord_sklev on zot_logrecord (sklev);
create index ind_zot_logrecord_title_id on zot_logrecord (title_id);
create index ind_zot_logrecord_ktyp_id on zot_logrecord (ktyp_id);
create index ind_zot_logrecord_killer_id on zot_logrecord (killer_id);
create index ind_zot_logrecord_ckiller_id on zot_logrecord (ckiller_id);
create index ind_zot_logrecord_ikiller_id on zot_logrecord (ikiller_id);
create index ind_zot_logrecord_kpath_id on zot_logrecord (kpath_id);
create index ind_zot_logrecord_kmod_id on zot_logrecord (kmod_id);
create index ind_zot_logrecord_kaux_id on zot_logrecord (kaux_id);
create index ind_zot_logrecord_ckaux_id on zot_logrecord (ckaux_id);
create index ind_zot_logrecord_place_id on zot_logrecord (place_id);
create index ind_zot_logrecord_br_id on zot_logrecord (br_id);
create index ind_zot_logrecord_ltyp_id on zot_logrecord (ltyp_id);
create index ind_zot_logrecord_hp on zot_logrecord (hp);
create index ind_zot_logrecord_mhp on zot_logrecord (mhp);
create index ind_zot_logrecord_god_id on zot_logrecord (god_id);
create index ind_zot_logrecord_tstart on zot_logrecord (tstart);
create index ind_zot_logrecord_tend on zot_logrecord (tend);
create index ind_zot_logrecord_dur on zot_logrecord (dur);
create index ind_zot_logrecord_turn on zot_logrecord (turn);
create index ind_zot_logrecord_urune on zot_logrecord (urune);
create index ind_zot_logrecord_nrune on zot_logrecord (nrune);
create index ind_zot_logrecord_tmsg_id on zot_logrecord (tmsg_id);
create index ind_zot_logrecord_vmsg_id on zot_logrecord (vmsg_id);
create index ind_zot_logrecord_rstart on zot_logrecord (rstart);
create index ind_zot_logrecord_rend on zot_logrecord (rend);
create index ind_zot_logrecord_ntv on zot_logrecord (ntv);
create index ind_zot_logrecord_mapname_id on zot_logrecord (mapname_id);
create index ind_zot_logrecord_killermap_id on zot_logrecord (killermap_id);
create index ind_zot_logrecord_mapdesc_id on zot_logrecord (mapdesc_id);
create index ind_zot_logrecord_maxskills_id on zot_logrecord (maxskills_id);
create index ind_zot_logrecord_status_id on zot_logrecord (status_id);
create table zot_milestone (
  id serial,
  game_key_id int,
  file_offset int default 0,
  file_id int,
  alpha boolean,
  src_id int,
  v_id int,
  cv_id int,
  pname_id int,
  race_id int,
  crace_id int,
  cls_id int,
  charabbrev_id int,
  xl int default 0,
  sk_id int,
  sklev int default 0,
  title_id int,
  place_id int,
  br_id int,
  lvl int default 0,
  absdepth int default 0,
  ltyp_id int,
  hp int default 0,
  mhp int default 0,
  mmhp int default 0,
  sstr int default 0,
  sint int default 0,
  sdex int default 0,
  god_id int,
  dur bigint default 0,
  turn int default 0,
  urune int default 0,
  nrune int default 0,
  ttime timestamp,
  rtime citext,
  tstart timestamp,
  rstart citext,
  verb_id int,
  noun_id int,
  milestone_id int,
  ntv int default 0,
  oplace_id int,
  tiles boolean,
  gold int default 0,
  goldfound int default 0,
  goldspent int default 0,
  kills int default 0,
  ac int default 0,
  ev int default 0,
  sh int default 0,
  aut int default 0,
  maxskills_id int,
  status_id int,
  primary key (id),
  foreign key (game_key_id) references l_game_key (id),
  foreign key (file_id) references l_file (id),
  foreign key (src_id) references l_src (id),
  foreign key (v_id) references l_version (id),
  foreign key (cv_id) references l_cversion (id),
  foreign key (pname_id) references l_name (id),
  foreign key (race_id) references l_race (id),
  foreign key (crace_id) references l_crace (id),
  foreign key (cls_id) references l_cls (id),
  foreign key (charabbrev_id) references l_char (id),
  foreign key (sk_id) references l_sk (id),
  foreign key (title_id) references l_title (id),
  foreign key (place_id) references l_place (id),
  foreign key (br_id) references l_br (id),
  foreign key (ltyp_id) references l_ltyp (id),
  foreign key (god_id) references l_god (id),
  foreign key (verb_id) references l_verb (id),
  foreign key (noun_id) references l_noun (id),
  foreign key (milestone_id) references l_milestone (id),
  foreign key (oplace_id) references l_oplace (id),
  foreign key (maxskills_id) references l_maxskills (id),
  foreign key (status_id) references l_status (id)
);
create index ind_zot_milestone_file_file_offset on zot_milestone (file, file_offset);
create index ind_zot_milestone_verb_noun on zot_milestone (verb, noun);
create index ind_zot_milestone_game_key_id on zot_milestone (game_key_id);
create index ind_zot_milestone_file_offset on zot_milestone (file_offset);
create index ind_zot_milestone_file_id on zot_milestone (file_id);
create index ind_zot_milestone_src_id on zot_milestone (src_id);
create index ind_zot_milestone_v_id on zot_milestone (v_id);
create index ind_zot_milestone_cv_id on zot_milestone (cv_id);
create index ind_zot_milestone_pname_id on zot_milestone (pname_id);
create index ind_zot_milestone_race_id on zot_milestone (race_id);
create index ind_zot_milestone_crace_id on zot_milestone (crace_id);
create index ind_zot_milestone_cls_id on zot_milestone (cls_id);
create index ind_zot_milestone_charabbrev_id on zot_milestone (charabbrev_id);
create index ind_zot_milestone_xl on zot_milestone (xl);
create index ind_zot_milestone_sk_id on zot_milestone (sk_id);
create index ind_zot_milestone_sklev on zot_milestone (sklev);
create index ind_zot_milestone_title_id on zot_milestone (title_id);
create index ind_zot_milestone_place_id on zot_milestone (place_id);
create index ind_zot_milestone_br_id on zot_milestone (br_id);
create index ind_zot_milestone_ltyp_id on zot_milestone (ltyp_id);
create index ind_zot_milestone_hp on zot_milestone (hp);
create index ind_zot_milestone_mhp on zot_milestone (mhp);
create index ind_zot_milestone_god_id on zot_milestone (god_id);
create index ind_zot_milestone_turn on zot_milestone (turn);
create index ind_zot_milestone_urune on zot_milestone (urune);
create index ind_zot_milestone_nrune on zot_milestone (nrune);
create index ind_zot_milestone_ttime on zot_milestone (ttime);
create index ind_zot_milestone_rtime on zot_milestone (rtime);
create index ind_zot_milestone_tstart on zot_milestone (tstart);
create index ind_zot_milestone_rstart on zot_milestone (rstart);
create index ind_zot_milestone_verb_id on zot_milestone (verb_id);
create index ind_zot_milestone_noun_id on zot_milestone (noun_id);
create index ind_zot_milestone_milestone_id on zot_milestone (milestone_id);
create index ind_zot_milestone_ntv on zot_milestone (ntv);
create index ind_zot_milestone_oplace_id on zot_milestone (oplace_id);
create index ind_zot_milestone_maxskills_id on zot_milestone (maxskills_id);
create index ind_zot_milestone_status_id on zot_milestone (status_id);
create table nostalgia_logrecord (
  id serial,
  file_offset bigint default 0,
  game_key_id int,
  file_id int,
  alpha boolean,
  src_id int,
  v_id int,
  cv_id int,
  lv_id int,
  sc bigint default 0,
  pname_id int,
  race_id int,
  crace_id int,
  cls_id int,
  charabbrev_id int,
  xl int default 0,
  sk_id int,
  sklev int default 0,
  title_id int,
  ktyp_id int,
  killer_id int,
  ckiller_id int,
  ikiller_id int,
  kpath_id int,
  kmod_id int,
  kaux_id int,
  ckaux_id int,
  place_id int,
  br_id int,
  lvl int default 0,
  absdepth int default 0,
  ltyp_id int,
  hp int default 0,
  mhp int default 0,
  mmhp int default 0,
  dam int default 0,
  sdam int default 0,
  tdam int default 0,
  sstr int default 0,
  sint int default 0,
  sdex int default 0,
  god_id int,
  piety int default 0,
  pen int default 0,
  wiz int default 0,
  tstart timestamp,
  tend timestamp,
  dur bigint default 0,
  turn int default 0,
  urune int default 0,
  nrune int default 0,
  tmsg_id int,
  vmsg_id int,
  splat boolean,
  rstart citext,
  rend citext,
  ntv int default 0,
  mapname_id int,
  killermap_id int,
  mapdesc_id int,
  tiles boolean,
  gold int default 0,
  goldfound int default 0,
  goldspent int default 0,
  kills int default 0,
  ac int default 0,
  ev int default 0,
  sh int default 0,
  aut int default 0,
  maxskills_id int,
  status_id int,
  primary key (id),
  foreign key (game_key_id) references l_game_key (id),
  foreign key (file_id) references l_file (id),
  foreign key (src_id) references l_src (id),
  foreign key (v_id) references l_version (id),
  foreign key (cv_id) references l_cversion (id),
  foreign key (lv_id) references l_lv (id),
  foreign key (pname_id) references l_name (id),
  foreign key (race_id) references l_race (id),
  foreign key (crace_id) references l_crace (id),
  foreign key (cls_id) references l_cls (id),
  foreign key (charabbrev_id) references l_char (id),
  foreign key (sk_id) references l_sk (id),
  foreign key (title_id) references l_title (id),
  foreign key (ktyp_id) references l_ktyp (id),
  foreign key (killer_id) references l_killer (id),
  foreign key (ckiller_id) references l_killer (id),
  foreign key (ikiller_id) references l_killer (id),
  foreign key (kpath_id) references l_kpath (id),
  foreign key (kmod_id) references l_kmod (id),
  foreign key (kaux_id) references l_kaux (id),
  foreign key (ckaux_id) references l_kaux (id),
  foreign key (place_id) references l_place (id),
  foreign key (br_id) references l_br (id),
  foreign key (ltyp_id) references l_ltyp (id),
  foreign key (god_id) references l_god (id),
  foreign key (tmsg_id) references l_msg (id),
  foreign key (vmsg_id) references l_msg (id),
  foreign key (mapname_id) references l_map (id),
  foreign key (killermap_id) references l_map (id),
  foreign key (mapdesc_id) references l_mapdesc (id),
  foreign key (maxskills_id) references l_maxskills (id),
  foreign key (status_id) references l_status (id)
);
create index ind_nostalgia_logrecord_file_file_offset on nostalgia_logrecord (file, file_offset);
create index ind_nostalgia_logrecord_file_offset on nostalgia_logrecord (file_offset);
create index ind_nostalgia_logrecord_game_key_id on nostalgia_logrecord (game_key_id);
create index ind_nostalgia_logrecord_file_id on nostalgia_logrecord (file_id);
create index ind_nostalgia_logrecord_src_id on nostalgia_logrecord (src_id);
create index ind_nostalgia_logrecord_v_id on nostalgia_logrecord (v_id);
create index ind_nostalgia_logrecord_cv_id on nostalgia_logrecord (cv_id);
create index ind_nostalgia_logrecord_lv_id on nostalgia_logrecord (lv_id);
create index ind_nostalgia_logrecord_sc on nostalgia_logrecord (sc);
create index ind_nostalgia_logrecord_pname_id on nostalgia_logrecord (pname_id);
create index ind_nostalgia_logrecord_race_id on nostalgia_logrecord (race_id);
create index ind_nostalgia_logrecord_crace_id on nostalgia_logrecord (crace_id);
create index ind_nostalgia_logrecord_cls_id on nostalgia_logrecord (cls_id);
create index ind_nostalgia_logrecord_charabbrev_id on nostalgia_logrecord (charabbrev_id);
create index ind_nostalgia_logrecord_xl on nostalgia_logrecord (xl);
create index ind_nostalgia_logrecord_sk_id on nostalgia_logrecord (sk_id);
create index ind_nostalgia_logrecord_sklev on nostalgia_logrecord (sklev);
create index ind_nostalgia_logrecord_title_id on nostalgia_logrecord (title_id);
create index ind_nostalgia_logrecord_ktyp_id on nostalgia_logrecord (ktyp_id);
create index ind_nostalgia_logrecord_killer_id on nostalgia_logrecord (killer_id);
create index ind_nostalgia_logrecord_ckiller_id on nostalgia_logrecord (ckiller_id);
create index ind_nostalgia_logrecord_ikiller_id on nostalgia_logrecord (ikiller_id);
create index ind_nostalgia_logrecord_kpath_id on nostalgia_logrecord (kpath_id);
create index ind_nostalgia_logrecord_kmod_id on nostalgia_logrecord (kmod_id);
create index ind_nostalgia_logrecord_kaux_id on nostalgia_logrecord (kaux_id);
create index ind_nostalgia_logrecord_ckaux_id on nostalgia_logrecord (ckaux_id);
create index ind_nostalgia_logrecord_place_id on nostalgia_logrecord (place_id);
create index ind_nostalgia_logrecord_br_id on nostalgia_logrecord (br_id);
create index ind_nostalgia_logrecord_ltyp_id on nostalgia_logrecord (ltyp_id);
create index ind_nostalgia_logrecord_hp on nostalgia_logrecord (hp);
create index ind_nostalgia_logrecord_mhp on nostalgia_logrecord (mhp);
create index ind_nostalgia_logrecord_god_id on nostalgia_logrecord (god_id);
create index ind_nostalgia_logrecord_tstart on nostalgia_logrecord (tstart);
create index ind_nostalgia_logrecord_tend on nostalgia_logrecord (tend);
create index ind_nostalgia_logrecord_dur on nostalgia_logrecord (dur);
create index ind_nostalgia_logrecord_turn on nostalgia_logrecord (turn);
create index ind_nostalgia_logrecord_urune on nostalgia_logrecord (urune);
create index ind_nostalgia_logrecord_nrune on nostalgia_logrecord (nrune);
create index ind_nostalgia_logrecord_tmsg_id on nostalgia_logrecord (tmsg_id);
create index ind_nostalgia_logrecord_vmsg_id on nostalgia_logrecord (vmsg_id);
create index ind_nostalgia_logrecord_rstart on nostalgia_logrecord (rstart);
create index ind_nostalgia_logrecord_rend on nostalgia_logrecord (rend);
create index ind_nostalgia_logrecord_ntv on nostalgia_logrecord (ntv);
create index ind_nostalgia_logrecord_mapname_id on nostalgia_logrecord (mapname_id);
create index ind_nostalgia_logrecord_killermap_id on nostalgia_logrecord (killermap_id);
create index ind_nostalgia_logrecord_mapdesc_id on nostalgia_logrecord (mapdesc_id);
create index ind_nostalgia_logrecord_maxskills_id on nostalgia_logrecord (maxskills_id);
create index ind_nostalgia_logrecord_status_id on nostalgia_logrecord (status_id);
create table nostalgia_milestone (
  id serial,
  game_key_id int,
  file_offset int default 0,
  file_id int,
  alpha boolean,
  src_id int,
  v_id int,
  cv_id int,
  pname_id int,
  race_id int,
  crace_id int,
  cls_id int,
  charabbrev_id int,
  xl int default 0,
  sk_id int,
  sklev int default 0,
  title_id int,
  place_id int,
  br_id int,
  lvl int default 0,
  absdepth int default 0,
  ltyp_id int,
  hp int default 0,
  mhp int default 0,
  mmhp int default 0,
  sstr int default 0,
  sint int default 0,
  sdex int default 0,
  god_id int,
  dur bigint default 0,
  turn int default 0,
  urune int default 0,
  nrune int default 0,
  ttime timestamp,
  rtime citext,
  tstart timestamp,
  rstart citext,
  verb_id int,
  noun_id int,
  milestone_id int,
  ntv int default 0,
  oplace_id int,
  tiles boolean,
  gold int default 0,
  goldfound int default 0,
  goldspent int default 0,
  kills int default 0,
  ac int default 0,
  ev int default 0,
  sh int default 0,
  aut int default 0,
  maxskills_id int,
  status_id int,
  primary key (id),
  foreign key (game_key_id) references l_game_key (id),
  foreign key (file_id) references l_file (id),
  foreign key (src_id) references l_src (id),
  foreign key (v_id) references l_version (id),
  foreign key (cv_id) references l_cversion (id),
  foreign key (pname_id) references l_name (id),
  foreign key (race_id) references l_race (id),
  foreign key (crace_id) references l_crace (id),
  foreign key (cls_id) references l_cls (id),
  foreign key (charabbrev_id) references l_char (id),
  foreign key (sk_id) references l_sk (id),
  foreign key (title_id) references l_title (id),
  foreign key (place_id) references l_place (id),
  foreign key (br_id) references l_br (id),
  foreign key (ltyp_id) references l_ltyp (id),
  foreign key (god_id) references l_god (id),
  foreign key (verb_id) references l_verb (id),
  foreign key (noun_id) references l_noun (id),
  foreign key (milestone_id) references l_milestone (id),
  foreign key (oplace_id) references l_oplace (id),
  foreign key (maxskills_id) references l_maxskills (id),
  foreign key (status_id) references l_status (id)
);
create index ind_nostalgia_milestone_file_file_offset on nostalgia_milestone (file, file_offset);
create index ind_nostalgia_milestone_verb_noun on nostalgia_milestone (verb, noun);
create index ind_nostalgia_milestone_game_key_id on nostalgia_milestone (game_key_id);
create index ind_nostalgia_milestone_file_offset on nostalgia_milestone (file_offset);
create index ind_nostalgia_milestone_file_id on nostalgia_milestone (file_id);
create index ind_nostalgia_milestone_src_id on nostalgia_milestone (src_id);
create index ind_nostalgia_milestone_v_id on nostalgia_milestone (v_id);
create index ind_nostalgia_milestone_cv_id on nostalgia_milestone (cv_id);
create index ind_nostalgia_milestone_pname_id on nostalgia_milestone (pname_id);
create index ind_nostalgia_milestone_race_id on nostalgia_milestone (race_id);
create index ind_nostalgia_milestone_crace_id on nostalgia_milestone (crace_id);
create index ind_nostalgia_milestone_cls_id on nostalgia_milestone (cls_id);
create index ind_nostalgia_milestone_charabbrev_id on nostalgia_milestone (charabbrev_id);
create index ind_nostalgia_milestone_xl on nostalgia_milestone (xl);
create index ind_nostalgia_milestone_sk_id on nostalgia_milestone (sk_id);
create index ind_nostalgia_milestone_sklev on nostalgia_milestone (sklev);
create index ind_nostalgia_milestone_title_id on nostalgia_milestone (title_id);
create index ind_nostalgia_milestone_place_id on nostalgia_milestone (place_id);
create index ind_nostalgia_milestone_br_id on nostalgia_milestone (br_id);
create index ind_nostalgia_milestone_ltyp_id on nostalgia_milestone (ltyp_id);
create index ind_nostalgia_milestone_hp on nostalgia_milestone (hp);
create index ind_nostalgia_milestone_mhp on nostalgia_milestone (mhp);
create index ind_nostalgia_milestone_god_id on nostalgia_milestone (god_id);
create index ind_nostalgia_milestone_turn on nostalgia_milestone (turn);
create index ind_nostalgia_milestone_urune on nostalgia_milestone (urune);
create index ind_nostalgia_milestone_nrune on nostalgia_milestone (nrune);
create index ind_nostalgia_milestone_ttime on nostalgia_milestone (ttime);
create index ind_nostalgia_milestone_rtime on nostalgia_milestone (rtime);
create index ind_nostalgia_milestone_tstart on nostalgia_milestone (tstart);
create index ind_nostalgia_milestone_rstart on nostalgia_milestone (rstart);
create index ind_nostalgia_milestone_verb_id on nostalgia_milestone (verb_id);
create index ind_nostalgia_milestone_noun_id on nostalgia_milestone (noun_id);
create index ind_nostalgia_milestone_milestone_id on nostalgia_milestone (milestone_id);
create index ind_nostalgia_milestone_ntv on nostalgia_milestone (ntv);
create index ind_nostalgia_milestone_oplace_id on nostalgia_milestone (oplace_id);
create index ind_nostalgia_milestone_maxskills_id on nostalgia_milestone (maxskills_id);
create index ind_nostalgia_milestone_status_id on nostalgia_milestone (status_id);
create table logrecord (
  id serial,
  file_offset bigint default 0,
  game_key_id int,
  file_id int,
  alpha boolean,
  src_id int,
  v_id int,
  cv_id int,
  lv_id int,
  sc bigint default 0,
  pname_id int,
  race_id int,
  crace_id int,
  cls_id int,
  charabbrev_id int,
  xl int default 0,
  sk_id int,
  sklev int default 0,
  title_id int,
  ktyp_id int,
  killer_id int,
  ckiller_id int,
  ikiller_id int,
  kpath_id int,
  kmod_id int,
  kaux_id int,
  ckaux_id int,
  place_id int,
  br_id int,
  lvl int default 0,
  absdepth int default 0,
  ltyp_id int,
  hp int default 0,
  mhp int default 0,
  mmhp int default 0,
  dam int default 0,
  sdam int default 0,
  tdam int default 0,
  sstr int default 0,
  sint int default 0,
  sdex int default 0,
  god_id int,
  piety int default 0,
  pen int default 0,
  wiz int default 0,
  tstart timestamp,
  tend timestamp,
  dur bigint default 0,
  turn int default 0,
  urune int default 0,
  nrune int default 0,
  tmsg_id int,
  vmsg_id int,
  splat boolean,
  rstart citext,
  rend citext,
  ntv int default 0,
  mapname_id int,
  killermap_id int,
  mapdesc_id int,
  tiles boolean,
  gold int default 0,
  goldfound int default 0,
  goldspent int default 0,
  kills int default 0,
  ac int default 0,
  ev int default 0,
  sh int default 0,
  aut int default 0,
  maxskills_id int,
  status_id int,
  primary key (id),
  foreign key (game_key_id) references l_game_key (id),
  foreign key (file_id) references l_file (id),
  foreign key (src_id) references l_src (id),
  foreign key (v_id) references l_version (id),
  foreign key (cv_id) references l_cversion (id),
  foreign key (lv_id) references l_lv (id),
  foreign key (pname_id) references l_name (id),
  foreign key (race_id) references l_race (id),
  foreign key (crace_id) references l_crace (id),
  foreign key (cls_id) references l_cls (id),
  foreign key (charabbrev_id) references l_char (id),
  foreign key (sk_id) references l_sk (id),
  foreign key (title_id) references l_title (id),
  foreign key (ktyp_id) references l_ktyp (id),
  foreign key (killer_id) references l_killer (id),
  foreign key (ckiller_id) references l_killer (id),
  foreign key (ikiller_id) references l_killer (id),
  foreign key (kpath_id) references l_kpath (id),
  foreign key (kmod_id) references l_kmod (id),
  foreign key (kaux_id) references l_kaux (id),
  foreign key (ckaux_id) references l_kaux (id),
  foreign key (place_id) references l_place (id),
  foreign key (br_id) references l_br (id),
  foreign key (ltyp_id) references l_ltyp (id),
  foreign key (god_id) references l_god (id),
  foreign key (tmsg_id) references l_msg (id),
  foreign key (vmsg_id) references l_msg (id),
  foreign key (mapname_id) references l_map (id),
  foreign key (killermap_id) references l_map (id),
  foreign key (mapdesc_id) references l_mapdesc (id),
  foreign key (maxskills_id) references l_maxskills (id),
  foreign key (status_id) references l_status (id)
);
create index ind_logrecord_file_file_offset on logrecord (file, file_offset);
create index ind_logrecord_file_offset on logrecord (file_offset);
create index ind_logrecord_game_key_id on logrecord (game_key_id);
create index ind_logrecord_file_id on logrecord (file_id);
create index ind_logrecord_src_id on logrecord (src_id);
create index ind_logrecord_v_id on logrecord (v_id);
create index ind_logrecord_cv_id on logrecord (cv_id);
create index ind_logrecord_lv_id on logrecord (lv_id);
create index ind_logrecord_sc on logrecord (sc);
create index ind_logrecord_pname_id on logrecord (pname_id);
create index ind_logrecord_race_id on logrecord (race_id);
create index ind_logrecord_crace_id on logrecord (crace_id);
create index ind_logrecord_cls_id on logrecord (cls_id);
create index ind_logrecord_charabbrev_id on logrecord (charabbrev_id);
create index ind_logrecord_xl on logrecord (xl);
create index ind_logrecord_sk_id on logrecord (sk_id);
create index ind_logrecord_sklev on logrecord (sklev);
create index ind_logrecord_title_id on logrecord (title_id);
create index ind_logrecord_ktyp_id on logrecord (ktyp_id);
create index ind_logrecord_killer_id on logrecord (killer_id);
create index ind_logrecord_ckiller_id on logrecord (ckiller_id);
create index ind_logrecord_ikiller_id on logrecord (ikiller_id);
create index ind_logrecord_kpath_id on logrecord (kpath_id);
create index ind_logrecord_kmod_id on logrecord (kmod_id);
create index ind_logrecord_kaux_id on logrecord (kaux_id);
create index ind_logrecord_ckaux_id on logrecord (ckaux_id);
create index ind_logrecord_place_id on logrecord (place_id);
create index ind_logrecord_br_id on logrecord (br_id);
create index ind_logrecord_ltyp_id on logrecord (ltyp_id);
create index ind_logrecord_hp on logrecord (hp);
create index ind_logrecord_mhp on logrecord (mhp);
create index ind_logrecord_god_id on logrecord (god_id);
create index ind_logrecord_tstart on logrecord (tstart);
create index ind_logrecord_tend on logrecord (tend);
create index ind_logrecord_dur on logrecord (dur);
create index ind_logrecord_turn on logrecord (turn);
create index ind_logrecord_urune on logrecord (urune);
create index ind_logrecord_nrune on logrecord (nrune);
create index ind_logrecord_tmsg_id on logrecord (tmsg_id);
create index ind_logrecord_vmsg_id on logrecord (vmsg_id);
create index ind_logrecord_rstart on logrecord (rstart);
create index ind_logrecord_rend on logrecord (rend);
create index ind_logrecord_ntv on logrecord (ntv);
create index ind_logrecord_mapname_id on logrecord (mapname_id);
create index ind_logrecord_killermap_id on logrecord (killermap_id);
create index ind_logrecord_mapdesc_id on logrecord (mapdesc_id);
create index ind_logrecord_maxskills_id on logrecord (maxskills_id);
create index ind_logrecord_status_id on logrecord (status_id);
create table milestone (
  id serial,
  game_key_id int,
  file_offset int default 0,
  file_id int,
  alpha boolean,
  src_id int,
  v_id int,
  cv_id int,
  pname_id int,
  race_id int,
  crace_id int,
  cls_id int,
  charabbrev_id int,
  xl int default 0,
  sk_id int,
  sklev int default 0,
  title_id int,
  place_id int,
  br_id int,
  lvl int default 0,
  absdepth int default 0,
  ltyp_id int,
  hp int default 0,
  mhp int default 0,
  mmhp int default 0,
  sstr int default 0,
  sint int default 0,
  sdex int default 0,
  god_id int,
  dur bigint default 0,
  turn int default 0,
  urune int default 0,
  nrune int default 0,
  ttime timestamp,
  rtime citext,
  tstart timestamp,
  rstart citext,
  verb_id int,
  noun_id int,
  milestone_id int,
  ntv int default 0,
  oplace_id int,
  tiles boolean,
  gold int default 0,
  goldfound int default 0,
  goldspent int default 0,
  kills int default 0,
  ac int default 0,
  ev int default 0,
  sh int default 0,
  aut int default 0,
  maxskills_id int,
  status_id int,
  primary key (id),
  foreign key (game_key_id) references l_game_key (id),
  foreign key (file_id) references l_file (id),
  foreign key (src_id) references l_src (id),
  foreign key (v_id) references l_version (id),
  foreign key (cv_id) references l_cversion (id),
  foreign key (pname_id) references l_name (id),
  foreign key (race_id) references l_race (id),
  foreign key (crace_id) references l_crace (id),
  foreign key (cls_id) references l_cls (id),
  foreign key (charabbrev_id) references l_char (id),
  foreign key (sk_id) references l_sk (id),
  foreign key (title_id) references l_title (id),
  foreign key (place_id) references l_place (id),
  foreign key (br_id) references l_br (id),
  foreign key (ltyp_id) references l_ltyp (id),
  foreign key (god_id) references l_god (id),
  foreign key (verb_id) references l_verb (id),
  foreign key (noun_id) references l_noun (id),
  foreign key (milestone_id) references l_milestone (id),
  foreign key (oplace_id) references l_oplace (id),
  foreign key (maxskills_id) references l_maxskills (id),
  foreign key (status_id) references l_status (id)
);
create index ind_milestone_file_file_offset on milestone (file, file_offset);
create index ind_milestone_verb_noun on milestone (verb, noun);
create index ind_milestone_game_key_id on milestone (game_key_id);
create index ind_milestone_file_offset on milestone (file_offset);
create index ind_milestone_file_id on milestone (file_id);
create index ind_milestone_src_id on milestone (src_id);
create index ind_milestone_v_id on milestone (v_id);
create index ind_milestone_cv_id on milestone (cv_id);
create index ind_milestone_pname_id on milestone (pname_id);
create index ind_milestone_race_id on milestone (race_id);
create index ind_milestone_crace_id on milestone (crace_id);
create index ind_milestone_cls_id on milestone (cls_id);
create index ind_milestone_charabbrev_id on milestone (charabbrev_id);
create index ind_milestone_xl on milestone (xl);
create index ind_milestone_sk_id on milestone (sk_id);
create index ind_milestone_sklev on milestone (sklev);
create index ind_milestone_title_id on milestone (title_id);
create index ind_milestone_place_id on milestone (place_id);
create index ind_milestone_br_id on milestone (br_id);
create index ind_milestone_ltyp_id on milestone (ltyp_id);
create index ind_milestone_hp on milestone (hp);
create index ind_milestone_mhp on milestone (mhp);
create index ind_milestone_god_id on milestone (god_id);
create index ind_milestone_turn on milestone (turn);
create index ind_milestone_urune on milestone (urune);
create index ind_milestone_nrune on milestone (nrune);
create index ind_milestone_ttime on milestone (ttime);
create index ind_milestone_rtime on milestone (rtime);
create index ind_milestone_tstart on milestone (tstart);
create index ind_milestone_rstart on milestone (rstart);
create index ind_milestone_verb_id on milestone (verb_id);
create index ind_milestone_noun_id on milestone (noun_id);
create index ind_milestone_milestone_id on milestone (milestone_id);
create index ind_milestone_ntv on milestone (ntv);
create index ind_milestone_oplace_id on milestone (oplace_id);
create index ind_milestone_maxskills_id on milestone (maxskills_id);
create index ind_milestone_status_id on milestone (status_id);
