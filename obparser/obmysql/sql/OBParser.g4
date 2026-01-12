parser grammar OBParser;


options { tokenVocab=OBLexer; }


@parser::members {
var is_pl_parse_ bool = false
var is_pl_parse_expr_ bool = false
}


// start rule: sql_stmt

sql_stmt
    : stmt_list
    ;

stmt_list
    : EOF
    | DELIMITER
    | stmt EOF
    | stmt DELIMITER EOF?
    ;

stmt
    : select_stmt
    | insert_stmt
    | create_table_stmt
    | create_function_stmt
    | drop_function_stmt
    | create_table_like_stmt
    | create_database_stmt
    | drop_database_stmt
    | alter_database_stmt
    | use_database_stmt
    | update_stmt
    | delete_stmt
    | drop_table_stmt
    | drop_view_stmt
    | explain_stmt
    | create_outline_stmt
    | alter_outline_stmt
    | drop_outline_stmt
    | show_stmt
    | prepare_stmt
    | variable_set_stmt
    | execute_stmt
    | alter_table_stmt
    | alter_system_stmt
    | deallocate_prepare_stmt
    | create_user_stmt
    | drop_user_stmt
    | set_password_stmt
    | rename_user_stmt
    | lock_user_stmt
    | grant_stmt
    | revoke_stmt
    | begin_stmt
    | commit_stmt
    | rollback_stmt
    | create_tablespace_stmt
    | drop_tablespace_stmt
    | alter_tablespace_stmt
    | rotate_master_key_stmt
    | create_index_stmt
    | drop_index_stmt
    | kill_stmt
    | help_stmt
    | create_view_stmt
    | create_tenant_stmt
    | alter_tenant_stmt
    | drop_tenant_stmt
    | create_resource_stmt
    | alter_resource_stmt
    | drop_resource_stmt
    | set_names_stmt
    | set_charset_stmt
    | create_tablegroup_stmt
    | drop_tablegroup_stmt
    | alter_tablegroup_stmt
    | rename_table_stmt
    | truncate_table_stmt
    | set_transaction_stmt
    | create_synonym_stmt
    | drop_synonym_stmt
    | create_savepoint_stmt
    | rollback_savepoint_stmt
    | release_savepoint_stmt
    | lock_tables_stmt
    | unlock_tables_stmt
    | flashback_stmt
    | purge_stmt
    | analyze_stmt
    | load_data_stmt
    | switchover_cluster_stmt
    | alter_cluster_stmt
    | optimize_stmt
    | dump_memory_stmt
    ;

expr_list
    : expr (Comma expr)*
    ;

expr_as_list
    : expr_with_opt_alias (Comma expr_with_opt_alias)*
    ;

expr_with_opt_alias
    : expr
    | expr AS? column_label
    | expr AS? STRING_VALUE
    ;

column_ref
    : column_name
    | relation_name Dot column_name
    | relation_name Dot Star
    | relation_name Dot relation_name Dot column_name
    | relation_name Dot relation_name Dot Star
    | Dot relation_name Dot column_name
    ;

complex_string_literal
    : charset_introducer? STRING_VALUE
    | charset_introducer HEX_STRING_VALUE
    ;

charset_introducer
    : UnderlineUTF8
    | UnderlineUTF8MB4
    | UnderlineBINARY
    | UnderlineGBK
    | UnderlineGB18030
    | UnderlineUTF16
    ;

literal
    : complex_string_literal
    | DATE_VALUE
    | TIMESTAMP_VALUE
    | INTNUM
    | APPROXNUM
    | DECIMAL_VAL
    | BOOL_VALUE
    | NULLX
    | HEX_STRING_VALUE
    ;

number_literal
    : INTNUM
    | DECIMAL_VAL
    ;

expr_const
    : literal
    | SYSTEM_VARIABLE
    | QUESTIONMARK
    | global_or_session_alias Dot column_name
    ;

conf_const
    : STRING_VALUE
    | DATE_VALUE
    | TIMESTAMP_VALUE
    | INTNUM
    | APPROXNUM
    | DECIMAL_VAL
    | BOOL_VALUE
    | NULLX
    | SYSTEM_VARIABLE
    | global_or_session_alias Dot column_name
    | Minus INTNUM
    | Minus DECIMAL_VAL
    ;

global_or_session_alias
    : GLOBAL_ALIAS
    | SESSION_ALIAS
    ;

bool_pri
    : bool_pri IS NULLX
    | bool_pri IS not NULLX
    | bool_pri COMP_LE predicate
    | bool_pri COMP_LE sub_query_flag LeftParen select_no_parens RightParen
    | bool_pri COMP_LT predicate
    | bool_pri COMP_LT sub_query_flag LeftParen select_no_parens RightParen
    | bool_pri COMP_EQ predicate
    | bool_pri COMP_EQ sub_query_flag LeftParen select_no_parens RightParen
    | bool_pri COMP_NSEQ predicate
    | bool_pri COMP_GE predicate
    | bool_pri COMP_GE sub_query_flag LeftParen select_no_parens RightParen
    | bool_pri COMP_GT predicate
    | bool_pri COMP_GT sub_query_flag LeftParen select_no_parens RightParen
    | bool_pri COMP_NE predicate
    | bool_pri COMP_NE sub_query_flag LeftParen select_no_parens RightParen
    | predicate
    ;

predicate
    : bit_expr IN in_expr
    | bit_expr not IN in_expr
    | bit_expr not BETWEEN bit_expr AND predicate
    | bit_expr BETWEEN bit_expr AND predicate
    | bit_expr LIKE simple_expr
    | bit_expr LIKE simple_expr ESCAPE simple_expr
    | bit_expr not LIKE simple_expr
    | bit_expr not LIKE simple_expr ESCAPE simple_expr
    | bit_expr REGEXP bit_expr
    | bit_expr not REGEXP bit_expr
    | bit_expr
    ;

bit_expr
    : bit_expr Or bit_expr
    | bit_expr And bit_expr
    | bit_expr SHIFT_LEFT bit_expr
    | bit_expr SHIFT_RIGHT bit_expr
    | bit_expr Plus bit_expr
    | bit_expr Minus bit_expr
    | bit_expr Plus INTERVAL expr date_unit
    | bit_expr Minus INTERVAL expr date_unit
    | bit_expr Star bit_expr
    | bit_expr Div bit_expr
    | bit_expr Mod bit_expr
    | bit_expr MOD bit_expr
    | bit_expr DIV bit_expr
    | bit_expr Caret bit_expr
    | simple_expr
    ;

simple_expr
    : simple_expr collation
    | BINARY simple_expr
    | column_ref
    | expr_const
    | simple_expr CNNOP simple_expr
    | Plus simple_expr
    | Minus simple_expr
    | Tilde simple_expr
    | not2 simple_expr
    | select_with_parens
    | LeftParen expr RightParen
    | LeftParen expr_list Comma expr RightParen
    | ROW LeftParen expr_list Comma expr RightParen
    | EXISTS select_with_parens
    | MATCH LeftParen column_list RightParen AGAINST LeftParen STRING_VALUE ((IN NATURAL LANGUAGE MODE) | (IN BOOLEAN MODE))? RightParen
    | case_expr
    | func_expr
    | window_function
    | USER_VARIABLE
    ;

expr
    : expr AND expr
    | expr AND_OP expr
    | expr OR expr
    | expr CNNOP expr
    | expr XOR expr
    | NOT expr
    | bool_pri IS BOOL_VALUE
    | bool_pri IS not BOOL_VALUE
    | bool_pri IS UNKNOWN
    | bool_pri IS not UNKNOWN
    | bool_pri
    | USER_VARIABLE SET_VAR expr
    ;

not
    : NOT
    | NOT
    ;

not2
    : Not
    | NOT
    ;

sub_query_flag
    : ALL
    | ANY
    | SOME
    ;

in_expr
    : select_with_parens
    | LeftParen expr_list RightParen
    ;

case_expr
    : CASE case_arg when_clause_list case_default END
    ;

window_function
    : COUNT LeftParen ALL? Star RightParen OVER new_generalized_window_clause
    | COUNT LeftParen ALL? expr RightParen OVER new_generalized_window_clause
    | COUNT LeftParen DISTINCT expr_list RightParen OVER new_generalized_window_clause
    | APPROX_COUNT_DISTINCT LeftParen expr_list RightParen OVER new_generalized_window_clause
    | APPROX_COUNT_DISTINCT_SYNOPSIS LeftParen expr_list RightParen OVER new_generalized_window_clause
    | APPROX_COUNT_DISTINCT_SYNOPSIS_MERGE LeftParen expr RightParen OVER new_generalized_window_clause
    | SUM LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen OVER new_generalized_window_clause
    | MAX LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen OVER new_generalized_window_clause
    | MIN LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen OVER new_generalized_window_clause
    | AVG LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen OVER new_generalized_window_clause
    | STDDEV LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen OVER new_generalized_window_clause
    | VARIANCE LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen OVER new_generalized_window_clause
    | STDDEV_POP LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen OVER new_generalized_window_clause
    | STDDEV_SAMP LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen OVER new_generalized_window_clause
    | GROUP_CONCAT LeftParen (DISTINCT | UNIQUE)? expr_list order_by? (SEPARATOR STRING_VALUE)? RightParen OVER new_generalized_window_clause
    | LISTAGG LeftParen (DISTINCT | UNIQUE)? expr_list order_by? (SEPARATOR STRING_VALUE)? RightParen OVER new_generalized_window_clause
    | RANK LeftParen RightParen OVER new_generalized_window_clause
    | DENSE_RANK LeftParen RightParen OVER new_generalized_window_clause
    | PERCENT_RANK LeftParen RightParen OVER new_generalized_window_clause
    | ROW_NUMBER LeftParen RightParen OVER new_generalized_window_clause
    | NTILE LeftParen expr RightParen OVER new_generalized_window_clause
    | CUME_DIST LeftParen RightParen OVER new_generalized_window_clause
    | FIRST_VALUE win_fun_first_last_params OVER new_generalized_window_clause
    | LAST_VALUE win_fun_first_last_params OVER new_generalized_window_clause
    | LEAD win_fun_lead_lag_params OVER new_generalized_window_clause
    | LAG win_fun_lead_lag_params OVER new_generalized_window_clause
    | NTH_VALUE LeftParen expr Comma expr RightParen (FROM first_or_last)? (respect_or_ignore NULLS)? OVER new_generalized_window_clause
    ;

first_or_last
    : FIRST
    | LAST
    ;

respect_or_ignore
    : RESPECT
    | IGNORE
    ;

win_fun_first_last_params
    : LeftParen expr respect_or_ignore NULLS RightParen
    | LeftParen expr RightParen (respect_or_ignore NULLS)?
    ;

win_fun_lead_lag_params
    : LeftParen expr respect_or_ignore NULLS RightParen
    | LeftParen expr respect_or_ignore NULLS Comma expr_list RightParen
    | LeftParen expr_list RightParen (respect_or_ignore NULLS)?
    ;

new_generalized_window_clause
    : NAME_OB
    | new_generalized_window_clause_with_blanket
    ;

new_generalized_window_clause_with_blanket
    : LeftParen NAME_OB generalized_window_clause RightParen
    | LeftParen generalized_window_clause RightParen
    ;

named_windows
    : named_window (Comma named_window)*
    ;

named_window
    : NAME_OB AS new_generalized_window_clause_with_blanket
    ;

generalized_window_clause
    : (PARTITION BY expr_list)? order_by? win_window?
    ;

win_rows_or_range
    : ROWS
    | RANGE
    ;

win_preceding_or_following
    : PRECEDING
    | FOLLOWING
    ;

win_interval
    : expr
    | INTERVAL expr date_unit
    ;

win_bounding
    : CURRENT ROW
    | win_interval win_preceding_or_following
    ;

win_window
    : win_rows_or_range BETWEEN win_bounding AND win_bounding
    | win_rows_or_range win_bounding
    ;

case_arg
    : expr?
    ;

when_clause_list
    : when_clause+
    ;

when_clause
    : WHEN expr THEN expr
    ;

case_default
    : ELSE expr
    | empty
    ;

func_expr
    : MOD LeftParen expr Comma expr RightParen
    | COUNT LeftParen ALL? Star RightParen
    | COUNT LeftParen ALL? expr RightParen
    | COUNT LeftParen DISTINCT expr_list RightParen
    | COUNT LeftParen UNIQUE expr_list RightParen
    | APPROX_COUNT_DISTINCT LeftParen expr_list RightParen
    | APPROX_COUNT_DISTINCT_SYNOPSIS LeftParen expr_list RightParen
    | APPROX_COUNT_DISTINCT_SYNOPSIS_MERGE LeftParen expr RightParen
    | SUM LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen
    | MAX LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen
    | MIN LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen
    | AVG LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen
    | STDDEV LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen
    | VARIANCE LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen
    | STDDEV_POP LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen
    | STDDEV_SAMP LeftParen (ALL | DISTINCT | UNIQUE)? expr RightParen
    | GROUPING LeftParen expr RightParen
    | GROUP_CONCAT LeftParen (DISTINCT | UNIQUE)? expr_list order_by? (SEPARATOR STRING_VALUE)? RightParen
    | IF LeftParen expr Comma expr Comma expr RightParen
    | ISNULL LeftParen expr RightParen
    | cur_timestamp_func
    | sysdate_func
    | cur_time_func
    | cur_date_func
    | utc_timestamp_func
    | CAST LeftParen expr AS cast_data_type RightParen
    | INSERT LeftParen expr Comma expr Comma expr Comma expr RightParen
    | LEFT LeftParen expr Comma expr RightParen
    | CONVERT LeftParen expr Comma cast_data_type RightParen
    | CONVERT LeftParen expr USING charset_name RightParen
    | POSITION LeftParen bit_expr IN expr RightParen
    | substr_or_substring LeftParen substr_params RightParen
    | TRIM LeftParen parameterized_trim RightParen
    | DATE LeftParen expr RightParen
    | YEAR LeftParen expr RightParen
    | TIME LeftParen expr RightParen
    | MONTH LeftParen expr RightParen
    | WEEK LeftParen expr RightParen
    | WEEK LeftParen expr Comma expr RightParen
    | SECOND LeftParen expr RightParen
    | MINUTE LeftParen expr RightParen
    | MICROSECOND LeftParen expr RightParen
    | HOUR LeftParen expr RightParen
    | DATE_ADD LeftParen date_params RightParen
    | DATE_SUB LeftParen date_params RightParen
    | ADDDATE LeftParen date_params RightParen
    | SUBDATE LeftParen date_params RightParen
    | ADDDATE LeftParen expr Comma expr RightParen
    | SUBDATE LeftParen expr Comma expr RightParen
    | TIMESTAMPDIFF LeftParen timestamp_params RightParen
    | TIMESTAMPADD LeftParen timestamp_params RightParen
    | EXTRACT LeftParen date_unit FROM expr RightParen
    | ASCII LeftParen expr RightParen
    | DEFAULT LeftParen column_definition_ref RightParen
    | VALUES LeftParen column_definition_ref RightParen
    | CHARACTER LeftParen expr_list RightParen
    | CHARACTER LeftParen expr_list USING charset_name RightParen
    | function_name LeftParen expr_as_list? RightParen
    | relation_name Dot function_name LeftParen expr_as_list? RightParen
    | sys_interval_func
    ;

sys_interval_func
    : INTERVAL LeftParen expr Comma expr RightParen
    | INTERVAL LeftParen expr Comma expr Comma expr_list RightParen
    ;

utc_timestamp_func
    : UTC_TIMESTAMP LeftParen RightParen
    | UTC_TIMESTAMP LeftParen INTNUM RightParen
    ;

sysdate_func
    : SYSDATE LeftParen RightParen
    | SYSDATE LeftParen INTNUM RightParen
    ;

cur_timestamp_func
    : NOW LeftParen RightParen
    | NOW LeftParen INTNUM RightParen
    | now_synonyms_func ((LeftParen INTNUM RightParen) | (LeftParen RightParen))?
    ;

now_synonyms_func
    : CURRENT_TIMESTAMP
    | LOCALTIME
    | LOCALTIMESTAMP
    ;

cur_time_func
    : CURTIME LeftParen RightParen
    | CURTIME LeftParen INTNUM RightParen
    | CURRENT_TIME ((LeftParen INTNUM RightParen) | (LeftParen RightParen))?
    ;

cur_date_func
    : CURDATE LeftParen RightParen
    | CURRENT_DATE
    | CURRENT_DATE LeftParen RightParen
    ;

substr_or_substring
    : SUBSTR
    | SUBSTRING
    ;

substr_params
    : expr Comma expr
    | expr Comma expr Comma expr
    | expr FROM expr
    | expr FROM expr FOR expr
    ;

date_params
    : expr Comma INTERVAL expr date_unit
    ;

timestamp_params
    : date_unit Comma expr Comma expr
    ;

delete_stmt
    : delete_with_opt_hint FROM tbl_name (WHERE opt_hint_value expr)? order_by? limit_clause?
    | delete_with_opt_hint multi_delete_table (WHERE opt_hint_value expr)?
    ;

multi_delete_table
    : relation_with_star_list FROM table_references
    | FROM relation_with_star_list USING table_references
    ;

update_stmt
    : update_with_opt_hint IGNORE? table_references SET update_asgn_list (WHERE opt_hint_value expr)? order_by? limit_clause?
    ;

update_asgn_list
    : update_asgn_factor (Comma update_asgn_factor)*
    ;

update_asgn_factor
    : column_definition_ref COMP_EQ expr_or_default
    ;

create_resource_stmt
    : CREATE RESOURCE UNIT (IF not EXISTS)? relation_name (resource_unit_option | (opt_resource_unit_option_list Comma resource_unit_option))?
    | CREATE RESOURCE POOL (IF not EXISTS)? relation_name (create_resource_pool_option | (opt_create_resource_pool_option_list Comma create_resource_pool_option))?
    ;

opt_resource_unit_option_list
    : resource_unit_option
    | opt_resource_unit_option_list Comma resource_unit_option
    | empty
    ;

resource_unit_option
    : MIN_CPU COMP_EQ? conf_const
    | MIN_IOPS COMP_EQ? conf_const
    | MIN_MEMORY COMP_EQ? conf_const
    | MAX_CPU COMP_EQ? conf_const
    | MAX_MEMORY COMP_EQ? conf_const
    | MAX_IOPS COMP_EQ? conf_const
    | MAX_DISK_SIZE COMP_EQ? conf_const
    | MAX_SESSION_NUM COMP_EQ? conf_const
    ;

opt_create_resource_pool_option_list
    : create_resource_pool_option
    | opt_create_resource_pool_option_list Comma create_resource_pool_option
    | empty
    ;

create_resource_pool_option
    : UNIT COMP_EQ? relation_name_or_string
    | UNIT_NUM COMP_EQ? INTNUM
    | ZONE_LIST COMP_EQ? LeftParen zone_list RightParen
    | REPLICA_TYPE COMP_EQ? STRING_VALUE
    | IS_TENANT_SYS_POOL COMP_EQ? BOOL_VALUE
    ;

alter_resource_pool_option_list
    : alter_resource_pool_option (Comma alter_resource_pool_option)*
    ;

unit_id_list
    : INTNUM (Comma INTNUM)*
    ;

alter_resource_pool_option
    : UNIT COMP_EQ? relation_name_or_string
    | UNIT_NUM COMP_EQ? INTNUM (DELETE UNIT opt_equal_mark LeftParen unit_id_list RightParen)?
    | ZONE_LIST COMP_EQ? LeftParen zone_list RightParen
    ;

alter_resource_stmt
    : ALTER RESOURCE UNIT relation_name (resource_unit_option | (opt_resource_unit_option_list Comma resource_unit_option))?
    | ALTER RESOURCE POOL relation_name alter_resource_pool_option_list
    | ALTER RESOURCE POOL relation_name SPLIT INTO LeftParen resource_pool_list RightParen ON LeftParen zone_list RightParen
    ;

drop_resource_stmt
    : DROP RESOURCE UNIT (IF EXISTS)? relation_name
    | DROP RESOURCE POOL (IF EXISTS)? relation_name
    ;

create_tenant_stmt
    : CREATE TENANT (IF not EXISTS)? relation_name (tenant_option | (opt_tenant_option_list Comma tenant_option))? ((SET sys_var_and_val_list) | (SET VARIABLES sys_var_and_val_list) | (VARIABLES sys_var_and_val_list))?
    ;

opt_tenant_option_list
    : tenant_option
    | opt_tenant_option_list Comma tenant_option
    | empty
    ;

tenant_option
    : LOGONLY_REPLICA_NUM COMP_EQ? INTNUM
    | LOCALITY COMP_EQ? STRING_VALUE FORCE?
    | REPLICA_NUM COMP_EQ? INTNUM
    | REWRITE_MERGE_VERSION COMP_EQ? INTNUM
    | STORAGE_FORMAT_VERSION COMP_EQ? INTNUM
    | STORAGE_FORMAT_WORK_VERSION COMP_EQ? INTNUM
    | PRIMARY_ZONE COMP_EQ? primary_zone_name
    | RESOURCE_POOL_LIST COMP_EQ? LeftParen resource_pool_list RightParen
    | ZONE_LIST COMP_EQ? LeftParen zone_list RightParen
    | charset_key COMP_EQ? charset_name
    | COLLATE COMP_EQ? collation_name
    | read_only_or_write
    | COMMENT COMP_EQ? STRING_VALUE
    | default_tablegroup
    | PROGRESSIVE_MERGE_NUM COMP_EQ? INTNUM
    ;

zone_list
    : STRING_VALUE (opt_comma STRING_VALUE)*
    ;

resource_pool_list
    : STRING_VALUE (Comma STRING_VALUE)*
    ;

alter_tenant_stmt
    : ALTER TENANT relation_name SET? (tenant_option | (opt_tenant_option_list Comma tenant_option))? (VARIABLES sys_var_and_val_list)?
    | ALTER TENANT ALL SET? (tenant_option | (opt_tenant_option_list Comma tenant_option))? (VARIABLES sys_var_and_val_list)?
    | ALTER TENANT relation_name RENAME GLOBAL_NAME TO relation_name
    | ALTER TENANT relation_name lock_spec_mysql57
    ;

drop_tenant_stmt
    : DROP TENANT (IF EXISTS)? relation_name FORCE?
    ;

create_database_stmt
    : CREATE database_key (IF not EXISTS)? database_factor database_option_list?
    ;

database_key
    : DATABASE
    | SCHEMA
    ;

database_factor
    : relation_name
    ;

database_option_list
    : database_option+
    ;

charset_key
    : CHARSET
    | CHARACTER SET
    ;

database_option
    : DEFAULT? charset_key COMP_EQ? charset_name
    | DEFAULT? COLLATE COMP_EQ? collation_name
    | REPLICA_NUM COMP_EQ? INTNUM
    | PRIMARY_ZONE COMP_EQ? primary_zone_name
    | read_only_or_write
    | default_tablegroup
    | DATABASE_ID COMP_EQ? INTNUM
    ;

read_only_or_write
    : READ ONLY
    | READ WRITE
    ;

drop_database_stmt
    : DROP database_key (IF EXISTS)? database_factor
    ;

alter_database_stmt
    : ALTER database_key NAME_OB? SET? database_option_list
    ;

load_data_stmt
    : load_data_with_opt_hint (LOCAL | REMOTE_OSS)? INFILE STRING_VALUE (IGNORE | REPLACE)? INTO TABLE relation_factor use_partition? (CHARACTER SET charset_name_or_default)? field_opt line_opt ((IGNORE INTNUM lines_or_rows) | (GENERATED INTNUM lines_or_rows))? ((LeftParen RightParen) | (LeftParen field_or_vars_list RightParen))? (SET load_set_list)?
    ;

load_data_with_opt_hint
    : LOAD DATA
    | LOAD_DATA_HINT_BEGIN hint_list_with_end
    ;

lines_or_rows
    : LINES
    | ROWS
    ;

field_or_vars_list
    : field_or_vars (Comma field_or_vars)*
    ;

field_or_vars
    : column_definition_ref
    | USER_VARIABLE
    ;

load_set_list
    : load_set_element (Comma load_set_element)*
    ;

load_set_element
    : column_definition_ref COMP_EQ expr_or_default
    ;

use_database_stmt
    : USE database_factor
    ;

create_synonym_stmt
    : CREATE (OR REPLACE)? PUBLIC? SYNONYM synonym_name FOR synonym_object (At ip_port)?
    | CREATE (OR REPLACE)? PUBLIC? SYNONYM database_factor Dot synonym_name FOR synonym_object (At ip_port)?
    | CREATE (OR REPLACE)? PUBLIC? SYNONYM synonym_name FOR database_factor Dot synonym_object (At ip_port)?
    | CREATE (OR REPLACE)? PUBLIC? SYNONYM database_factor Dot synonym_name FOR database_factor Dot synonym_object (At ip_port)?
    ;

synonym_name
    : NAME_OB
    | unreserved_keyword
    ;

synonym_object
    : NAME_OB
    | unreserved_keyword
    ;

drop_synonym_stmt
    : DROP PUBLIC? SYNONYM synonym_name FORCE?
    | DROP PUBLIC? SYNONYM database_factor Dot synonym_name FORCE?
    ;

temporary_option
    : TEMPORARY?
    ;

create_table_like_stmt
    : CREATE temporary_option TABLE (IF not EXISTS)? relation_factor LIKE relation_factor
    | CREATE temporary_option TABLE (IF not EXISTS)? relation_factor LeftParen LIKE relation_factor RightParen
    ;

create_table_stmt
    : CREATE temporary_option TABLE (IF not EXISTS)? relation_factor LeftParen table_element_list RightParen table_option_list? opt_partition_option
    | CREATE temporary_option TABLE (IF not EXISTS)? relation_factor LeftParen table_element_list RightParen table_option_list? opt_partition_option AS? select_stmt
    | CREATE temporary_option TABLE (IF not EXISTS)? relation_factor table_option_list opt_partition_option AS? select_stmt
    | CREATE temporary_option TABLE (IF not EXISTS)? relation_factor partition_option AS? select_stmt
    | CREATE temporary_option TABLE (IF not EXISTS)? relation_factor select_stmt
    | CREATE temporary_option TABLE (IF not EXISTS)? relation_factor AS select_stmt
    ;

ret_type
    : STRING
    | INTEGER
    | REAL
    | DECIMAL
    ;

create_function_stmt
    : CREATE AGGREGATE? FUNCTION NAME_OB RETURNS ret_type SONAME STRING_VALUE
    ;

drop_function_stmt
    : DROP FUNCTION (IF EXISTS)? NAME_OB
    ;

table_element_list
    : table_element (Comma table_element)*
    ;

table_element
    : column_definition
    | (CONSTRAINT opt_constraint_name)? PRIMARY KEY index_using_algorithm? LeftParen column_name_list RightParen index_using_algorithm? (COMMENT STRING_VALUE)?
    | key_or_index index_name? index_using_algorithm? LeftParen sort_column_list RightParen opt_index_options?
    | UNIQUE key_or_index? index_name? index_using_algorithm? LeftParen sort_column_list RightParen opt_index_options?
    | CONSTRAINT constraint_name? UNIQUE key_or_index? index_name? index_using_algorithm? LeftParen sort_column_list RightParen opt_index_options?
    | FULLTEXT key_or_index? index_name? index_using_algorithm? LeftParen sort_column_list RightParen opt_index_options?
    | CONSTRAINT constraint_name CHECK LeftParen expr RightParen
    | (CONSTRAINT opt_constraint_name)? FOREIGN KEY index_name? LeftParen column_name_list RightParen REFERENCES relation_factor LeftParen column_name_list RightParen (MATCH match_action)? (opt_reference_option_list reference_option)?
    ;

opt_reference_option_list
    : opt_reference_option_list reference_option
    | empty
    ;

reference_option
    : ON UPDATE reference_action
    | ON DELETE reference_action
    ;

reference_action
    : RESTRICT
    | CASCADE
    | SET NULLX
    | NO ACTION
    | SET DEFAULT
    ;

match_action
    : SIMPLE
    | FULL
    | PARTIAL
    ;

column_definition
    : column_definition_ref data_type (opt_column_attribute_list column_attribute)? (FIRST | (BEFORE column_name) | (AFTER column_name))?
    | column_definition_ref data_type (GENERATED ALWAYS)? AS LeftParen expr RightParen (VIRTUAL | STORED)? (opt_generated_column_attribute_list generated_column_attribute)? (FIRST | (BEFORE column_name) | (AFTER column_name))?
    ;

opt_generated_column_attribute_list
    : opt_generated_column_attribute_list generated_column_attribute
    | empty
    ;

generated_column_attribute
    : NOT NULLX
    | NULLX
    | UNIQUE KEY
    | PRIMARY? KEY
    | UNIQUE
    | COMMENT STRING_VALUE
    | ID INTNUM
    ;

column_definition_ref
    : column_name
    | relation_name Dot column_name
    | relation_name Dot relation_name Dot column_name
    ;

column_definition_list
    : column_definition (Comma column_definition)*
    ;

cast_data_type
    : BINARY string_length_i?
    | CHARACTER string_length_i? BINARY?
    | cast_datetime_type_i (LeftParen INTNUM RightParen)?
    | NUMBER ((LeftParen INTNUM Comma INTNUM RightParen) | (LeftParen INTNUM RightParen))?
    | DECIMAL ((LeftParen INTNUM Comma INTNUM RightParen) | (LeftParen INTNUM RightParen))?
    | SIGNED INTEGER?
    | UNSIGNED INTEGER?
    | DOUBLE
    ;

cast_datetime_type_i
    : DATETIME
    | DATE
    | TIME
    ;

data_type
    : int_type_i (LeftParen INTNUM RightParen)? (UNSIGNED | SIGNED)? ZEROFILL?
    | float_type_i ((LeftParen INTNUM Comma INTNUM RightParen) | (LeftParen INTNUM RightParen))? (UNSIGNED | SIGNED)? ZEROFILL?
    | NUMBER ((LeftParen INTNUM Comma INTNUM RightParen) | (LeftParen INTNUM RightParen))? (UNSIGNED | SIGNED)? ZEROFILL?
    | DECIMAL ((LeftParen INTNUM Comma INTNUM RightParen) | (LeftParen INTNUM RightParen))? (UNSIGNED | SIGNED)? ZEROFILL?
    | BOOL
    | BOOLEAN
    | datetime_type_i (LeftParen INTNUM RightParen)?
    | date_year_type_i
    | CHARACTER string_length_i? BINARY? (charset_key charset_name)? collation?
    | VARCHAR string_length_i BINARY? (charset_key charset_name)? collation?
    | blob_type_i string_length_i?
    | text_type_i string_length_i? BINARY? (charset_key charset_name)? collation?
    | BINARY string_length_i?
    | VARBINARY string_length_i
    | STRING_VALUE
    | BIT (LeftParen INTNUM RightParen)?
    | ENUM LeftParen string_list RightParen BINARY? (charset_key charset_name)? collation?
    | SET LeftParen string_list RightParen BINARY? (charset_key charset_name)? collation?
    ;

string_list
    : text_string (Comma text_string)*
    ;

text_string
    : STRING_VALUE
    | HEX_STRING_VALUE
    ;

int_type_i
    : TINYINT
    | SMALLINT
    | MEDIUMINT
    | INTEGER
    | BIGINT
    ;

float_type_i
    : FLOAT
    | DOUBLE
    | REAL
    | DOUBLE PRECISION
    | REAL PRECISION
    ;

datetime_type_i
    : DATETIME
    | TIMESTAMP
    | TIME
    ;

date_year_type_i
    : DATE
    | YEAR (LeftParen INTNUM RightParen)?
    ;

text_type_i
    : TINYTEXT
    | TEXT
    | MEDIUMTEXT
    | LONGTEXT
    ;

blob_type_i
    : TINYBLOB
    | BLOB
    | MEDIUMBLOB
    | LONGBLOB
    ;

string_length_i
    : LeftParen number_literal RightParen
    ;

collation_name
    : NAME_OB
    | STRING_VALUE
    ;

trans_param_name
    : Quote STRING_VALUE Quote
    ;

trans_param_value
    : Quote STRING_VALUE Quote
    | INTNUM
    ;

charset_name
    : NAME_OB
    | STRING_VALUE
    | BINARY
    ;

charset_name_or_default
    : charset_name
    | DEFAULT
    ;

collation
    : COLLATE collation_name
    ;

opt_column_attribute_list
    : opt_column_attribute_list column_attribute
    | empty
    ;

column_attribute
    : not NULLX
    | NULLX
    | DEFAULT now_or_signed_literal
    | ORIG_DEFAULT now_or_signed_literal
    | AUTO_INCREMENT
    | UNIQUE KEY
    | PRIMARY? KEY
    | UNIQUE
    | COMMENT STRING_VALUE
    | ON UPDATE cur_timestamp_func
    | ID INTNUM
    ;

now_or_signed_literal
    : cur_timestamp_func
    | signed_literal
    ;

signed_literal
    : literal
    | Plus number_literal
    | Minus number_literal
    ;

opt_comma
    : Comma?
    ;

table_option_list_space_seperated
    : table_option
    | table_option table_option_list_space_seperated
    ;

table_option_list
    : table_option_list_space_seperated
    | table_option Comma table_option_list
    ;

primary_zone_name
    : DEFAULT
    | RANDOM
    | relation_name_or_string
    ;

tablespace
    : NAME_OB
    ;

locality_name
    : STRING_VALUE
    | DEFAULT
    ;

table_option
    : TABLE_MODE COMP_EQ? STRING_VALUE
    | DUPLICATE_SCOPE COMP_EQ? STRING_VALUE
    | LOCALITY COMP_EQ? locality_name FORCE?
    | EXPIRE_INFO COMP_EQ? LeftParen expr RightParen
    | PROGRESSIVE_MERGE_NUM COMP_EQ? INTNUM
    | BLOCK_SIZE COMP_EQ? INTNUM
    | TABLE_ID COMP_EQ? INTNUM
    | REPLICA_NUM COMP_EQ? INTNUM
    | COMPRESSION COMP_EQ? STRING_VALUE
    | ROW_FORMAT COMP_EQ? row_format_option
    | STORAGE_FORMAT_VERSION COMP_EQ? INTNUM
    | USE_BLOOM_FILTER COMP_EQ? BOOL_VALUE
    | DEFAULT? charset_key COMP_EQ? charset_name
    | DEFAULT? COLLATE COMP_EQ? collation_name
    | COMMENT COMP_EQ? STRING_VALUE
    | PRIMARY_ZONE COMP_EQ? primary_zone_name
    | TABLEGROUP COMP_EQ? relation_name_or_string
    | AUTO_INCREMENT COMP_EQ? int_or_decimal
    | read_only_or_write
    | ENGINE_ COMP_EQ? relation_name_or_string
    | TABLET_SIZE COMP_EQ? INTNUM
    | PCTFREE COMP_EQ? INTNUM
    | MAX_USED_PART_ID COMP_EQ? INTNUM
    | TABLESPACE tablespace
    ;

relation_name_or_string
    : relation_name
    | STRING_VALUE
    | ALL
    ;

opt_equal_mark
    : COMP_EQ?
    ;

partition_option
    : hash_partition_option
    | key_partition_option
    | range_partition_option
    | list_partition_option
    ;

opt_partition_option
    : partition_option opt_column_partition_option
    | opt_column_partition_option
    ;

hash_partition_option
    : PARTITION BY HASH LeftParen expr RightParen subpartition_option (PARTITIONS INTNUM)?
    ;

list_partition_option
    : PARTITION BY BISON_LIST LeftParen expr RightParen subpartition_option (PARTITIONS INTNUM)? opt_list_partition_list
    | PARTITION BY BISON_LIST COLUMNS LeftParen column_name_list RightParen subpartition_option (PARTITIONS INTNUM)? opt_list_partition_list
    ;

key_partition_option
    : PARTITION BY KEY LeftParen column_name_list RightParen subpartition_option (PARTITIONS INTNUM)?
    | PARTITION BY KEY LeftParen RightParen subpartition_option (PARTITIONS INTNUM)?
    ;

range_partition_option
    : PARTITION BY RANGE LeftParen expr RightParen subpartition_option (PARTITIONS INTNUM)? opt_range_partition_list
    | PARTITION BY RANGE COLUMNS LeftParen column_name_list RightParen subpartition_option (PARTITIONS INTNUM)? opt_range_partition_list
    ;

opt_column_partition_option
    : column_partition_option?
    ;

column_partition_option
    : PARTITION BY COLUMN LeftParen vertical_column_name RightParen
    | PARTITION BY COLUMN LeftParen vertical_column_name Comma aux_column_list RightParen
    ;

aux_column_list
    : vertical_column_name (Comma vertical_column_name)*
    ;

vertical_column_name
    : column_name
    | LeftParen column_name_list RightParen
    ;

column_name_list
    : column_name (Comma column_name)*
    ;

subpartition_option
    : SUBPARTITION BY RANGE LeftParen expr RightParen SUBPARTITION TEMPLATE opt_range_subpartition_list
    | SUBPARTITION BY RANGE COLUMNS LeftParen column_name_list RightParen SUBPARTITION TEMPLATE opt_range_subpartition_list
    | SUBPARTITION BY HASH LeftParen expr RightParen (SUBPARTITIONS INTNUM)?
    | SUBPARTITION BY BISON_LIST LeftParen expr RightParen SUBPARTITION TEMPLATE opt_list_subpartition_list
    | SUBPARTITION BY BISON_LIST COLUMNS LeftParen column_name_list RightParen SUBPARTITION TEMPLATE opt_list_subpartition_list
    | SUBPARTITION BY KEY LeftParen column_name_list RightParen (SUBPARTITIONS INTNUM)?
    | empty
    ;

opt_list_partition_list
    : LeftParen list_partition_list RightParen
    ;

opt_list_subpartition_list
    : LeftParen list_subpartition_list RightParen
    ;

opt_range_partition_list
    : LeftParen range_partition_list RightParen
    ;

opt_range_subpartition_list
    : LeftParen range_subpartition_list RightParen
    ;

list_partition_list
    : list_partition_element (Comma list_partition_element)*
    ;

list_subpartition_list
    : list_subpartition_element (Comma list_subpartition_element)*
    ;

list_subpartition_element
    : SUBPARTITION relation_factor VALUES IN list_partition_expr
    ;

list_partition_element
    : PARTITION relation_factor VALUES IN list_partition_expr (ID INTNUM)?
    ;

list_partition_expr
    : LeftParen list_expr RightParen
    | LeftParen DEFAULT RightParen
    ;

list_expr
    : expr (Comma expr)*
    ;

range_partition_list
    : range_partition_element (Comma range_partition_element)*
    ;

range_partition_element
    : PARTITION relation_factor VALUES LESS THAN range_partition_expr (ID INTNUM)?
    ;

range_subpartition_element
    : SUBPARTITION relation_factor VALUES LESS THAN range_partition_expr
    ;

range_subpartition_list
    : range_subpartition_element (Comma range_subpartition_element)*
    ;

range_partition_expr
    : LeftParen range_expr_list RightParen
    | MAXVALUE
    ;

range_expr_list
    : range_expr (Comma range_expr)*
    ;

range_expr
    : expr
    | MAXVALUE
    ;

int_or_decimal
    : INTNUM
    | DECIMAL_VAL
    ;

tg_hash_partition_option
    : PARTITION BY HASH tg_subpartition_option (PARTITIONS INTNUM)?
    ;

tg_key_partition_option
    : PARTITION BY KEY INTNUM tg_subpartition_option (PARTITIONS INTNUM)?
    ;

tg_range_partition_option
    : PARTITION BY RANGE tg_subpartition_option (PARTITIONS INTNUM)? opt_range_partition_list
    | PARTITION BY RANGE COLUMNS INTNUM tg_subpartition_option (PARTITIONS INTNUM)? opt_range_partition_list
    ;

tg_list_partition_option
    : PARTITION BY BISON_LIST tg_subpartition_option (PARTITIONS INTNUM)? opt_list_partition_list
    | PARTITION BY BISON_LIST COLUMNS INTNUM tg_subpartition_option (PARTITIONS INTNUM)? opt_list_partition_list
    ;

tg_subpartition_option
    : SUBPARTITION BY RANGE SUBPARTITION TEMPLATE opt_range_subpartition_list
    | SUBPARTITION BY RANGE COLUMNS INTNUM SUBPARTITION TEMPLATE opt_range_subpartition_list
    | SUBPARTITION BY HASH (SUBPARTITIONS INTNUM)?
    | SUBPARTITION BY KEY INTNUM (SUBPARTITIONS INTNUM)?
    | SUBPARTITION BY BISON_LIST SUBPARTITION TEMPLATE opt_list_subpartition_list
    | SUBPARTITION BY BISON_LIST COLUMNS INTNUM SUBPARTITION TEMPLATE opt_list_subpartition_list
    | empty
    ;

row_format_option
    : REDUNDANT
    | COMPACT
    | DYNAMIC
    | COMPRESSED
    | DEFAULT
    ;

create_tablegroup_stmt
    : CREATE TABLEGROUP (IF not EXISTS)? relation_name tablegroup_option_list? (tg_hash_partition_option | tg_key_partition_option | tg_range_partition_option | tg_list_partition_option)?
    ;

drop_tablegroup_stmt
    : DROP TABLEGROUP (IF EXISTS)? relation_name
    ;

alter_tablegroup_stmt
    : ALTER TABLEGROUP relation_name ADD TABLE? table_list
    | ALTER TABLEGROUP relation_name alter_tablegroup_actions
    | ALTER TABLEGROUP relation_name alter_tg_partition_option
    ;

tablegroup_option_list_space_seperated
    : tablegroup_option
    | tablegroup_option tablegroup_option_list_space_seperated
    ;

tablegroup_option_list
    : tablegroup_option_list_space_seperated
    | tablegroup_option Comma tablegroup_option_list
    ;

tablegroup_option
    : LOCALITY COMP_EQ? locality_name FORCE?
    | PRIMARY_ZONE COMP_EQ? primary_zone_name
    | TABLEGROUP_ID COMP_EQ? INTNUM
    | BINDING COMP_EQ? BOOL_VALUE
    | MAX_USED_PART_ID COMP_EQ? INTNUM
    ;

alter_tablegroup_actions
    : alter_tablegroup_action (Comma alter_tablegroup_action)*
    ;

alter_tablegroup_action
    : SET? tablegroup_option_list_space_seperated
    ;

default_tablegroup
    : DEFAULT? TABLEGROUP COMP_EQ? relation_name
    | DEFAULT? TABLEGROUP COMP_EQ? NULLX
    ;

create_view_stmt
    : CREATE (OR REPLACE)? MATERIALIZED? VIEW view_name (LeftParen column_list RightParen)? (TABLE_ID COMP_EQ INTNUM)? AS view_select_stmt
    ;

view_select_stmt
    : select_stmt
    ;

view_name
    : relation_factor
    ;

create_index_stmt
    : CREATE (FULLTEXT | UNIQUE)? INDEX (IF not EXISTS)? normal_relation_factor index_using_algorithm? ON relation_factor LeftParen sort_column_list RightParen opt_index_options? opt_partition_option
    ;

index_name
    : relation_name
    ;

opt_constraint_name
    : constraint_name?
    ;

constraint_name
    : relation_name
    ;

sort_column_list
    : sort_column_key (Comma sort_column_key)*
    ;

sort_column_key
    : column_name (LeftParen INTNUM RightParen)? (ASC | DESC)? (ID INTNUM)?
    ;

opt_index_options
    : index_option+
    ;

index_option
    : GLOBAL
    | LOCAL
    | BLOCK_SIZE COMP_EQ? INTNUM
    | COMMENT STRING_VALUE
    | STORING LeftParen column_name_list RightParen
    | CTXCAT LeftParen column_name_list RightParen
    | WITH_ROWID
    | WITH PARSER STRING_VALUE
    | index_using_algorithm
    | visibility_option
    | DATA_TABLE_ID COMP_EQ? INTNUM
    | INDEX_TABLE_ID COMP_EQ? INTNUM
    | VIRTUAL_COLUMN_ID COMP_EQ? INTNUM
    | MAX_USED_PART_ID COMP_EQ? INTNUM
    ;

index_using_algorithm
    : USING BTREE
    | USING HASH
    ;

drop_table_stmt
    : DROP (TEMPORARY | MATERIALIZED)? table_or_tables (IF EXISTS)? table_list (CASCADE | RESTRICT)?
    ;

table_or_tables
    : TABLE
    | TABLES
    ;

drop_view_stmt
    : DROP MATERIALIZED? VIEW (IF EXISTS)? table_list (CASCADE | RESTRICT)?
    ;

table_list
    : relation_factor (Comma relation_factor)*
    ;

drop_index_stmt
    : DROP INDEX relation_name ON relation_factor
    ;

insert_stmt
    : insert_with_opt_hint IGNORE? INTO? single_table_insert (ON DUPLICATE KEY UPDATE update_asgn_list)?
    | replace_with_opt_hint IGNORE? INTO? single_table_insert
    ;

single_table_insert
    : dml_table_name values_clause
    | dml_table_name LeftParen RightParen values_clause
    | dml_table_name LeftParen column_list RightParen values_clause
    | dml_table_name SET update_asgn_list
    ;

values_clause
    : value_or_values insert_vals_list
    | select_stmt
    ;

value_or_values
    : VALUE
    | VALUES
    ;

replace_with_opt_hint
    : REPLACE
    | REPLACE_HINT_BEGIN hint_list_with_end
    ;

insert_with_opt_hint
    : INSERT
    | INSERT_HINT_BEGIN hint_list_with_end
    ;

column_list
    : column_definition_ref (Comma column_definition_ref)*
    ;

insert_vals_list
    : LeftParen insert_vals RightParen
    | insert_vals_list Comma LeftParen insert_vals RightParen
    ;

insert_vals
    : expr_or_default
    | insert_vals Comma expr_or_default
    | empty
    ;

expr_or_default
    : expr
    | DEFAULT
    ;

select_stmt
    : select_no_parens
    | select_with_parens
    | select_into
    ;

select_into
    : select_no_parens into_clause
    ;

select_with_parens
    : LeftParen select_no_parens RightParen
    | LeftParen select_with_parens RightParen
    ;

select_no_parens
    : select_clause (FOR UPDATE opt_for_update_wait)?
    | select_clause_set (FOR UPDATE opt_for_update_wait)?
    | select_clause_set_with_order_and_limit (FOR UPDATE opt_for_update_wait)?
    ;

no_table_select
    : select_with_opt_hint query_expression_option_list? select_expr_list into_opt
    | select_with_opt_hint query_expression_option_list? select_expr_list into_opt FROM DUAL (WHERE opt_hint_value expr)? (WINDOW named_windows)?
    ;

select_clause
    : no_table_select
    | no_table_select_with_order_and_limit
    | simple_select
    | simple_select_with_order_and_limit
    | select_with_parens_with_order_and_limit
    ;

select_clause_set_with_order_and_limit
    : select_clause_set order_by
    | select_clause_set order_by? limit_clause
    ;

select_clause_set
    : select_clause_set set_type select_clause_set_right
    | select_clause_set order_by set_type select_clause_set_right
    | select_clause_set order_by? limit_clause set_type select_clause_set_right
    | select_clause_set_left set_type select_clause_set_right
    ;

select_clause_set_right
    : no_table_select
    | simple_select
    | select_with_parens
    ;

select_clause_set_left
    : no_table_select_with_order_and_limit
    | simple_select_with_order_and_limit
    | select_clause_set_right
    ;

no_table_select_with_order_and_limit
    : no_table_select order_by
    | no_table_select order_by? limit_clause
    ;

simple_select_with_order_and_limit
    : simple_select order_by
    | simple_select order_by? limit_clause
    ;

select_with_parens_with_order_and_limit
    : select_with_parens order_by
    | select_with_parens order_by? limit_clause
    ;

select_with_opt_hint
    : SELECT
    | SELECT_HINT_BEGIN hint_list_with_end
    ;

update_with_opt_hint
    : UPDATE
    | UPDATE_HINT_BEGIN hint_list_with_end
    ;

delete_with_opt_hint
    : DELETE
    | DELETE_HINT_BEGIN hint_list_with_end
    ;

simple_select
    : select_with_opt_hint query_expression_option_list? select_expr_list into_opt FROM from_list (WHERE opt_hint_value expr)? (GROUP BY groupby_clause)? (HAVING expr)? (WINDOW named_windows)?
    ;

set_type_union
    : UNION
    ;

set_type_other
    : INTERSECT
    | EXCEPT
    | MINUS
    ;

set_type
    : set_type_union set_expression_option
    | set_type_other
    ;

set_expression_option
    : (ALL | DISTINCT | UNIQUE)?
    ;

opt_hint_value
    : HINT_VALUE?
    ;

limit_clause
    : LIMIT limit_expr OFFSET limit_expr
    | LIMIT limit_expr
    | LIMIT limit_expr Comma limit_expr
    ;

into_clause
    : INTO OUTFILE STRING_VALUE (charset_key charset_name)? field_opt line_opt
    | INTO DUMPFILE STRING_VALUE
    | INTO into_var_list
    ;

into_opt
    : into_clause?
    ;

into_var_list
    : into_var (Comma into_var)*
    ;

into_var
    : USER_VARIABLE
    | var_name
    ;

field_opt
    : columns_or_fields field_term_list
    | empty
    ;

field_term_list
    : field_term+
    ;

field_term
    : TERMINATED BY STRING_VALUE
    | OPTIONALLY ENCLOSED BY STRING_VALUE
    | ENCLOSED BY STRING_VALUE
    | ESCAPED BY STRING_VALUE
    ;

line_opt
    : LINES line_term_list
    | empty
    ;

line_term_list
    : line_term+
    ;

line_term
    : TERMINATED BY STRING_VALUE
    | STARTING BY STRING_VALUE
    ;

hint_list_with_end
    : (hint_options | (opt_hint_list Comma hint_options))? HINT_END
    ;

opt_hint_list
    : hint_options
    | opt_hint_list Comma hint_options
    | empty
    ;

hint_options
    : hint_option+
    ;

name_list
    : NAME_OB
    | name_list NAME_OB
    | name_list Comma NAME_OB
    ;

hint_option
    : NO_REWRITE
    | READ_CONSISTENCY LeftParen consistency_level RightParen
    | INDEX_HINT LeftParen qb_name_option relation_factor_in_hint NAME_OB RightParen
    | QUERY_TIMEOUT LeftParen INTNUM RightParen
    | FROZEN_VERSION LeftParen INTNUM RightParen
    | TOPK LeftParen INTNUM INTNUM RightParen
    | HOTSPOT
    | LOG_LEVEL LeftParen NAME_OB RightParen
    | LOG_LEVEL LeftParen Quote STRING_VALUE Quote RightParen
    | LEADING_HINT LeftParen qb_name_option relation_factor_in_leading_hint_list_entry RightParen
    | LEADING_HINT LeftParen qb_name_option relation_factor_in_hint_list RightParen
    | ORDERED
    | FULL_HINT LeftParen qb_name_option relation_factor_in_hint RightParen
    | USE_PLAN_CACHE LeftParen use_plan_cache_type RightParen
    | USE_MERGE LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | NO_USE_MERGE LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | USE_HASH LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | NO_USE_HASH LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | USE_NL LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | NO_USE_NL LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | USE_BNL LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | NO_USE_BNL LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | USE_NL_MATERIALIZATION LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | NO_USE_NL_MATERIALIZATION LeftParen qb_name_option relation_factor_in_use_join_hint_list RightParen
    | USE_HASH_AGGREGATION
    | NO_USE_HASH_AGGREGATION
    | MERGE_HINT (LeftParen qb_name_option RightParen)?
    | NO_MERGE_HINT (LeftParen qb_name_option RightParen)?
    | NO_EXPAND (LeftParen qb_name_option RightParen)?
    | USE_CONCAT (LeftParen qb_name_option RightParen)?
    | UNNEST (LeftParen qb_name_option RightParen)?
    | NO_UNNEST (LeftParen qb_name_option RightParen)?
    | PLACE_GROUP_BY (LeftParen qb_name_option RightParen)?
    | NO_PLACE_GROUP_BY (LeftParen qb_name_option RightParen)?
    | NO_PRED_DEDUCE (LeftParen qb_name_option RightParen)?
    | USE_JIT LeftParen use_jit_type RightParen
    | NO_USE_JIT
    | USE_LATE_MATERIALIZATION
    | NO_USE_LATE_MATERIALIZATION
    | TRACE_LOG
    | STAT LeftParen tracing_num_list RightParen
    | TRACING LeftParen tracing_num_list RightParen
    | USE_PX
    | NO_USE_PX
    | TRANS_PARAM LeftParen trans_param_name Comma? trans_param_value RightParen
    | PX_JOIN_FILTER
    | FORCE_REFRESH_LOCATION_CACHE
    | QB_NAME LeftParen NAME_OB RightParen
    | MAX_CONCURRENT LeftParen INTNUM RightParen
    | PARALLEL LeftParen INTNUM RightParen
    | PQ_DISTRIBUTE LeftParen qb_name_option relation_factor_in_pq_hint Comma? distribute_method (opt_comma distribute_method)? RightParen
    | LOAD_BATCH_SIZE LeftParen INTNUM RightParen
    | NAME_OB
    | PARSER_SYNTAX_ERROR
    ;

consistency_level
    : WEAK
    | STRONG
    | FROZEN
    ;

use_plan_cache_type
    : NONE
    | DEFAULT
    ;

use_jit_type
    : AUTO
    | FORCE
    ;

distribute_method
    : NONE
    | PARTITION
    | RANDOM
    | RANDOM_LOCAL
    | HASH
    | BROADCAST
    ;

limit_expr
    : INTNUM
    | QUESTIONMARK
    ;

opt_for_update_wait
    : empty
    | WAIT DECIMAL_VAL
    | WAIT INTNUM
    | NOWAIT
    | NO_WAIT
    ;

parameterized_trim
    : expr
    | expr FROM expr
    | BOTH expr FROM expr
    | LEADING expr FROM expr
    | TRAILING expr FROM expr
    | BOTH FROM expr
    | LEADING FROM expr
    | TRAILING FROM expr
    ;

groupby_clause
    : sort_list_for_group_by (WITH ROLLUP)?
    ;

sort_list_for_group_by
    : sort_key_for_group_by (Comma sort_key_for_group_by)*
    ;

sort_key_for_group_by
    : expr (ASC | DESC)?
    ;

order_by
    : ORDER BY sort_list
    ;

sort_list
    : sort_key (Comma sort_key)*
    ;

sort_key
    : expr (ASC | DESC)?
    ;

query_expression_option_list
    : query_expression_option
    | query_expression_option query_expression_option
    ;

query_expression_option
    : ALL
    | DISTINCT
    | UNIQUE
    | SQL_CALC_FOUND_ROWS
    ;

projection
    : expr
    | expr AS? column_label
    | expr AS? STRING_VALUE
    | Star
    ;

select_expr_list
    : projection (Comma projection)*
    ;

from_list
    : table_references
    ;

table_references
    : table_reference (Comma table_reference)*
    ;

table_reference
    : table_factor
    | joined_table
    ;

table_factor
    : tbl_name
    | table_subquery
    | select_with_parens
    | LeftParen table_references RightParen
    ;

tbl_name
    : relation_factor use_partition? (sample_clause opt_seed)? index_hint_list?
    | relation_factor use_partition? (sample_clause opt_seed)? AS relation_name index_hint_list?
    | relation_factor use_partition? relation_name index_hint_list?
    | relation_factor use_partition? sample_clause seed relation_name index_hint_list?
    | relation_factor use_partition? sample_clause relation_name index_hint_list?
    ;

dml_table_name
    : relation_factor use_partition?
    ;

seed
    : SEED LeftParen INTNUM RightParen
    ;

opt_seed
    : seed?
    ;

sample_percent
    : INTNUM
    | DECIMAL_VAL
    ;

sample_clause
    : SAMPLE BLOCK? (ALL | BASE | INCR)? LeftParen sample_percent RightParen
    ;

table_subquery
    : select_with_parens AS? relation_name
    ;

use_partition
    : PARTITION LeftParen name_list RightParen
    ;

index_hint_type
    : FORCE
    | IGNORE
    ;

key_or_index
    : KEY
    | INDEX
    ;

index_hint_scope
    : empty
    | FOR JOIN
    | FOR ORDER BY
    | FOR GROUP BY
    ;

index_element
    : NAME_OB
    | PRIMARY
    ;

index_list
    : index_element (Comma index_element)*
    ;

index_hint_definition
    : USE key_or_index index_hint_scope LeftParen index_list? RightParen
    | index_hint_type key_or_index index_hint_scope LeftParen index_list RightParen
    ;

index_hint_list
    : index_hint_definition
    | index_hint_definition index_hint_list
    ;

relation_factor
    : normal_relation_factor
    | dot_relation_factor
    ;

relation_with_star_list
    : relation_factor_with_star (Comma relation_factor_with_star)*
    ;

relation_factor_with_star
    : relation_name (Dot Star)?
    | relation_name Dot relation_name (Dot Star)?
    ;

normal_relation_factor
    : relation_name
    | relation_name Dot relation_name
    ;

dot_relation_factor
    : Dot relation_name
    ;

relation_factor_in_hint
    : normal_relation_factor qb_name_option
    ;

qb_name_option
    : At NAME_OB
    | empty
    ;

relation_factor_in_hint_list
    : relation_factor_in_hint (relation_sep_option relation_factor_in_hint)*
    ;

relation_sep_option
    : Comma?
    ;

relation_factor_in_pq_hint
    : relation_factor_in_hint
    | LeftParen relation_factor_in_hint_list RightParen
    ;

relation_factor_in_leading_hint
    : LeftParen relation_factor_in_hint_list RightParen
    ;

relation_factor_in_leading_hint_list
    : relation_factor_in_leading_hint
    | relation_factor_in_leading_hint_list relation_sep_option relation_factor_in_leading_hint
    | relation_factor_in_leading_hint_list relation_sep_option relation_factor_in_hint
    | LeftParen relation_factor_in_leading_hint_list RightParen
    | LeftParen relation_factor_in_hint_list relation_sep_option relation_factor_in_leading_hint_list RightParen
    | relation_factor_in_leading_hint_list relation_sep_option LeftParen relation_factor_in_hint_list relation_sep_option relation_factor_in_leading_hint_list RightParen
    ;

relation_factor_in_leading_hint_list_entry
    : relation_factor_in_leading_hint_list
    | relation_factor_in_hint_list relation_sep_option relation_factor_in_leading_hint_list
    ;

relation_factor_in_use_join_hint_list
    : relation_factor_in_hint
    | LeftParen relation_factor_in_hint_list RightParen
    | relation_factor_in_use_join_hint_list relation_sep_option relation_factor_in_hint
    | relation_factor_in_use_join_hint_list relation_sep_option LeftParen relation_factor_in_hint_list RightParen
    ;

tracing_num_list
    : INTNUM relation_sep_option tracing_num_list
    | INTNUM
    ;

join_condition
    : ON expr
    | USING LeftParen column_list RightParen
    ;

joined_table
    : table_factor inner_join_type table_factor
    | joined_table inner_join_type table_factor
    | table_factor inner_join_type table_factor ON expr
    | joined_table inner_join_type table_factor ON expr
    | table_factor inner_join_type table_factor USING LeftParen column_list RightParen
    | joined_table inner_join_type table_factor USING LeftParen column_list RightParen
    | table_factor outer_join_type table_factor join_condition
    | joined_table outer_join_type table_factor join_condition
    | table_factor natural_join_type table_factor
    | joined_table natural_join_type table_factor
    ;

natural_join_type
    : NATURAL outer_join_type
    | NATURAL INNER? JOIN
    ;

inner_join_type
    : JOIN
    | INNER JOIN
    | CROSS JOIN
    ;

outer_join_type
    : FULL OUTER? JOIN
    | LEFT OUTER? JOIN
    | RIGHT OUTER? JOIN
    ;

analyze_stmt
    : ANALYZE TABLE relation_factor UPDATE HISTOGRAM ON column_name_list WITH INTNUM BUCKETS
    | ANALYZE TABLE relation_factor DROP HISTOGRAM ON column_name_list
    ;

create_outline_stmt
    : CREATE (OR REPLACE)? OUTLINE relation_name ON explainable_stmt (TO explainable_stmt)?
    | CREATE (OR REPLACE)? OUTLINE relation_name ON STRING_VALUE USING HINT_HINT_BEGIN hint_list_with_end
    ;

alter_outline_stmt
    : ALTER OUTLINE relation_name ADD explainable_stmt (TO explainable_stmt)?
    ;

drop_outline_stmt
    : DROP OUTLINE relation_factor
    ;

explain_stmt
    : explain_or_desc relation_factor (STRING_VALUE | column_name)?
    | explain_or_desc explainable_stmt
    | explain_or_desc BASIC explainable_stmt
    | explain_or_desc OUTLINE explainable_stmt
    | explain_or_desc EXTENDED explainable_stmt
    | explain_or_desc EXTENDED_NOADDR explainable_stmt
    | explain_or_desc PLANREGRESS explainable_stmt
    | explain_or_desc PARTITIONS explainable_stmt
    | explain_or_desc FORMAT COMP_EQ format_name explainable_stmt
    ;

explain_or_desc
    : EXPLAIN
    | DESCRIBE
    | DESC
    ;

explainable_stmt
    : select_stmt
    | delete_stmt
    | insert_stmt
    | update_stmt
    ;

format_name
    : TRADITIONAL
    | JSON
    ;

show_stmt
    : SHOW FULL? TABLES (from_or_in database_factor)? ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW databases_or_schemas STATUS? ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW FULL? columns_or_fields from_or_in relation_factor (from_or_in database_factor)? ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW TABLE STATUS (from_or_in database_factor)? ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW PROCEDURE STATUS (from_or_in database_factor)? ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW FUNCTION STATUS (from_or_in database_factor)? ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW SERVER STATUS ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW (GLOBAL | SESSION | LOCAL)? VARIABLES ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW SCHEMA
    | SHOW CREATE database_or_schema (IF not EXISTS)? database_factor
    | SHOW CREATE TABLE relation_factor
    | SHOW CREATE VIEW relation_factor
    | SHOW CREATE PROCEDURE relation_factor
    | SHOW CREATE FUNCTION relation_factor
    | SHOW WARNINGS ((LIMIT INTNUM Comma INTNUM) | (LIMIT INTNUM))?
    | SHOW ERRORS ((LIMIT INTNUM Comma INTNUM) | (LIMIT INTNUM))?
    | SHOW COUNT LeftParen Star RightParen WARNINGS
    | SHOW COUNT LeftParen Star RightParen ERRORS
    | SHOW GRANTS opt_for_grant_user
    | SHOW charset_key ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW TRACE ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW COLLATION ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW PARAMETERS ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))? tenant_name?
    | SHOW index_or_indexes_or_keys from_or_in relation_factor (from_or_in database_factor)? (WHERE opt_hint_value expr)?
    | SHOW FULL? PROCESSLIST
    | SHOW TABLEGROUPS ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW (GLOBAL | SESSION | LOCAL)? STATUS ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW TENANT STATUS?
    | SHOW CREATE TENANT relation_name
    | SHOW STORAGE? ENGINES
    | SHOW PRIVILEGES
    | SHOW RECYCLEBIN
    | SHOW CREATE TABLEGROUP relation_name
    ;

databases_or_schemas
    : DATABASES
    | SCHEMAS
    ;

opt_for_grant_user
    : opt_for_user
    | FOR CURRENT_USER
    | FOR CURRENT_USER LeftParen RightParen
    ;

columns_or_fields
    : COLUMNS
    | FIELDS
    ;

database_or_schema
    : DATABASE
    | SCHEMA
    ;

index_or_indexes_or_keys
    : INDEX
    | INDEXES
    | KEYS
    ;

from_or_in
    : FROM
    | IN
    ;

help_stmt
    : HELP STRING_VALUE
    | HELP NAME_OB
    ;

create_tablespace_stmt
    : CREATE TABLESPACE tablespace permanent_tablespace
    ;

permanent_tablespace
    : permanent_tablespace_options?
    ;

permanent_tablespace_option
    : ENCRYPTION COMP_EQ? STRING_VALUE
    ;

drop_tablespace_stmt
    : DROP TABLESPACE tablespace
    ;

alter_tablespace_actions
    : alter_tablespace_action
    | alter_tablespace_action Comma alter_tablespace_action
    ;

alter_tablespace_action
    : SET? permanent_tablespace_option
    ;

alter_tablespace_stmt
    : ALTER TABLESPACE tablespace alter_tablespace_actions
    ;

rotate_master_key_stmt
    : ALTER INSTANCE ROTATE INNODB MASTER KEY
    ;

permanent_tablespace_options
    : permanent_tablespace_option (Comma permanent_tablespace_option)*
    ;

create_user_stmt
    : CREATE USER (IF not EXISTS)? user_specification_list
    | CREATE USER (IF not EXISTS)? user_specification_list require_specification
    ;

user_specification_list
    : user_specification (Comma user_specification)*
    ;

user_specification
    : user USER_VARIABLE?
    | user USER_VARIABLE? IDENTIFIED BY password
    | user USER_VARIABLE? IDENTIFIED BY PASSWORD password
    ;

require_specification
    : REQUIRE NONE
    | REQUIRE SSL
    | REQUIRE X509
    | REQUIRE tls_option_list
    ;

tls_option_list
    : tls_option
    | tls_option_list tls_option
    | tls_option_list AND tls_option
    ;

tls_option
    : CIPHER STRING_VALUE
    | ISSUER STRING_VALUE
    | SUBJECT STRING_VALUE
    ;

user
    : STRING_VALUE
    | NAME_OB
    | unreserved_keyword
    ;

opt_host_name
    : USER_VARIABLE?
    ;

user_with_host_name
    : user USER_VARIABLE?
    ;

password
    : STRING_VALUE
    ;

drop_user_stmt
    : DROP USER user_list
    ;

user_list
    : user_with_host_name (Comma user_with_host_name)*
    ;

set_password_stmt
    : SET PASSWORD (FOR user opt_host_name)? COMP_EQ STRING_VALUE
    | SET PASSWORD (FOR user opt_host_name)? COMP_EQ PASSWORD LeftParen password RightParen
    | ALTER USER user_with_host_name IDENTIFIED BY password
    | ALTER USER user_with_host_name require_specification
    ;

opt_for_user
    : FOR user opt_host_name
    | empty
    ;

rename_user_stmt
    : RENAME USER rename_list
    ;

rename_info
    : user USER_VARIABLE? TO user USER_VARIABLE?
    ;

rename_list
    : rename_info (Comma rename_info)*
    ;

lock_user_stmt
    : ALTER USER user_list ACCOUNT lock_spec_mysql57
    ;

lock_spec_mysql57
    : LOCK_
    | UNLOCK
    ;

lock_tables_stmt
    : LOCK_ table_or_tables lock_table_list
    ;

unlock_tables_stmt
    : UNLOCK TABLES
    ;

lock_table_list
    : lock_table (Comma lock_table)*
    ;

lock_table
    : relation_factor lock_type
    | relation_factor AS? relation_name lock_type
    ;

lock_type
    : READ LOCAL?
    | WRITE
    | LOW_PRIORITY WRITE
    ;

begin_stmt
    : BEGI WORK?
    | START TRANSACTION ((WITH CONSISTENT SNAPSHOT) | transaction_access_mode | (WITH CONSISTENT SNAPSHOT Comma transaction_access_mode) | (transaction_access_mode Comma WITH CONSISTENT SNAPSHOT))?
    ;

commit_stmt
    : COMMIT WORK?
    ;

rollback_stmt
    : ROLLBACK WORK?
    ;

kill_stmt
    : KILL expr
    | KILL CONNECTION expr
    | KILL QUERY expr
    ;

grant_stmt
    : GRANT grant_privileges ON priv_level TO user_specification_list grant_options
    ;

grant_privileges
    : priv_type_list
    | ALL PRIVILEGES?
    ;

priv_type_list
    : priv_type (Comma priv_type)*
    ;

priv_type
    : ALTER
    | CREATE
    | CREATE USER
    | DELETE
    | DROP
    | GRANT OPTION
    | INSERT
    | UPDATE
    | SELECT
    | INDEX
    | CREATE VIEW
    | SHOW VIEW
    | SHOW DATABASES
    | SUPER
    | PROCESS
    | USAGE
    | CREATE SYNONYM
    ;

priv_level
    : Star
    | Star Dot Star
    | relation_name Dot Star
    | relation_name
    | relation_name Dot relation_name
    ;

grant_options
    : WITH GRANT OPTION
    | empty
    ;

revoke_stmt
    : REVOKE grant_privileges ON priv_level FROM user_list
    | REVOKE ALL PRIVILEGES? Comma GRANT OPTION FROM user_list
    ;

prepare_stmt
    : PREPARE stmt_name FROM preparable_stmt
    ;

stmt_name
    : column_label
    ;

preparable_stmt
    : text_string
    | USER_VARIABLE
    ;

variable_set_stmt
    : SET var_and_val_list
    ;

sys_var_and_val_list
    : sys_var_and_val (Comma sys_var_and_val)*
    ;

var_and_val_list
    : var_and_val (Comma var_and_val)*
    ;

set_expr_or_default
    : expr
    | ON
    | OFF
    | BINARY
    | DEFAULT
    ;

var_and_val
    : USER_VARIABLE to_or_eq expr
    | USER_VARIABLE SET_VAR expr
    | sys_var_and_val
    | scope_or_scope_alias column_name to_or_eq set_expr_or_default
    | SYSTEM_VARIABLE to_or_eq set_expr_or_default
    ;

sys_var_and_val
    : var_name to_or_eq set_expr_or_default
    | var_name SET_VAR set_expr_or_default
    ;

scope_or_scope_alias
    : GLOBAL
    | SESSION
    | GLOBAL_ALIAS Dot
    | SESSION_ALIAS Dot
    ;

to_or_eq
    : TO
    | COMP_EQ
    ;

execute_stmt
    : EXECUTE stmt_name (USING argument_list)?
    ;

argument_list
    : argument (Comma argument)*
    ;

argument
    : USER_VARIABLE
    ;

deallocate_prepare_stmt
    : deallocate_or_drop PREPARE stmt_name
    ;

deallocate_or_drop
    : DEALLOCATE
    | DROP
    ;

truncate_table_stmt
    : TRUNCATE TABLE? relation_factor
    ;

rename_table_stmt
    : RENAME TABLE rename_table_actions
    ;

rename_table_actions
    : rename_table_action (Comma rename_table_action)*
    ;

rename_table_action
    : relation_factor TO relation_factor
    ;

alter_table_stmt
    : ALTER TABLE relation_factor alter_table_actions
    ;

alter_table_actions
    : alter_table_action
    | alter_table_actions Comma alter_table_action
    | empty
    ;

alter_table_action
    : SET? table_option_list_space_seperated
    | alter_column_option
    | alter_tablegroup_option
    | RENAME TO? relation_factor
    | alter_index_option
    | alter_partition_option
    | alter_constraint_option
    | alter_foreign_key_action
    ;

alter_constraint_option
    : DROP CONSTRAINT LeftParen name_list RightParen
    | ADD CONSTRAINT constraint_name CHECK LeftParen expr RightParen
    ;

alter_partition_option
    : DROP PARTITION drop_partition_name_list
    | ADD PARTITION opt_partition_range_or_list
    | modify_partition_info
    | REORGANIZE PARTITION name_list INTO opt_partition_range_or_list
    | TRUNCATE PARTITION name_list
    ;

opt_partition_range_or_list
    : opt_range_partition_list
    | opt_list_partition_list
    ;

alter_tg_partition_option
    : DROP PARTITION drop_partition_name_list
    | ADD PARTITION opt_partition_range_or_list
    | modify_tg_partition_info
    | REORGANIZE PARTITION name_list INTO opt_partition_range_or_list
    | TRUNCATE PARTITION name_list
    ;

drop_partition_name_list
    : name_list
    | LeftParen name_list RightParen
    ;

modify_partition_info
    : hash_partition_option
    | key_partition_option
    | range_partition_option
    | list_partition_option
    ;

modify_tg_partition_info
    : tg_hash_partition_option
    | tg_key_partition_option
    | tg_range_partition_option
    | tg_list_partition_option
    ;

alter_index_option
    : ADD key_or_index index_name? index_using_algorithm? LeftParen sort_column_list RightParen opt_index_options?
    | ADD UNIQUE key_or_index? index_name? index_using_algorithm? LeftParen sort_column_list RightParen opt_index_options?
    | ADD CONSTRAINT constraint_name? UNIQUE key_or_index? index_name? index_using_algorithm? LeftParen sort_column_list RightParen opt_index_options?
    | ADD FULLTEXT key_or_index? index_name? index_using_algorithm? LeftParen sort_column_list RightParen opt_index_options?
    | DROP key_or_index index_name
    | ADD (CONSTRAINT opt_constraint_name)? PRIMARY KEY LeftParen column_name_list RightParen opt_index_options?
    | ALTER INDEX index_name visibility_option
    | RENAME key_or_index index_name TO index_name
    ;

alter_foreign_key_action
    : DROP FOREIGN KEY index_name
    | ADD (CONSTRAINT opt_constraint_name)? FOREIGN KEY index_name? LeftParen column_name_list RightParen REFERENCES relation_factor LeftParen column_name_list RightParen (MATCH match_action)? (opt_reference_option_list reference_option)?
    ;

visibility_option
    : VISIBLE
    | INVISIBLE
    ;

alter_column_option
    : ADD COLUMN? column_definition
    | ADD COLUMN? LeftParen column_definition_list RightParen
    | DROP COLUMN? column_name (CASCADE | RESTRICT)?
    | ALTER COLUMN? column_name alter_column_behavior
    | CHANGE COLUMN? column_name column_definition
    | MODIFY COLUMN? column_definition
    ;

alter_tablegroup_option
    : DROP TABLEGROUP
    ;

alter_column_behavior
    : SET DEFAULT signed_literal
    | DROP DEFAULT
    ;

flashback_stmt
    : FLASHBACK TABLE relation_factor TO BEFORE DROP (RENAME TO relation_factor)?
    | FLASHBACK database_key database_factor TO BEFORE DROP (RENAME TO database_factor)?
    | FLASHBACK TENANT relation_name TO BEFORE DROP (RENAME TO relation_name)?
    ;

purge_stmt
    : PURGE TABLE relation_factor
    | PURGE INDEX relation_factor
    | PURGE database_key database_factor
    | PURGE TENANT relation_name
    | PURGE RECYCLEBIN
    ;

optimize_stmt
    : OPTIMIZE TABLE table_list
    | OPTIMIZE TENANT relation_name
    | OPTIMIZE TENANT ALL
    ;

dump_memory_stmt
    : DUMP ENTITY ALL
    | DUMP ENTITY P_ENTITY COMP_EQ STRING_VALUE Comma SLOT_IDX COMP_EQ INTNUM
    | DUMP CHUNK ALL
    | DUMP CHUNK TENANT_ID COMP_EQ INTNUM Comma CTX_ID COMP_EQ INTNUM
    | DUMP CHUNK P_CHUNK COMP_EQ STRING_VALUE
    | SET LEAK_MOD COMP_EQ STRING_VALUE
    | DUMP MEMORY LEAK
    ;

alter_system_stmt
    : ALTER SYSTEM BOOTSTRAP (CLUSTER cluster_role)? server_info_list
    | ALTER SYSTEM FLUSH cache_type CACHE (TENANT COMP_EQ tenant_name_list)? flush_scope
    | ALTER SYSTEM FLUSH SQL cache_type (TENANT COMP_EQ tenant_name_list)? flush_scope
    | ALTER SYSTEM FLUSH KVCACHE tenant_name? cache_name?
    | ALTER SYSTEM FLUSH ILOGCACHE file_id?
    | ALTER SYSTEM ALTER PLAN BASELINE tenant_name? sql_id_expr? baseline_id_expr? SET baseline_asgn_factor
    | ALTER SYSTEM LOAD PLAN BASELINE FROM PLAN CACHE (TENANT COMP_EQ tenant_name_list)? sql_id_expr?
    | ALTER SYSTEM SWITCH REPLICA partition_role partition_id_or_server_or_zone
    | ALTER SYSTEM SWITCH ROOTSERVICE partition_role server_or_zone
    | ALTER SYSTEM alter_or_change_or_modify REPLICA partition_id_desc ip_port alter_or_change_or_modify change_actions FORCE?
    | ALTER SYSTEM DROP REPLICA partition_id_desc ip_port (CREATE_TIMESTAMP opt_equal_mark INTNUM)? zone_desc? FORCE?
    | ALTER SYSTEM migrate_action REPLICA partition_id_desc SOURCE COMP_EQ? STRING_VALUE DESTINATION COMP_EQ? STRING_VALUE FORCE?
    | ALTER SYSTEM REPORT REPLICA server_or_zone?
    | ALTER SYSTEM RECYCLE REPLICA server_or_zone?
    | ALTER SYSTEM START MERGE zone_desc
    | ALTER SYSTEM suspend_or_resume MERGE zone_desc?
    | ALTER SYSTEM CLEAR MERGE ERROR_P
    | ALTER SYSTEM CANCEL cancel_task_type TASK STRING_VALUE
    | ALTER SYSTEM MAJOR FREEZE (IGNORE server_list)?
    | ALTER SYSTEM CHECKPOINT
    | ALTER SYSTEM MINOR FREEZE (tenant_list_tuple | partition_id_desc)? (SERVER opt_equal_mark LeftParen server_list RightParen)? zone_desc?
    | ALTER SYSTEM CLEAR ROOTTABLE tenant_name?
    | ALTER SYSTEM server_action SERVER server_list zone_desc?
    | ALTER SYSTEM ADD ZONE relation_name_or_string add_or_alter_zone_options
    | ALTER SYSTEM zone_action ZONE relation_name_or_string
    | ALTER SYSTEM alter_or_change_or_modify ZONE relation_name_or_string SET? add_or_alter_zone_options
    | ALTER SYSTEM REFRESH SCHEMA server_or_zone?
    | ALTER SYSTEM REFRESH MEMORY STAT server_or_zone?
    | ALTER SYSTEM SET? alter_system_set_parameter_actions
    | ALTER SYSTEM SET_TP alter_system_settp_actions
    | ALTER SYSTEM CLEAR LOCATION CACHE server_or_zone?
    | ALTER SYSTEM REMOVE BALANCE TASK (TENANT COMP_EQ tenant_name_list)? (ZONE COMP_EQ zone_list)? (TYPE opt_equal_mark balance_task_type)?
    | ALTER SYSTEM RELOAD GTS
    | ALTER SYSTEM RELOAD UNIT
    | ALTER SYSTEM RELOAD SERVER
    | ALTER SYSTEM RELOAD ZONE
    | ALTER SYSTEM MIGRATE UNIT COMP_EQ? INTNUM DESTINATION COMP_EQ? STRING_VALUE
    | ALTER SYSTEM CANCEL MIGRATE UNIT INTNUM
    | ALTER SYSTEM UPGRADE VIRTUAL SCHEMA
    | ALTER SYSTEM RUN JOB STRING_VALUE server_or_zone?
    | ALTER SYSTEM upgrade_action UPGRADE
    | ALTER SYSTEM upgrade_action ROLLING UPGRADE
    | ALTER SYSTEM REFRESH TIME_ZONE_INFO
    | ALTER SYSTEM SET DISK VALID ip_port
    | ALTER SYSTEM RESTORE tenant_name FROM STRING_VALUE
    | ALTER SYSTEM RESTORE relation_name FROM relation_name AT STRING_VALUE UNTIL STRING_VALUE WITH STRING_VALUE
    | ALTER SYSTEM CHANGE TENANT change_tenant_name_or_tenant_id
    | ALTER SYSTEM DROP TABLES IN SESSION INTNUM
    | ALTER SYSTEM REFRESH TABLES IN SESSION INTNUM
    | ALTER DISKGROUP relation_name ADD DISK STRING_VALUE (NAME opt_equal_mark relation_name_or_string)? ip_port zone_desc?
    | ALTER DISKGROUP relation_name DROP DISK STRING_VALUE ip_port zone_desc?
    | ALTER SYSTEM ARCHIVELOG
    | ALTER SYSTEM NOARCHIVELOG
    | ALTER SYSTEM BACKUP DATABASE
    | ALTER SYSTEM BACKUP INCREMENTAL DATABASE
    | ALTER SYSTEM CANCEL BACKUP
    | ALTER SYSTEM SUSPEND BACKUP
    | ALTER SYSTEM RESUME BACKUP
    | ALTER SYSTEM DELETE EXPIRED BACKUP
    | ALTER SYSTEM DELETE BACKUP INTNUM
    | ALTER SYSTEM VALIDATE DATABASE
    | ALTER SYSTEM VALIDATE BACKUPSET INTNUM
    ;

change_tenant_name_or_tenant_id
    : relation_name_or_string
    | TENANT_ID COMP_EQ? INTNUM
    ;

cache_type
    : ALL
    | LOCATION
    | CLOG
    | ILOG
    | COLUMN_STAT
    | BLOCK_INDEX
    | BLOCK
    | ROW
    | BLOOM_FILTER
    | SCHEMA
    | PLAN
    | AUDIT
    | PL
    ;

balance_task_type
    : AUTO
    | MANUAL
    | ALL
    ;

tenant_list_tuple
    : TENANT COMP_EQ? LeftParen tenant_name_list RightParen
    ;

tenant_name_list
    : relation_name_or_string (Comma relation_name_or_string)*
    ;

flush_scope
    : GLOBAL?
    ;

server_info_list
    : server_info (Comma server_info)*
    ;

server_info
    : REGION COMP_EQ? relation_name_or_string ZONE COMP_EQ? relation_name_or_string SERVER COMP_EQ? STRING_VALUE
    | ZONE COMP_EQ? relation_name_or_string SERVER COMP_EQ? STRING_VALUE
    ;

server_action
    : ADD
    | DELETE
    | CANCEL DELETE
    | START
    | STOP
    | FORCE STOP
    | ISOLATE
    ;

server_list
    : STRING_VALUE (Comma STRING_VALUE)*
    ;

zone_action
    : DELETE
    | START
    | STOP
    | FORCE STOP
    ;

ip_port
    : SERVER COMP_EQ? STRING_VALUE
    ;

zone_desc
    : ZONE COMP_EQ? relation_name_or_string
    ;

server_or_zone
    : ip_port
    | zone_desc
    ;

add_or_alter_zone_option
    : REGION COMP_EQ? relation_name_or_string
    | IDC COMP_EQ? relation_name_or_string
    | ZONE_TYPE COMP_EQ? relation_name_or_string
    ;

add_or_alter_zone_options
    : add_or_alter_zone_option
    | add_or_alter_zone_options Comma add_or_alter_zone_option
    | empty
    ;

alter_or_change_or_modify
    : ALTER
    | CHANGE
    | MODIFY
    ;

partition_id_desc
    : PARTITION_ID COMP_EQ? STRING_VALUE
    ;

partition_id_or_server_or_zone
    : partition_id_desc ip_port
    | ip_port tenant_name?
    | zone_desc tenant_name?
    ;

migrate_action
    : MOVE
    | COPY
    ;

change_actions
    : change_action
    | change_action change_actions
    ;

change_action
    : replica_type
    ;

replica_type
    : REPLICA_TYPE COMP_EQ? STRING_VALUE
    ;

suspend_or_resume
    : SUSPEND
    | RESUME
    ;

baseline_id_expr
    : BASELINE_ID COMP_EQ? INTNUM
    ;

sql_id_expr
    : SQL_ID COMP_EQ? STRING_VALUE
    ;

baseline_asgn_factor
    : column_name COMP_EQ literal
    ;

tenant_name
    : TENANT COMP_EQ? relation_name_or_string
    ;

cache_name
    : CACHE COMP_EQ? relation_name_or_string
    ;

file_id
    : FILE_ID COMP_EQ? INTNUM
    ;

cancel_task_type
    : PARTITION MIGRATION
    | empty
    ;

alter_system_set_parameter_actions
    : alter_system_set_parameter_action (Comma alter_system_set_parameter_action)*
    ;

alter_system_set_parameter_action
    : NAME_OB COMP_EQ conf_const (COMMENT STRING_VALUE)? ((SCOPE COMP_EQ MEMORY) | (SCOPE COMP_EQ SPFILE) | (SCOPE COMP_EQ BOTH))? server_or_zone? tenant_name?
    | TABLET_SIZE COMP_EQ conf_const (COMMENT STRING_VALUE)? ((SCOPE COMP_EQ MEMORY) | (SCOPE COMP_EQ SPFILE) | (SCOPE COMP_EQ BOTH))? server_or_zone? tenant_name?
    | CLUSTER_ID COMP_EQ conf_const (COMMENT STRING_VALUE)? ((SCOPE COMP_EQ MEMORY) | (SCOPE COMP_EQ SPFILE) | (SCOPE COMP_EQ BOTH))? server_or_zone? tenant_name?
    ;

alter_system_settp_actions
    : settp_option
    | alter_system_settp_actions Comma settp_option
    | empty
    ;

settp_option
    : TP_NO COMP_EQ? INTNUM
    | TP_NAME COMP_EQ? relation_name_or_string
    | OCCUR COMP_EQ? INTNUM
    | FREQUENCY COMP_EQ? INTNUM
    | ERROR_CODE COMP_EQ? INTNUM
    ;

cluster_role
    : PRIMARY
    | STANDBY
    ;

partition_role
    : LEADER
    | FOLLOWER
    ;

upgrade_action
    : BEGI
    | END
    ;

set_names_stmt
    : SET NAMES charset_name_or_default collation?
    ;

set_charset_stmt
    : SET charset_key charset_name_or_default
    ;

set_transaction_stmt
    : SET (GLOBAL | SESSION | LOCAL)? TRANSACTION transaction_characteristics
    ;

transaction_characteristics
    : transaction_access_mode
    | ISOLATION LEVEL isolation_level
    | transaction_access_mode Comma ISOLATION LEVEL isolation_level
    | ISOLATION LEVEL isolation_level Comma transaction_access_mode
    ;

transaction_access_mode
    : READ ONLY
    | READ WRITE
    ;

isolation_level
    : READ UNCOMMITTED
    | READ COMMITTED
    | REPEATABLE READ
    | SERIALIZABLE
    ;

create_savepoint_stmt
    : SAVEPOINT var_name
    ;

rollback_savepoint_stmt
    : ROLLBACK TO var_name
    | ROLLBACK WORK TO var_name
    | ROLLBACK TO SAVEPOINT var_name
    ;

release_savepoint_stmt
    : RELEASE SAVEPOINT var_name
    ;

alter_cluster_stmt
    : ALTER SYSTEM cluster_action cluster_name CLUSTER_ID COMP_EQ? conf_const FORCE?
    ;

cluster_action
    : ADD CLUSTER
    | REMOVE CLUSTER
    | ENABLE CLUSTER SYNCHRONIZATION
    | DISABLE CLUSTER SYNCHRONIZATION
    ;

switchover_cluster_stmt
    : ALTER SYSTEM commit_switchover_clause
    ;

commit_switchover_clause
    : COMMIT TO SWITCHOVER TO PRIMARY (WITH SESSION SHUTDOWN)?
    | COMMIT TO SWITCHOVER TO PHYSICAL STANDBY (WITH SESSION SHUTDOWN)?
    | ACTIVATE PHYSICAL STANDBY CLUSTER
    | CONVERT TO PHYSICAL STANDBY
    ;

cluster_name
    : relation_name
    | STRING_VALUE
    ;

var_name
    : NAME_OB
    | unreserved_keyword_normal
    ;

column_name
    : NAME_OB
    | unreserved_keyword
    ;

relation_name
    : NAME_OB
    | unreserved_keyword
    ;

function_name
    : NAME_OB
    | DUMP
    | CHARSET
    | COLLATION
    | KEY_VERSION
    | USER
    | DATABASE
    | COALESCE
    | REPEAT
    | ROW_COUNT
    | REVERSE
    | RIGHT
    | CURRENT_USER
    | SYSTEM_USER
    | SESSION_USER
    | REPLACE
    | TRUNCATE
    ;

column_label
    : NAME_OB
    | unreserved_keyword
    ;

date_unit
    : DAY
    | DAY_HOUR
    | DAY_MICROSECOND
    | DAY_MINUTE
    | DAY_SECOND
    | HOUR
    | HOUR_MICROSECOND
    | HOUR_MINUTE
    | HOUR_SECOND
    | MICROSECOND
    | MINUTE
    | MINUTE_MICROSECOND
    | MINUTE_SECOND
    | MONTH
    | QUARTER
    | SECOND
    | SECOND_MICROSECOND
    | WEEK
    | YEAR
    | YEAR_MONTH
    ;

unreserved_keyword
    : unreserved_keyword_normal
    | unreserved_keyword_special
    ;

unreserved_keyword_normal
    : ACCOUNT
    | ACTION
    | ACTIVE
    | ADDDATE
    | AFTER
    | AGGREGATE
    | ALGORITHM
    | ANALYSE
    | ANY
    | APPROX_COUNT_DISTINCT
    | APPROX_COUNT_DISTINCT_SYNOPSIS
    | APPROX_COUNT_DISTINCT_SYNOPSIS_MERGE
    | ASCII
    | AT
    | AUTHORS
    | AUTO
    | AUTOEXTEND_SIZE
    | AUTO_INCREMENT
    | AVG
    | AVG_ROW_LENGTH
    | BACKUP
    | BASE
    | BASELINE
    | BASELINE_ID
    | BASIC
    | BALANCE
    | BINDING
    | BINLOG
    | BIT
    | BLOCK
    | BLOCK_SIZE
    | BLOCK_INDEX
    | BLOOM_FILTER
    | BOOL
    | BOOLEAN
    | BOOTSTRAP
    | BTREE
    | BYTE
    | BREADTH
    | CACHE
    | KVCACHE
    | ILOGCACHE
    | CANCEL
    | CASCADED
    | CAST
    | CATALOG_NAME
    | CHAIN
    | CHANGED
    | CHARSET
    | CHECKSUM
    | CHECKPOINT
    | CHUNK
    | CIPHER
    | CLASS_ORIGIN
    | CLEAN
    | CLEAR
    | CLIENT
    | CLOSE
    | CLOG
    | CLUSTER
    | CLUSTER_ID
    | CLUSTER_NAME
    | COALESCE
    | CODE
    | COLLATION
    | COLUMN_FORMAT
    | COLUMN_NAME
    | COLUMN_STAT
    | COLUMNS
    | COMMENT
    | COMMIT
    | COMMITTED
    | COMPACT
    | COMPLETION
    | COMPRESSED
    | COMPRESSION
    | CONCURRENT
    | CONNECTION
    | CONSISTENT
    | CONSISTENT_MODE
    | CONSTRAINT_CATALOG
    | CONSTRAINT_NAME
    | CONSTRAINT_SCHEMA
    | CONTAINS
    | CONTEXT
    | CONTRIBUTORS
    | COPY
    | COUNT
    | CPU
    | CREATE_TIMESTAMP
    | CTXCAT
    | CTX_ID
    | CUBE
    | CUME_DIST
    | CURDATE
    | CURRENT
    | CURSOR_NAME
    | CURTIME
    | DATA
    | DATABASE_ID
    | DATAFILE
    | DATA_TABLE_ID
    | DATE
    | DATE_ADD
    | DATE_SUB
    | DATETIME
    | DAY
    | DEALLOCATE
    | DEFAULT_AUTH
    | DEFINER
    | DELAY
    | DELAY_KEY_WRITE
    | DENSE_RANK
    | DEPTH
    | DES_KEY_FILE
    | DESTINATION
    | DIAGNOSTICS
    | DIRECTORY
    | DISABLE
    | DISCARD
    | DISK
    | DISKGROUP
    | DO
    | DUMP
    | DUMPFILE
    | DUPLICATE
    | DUPLICATE_SCOPE
    | DYNAMIC
    | DEFAULT_TABLEGROUP
    | EFFECTIVE
    | ENABLE
    | ENCRYPTION
    | END
    | ENDS
    | ENGINE_
    | ENGINES
    | ENUM
    | ENTITY
    | ERROR_CODE
    | ERROR_P
    | ERRORS
    | ESCAPE
    | EVENT
    | EVENTS
    | EVERY
    | EXCHANGE
    | EXECUTE
    | EXPANSION
    | EXPIRE
    | EXPIRE_INFO
    | EXPORT
    | EXTENDED
    | EXTENDED_NOADDR
    | EXTENT_SIZE
    | EXTRACT
    | FAST
    | FAULTS
    | FIELDS
    | FILEX
    | FILE_ID
    | FINAL_COUNT
    | FIRST
    | FIRST_VALUE
    | FIXED
    | FLUSH
    | FOLLOWER
    | FORMAT
    | FOUND
    | FREEZE
    | FREQUENCY
    | FUNCTION
    | GENERAL
    | GEOMETRY
    | GEOMETRYCOLLECTION
    | GET_FORMAT
    | GLOBAL
    | GLOBAL_NAME
    | GRANTS
    | GROUPING
    | GROUP_CONCAT
    | GTS
    | HANDLER
    | HASH
    | HELP
    | HOST
    | HOSTS
    | HOUR
    | ID
    | IDC
    | IDENTIFIED
    | IGNORE_SERVER_IDS
    | ILOG
    | IMPORT
    | INDEXES
    | INDEX_TABLE_ID
    | INCR
    | INFO
    | INITIAL_SIZE
    | INNODB
    | INSERT_METHOD
    | INSTALL
    | INSTANCE
    | INVOKER
    | IO
    | IO_THREAD
    | IPC
    | ISOLATION
    | ISOLATE
    | ISSUER
    | IS_TENANT_SYS_POOL
    | JOB
    | JSON
    | KEY_BLOCK_SIZE
    | KEY_VERSION
    | LAG
    | LANGUAGE
    | LAST
    | LAST_VALUE
    | LEAD
    | LEADER
    | LEAVES
    | LESS
    | LEVEL
    | LINESTRING
    | LIST_
    | LISTAGG
    | LOCAL
    | LOCALITY
    | LOCKED
    | LOCKS
    | LOGFILE
    | LOGONLY_REPLICA_NUM
    | LOGS
    | MAJOR
    | MANUAL
    | MASTER
    | MASTER_AUTO_POSITION
    | MASTER_CONNECT_RETRY
    | MASTER_DELAY
    | MASTER_HEARTBEAT_PERIOD
    | MASTER_HOST
    | MASTER_LOG_FILE
    | MASTER_LOG_POS
    | MASTER_PASSWORD
    | MASTER_PORT
    | MASTER_RETRY_COUNT
    | MASTER_SERVER_ID
    | MASTER_SSL
    | MASTER_SSL_CA
    | MASTER_SSL_CAPATH
    | MASTER_SSL_CERT
    | MASTER_SSL_CIPHER
    | MASTER_SSL_CRL
    | MASTER_SSL_CRLPATH
    | MASTER_SSL_KEY
    | MASTER_USER
    | MAX
    | MAX_CONNECTIONS_PER_HOUR
    | MAX_CPU
    | MAX_DISK_SIZE
    | MAX_IOPS
    | MAX_MEMORY
    | MAX_QUERIES_PER_HOUR
    | MAX_ROWS
    | MAX_SESSION_NUM
    | MAX_SIZE
    | MAX_UPDATES_PER_HOUR
    | MAX_USER_CONNECTIONS
    | MEDIUM
    | MEMORY
    | MEMTABLE
    | MESSAGE_TEXT
    | META
    | MICROSECOND
    | MIGRATE
    | MIGRATION
    | MIN
    | MIN_CPU
    | MIN_IOPS
    | MIN_MEMORY
    | MINOR
    | MIN_ROWS
    | MINUTE
    | MODE
    | MODIFY
    | MONTH
    | MOVE
    | MULTILINESTRING
    | MULTIPOINT
    | MULTIPOLYGON
    | MUTEX
    | MYSQL_ERRNO
    | MAX_USED_PART_ID
    | NAME
    | NAMES
    | NATIONAL
    | NCHAR
    | NDB
    | NDBCLUSTER
    | NEW
    | NEXT
    | NO
    | NODEGROUP
    | NONE
    | NORMAL
    | NOW
    | NOWAIT
    | NO_WAIT
    | NTILE
    | NTH_VALUE
    | NUMBER
    | NVARCHAR
    | OCCUR
    | OFFSET
    | OLD_PASSWORD
    | OLD_KEY
    | OVER
    | ONE
    | ONE_SHOT
    | ONLY
    | OPEN
    | OPTIONS
    | REMOTE_OSS
    | OUTLINE
    | OWNER
    | PACK_KEYS
    | PAGE
    | PARAMETERS
    | PARSER
    | PARTIAL
    | PARTITION_ID
    | PARTITIONING
    | PARTITIONS
    | PERCENT_RANK
    | PAUSE
    | PHASE
    | PHYSICAL
    | PLANREGRESS
    | PLUGIN
    | PLUGIN_DIR
    | PLUGINS
    | POINT
    | POLYGON
    | POOL
    | PORT
    | POSITION
    | PREPARE
    | PRESERVE
    | PREV
    | PRIMARY_ZONE
    | PRIVILEGES
    | PROCESS
    | PROCESSLIST
    | PROFILE
    | PROFILES
    | PROGRESSIVE_MERGE_NUM
    | PROXY
    | PCTFREE
    | P_ENTITY
    | P_CHUNK
    | QUARTER
    | QUERY
    | QUICK
    | RANK
    | READ_ONLY
    | REBUILD
    | RECOVER
    | RECYCLE
    | RECYCLEBIN
    | ROTATE
    | ROW_NUMBER
    | REDO_BUFFER_SIZE
    | REDOFILE
    | REDUNDANT
    | REFRESH
    | REGION
    | RELAY
    | RELAYLOG
    | RELAY_LOG_FILE
    | RELAY_LOG_POS
    | RELAY_THREAD
    | RELOAD
    | REMOVE
    | REORGANIZE
    | REPAIR
    | REPEATABLE
    | REPLICA
    | REPLICA_NUM
    | REPLICA_TYPE
    | REPLICATION
    | REPORT
    | RESET
    | RESOURCE
    | RESOURCE_POOL_LIST
    | RESPECT
    | RESTART
    | RESTORE
    | RESUME
    | RETURNED_SQLSTATE
    | RETURNS
    | REVERSE
    | REWRITE_MERGE_VERSION
    | ROLLBACK
    | ROLLING
    | ROLLUP
    | ROOT
    | ROOTTABLE
    | ROUTINE
    | ROW
    | ROW_COUNT
    | ROW_FORMAT
    | ROWS
    | RTREE
    | RUN
    | SAMPLE
    | SAVEPOINT
    | SCHEDULE
    | SCHEMA_NAME
    | SCOPE
    | SECOND
    | SECURITY
    | SEED
    | SERIAL
    | SERIALIZABLE
    | SERVER
    | SERVER_IP
    | SERVER_PORT
    | SERVER_TYPE
    | SESSION
    | SESSION_USER
    | SET_MASTER_CLUSTER
    | SET_SLAVE_CLUSTER
    | SET_TP
    | SHARE
    | SHUTDOWN
    | SIGNED
    | SIMPLE
    | SLAVE
    | SLOW
    | SNAPSHOT
    | SOCKET
    | SOME
    | SONAME
    | SOUNDS
    | SOURCE
    | SPFILE
    | SPLIT
    | SQL_AFTER_GTIDS
    | SQL_AFTER_MTS_GAPS
    | SQL_BEFORE_GTIDS
    | SQL_BUFFER_RESULT
    | SQL_CACHE
    | SQL_ID
    | SQL_NO_CACHE
    | SQL_THREAD
    | SQL_TSI_DAY
    | SQL_TSI_HOUR
    | SQL_TSI_MINUTE
    | SQL_TSI_MONTH
    | SQL_TSI_QUARTER
    | SQL_TSI_SECOND
    | SQL_TSI_WEEK
    | SQL_TSI_YEAR
    | STANDBY
    | START
    | STARTS
    | STAT
    | STATS_AUTO_RECALC
    | STATS_PERSISTENT
    | STATS_SAMPLE_PAGES
    | STATUS
    | STDDEV
    | STDDEV_POP
    | STDDEV_SAMP
    | STOP
    | STORAGE
    | STORAGE_FORMAT_VERSION
    | STORAGE_FORMAT_WORK_VERSION
    | STORING
    | STRING
    | SUBCLASS_ORIGIN
    | SUBDATE
    | SUBJECT
    | SUBPARTITION
    | SUBPARTITIONS
    | SUBSTR
    | SUBSTRING
    | SUM
    | SUPER
    | SUSPEND
    | SWAPS
    | SWITCH
    | SWITCHES
    | SWITCHOVER
    | SYSTEM
    | SYSTEM_USER
    | SYSDATE
    | TABLE_CHECKSUM
    | TABLE_MODE
    | TABLEGROUPS
    | TABLE_ID
    | TABLE_NAME
    | TABLES
    | TABLESPACE
    | TABLET
    | TABLET_SIZE
    | TABLET_MAX_SIZE
    | TASK
    | TEMPLATE
    | TEMPORARY
    | TEMPTABLE
    | TENANT
    | TENANT_ID
    | SLOT_IDX
    | TEXT
    | THAN
    | TIME
    | TIMESTAMP
    | TIMESTAMPADD
    | TIMESTAMPDIFF
    | TIME_ZONE_INFO
    | TP_NAME
    | TP_NO
    | TRACE
    | TRADITIONAL
    | TRIGGERS
    | TRIM
    | TRUNCATE
    | TYPE
    | TYPES
    | TABLEGROUP_ID
    | UNCOMMITTED
    | UNDEFINED
    | UNDO_BUFFER_SIZE
    | UNDOFILE
    | UNICODE
    | UNKNOWN
    | UNINSTALL
    | UNIT
    | UNIT_NUM
    | UNLOCKED
    | UNTIL
    | UNUSUAL
    | UPGRADE
    | USE_BLOOM_FILTER
    | USE_FRM
    | USER
    | USER_RESOURCES
    | UNBOUNDED
    | VALID
    | VALUE
    | VARIANCE
    | VARIABLES
    | VERBOSE
    | VIRTUAL_COLUMN_ID
    | MATERIALIZED
    | VIEW
    | WAIT
    | WARNINGS
    | WEEK
    | WEIGHT_STRING
    | WORK
    | WRAPPER
    | X509
    | XA
    | XML
    | YEAR
    | ZONE
    | ZONE_LIST
    | ZONE_TYPE
    | LOCATION
    | PLAN
    | VISIBLE
    | INVISIBLE
    | ACTIVATE
    | SYNCHRONIZATION
    ;

unreserved_keyword_special
    : PASSWORD
    ;

empty
    :
    ;

forward_expr
    : expr EOF
    ;

forward_sql_stmt
    : stmt EOF
    ;
