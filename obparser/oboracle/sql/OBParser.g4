parser grammar OBParser;


options { tokenVocab=OBLexer; }


@parser::members {
public boolean is_pl_parse_ = false;
public boolean is_pl_parse_expr_ = false;
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
    | merge_stmt
    | create_table_stmt
    | alter_database_stmt
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
    | audit_stmt
    | deallocate_prepare_stmt
    | create_user_stmt
    | alter_user_profile_stmt
    | drop_user_stmt
    | set_password_stmt
    | lock_user_stmt
    | grant_stmt
    | revoke_stmt
    | begin_stmt
    | commit_stmt
    | rollback_stmt
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
    | create_tablespace_stmt
    | drop_tablespace_stmt
    | create_keystore_stmt
    | alter_keystore_stmt
    | lock_tables_stmt
    | unlock_tables_stmt
    | flashback_stmt
    | purge_stmt
    | create_sequence_stmt
    | alter_sequence_stmt
    | drop_sequence_stmt
    | alter_session_stmt
    | analyze_stmt
    | set_comment_stmt
    | pl_expr_stmt
    | shrink_space_stmt
    | load_data_stmt
    | create_role_stmt
    | drop_role_stmt
    | create_profile_stmt
    | alter_profile_stmt
    | drop_profile_stmt
    ;

pl_expr_stmt
    : {this.is_pl_parse_ && this.is_pl_parse_expr_}? DO expr
    ;

expr_list
    : bit_expr (Comma bit_expr)*
    ;

column_ref
    : column_name
    ;

complex_string_literal
    : STRING_VALUE
    ;

literal
    : complex_string_literal
    | DATE_VALUE
    | TIMESTAMP_VALUE
    | INTNUM
    | APPROXNUM
    | DECIMAL_VAL
    | NULLX
    | INTERVAL_VALUE
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
    : bit_expr IS NULLX
    | bit_expr IS not NULLX
    | bit_expr COMP_LE bit_expr
    | bit_expr COMP_LE sub_query_flag bit_expr
    | bit_expr COMP_LT bit_expr
    | bit_expr COMP_LT sub_query_flag bit_expr
    | bit_expr COMP_EQ bit_expr
    | bit_expr COMP_EQ sub_query_flag bit_expr
    | bit_expr COMP_GE bit_expr
    | bit_expr COMP_GE sub_query_flag bit_expr
    | bit_expr COMP_GT bit_expr
    | bit_expr COMP_GT sub_query_flag bit_expr
    | bit_expr COMP_NE bit_expr
    | bit_expr COMP_NE sub_query_flag bit_expr
    | predicate
    ;

predicate
    : LNNVL LeftParen bool_pri RightParen
    | bit_expr IN in_expr
    | bit_expr not IN in_expr
    | bit_expr not BETWEEN bit_expr AND bit_expr
    | bit_expr BETWEEN bit_expr AND bit_expr
    | bit_expr LIKE bit_expr
    | bit_expr LIKE bit_expr ESCAPE bit_expr
    | bit_expr not LIKE bit_expr
    | bit_expr not LIKE bit_expr ESCAPE bit_expr
    | REGEXP_LIKE LeftParen substr_params RightParen
    | EXISTS select_with_parens
    ;

bit_expr
    : bit_expr Plus bit_expr
    | bit_expr Minus bit_expr
    | bit_expr Star bit_expr
    | bit_expr Div bit_expr
    | bit_expr CNNOP bit_expr
    | unary_expr
    ;

unary_expr
    : Plus simple_expr
    | Minus simple_expr
    | simple_expr
    ;

simple_expr
    : simple_expr collation
    | ROWNUM
    | obj_access_ref COLUMN_OUTER_JOIN_SYMBOL
    | expr_const
    | select_with_parens
    | LeftParen bit_expr RightParen
    | LeftParen expr_list Comma bit_expr RightParen
    | MATCH LeftParen column_list RightParen AGAINST LeftParen STRING_VALUE ((IN NATURAL LANGUAGE MODE) | (IN BOOLEAN MODE))? RightParen
    | case_expr
    | obj_access_ref
    | sql_function
    | cursor_attribute_expr
    | window_function
    | USER_VARIABLE
    | PRIOR unary_expr
    | CONNECT_BY_ROOT unary_expr
    | LEVEL
    | CONNECT_BY_ISLEAF
    | CONNECT_BY_ISCYCLE
    | {this.is_pl_parse_}? QUESTIONMARK Dot column_name
    ;

common_cursor_attribute
    : ISOPEN
    | FOUND
    | NOTFOUND
    | ROWCOUNT
    ;

cursor_attribute_bulk_rowcount
    : BULK_ROWCOUNT LeftParen bit_expr RightParen
    ;

cursor_attribute_bulk_exceptions
    : BULK_EXCEPTIONS Dot COUNT
    | BULK_EXCEPTIONS LeftParen bit_expr RightParen Dot ERROR_INDEX
    | BULK_EXCEPTIONS LeftParen bit_expr RightParen Dot ERROR_CODE
    ;

implicit_cursor_attribute
    : SQL Mod common_cursor_attribute
    | SQL Mod cursor_attribute_bulk_rowcount
    | SQL Mod cursor_attribute_bulk_exceptions
    ;

explicit_cursor_attribute
    : column_ref Mod common_cursor_attribute
    ;

cursor_attribute_expr
    : {this.is_pl_parse_}? explicit_cursor_attribute
    | {this.is_pl_parse_}? implicit_cursor_attribute
    ;

obj_access_ref
    : column_ref ((Dot obj_access_ref) | (Dot Star))?
    | access_func_expr ((Dot obj_access_ref) | table_element_access_list)?
    ;

obj_access_ref_normal
    : var_name (Dot obj_access_ref_normal)?
    | access_func_expr ((Dot obj_access_ref_normal) | table_element_access_list)?
    ;

table_element_access_list
    : LeftParen table_index RightParen
    | table_element_access_list LeftParen table_index RightParen
    ;

table_index
    : INTNUM
    | var_name
    ;

expr
    : expr AND expr
    | expr OR expr
    | NOT expr
    | {this.is_pl_parse_}? bit_expr
    | bool_pri
    | USER_VARIABLE SET_VAR bit_expr
    | {this.is_pl_parse_}? BOOL_VALUE
    | USER_VARIABLE SET_VAR BOOL_VALUE
    | LeftParen expr RightParen
    ;

not
    : NOT
    ;

sub_query_flag
    : ALL
    | ANY
    | SOME
    ;

in_expr
    : bit_expr
    ;

case_expr
    : CASE bit_expr simple_when_clause_list case_default END
    | CASE bool_when_clause_list case_default END
    ;

window_function
    : COUNT LeftParen ALL? Star RightParen OVER LeftParen generalized_window_clause RightParen
    | COUNT LeftParen ALL? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | COUNT LeftParen DISTINCT bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | APPROX_COUNT_DISTINCT LeftParen expr_list RightParen OVER LeftParen generalized_window_clause RightParen
    | APPROX_COUNT_DISTINCT_SYNOPSIS LeftParen expr_list RightParen OVER LeftParen generalized_window_clause RightParen
    | APPROX_COUNT_DISTINCT_SYNOPSIS_MERGE LeftParen bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | SUM LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | MAX LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | MIN LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | AVG LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | STDDEV LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | VARIANCE LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | STDDEV_POP LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | STDDEV_SAMP LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | LISTAGG LeftParen ALL? expr_list (SEPARATOR STRING_VALUE)? RightParen WITHIN GROUP LeftParen order_by RightParen OVER LeftParen generalized_window_clause RightParen
    | RANK LeftParen RightParen OVER LeftParen generalized_window_clause RightParen
    | DENSE_RANK LeftParen RightParen OVER LeftParen generalized_window_clause RightParen
    | PERCENT_RANK LeftParen RightParen OVER LeftParen generalized_window_clause RightParen
    | ROW_NUMBER LeftParen RightParen OVER LeftParen generalized_window_clause RightParen
    | NTILE LeftParen bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
    | CUME_DIST LeftParen RightParen OVER LeftParen generalized_window_clause RightParen
    | FIRST_VALUE win_fun_first_last_params OVER LeftParen generalized_window_clause RightParen
    | LAST_VALUE win_fun_first_last_params OVER LeftParen generalized_window_clause RightParen
    | LEAD win_fun_lead_lag_params OVER LeftParen generalized_window_clause RightParen
    | LAG win_fun_lead_lag_params OVER LeftParen generalized_window_clause RightParen
    | NTH_VALUE LeftParen bit_expr Comma bit_expr RightParen (FROM first_or_last)? (respect_or_ignore NULLS)? OVER LeftParen generalized_window_clause RightParen
    | RATIO_TO_REPORT LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen OVER LeftParen generalized_window_clause RightParen
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
    : LeftParen bit_expr respect_or_ignore NULLS RightParen
    | LeftParen bit_expr RightParen (respect_or_ignore NULLS)?
    ;

win_fun_lead_lag_params
    : LeftParen bit_expr respect_or_ignore NULLS RightParen
    | LeftParen bit_expr respect_or_ignore NULLS Comma expr_list RightParen
    | LeftParen expr_list RightParen (respect_or_ignore NULLS)?
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
    : bit_expr
    ;

win_bounding
    : CURRENT ROW
    | win_interval win_preceding_or_following
    ;

win_window
    : win_rows_or_range BETWEEN win_bounding AND win_bounding
    | win_rows_or_range win_bounding
    ;

simple_when_clause_list
    : simple_when_clause+
    ;

simple_when_clause
    : WHEN bit_expr THEN bit_expr
    ;

bool_when_clause_list
    : bool_when_clause+
    ;

bool_when_clause
    : WHEN expr THEN bit_expr
    ;

case_default
    : ELSE bit_expr
    | empty
    ;

sql_function
    : single_row_function
    | aggregate_function
    | special_func_expr
    ;

single_row_function
    : numeric_function
    | character_function
    | extract_function
    | conversion_function
    | hierarchical_function
    | environment_id_function
    ;

numeric_function
    : MOD LeftParen bit_expr Comma bit_expr RightParen
    ;

character_function
    : SUBSTR LeftParen substr_params RightParen
    | TRIM LeftParen parameterized_trim RightParen
    ;

extract_function
    : EXTRACT LeftParen date_unit_for_extract FROM bit_expr RightParen
    | SESSIONTIMEZONE
    | DBTIMEZONE
    ;

conversion_function
    : CAST LeftParen bit_expr AS cast_data_type RightParen
    | CONVERT LeftParen bit_expr Comma cast_data_type RightParen
    | CONVERT LeftParen bit_expr USING charset_name RightParen
    ;

hierarchical_function
    : SYS_CONNECT_BY_PATH LeftParen bit_expr Comma signed_literal RightParen
    ;

environment_id_function
    : USER
    | UID
    ;

aggregate_function
    : APPROX_COUNT_DISTINCT LeftParen expr_list RightParen
    | APPROX_COUNT_DISTINCT_SYNOPSIS LeftParen expr_list RightParen
    | APPROX_COUNT_DISTINCT_SYNOPSIS_MERGE LeftParen bit_expr RightParen
    | SUM LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen
    | MAX LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen
    | MIN LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen
    | AVG LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen
    | STDDEV LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen
    | VARIANCE LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen
    | STDDEV_POP LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen
    | STDDEV_SAMP LeftParen (ALL | DISTINCT | UNIQUE)? bit_expr RightParen
    | GROUPING LeftParen bit_expr RightParen
    | LISTAGG LeftParen ALL? expr_list (SEPARATOR STRING_VALUE)? RightParen WITHIN GROUP LeftParen order_by RightParen
    ;

special_func_expr
    : ISNULL LeftParen bit_expr RightParen
    | cur_timestamp_func
    | INSERT LeftParen bit_expr Comma bit_expr Comma bit_expr Comma bit_expr RightParen
    | LEFT LeftParen bit_expr Comma bit_expr RightParen
    | POSITION LeftParen bit_expr IN bit_expr RightParen
    | DATE LeftParen bit_expr RightParen
    | YEAR LeftParen bit_expr RightParen
    | TIME LeftParen bit_expr RightParen
    | MONTH LeftParen bit_expr RightParen
    | DEFAULT LeftParen column_definition_ref RightParen
    | VALUES LeftParen column_definition_ref RightParen
    | CHARACTER LeftParen expr_list RightParen
    | CHARACTER LeftParen expr_list USING charset_name RightParen
    ;

access_func_expr
    : COUNT LeftParen ALL? Star RightParen
    | COUNT LeftParen ALL? bit_expr RightParen
    | COUNT LeftParen DISTINCT bit_expr RightParen
    | COUNT LeftParen UNIQUE bit_expr RightParen
    | function_name LeftParen func_param_list? RightParen
    ;

func_param_list
    : func_param (Comma func_param)*
    ;

func_param
    : func_param_with_assign
    | bit_expr
    ;

func_param_with_assign
    : var_name PARAM_ASSIGN_OPERATOR bit_expr
    ;

cur_timestamp_func
    : SYSDATE
    | SYSTIMESTAMP
    | SYSTIMESTAMP LeftParen INTNUM RightParen
    | CURRENT_DATE
    | LOCALTIMESTAMP
    | LOCALTIMESTAMP LeftParen INTNUM RightParen
    | CURRENT_TIMESTAMP
    | CURRENT_TIMESTAMP LeftParen INTNUM RightParen
    ;

substr_params
    : bit_expr Comma bit_expr
    | bit_expr Comma bit_expr Comma bit_expr
    ;

delete_stmt
    : delete_with_opt_hint FROM table_factor (WHERE opt_hint_value expr)? ((RETURNING returning_exprs opt_into_clause) | (RETURN returning_exprs opt_into_clause))?
    | delete_with_opt_hint table_factor (WHERE opt_hint_value expr)? ((RETURNING returning_exprs opt_into_clause) | (RETURN returning_exprs opt_into_clause))?
    ;

update_stmt
    : update_with_opt_hint dml_table_clause SET update_asgn_list (WHERE opt_hint_value expr)? ((RETURNING returning_exprs opt_into_clause) | (RETURN returning_exprs opt_into_clause))?
    ;

update_asgn_list
    : normal_asgn_list
    ;

normal_asgn_list
    : update_asgn_factor (Comma update_asgn_factor)*
    ;

update_asgn_factor
    : column_definition_ref COMP_EQ expr_or_default
    | LeftParen column_list RightParen COMP_EQ LeftParen subquery RightParen
    ;

create_resource_stmt
    : CREATE RESOURCE UNIT relation_name (resource_unit_option | (opt_resource_unit_option_list Comma resource_unit_option))?
    | CREATE RESOURCE POOL relation_name (create_resource_pool_option | (opt_create_resource_pool_option_list Comma create_resource_pool_option))?
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
    : DROP RESOURCE UNIT relation_name
    | DROP RESOURCE POOL relation_name
    ;

create_tenant_stmt
    : CREATE TENANT relation_name (tenant_option | (opt_tenant_option_list Comma tenant_option))? ((SET sys_var_and_val_list) | (SET VARIABLES sys_var_and_val_list) | (VARIABLES sys_var_and_val_list))?
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
    | read_only_or_write
    | COMMENT COMP_EQ? STRING_VALUE
    | default_tablegroup
    ;

zone_list
    : STRING_VALUE (opt_comma STRING_VALUE)*
    ;

resource_pool_list
    : STRING_VALUE (Comma STRING_VALUE)*
    ;

alter_tenant_stmt
    : ALTER TENANT relation_name SET? (tenant_option | (opt_tenant_option_list Comma tenant_option))? (VARIABLES sys_var_and_val_list)?
    | ALTER TENANT relation_name lock_spec_mysql57
    ;

drop_tenant_stmt
    : DROP TENANT relation_name
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

alter_database_stmt
    : ALTER database_key database_name? SET? database_option_list
    ;

database_name
    : NAME_OB
    ;

load_data_stmt
    : load_data_with_opt_hint LOCAL? INFILE STRING_VALUE (IGNORE | REPLACE)? INTO TABLE relation_factor use_partition? (CHARACTER SET charset_name_or_default)? field_opt line_opt (IGNORE INTNUM lines_or_rows)? ((LeftParen RightParen) | (LeftParen field_or_vars_list RightParen))? (SET load_set_list)?
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
    : GLOBAL TEMPORARY
    | empty
    ;

on_commit_option
    : ON COMMIT DELETE ROWS
    | ON COMMIT PRESERVE ROWS
    | empty
    ;

create_keystore_stmt
    : ADMINISTER KEY MANAGEMENT CREATE KEYSTORE keystore_name IDENTIFIED BY password
    ;

alter_keystore_stmt
    : ADMINISTER KEY MANAGEMENT ALTER KEYSTORE PASSWORD IDENTIFIED BY password SET password
    | ADMINISTER KEY MANAGEMENT SET KEY IDENTIFIED BY password
    | ADMINISTER KEY MANAGEMENT SET KEYSTORE CLOSE IDENTIFIED BY password
    | ADMINISTER KEY MANAGEMENT SET KEYSTORE OPEN IDENTIFIED BY password
    ;

create_table_stmt
    : CREATE temporary_option TABLE relation_factor LeftParen table_element_list RightParen table_option_list? opt_partition_option on_commit_option
    | CREATE temporary_option TABLE relation_factor LeftParen table_element_list RightParen table_option_list? opt_partition_option AS subquery order_by?
    | CREATE temporary_option TABLE relation_factor table_option_list opt_partition_option AS subquery order_by?
    | CREATE temporary_option TABLE relation_factor partition_option AS subquery order_by?
    | CREATE temporary_option TABLE relation_factor AS subquery order_by?
    ;

table_element_list
    : table_element (Comma table_element)*
    ;

table_element
    : column_definition
    | out_of_line_constraint
    | INDEX index_name? index_using_algorithm? LeftParen sort_column_list RightParen opt_index_options?
    ;

column_definition
    : column_definition_ref data_type visibility_option? (opt_column_attribute_list column_attribute)?
    | column_definition_ref data_type visibility_option? (GENERATED ALWAYS)? AS LeftParen bit_expr RightParen VIRTUAL? (opt_generated_column_attribute_list generated_column_attribute)?
    | column_definition_ref visibility_option? (GENERATED ALWAYS)? AS LeftParen bit_expr RightParen VIRTUAL? (opt_generated_column_attribute_list generated_column_attribute)?
    ;

column_definition_opt_datatype
    : column_definition_ref data_type? visibility_option? (opt_column_attribute_list column_attribute)?
    | column_definition_ref data_type? visibility_option? (GENERATED ALWAYS)? AS LeftParen bit_expr RightParen VIRTUAL? (opt_generated_column_attribute_list generated_column_attribute)?
    ;

out_of_line_constraint
    : constraint_and_name? UNIQUE LeftParen sort_column_list RightParen (USING INDEX opt_index_option_list)?
    | PRIMARY KEY LeftParen column_name_list RightParen (USING INDEX opt_index_option_list)?
    | constraint_and_name PRIMARY KEY LeftParen column_name_list RightParen (USING INDEX opt_index_option_list)?
    | constraint_and_name FOREIGN KEY LeftParen column_name_list RightParen references_clause (USING INDEX opt_index_option_list)? (ENABLE | DISABLE)?
    | FOREIGN KEY LeftParen column_name_list RightParen references_clause (USING INDEX opt_index_option_list)? (ENABLE | DISABLE)?
    | constraint_and_name? CHECK LeftParen expr RightParen constranit_state
    ;

constranit_state
    : (RELY | NORELY)? (USING INDEX opt_index_option_list)? (ENABLE | DISABLE)? (VALIDATE | NOVALIDATE)?
    ;

references_clause
    : REFERENCES normal_relation_factor LeftParen column_name_list RightParen reference_option?
    ;

reference_option
    : ON DELETE reference_action
    ;

reference_action
    : CASCADE
    | SET NULLX
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

column_definition_opt_datatype_list
    : column_definition_opt_datatype (Comma column_definition_opt_datatype)*
    ;

column_name_list
    : column_name (Comma column_name)*
    ;

cast_data_type
    : RAW string_length_i
    | CHARACTER string_length_i? BINARY?
    | {!this.is_pl_parse_ || !this.is_pl_parse_expr_}? varchar_type_i string_length_i BINARY?
    | {this.is_pl_parse_ && this.is_pl_parse_expr_}? varchar_type_i
    | {!this.is_pl_parse_ || !this.is_pl_parse_expr_}? NVARCHAR2 string_length_i
    | {!this.is_pl_parse_ || !this.is_pl_parse_expr_}? NCHAR string_length_i
    | cast_datetime_type_i
    | TIMESTAMP (LeftParen precision_int_num RightParen)?
    | TIMESTAMP (LeftParen precision_int_num RightParen)? WITH TIME ZONE
    | TIMESTAMP (LeftParen precision_int_num RightParen)? WITH LOCAL TIME ZONE
    | int_type_i
    | number_type_i number_precision
    | NUMBER number_precision?
    | FLOAT ((LeftParen INTNUM RightParen) | (LeftParen RightParen))?
    | double_type_i
    | INTERVAL YEAR (LeftParen precision_int_num RightParen)? TO MONTH
    | INTERVAL DAY (LeftParen precision_int_num RightParen)? TO SECOND (LeftParen precision_int_num RightParen)?
    | udt_type
    ;

udt_type
    : type_name
    | database_name Dot type_name
    ;

type_name
    : NAME_OB
    ;

cast_datetime_type_i
    : DATE
    ;

data_type
    : int_type_i
    | FLOAT ((LeftParen INTNUM RightParen) | (LeftParen RightParen))?
    | double_type_i
    | number_type_i number_precision
    | NUMBER number_precision?
    | TIMESTAMP (LeftParen precision_int_num RightParen)?
    | TIMESTAMP (LeftParen precision_int_num RightParen)? WITH TIME ZONE
    | TIMESTAMP (LeftParen precision_int_num RightParen)? WITH LOCAL TIME ZONE
    | datetime_type_i
    | CHARACTER string_length_i? BINARY? (charset_key charset_name)? collation?
    | varchar_type_i string_length_i BINARY? (charset_key charset_name)? collation?
    | RAW string_length_i
    | STRING_VALUE
    | BLOB
    | CLOB BINARY? (charset_key charset_name)? collation?
    | INTERVAL YEAR (LeftParen precision_int_num RightParen)? TO MONTH
    | INTERVAL DAY (LeftParen precision_int_num RightParen)? TO SECOND (LeftParen precision_int_num RightParen)?
    | NVARCHAR2 string_length_i
    | NCHAR string_length_i
    ;

int_type_i
    : SMALLINT
    | INT
    | INTEGER
    | NUMERIC
    | DECIMAL
    ;

varchar_type_i
    : VARCHAR
    | VARCHAR2
    ;

number_type_i
    : DECIMAL
    | NUMERIC
    ;

double_type_i
    : BINARY_DOUBLE
    | BINARY_FLOAT
    ;

datetime_type_i
    : DATE
    ;

number_precision
    : LeftParen signed_int_num Comma signed_int_num RightParen
    | LeftParen Star Comma signed_int_num RightParen
    | LeftParen Star RightParen
    | LeftParen signed_int_num RightParen
    ;

signed_int_num
    : INTNUM
    | Minus INTNUM
    ;

precision_int_num
    : INTNUM
    ;

string_length_i
    : LeftParen INTNUM (CHARACTER | BYTE)? RightParen
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
    | DEFAULT bit_expr
    | ORIG_DEFAULT now_or_signed_literal
    | PRIMARY KEY
    | UNIQUE
    | ON UPDATE cur_timestamp_func
    | ID INTNUM
    | constraint_and_name? CHECK LeftParen expr RightParen constranit_state
    ;

now_or_signed_literal
    : cur_timestamp_func_params
    | signed_literal_params
    ;

cur_timestamp_func_params
    : LeftParen cur_timestamp_func_params RightParen
    | cur_timestamp_func
    ;

signed_literal_params
    : LeftParen signed_literal_params RightParen
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
    | EXPIRE_INFO COMP_EQ? LeftParen bit_expr RightParen
    | PROGRESSIVE_MERGE_NUM COMP_EQ? INTNUM
    | BLOCK_SIZE COMP_EQ? INTNUM
    | TABLE_ID COMP_EQ? INTNUM
    | REPLICA_NUM COMP_EQ? INTNUM
    | compress_option
    | USE_BLOOM_FILTER COMP_EQ? BOOL_VALUE
    | PRIMARY_ZONE COMP_EQ? primary_zone_name
    | TABLEGROUP COMP_EQ? relation_name_or_string
    | read_only_or_write
    | ENGINE_ COMP_EQ? relation_name_or_string
    | TABLET_SIZE COMP_EQ? INTNUM
    | MAX_USED_PART_ID COMP_EQ? INTNUM
    | ENABLE ROW MOVEMENT
    | DISABLE ROW MOVEMENT
    | physical_attributes_option
    ;

storage_options_list
    : storage_option+
    ;

storage_option
    : INITIAL_ size_option
    | NEXT size_option
    | MINEXTENTS INTNUM
    | MAXEXTENTS int_or_unlimited
    ;

size_option
    : INTNUM unit_of_size?
    ;

int_or_unlimited
    : INTNUM
    | UNLIMITED
    ;

unit_of_size
    : K_SIZE
    | M_SIZE
    | G_SIZE
    | T_SIZE
    | P_SIZE
    | E_SIZE
    ;

relation_name_or_string
    : relation_name
    | STRING_VALUE
    ;

opt_equal_mark
    : COMP_EQ?
    ;

partition_option_inner
    : hash_partition_option
    | range_partition_option
    | list_partition_option
    ;

opt_partition_option
    : partition_option
    | opt_column_partition_option
    ;

partition_option
    : partition_option_inner column_partition_option?
    ;

hash_partition_option
    : PARTITION BY HASH LeftParen column_name_list RightParen subpartition_option (PARTITIONS INTNUM)? (TABLESPACE tablespace)? compress_option?
    ;

list_partition_option
    : PARTITION BY LIST LeftParen column_name_list RightParen subpartition_option (PARTITIONS INTNUM)? opt_list_partition_list
    ;

range_partition_option
    : PARTITION BY RANGE LeftParen column_name_list RightParen subpartition_option (PARTITIONS INTNUM)? opt_range_partition_list
    ;

subpartition_option
    : SUBPARTITION BY RANGE LeftParen column_name_list RightParen SUBPARTITION TEMPLATE opt_range_subpartition_list
    | SUBPARTITION BY HASH LeftParen column_name_list RightParen (SUBPARTITIONS INTNUM)?
    | SUBPARTITION BY LIST LeftParen column_name_list RightParen SUBPARTITION TEMPLATE opt_list_subpartition_list
    | empty
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
    : SUBPARTITION relation_factor VALUES list_partition_expr physical_attributes_option_list?
    ;

list_partition_element
    : PARTITION relation_factor VALUES list_partition_expr (ID INTNUM)? physical_attributes_option_list? compress_option?
    | PARTITION VALUES list_partition_expr (ID INTNUM)? physical_attributes_option_list? compress_option?
    ;

list_partition_expr
    : LeftParen list_expr RightParen
    | LeftParen DEFAULT RightParen
    ;

list_expr
    : bit_expr (Comma bit_expr)*
    ;

range_partition_list
    : range_partition_element (Comma range_partition_element)*
    ;

range_partition_element
    : PARTITION relation_factor VALUES LESS THAN range_partition_expr (ID INTNUM)? physical_attributes_option_list? compress_option?
    | PARTITION VALUES LESS THAN range_partition_expr (ID INTNUM)? physical_attributes_option_list? compress_option?
    ;

physical_attributes_option_list
    : physical_attributes_option+
    ;

physical_attributes_option
    : PCTFREE COMP_EQ? INTNUM
    | PCTUSED INTNUM
    | INITRANS INTNUM
    | MAXTRANS INTNUM
    | STORAGE LeftParen storage_options_list RightParen
    | TABLESPACE tablespace
    ;

opt_special_partition_list
    : LeftParen special_partition_list RightParen
    ;

special_partition_list
    : special_partition_define (Comma special_partition_define)*
    ;

special_partition_define
    : PARTITION (ID INTNUM)?
    | PARTITION relation_factor (ID INTNUM)?
    ;

range_subpartition_element
    : SUBPARTITION relation_factor VALUES LESS THAN range_partition_expr physical_attributes_option_list?
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
    : bit_expr
    | MAXVALUE
    ;

tg_hash_partition_option
    : PARTITION BY HASH tg_subpartition_option (PARTITIONS INTNUM)?
    | PARTITION BY HASH INTNUM tg_subpartition_option (PARTITIONS INTNUM)?
    ;

tg_range_partition_option
    : PARTITION BY RANGE COLUMNS INTNUM tg_subpartition_option (PARTITIONS INTNUM)? opt_range_partition_list
    ;

tg_list_partition_option
    : PARTITION BY LIST COLUMNS INTNUM tg_subpartition_option (PARTITIONS INTNUM)? opt_list_partition_list
    ;

tg_subpartition_option
    : SUBPARTITION BY RANGE COLUMNS INTNUM SUBPARTITION TEMPLATE opt_range_subpartition_list
    | SUBPARTITION BY HASH (SUBPARTITIONS INTNUM)?
    | SUBPARTITION BY LIST COLUMNS INTNUM SUBPARTITION TEMPLATE opt_list_subpartition_list
    | empty
    ;

opt_alter_compress_option
    : MOVE compress_option
    ;

compress_option
    : NOCOMPRESS
    | COMPRESS (BASIC | (FOR OLTP) | (FOR QUERY opt_compress_level) | (FOR ARCHIVE opt_compress_level))?
    ;

opt_compress_level
    : (LOW | HIGH)?
    ;

create_tablegroup_stmt
    : CREATE TABLEGROUP relation_name tablegroup_option_list? (tg_hash_partition_option | tg_range_partition_option | tg_list_partition_option)?
    ;

drop_tablegroup_stmt
    : DROP TABLEGROUP relation_name
    ;

alter_tablegroup_stmt
    : ALTER TABLEGROUP relation_name ADD TABLE? table_list
    | ALTER TABLEGROUP relation_name alter_tablegroup_actions
    | ALTER TABLEGROUP relation_name alter_partition_option
    | ALTER TABLEGROUP relation_name tg_modify_partition_info
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
    : DEFAULT_TABLEGROUP COMP_EQ? relation_name
    | DEFAULT_TABLEGROUP COMP_EQ? NULLX
    ;

create_view_stmt
    : CREATE (OR REPLACE)? VIEW view_name (LeftParen column_list RightParen)? (TABLE_ID COMP_EQ INTNUM)? AS view_subquery view_with_opt
    ;

view_subquery
    : subquery
    | subquery order_by
    ;

view_with_opt
    : WITH READ ONLY
    | empty
    ;

view_name
    : relation_factor
    ;

create_index_stmt
    : CREATE UNIQUE? INDEX normal_relation_factor index_using_algorithm? ON relation_factor LeftParen sort_column_list RightParen opt_index_options? opt_partition_option
    ;

index_name
    : relation_name
    ;

constraint_and_name
    : CONSTRAINT constraint_name
    ;

constraint_name
    : relation_name
    ;

sort_column_list
    : sort_column_key (Comma sort_column_key)*
    ;

sort_column_key
    : index_expr opt_asc_desc (ID INTNUM)?
    ;

index_expr
    : bit_expr
    ;

opt_index_option_list
    : opt_index_options?
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
    | WITH_ROWID
    | WITH PARSER STRING_VALUE
    | index_using_algorithm
    | visibility_option
    | DATA_TABLE_ID COMP_EQ? INTNUM
    | INDEX_TABLE_ID COMP_EQ? INTNUM
    | MAX_USED_PART_ID COMP_EQ? INTNUM
    | physical_attributes_option
    ;

index_using_algorithm
    : USING BTREE
    | USING HASH
    ;

drop_table_stmt
    : DROP TABLE relation_factor (CASCADE CONSTRAINTS)? PURGE?
    ;

table_or_tables
    : TABLE
    | TABLES
    ;

drop_view_stmt
    : DROP MATERIALIZED? VIEW relation_factor (CASCADE CONSTRAINTS)?
    ;

table_list
    : relation_factor (Comma relation_factor)*
    ;

drop_index_stmt
    : DROP INDEX relation_name
    | DROP INDEX relation_name Dot relation_name
    ;

insert_stmt
    : insert_with_opt_hint single_table_insert
    ;

single_table_insert
    : INTO insert_table_clause NOLOGGING? LeftParen column_list RightParen values_clause ((RETURNING returning_exprs opt_into_clause) | (RETURN returning_exprs opt_into_clause))?
    | INTO insert_table_clause NOLOGGING? LeftParen RightParen values_clause ((RETURNING returning_exprs opt_into_clause) | (RETURN returning_exprs opt_into_clause))?
    | INTO insert_table_clause NOLOGGING? values_clause ((RETURNING returning_exprs opt_into_clause) | (RETURN returning_exprs opt_into_clause))?
    ;

values_clause
    : VALUES insert_vals_list
    | subquery order_by?
    ;

opt_into_clause
    : into_clause?
    ;

returning_exprs
    : projection (Comma projection)*
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
    : expr_or_default (Comma expr_or_default)*
    ;

expr_or_default
    : bit_expr
    | DEFAULT
    ;

merge_with_opt_hint
    : MERGE
    | MERGE_HINT_BEGIN hint_list_with_end
    ;

merge_stmt
    : merge_with_opt_hint INTO source_relation_factor relation_name? USING source_relation_factor relation_name? ON LeftParen expr RightParen merge_update_clause merge_insert_clause
    | merge_with_opt_hint INTO source_relation_factor relation_name? USING source_relation_factor relation_name? ON LeftParen expr RightParen merge_insert_clause
    | merge_with_opt_hint INTO source_relation_factor relation_name? USING source_relation_factor relation_name? ON LeftParen expr RightParen merge_update_clause
    ;

merge_update_clause
    : WHEN MATCHED THEN UPDATE SET update_asgn_list (WHERE opt_hint_value expr)? (DELETE WHERE expr)?
    ;

merge_insert_clause
    : WHEN NOT MATCHED THEN INSERT (LeftParen column_list RightParen)? VALUES LeftParen insert_vals RightParen (WHERE opt_hint_value expr)?
    ;

source_relation_factor
    : relation_factor
    | select_with_parens
    ;

select_stmt
    : subquery
    | subquery for_update
    | subquery order_by
    | subquery order_by for_update
    | subquery for_update order_by
    ;

subquery
    : select_no_parens
    | select_with_parens
    | with_select
    ;

select_with_parens
    : LeftParen select_no_parens RightParen
    | LeftParen select_with_parens RightParen
    | LeftParen with_select RightParen
    ;

select_no_parens
    : select_clause
    | select_clause_set
    ;

no_table_select
    : select_with_opt_hint query_expression_option_list? select_expr_list into_opt FROM DUAL (WHERE opt_hint_value expr)?
    ;

no_table_select_with_hierarchical_query
    : select_with_opt_hint query_expression_option_list? select_expr_list into_opt FROM DUAL (WHERE opt_hint_value expr)? start_with connect_by
    | select_with_opt_hint query_expression_option_list? select_expr_list into_opt FROM DUAL (WHERE opt_hint_value expr)? connect_by start_with?
    ;

select_clause
    : no_table_select
    | no_table_select_with_hierarchical_query
    | simple_select
    | select_with_hierarchical_query
    ;

select_clause_set
    : select_clause_set set_type select_clause_set_right
    | select_clause_set_left set_type select_clause_set_right
    ;

select_clause_set_right
    : no_table_select
    | simple_select
    | select_with_parens
    ;

select_clause_set_left
    : select_clause_set_right
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
    : select_with_opt_hint query_expression_option_list? select_expr_list into_opt FROM from_list (WHERE opt_hint_value expr)? (GROUP BY groupby_clause)? (HAVING expr)?
    ;

select_with_hierarchical_query
    : select_with_opt_hint query_expression_option_list? select_expr_list into_opt FROM from_list (WHERE opt_hint_value expr)? start_with connect_by (GROUP BY groupby_clause)? (HAVING expr)?
    | select_with_opt_hint query_expression_option_list? select_expr_list into_opt FROM from_list (WHERE opt_hint_value expr)? connect_by start_with? (GROUP BY groupby_clause)? (HAVING expr)?
    ;

start_with
    : START WITH expr
    ;

connect_by
    : CONNECT BY NOCYCLE? expr
    ;

set_type_union
    : UNION
    ;

set_type_other
    : INTERSECT
    | MINUS
    ;

set_type
    : set_type_union set_expression_option
    | set_type_other
    ;

set_expression_option
    : ALL?
    ;

opt_hint_value
    : HINT_VALUE?
    ;

into_clause
    : INTO into_var_list
    | BULK COLLECT INTO into_var_list
    ;

into_opt
    : INTO OUTFILE STRING_VALUE (charset_key charset_name)? field_opt line_opt
    | INTO DUMPFILE STRING_VALUE
    | into_clause
    | empty
    ;

into_var_list
    : into_var (Comma into_var)*
    ;

into_var
    : USER_VARIABLE
    | obj_access_ref_normal
    | {this.is_pl_parse_}? QUESTIONMARK Dot column_name
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
    | USE_JIT
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

distribute_method
    : NONE
    | PARTITION
    | RANDOM
    | RANDOM_LOCAL
    | HASH
    | BROADCAST
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

for_update
    : FOR UPDATE ((WAIT DECIMAL_VAL) | (WAIT INTNUM) | NOWAIT)?
    ;

parameterized_trim
    : bit_expr
    | bit_expr FROM bit_expr
    | BOTH bit_expr FROM bit_expr
    | LEADING bit_expr FROM bit_expr
    | TRAILING bit_expr FROM bit_expr
    | BOTH FROM bit_expr
    | LEADING FROM bit_expr
    | TRAILING FROM bit_expr
    ;

groupby_clause
    : key_for_group_by
    | opt_rollup
    | groupby_clause Comma key_for_group_by
    | groupby_clause Comma opt_rollup
    ;

list_for_group_by
    : key_for_group_by (Comma key_for_group_by)*
    ;

key_for_group_by
    : bit_expr
    ;

opt_rollup
    : ROLLUP LeftParen list_for_group_by RightParen
    ;

order_by
    : ORDER SIBLINGS? BY sort_list
    ;

sort_list
    : sort_key (Comma sort_key)*
    ;

sort_key
    : bit_expr opt_asc_desc
    ;

opt_null_pos
    : empty
    | NULLS LAST
    | NULLS FIRST
    ;

opt_ascending_type
    : (ASC | DESC)?
    ;

opt_asc_desc
    : opt_ascending_type opt_null_pos
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
    : bit_expr
    | bit_expr AS? column_label
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
    | LeftParen subquery order_by RightParen use_flashback?
    | select_with_parens
    | select_with_parens use_flashback
    | LeftParen table_reference RightParen
    | TABLE LeftParen simple_expr RightParen relation_name?
    ;

tbl_name
    : relation_factor use_partition? use_flashback?
    | relation_factor use_partition? sample_clause
    | relation_factor use_partition? sample_clause use_flashback
    | relation_factor use_partition? sample_clause seed
    | relation_factor use_partition? sample_clause seed use_flashback
    | relation_factor use_partition? sample_clause seed relation_name
    | relation_factor use_partition? sample_clause seed use_flashback relation_name
    | relation_factor use_partition? sample_clause relation_name
    | relation_factor use_partition? sample_clause use_flashback relation_name
    | relation_factor use_partition? relation_name
    | relation_factor use_partition? use_flashback relation_name
    ;

dml_table_name
    : relation_factor use_partition?
    ;

insert_table_clause
    : dml_table_name relation_name?
    | select_with_parens relation_name?
    | LeftParen subquery order_by RightParen relation_name?
    ;

dml_table_clause
    : dml_table_name relation_name?
    | ONLY LeftParen dml_table_name RightParen relation_name?
    | select_with_parens relation_name?
    | LeftParen subquery order_by RightParen relation_name?
    ;

seed
    : SEED LeftParen INTNUM RightParen
    ;

sample_percent
    : INTNUM
    | DECIMAL_VAL
    ;

sample_clause
    : SAMPLE BLOCK? (ALL | BASE | INCR)? LeftParen sample_percent RightParen
    ;

table_subquery
    : select_with_parens use_flashback? relation_name
    | LeftParen subquery order_by RightParen use_flashback? relation_name
    ;

use_partition
    : PARTITION LeftParen name_list RightParen
    | SUBPARTITION LeftParen name_list RightParen
    ;

use_flashback
    : AS OF TIMESTAMP simple_expr
    | AS OF SCN simple_expr
    ;

relation_factor
    : normal_relation_factor
    | dot_relation_factor
    ;

normal_relation_factor
    : relation_name
    | database_factor Dot relation_name
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

tracing_num_list
    : INTNUM relation_sep_option tracing_num_list
    | INTNUM
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

join_condition
    : ON expr
    | USING LeftParen column_list RightParen
    ;

joined_table
    : table_factor outer_join_type JOIN table_factor join_condition
    | joined_table outer_join_type JOIN table_factor join_condition
    | table_factor INNER JOIN table_factor ON expr
    | joined_table INNER JOIN table_factor ON expr
    | table_factor INNER JOIN table_factor USING LeftParen column_list RightParen
    | joined_table INNER JOIN table_factor USING LeftParen column_list RightParen
    | table_factor JOIN table_factor ON expr
    | joined_table JOIN table_factor ON expr
    | table_factor JOIN table_factor USING LeftParen column_list RightParen
    | joined_table JOIN table_factor USING LeftParen column_list RightParen
    | table_factor natural_join_type table_factor
    | joined_table natural_join_type table_factor
    | table_factor CROSS JOIN table_factor
    | joined_table CROSS JOIN table_factor
    ;

natural_join_type
    : NATURAL outer_join_type JOIN
    | NATURAL JOIN
    | NATURAL INNER JOIN
    ;

outer_join_type
    : FULL join_outer
    | LEFT join_outer
    | RIGHT join_outer
    ;

join_outer
    : OUTER?
    ;

with_select
    : with_clause select_no_parens
    | with_clause select_with_parens
    ;

with_clause
    : WITH with_list
    | WITH RECURSIVE common_table_expr
    ;

with_list
    : common_table_expr (Comma common_table_expr)*
    ;

common_table_expr
    : relation_name (LeftParen alias_name_list RightParen)? AS select_with_parens ((SEARCH DEPTH FIRST BY sort_list search_set_value) | (SEARCH BREADTH FIRST BY sort_list search_set_value))? (CYCLE alias_name_list SET var_name TO STRING_VALUE DEFAULT STRING_VALUE)?
    | relation_name (LeftParen alias_name_list RightParen)? AS LeftParen subquery order_by RightParen ((SEARCH DEPTH FIRST BY sort_list search_set_value) | (SEARCH BREADTH FIRST BY sort_list search_set_value))? (CYCLE alias_name_list SET var_name TO STRING_VALUE DEFAULT STRING_VALUE)?
    ;

alias_name_list
    : column_alias_name (Comma column_alias_name)*
    ;

column_alias_name
    : column_name
    ;

search_set_value
    : SET var_name
    ;

analyze_stmt
    : ANALYZE TABLE relation_factor use_partition? analyze_statistics_clause
    ;

analyze_statistics_clause
    : COMPUTE STATISTICS opt_analyze_for_clause_list?
    | ESTIMATE STATISTICS opt_analyze_for_clause_list? (SAMPLE INTNUM sample_option)?
    ;

opt_analyze_for_clause_list
    : opt_analyze_for_clause_element+
    ;

opt_analyze_for_clause_element
    : FOR TABLE
    | FOR ALL opt_analyze_index COLUMNS opt_bucket_num
    | FOR COLUMNS opt_bucket_num analyze_column_list
    ;

opt_analyze_index
    : INDEX?
    ;

analyze_column_list
    : analyze_column_info (Comma analyze_column_info)*
    ;

analyze_column_info
    : column_name (SIZE INTNUM)?
    ;

opt_bucket_num
    : SIZE INTNUM
    | empty
    ;

sample_option
    : ROWS
    | PERCENTAGE
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
    | merge_stmt
    | update_stmt
    ;

format_name
    : TRADITIONAL
    | JSON
    ;

show_stmt
    : SHOW FULL? columns_or_fields from_or_in relation_factor (from_or_in database_factor)? ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW TABLE STATUS (from_or_in database_factor)? ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW (GLOBAL | SESSION | LOCAL)? VARIABLES ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW CREATE TABLE relation_factor
    | SHOW CREATE VIEW relation_factor
    | SHOW CREATE PROCEDURE relation_factor
    | SHOW CREATE FUNCTION relation_factor
    | SHOW GRANTS opt_for_grant_user
    | SHOW charset_key ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW TRACE ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW COLLATION ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW PARAMETERS ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))? tenant_name?
    | SHOW FULL? PROCESSLIST
    | SHOW TABLEGROUPS ((LIKE STRING_VALUE) | (LIKE STRING_VALUE ESCAPE STRING_VALUE) | (WHERE expr))?
    | SHOW PRIVILEGES
    | SHOW RECYCLEBIN
    | SHOW CREATE TABLEGROUP relation_name
    ;

opt_for_grant_user
    : opt_for_user
    | FOR CURRENT_USER LeftParen RightParen
    ;

columns_or_fields
    : COLUMNS
    | FIELDS
    ;

from_or_in
    : FROM
    | IN
    ;

help_stmt
    : HELP STRING_VALUE
    | HELP NAME_OB
    ;

create_user_stmt
    : CREATE USER user_specification user_profile?
    | CREATE USER user_specification require_specification user_profile?
    ;

alter_user_profile_stmt
    : ALTER USER user_with_host_name user_profile
    ;

user_specification
    : user USER_VARIABLE? IDENTIFIED BY password
    | user USER_VARIABLE? IDENTIFIED BY VALUES password
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

grant_user
    : user USER_VARIABLE?
    | CONNECT
    | RESOURCE
    ;

grant_user_list
    : grant_user (Comma grant_user)*
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
    : INTNUM
    | NAME_OB
    | unreserved_keyword
    ;

drop_user_stmt
    : DROP USER user_list CASCADE?
    ;

user_list
    : user_with_host_name (Comma user_with_host_name)*
    ;

set_password_stmt
    : SET PASSWORD (FOR user opt_host_name)? COMP_EQ STRING_VALUE
    | SET PASSWORD (FOR user opt_host_name)? COMP_EQ PASSWORD LeftParen password RightParen
    | ALTER USER user_with_host_name IDENTIFIED BY password
    | ALTER USER user_with_host_name IDENTIFIED BY VALUES STRING_VALUE
    | ALTER USER user_with_host_name require_specification
    ;

opt_for_user
    : FOR user opt_host_name
    | empty
    ;

lock_user_stmt
    : ALTER USER user_list ACCOUNT lock_spec_mysql57
    ;

lock_spec_mysql57
    : LOCK
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

create_sequence_stmt
    : CREATE SEQUENCE relation_factor sequence_option_list?
    ;

sequence_option_list
    : sequence_option+
    ;

sequence_option
    : INCREMENT BY simple_num
    | START WITH simple_num
    | MAXVALUE simple_num
    | NOMAXVALUE
    | MINVALUE simple_num
    | NOMINVALUE
    | CYCLE
    | NOCYCLE
    | CACHE simple_num
    | NOCACHE
    | ORDER
    | NOORDER
    ;

simple_num
    : Plus INTNUM
    | Minus INTNUM
    | INTNUM
    | Plus DECIMAL_VAL
    | Minus DECIMAL_VAL
    | DECIMAL_VAL
    ;

drop_sequence_stmt
    : DROP SEQUENCE relation_factor
    ;

alter_sequence_stmt
    : ALTER SEQUENCE relation_factor sequence_option_list?
    ;

begin_stmt
    : BEGI WORK?
    | START TRANSACTION ((WITH CONSISTENT SNAPSHOT) | transaction_access_mode | (WITH CONSISTENT SNAPSHOT Comma transaction_access_mode) | (transaction_access_mode Comma WITH CONSISTENT SNAPSHOT))?
    ;

commit_stmt
    : COMMIT WORK?
    | COMMIT COMMENT STRING_VALUE
    ;

rollback_stmt
    : ROLLBACK WORK?
    ;

kill_stmt
    : KILL bit_expr
    | KILL CONNECTION bit_expr
    | KILL QUERY bit_expr
    ;

create_role_stmt
    : CREATE ROLE role (NOT IDENTIFIED)?
    ;

role_list
    : role (Comma role)*
    ;

role
    : STRING_VALUE
    | NAME_OB
    ;

drop_role_stmt
    : DROP ROLE role
    ;

system_privilege
    : CREATE SESSION
    | EXEMPT REDACTION POLICY
    ;

system_privilege_list
    : system_privilege (Comma system_privilege)*
    ;

grant_stmt
    : GRANT grant_privileges ON priv_level TO grant_user_list grant_options
    | GRANT grant_system_privileges
    ;

grant_system_privileges
    : role_list TO grantee_clause (WITH ADMIN OPTION)?
    | system_privilege_list TO grantee_clause (WITH ADMIN OPTION)?
    ;

grantee_clause
    : grant_user
    | grant_user IDENTIFIED BY password
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
    | REVOKE revoke_system_privilege
    ;

revoke_system_privilege
    : role_list FROM grantee_clause
    ;

prepare_stmt
    : PREPARE stmt_name FROM preparable_stmt
    ;

stmt_name
    : column_label
    ;

preparable_stmt
    : select_stmt
    | insert_stmt
    | merge_stmt
    | update_stmt
    | delete_stmt
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
    : bit_expr
    | BOOL_VALUE
    | ON
    | OFF
    | BINARY
    | DEFAULT
    ;

var_and_val
    : USER_VARIABLE to_or_eq bit_expr
    | USER_VARIABLE SET_VAR bit_expr
    | USER_VARIABLE to_or_eq BOOL_VALUE
    | USER_VARIABLE SET_VAR BOOL_VALUE
    | sys_var_and_val
    | scope_or_scope_alias column_name to_or_eq set_expr_or_default
    | SYSTEM_VARIABLE to_or_eq set_expr_or_default
    ;

sys_var_and_val
    : obj_access_ref_normal to_or_eq set_expr_or_default
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

argument
    : USER_VARIABLE
    ;

execute_stmt
    : EXECUTE stmt_name (USING argument_list)?
    ;

argument_list
    : argument (Comma argument)*
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
    : RENAME rename_table_actions
    ;

rename_table_actions
    : rename_table_action
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
    | opt_alter_compress_option
    | alter_column_option
    | alter_tablegroup_option
    | RENAME TO? relation_factor
    | alter_index_option
    | alter_partition_option
    | modify_partition_info
    | DROP CONSTRAINT constraint_name
    ;

alter_partition_option
    : DROP PARTITION drop_partition_name_list
    | add_range_or_list_partition
    | SPLIT PARTITION relation_factor split_actions
    | TRUNCATE PARTITION name_list
    ;

drop_partition_name_list
    : name_list
    | LeftParen name_list RightParen
    ;

split_actions
    : VALUES LeftParen list_expr RightParen modify_special_partition
    | AT LeftParen range_expr_list RightParen modify_special_partition
    | split_range_partition
    | split_list_partition
    ;

add_range_or_list_partition
    : ADD range_partition_list
    | ADD list_partition_list
    ;

modify_special_partition
    : INTO opt_special_partition_list
    | empty
    ;

split_range_partition
    : INTO opt_range_partition_list
    | INTO LeftParen range_partition_list Comma special_partition_list RightParen
    ;

split_list_partition
    : INTO opt_list_partition_list
    | INTO LeftParen list_partition_list Comma special_partition_list RightParen
    ;

modify_partition_info
    : modify hash_partition_option
    | modify list_partition_option
    | modify range_partition_option
    ;

tg_modify_partition_info
    : modify tg_hash_partition_option
    | modify tg_range_partition_option
    | modify tg_list_partition_option
    ;

alter_index_option
    : ADD out_of_line_constraint
    | ALTER INDEX index_name visibility_option
    | RENAME CONSTRAINT index_name TO index_name
    | MODIFY CONSTRAINT constraint_name (RELY | NORELY)? (ENABLE | DISABLE)? (VALIDATE | NOVALIDATE)?
    ;

visibility_option
    : VISIBLE
    | INVISIBLE
    ;

alter_column_option
    : ADD column_definition
    | ADD LeftParen column_definition_list RightParen
    | DROP COLUMN column_name (CASCADE | RESTRICT)?
    | DROP LeftParen column_name_list RightParen
    | ALTER COLUMN? column_name alter_column_behavior
    | RENAME COLUMN column_name TO column_name
    | MODIFY column_definition_opt_datatype
    | MODIFY LeftParen column_definition_opt_datatype_list RightParen
    ;

alter_tablegroup_option
    : DROP TABLEGROUP
    ;

alter_column_behavior
    : DROP DEFAULT
    ;

flashback_stmt
    : FLASHBACK TABLE relation_factors TO BEFORE DROP (RENAME TO relation_factor)?
    | FLASHBACK database_key database_factor TO BEFORE DROP (RENAME TO database_factor)?
    | FLASHBACK TENANT relation_name TO BEFORE DROP (RENAME TO relation_name)?
    | FLASHBACK TABLE relation_factors TO TIMESTAMP simple_expr
    | FLASHBACK TABLE relation_factors TO SCN simple_expr
    ;

relation_factors
    : relation_factor (Comma relation_factor)*
    ;

purge_stmt
    : PURGE TABLE relation_factor
    | PURGE INDEX relation_factor
    | PURGE database_key database_factor
    | PURGE TENANT relation_name
    | PURGE RECYCLEBIN
    ;

shrink_space_stmt
    : ALTER TABLE relation_factor SHRINK SPACE
    | ALTER TENANT relation_name SHRINK SPACE
    | ALTER TENANT ALL SHRINK SPACE
    ;

audit_stmt
    : audit_or_noaudit audit_clause
    ;

audit_or_noaudit
    : AUDIT
    | NOAUDIT
    ;

audit_clause
    : audit_operation_clause auditing_on_clause op_audit_tail_clause
    | audit_operation_clause op_audit_tail_clause
    | audit_operation_clause auditing_by_user_clause op_audit_tail_clause
    ;

audit_operation_clause
    : audit_all_shortcut_list
    | ALL
    | ALL STATEMENTS
    ;

audit_all_shortcut_list
    : audit_all_shortcut (Comma audit_all_shortcut)*
    ;

auditing_on_clause
    : ON normal_relation_factor
    | ON DEFAULT
    ;

auditing_by_user_clause
    : BY user_list
    ;

op_audit_tail_clause
    : empty
    | audit_by_session_access_option
    | audit_whenever_option
    | audit_by_session_access_option audit_whenever_option
    ;

audit_by_session_access_option
    : BY ACCESS
    ;

audit_whenever_option
    : WHENEVER NOT SUCCESSFUL
    | WHENEVER SUCCESSFUL
    ;

audit_all_shortcut
    : ALTER SYSTEM
    | CLUSTER
    | CONTEXT
    | DATABASE LINK
    | MATERIALIZED VIEW
    | NOT EXISTS
    | OUTLINE
    | PROCEDURE
    | PROFILE
    | PUBLIC DATABASE LINK
    | PUBLIC SYNONYM
    | ROLE
    | SEQUENCE
    | SESSION
    | SYNONYM
    | SYSTEM AUDIT
    | SYSTEM GRANT
    | TABLE
    | TABLESPACE
    | TRIGGER
    | TYPE
    | USER
    | VIEW
    | ALTER SEQUENCE
    | ALTER TABLE
    | COMMENT TABLE
    | DELETE TABLE
    | EXECUTE PROCEDURE
    | GRANT PROCEDURE
    | GRANT SEQUENCE
    | GRANT TABLE
    | GRANT TYPE
    | INSERT TABLE
    | SELECT SEQUENCE
    | SELECT TABLE
    | UPDATE TABLE
    | ALTER
    | AUDIT
    | COMMENT
    | DELETE
    | EXECUTE
    | FLASHBACK
    | GRANT
    | INDEX
    | INSERT
    | RENAME
    | SELECT
    | UPDATE
    ;

alter_system_stmt
    : ALTER SYSTEM BOOTSTRAP (CLUSTER partition_role)? server_info_list (USER user PASSWORD password)?
    | ALTER SYSTEM FLUSH cache_type CACHE (TENANT COMP_EQ tenant_name_list)? flush_scope
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
    | ALTER SYSTEM REFRESH TIME_ZONE_INFO
    | ALTER SYSTEM SET DISK VALID ip_port
    | ALTER SYSTEM DROP TABLES IN SESSION INTNUM
    | ALTER SYSTEM REFRESH TABLES IN SESSION INTNUM
    | ALTER SYSTEM SET alter_system_set_clause_list
    ;

alter_system_set_clause_list
    : alter_system_set_clause+
    ;

alter_system_set_clause
    : set_system_parameter_clause
    ;

set_system_parameter_clause
    : var_name COMP_EQ bit_expr
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

modify
    : MODIFY
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

partition_role
    : LEADER
    | FOLLOWER
    ;

upgrade_action
    : BEGI
    | END
    ;

alter_session_stmt
    : ALTER SESSION SET CURRENT_SCHEMA COMP_EQ current_schema
    | ALTER SESSION SET ISOLATION_LEVEL COMP_EQ session_isolation_level
    | ALTER SESSION SET alter_session_set_clause
    ;

session_isolation_level
    : isolation_level
    ;

alter_session_set_clause
    : set_system_parameter_clause_list
    ;

set_system_parameter_clause_list
    : set_system_parameter_clause+
    ;

current_schema
    : relation_name
    ;

set_comment_stmt
    : COMMENT ON TABLE normal_relation_factor IS STRING_VALUE
    | COMMENT ON COLUMN column_definition_ref IS STRING_VALUE
    ;

create_tablespace_stmt
    : CREATE TABLESPACE tablespace permanent_tablespace
    ;

drop_tablespace_stmt
    : DROP TABLESPACE tablespace
    ;

permanent_tablespace
    : permanent_tablespace_options?
    ;

permanent_tablespace_options
    : permanent_tablespace_option (Comma permanent_tablespace_option)*
    ;

permanent_tablespace_option
    : ENCRYPTION USING STRING_VALUE
    ;

create_profile_stmt
    : CREATE PROFILE profile_name LIMIT password_parameters
    ;

alter_profile_stmt
    : ALTER PROFILE profile_name LIMIT password_parameters
    ;

drop_profile_stmt
    : DROP PROFILE profile_name
    ;

profile_name
    : NAME_OB
    | unreserved_keyword
    | DEFAULT
    ;

password_parameters
    : password_parameter+
    ;

password_parameter
    : password_parameter_type password_parameter_value
    ;

verify_function_name
    : relation_name
    | NULLX
    ;

password_parameter_value
    : number_literal
    | UNLIMITED
    | verify_function_name
    | DEFAULT
    ;

password_parameter_type
    : FAILED_LOGIN_ATTEMPTS
    | PASSWORD_LOCK_TIME
    | PASSWORD_VERIFY_FUNCTION
    ;

user_profile
    : PROFILE profile_name
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

var_name
    : NAME_OB
    | oracle_unreserved_keyword
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
    | E_SIZE
    | G_SIZE
    | K_SIZE
    | M_SIZE
    | P_SIZE
    | T_SIZE
    ;

column_label
    : NAME_OB
    | unreserved_keyword
    ;

keystore_name
    : NAME_OB
    | unreserved_keyword
    ;

date_unit
    : YEAR
    | MONTH
    | DAY
    | HOUR
    | MINUTE
    | SECOND
    ;

timezone_unit
    : TIMEZONE_HOUR
    | TIMEZONE_MINUTE
    | TIMEZONE_REGION
    | TIMEZONE_ABBR
    ;

date_unit_for_extract
    : date_unit
    | timezone_unit
    ;

unreserved_keyword
    : oracle_unreserved_keyword
    | unreserved_keyword_normal
    | unreserved_keyword_special
    | STAT
    | LOG_LEVEL
    | CLIENT_VERSION
    ;

oracle_unreserved_keyword
    : ADMIN
    | AFTER
    | ALLOCATE
    | ANALYZE
    | ARCHIVE
    | ARCHIVELOG
    | AUTHORIZATION
    | AVG
    | BACKUP
    | BECOME
    | BEGIN_KEY
    | BLOCK
    | BODY
    | CACHE
    | CANCEL
    | CHECKPOINT
    | CLOSE
    | COBOL
    | COMMIT
    | COMPILE
    | CONSTRAINTS
    | CONTENTS
    | CONTROLFILE
    | COUNT
    | CYCLE
    | CURRENT_USER
    | DATABASE
    | DATAFILE
    | DBA
    | DISABLE
    | DISMOUNT
    | DUMP
    | ENABLE
    | ENCRYPTION
    | END
    | ESCAPE
    | EVENTS
    | EXCEPTIONS
    | EXEC
    | EXECUTE
    | EXTENT
    | EXTERNALLY
    | FLUSH
    | FOREIGN
    | FORTRAN
    | FOUND
    | FREELIST
    | FREELISTS
    | FUNCTION
    | GO
    | GOTO
    | GROUPS
    | INCLUDING
    | INDICATOR
    | INITRANS
    | INSTANCE
    | HIGH
    | KEY
    | LANGUAGE
    | LAYER
    | LINK
    | LISTS
    | LOGFILE
    | LOW
    | MANAGE
    | MANUAL
    | MAX
    | MAXDATAFILES
    | MAXINSTANCES
    | MAXLOGFILES
    | MAXLOGHISTORY
    | MAXLOGMEMBERS
    | MAXTRANS
    | MIN
    | MINEXTENTS
    | MINVALUE
    | MODULE
    | MOUNT
    | NEW
    | NEXT
    | NOARCHIVELOG
    | NOCACHE
    | NOMAXVALUE
    | NOMINVALUE
    | NONE
    | NOORDER
    | NORESETLOGS
    | NOSORT
    | OLD
    | ONLY
    | OPEN
    | OPTIMAL
    | OWN
    | PACKAGE_KEY
    | PCTINCREASE
    | PCTUSED
    | PLAN
    | PLI
    | PRECISION
    | PRIMARY
    | PRIVATE
    | PROFILE
    | QUOTA
    | RECOVER
    | REFERENCING
    | RESETLOGS
    | RESTRICTED
    | REUSE
    | ROLE
    | ROLES
    | ROLLBACK
    | SAVEPOINT
    | SCN
    | SECTION
    | SEGMENT
    | SEQUENCE
    | SHARED
    | SNAPSHOT
    | SORT
    | SQLCODE
    | SQLERROR
    | STATEMENT_ID
    | STATISTICS
    | STDDEV
    | STDDEV_POP
    | STDDEV_SAMP
    | STOP
    | STORAGE
    | SUM
    | SWITCH
    | SYSTEM
    | TABLES
    | TABLESPACE
    | TEMPORARY
    | THREAD
    | TIME
    | TRACING
    | TRIGGERS
    | TRUNCATE
    | UNDER
    | UNTIL
    | USE
    | VARIANCE
    | WORK
    | WITHIN
    | ORA_ROWSCN
    ;

unreserved_keyword_normal
    : ACCOUNT
    | ACTION
    | ACTIVE
    | ADDDATE
    | ADMINISTER
    | AGGREGATE
    | ALGORITHM
    | ANALYSE
    | APPROX_COUNT_DISTINCT
    | APPROX_COUNT_DISTINCT_SYNOPSIS
    | APPROX_COUNT_DISTINCT_SYNOPSIS_MERGE
    | AT
    | AUTHORS
    | AUTO
    | AUTOEXTEND_SIZE
    | AVG_ROW_LENGTH
    | BASE
    | BASELINE
    | BASELINE_ID
    | BASIC
    | BALANCE
    | BINDING
    | BINLOG
    | BIT
    | BLOCK_SIZE
    | BLOCK_INDEX
    | BLOOM_FILTER
    | BOOL
    | BOOLEAN
    | BOOTSTRAP
    | BTREE
    | BYTE
    | BREADTH
    | CASCADED
    | CAST
    | CATALOG_NAME
    | CHAIN
    | CHANGED
    | CHARSET
    | CHECKSUM
    | CIPHER
    | CLASS_ORIGIN
    | CLEAN
    | CLEAR
    | CLIENT
    | CLOG
    | CLUSTER_ID
    | CLUSTER_NAME
    | COALESCE
    | CODE
    | COLLATION
    | COLUMN_FORMAT
    | COLUMN_NAME
    | COLUMN_STAT
    | COLUMNS
    | COMMITTED
    | COMPACT
    | COMPLETION
    | COMPRESSED
    | COMPRESSION
    | COMPUTE
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
    | CPU
    | CREATE_TIMESTAMP
    | CUBE
    | CUME_DIST
    | CURSOR_NAME
    | DATA
    | DATABASE_ID
    | DATA_TABLE_ID
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
    | DESCRIBE
    | DESTINATION
    | DIAGNOSTICS
    | DIRECTORY
    | DISCARD
    | DISK
    | DO
    | DUMPFILE
    | DUPLICATE
    | DUPLICATE_SCOPE
    | DYNAMIC
    | DEFAULT_TABLEGROUP
    | E_SIZE
    | EFFECTIVE
    | ENDS
    | ENGINE_
    | ENGINES
    | ENUM
    | ERROR_CODE
    | ERROR_P
    | ERRORS
    | ESTIMATE
    | EVENT
    | EVERY
    | EXCHANGE
    | EXEMPT
    | EXPANSION
    | EXPIRE
    | EXPIRE_INFO
    | EXPORT
    | EXTENDED
    | EXTENDED_NOADDR
    | EXTENT_SIZE
    | EXTRACT
    | FAILED_LOGIN_ATTEMPTS
    | FAST
    | FAULTS
    | FIELDS
    | FILE_ID
    | FINAL_COUNT
    | FIRST
    | FIRST_VALUE
    | FIXED
    | FOLLOWER
    | FORMAT
    | FREEZE
    | FREQUENCY
    | G_SIZE
    | GENERAL
    | GENERATED
    | GEOMETRY
    | GEOMETRYCOLLECTION
    | GET_FORMAT
    | GLOBAL
    | GRANTS
    | GROUPING
    | GTS
    | HANDLER
    | HASH
    | HELP
    | HOST
    | HOSTS
    | HOUR
    | ID
    | IDC
    | IGNORE_SERVER_IDS
    | ILOG
    | ILOGCACHE
    | IMPORT
    | INDEXES
    | INDEX_TABLE_ID
    | INCR
    | INFO
    | INITIAL_SIZE
    | INSERT_METHOD
    | INSTALL
    | INTERVAL
    | INVOKER
    | IO
    | IO_THREAD
    | IPC
    | ISOLATION
    | ISSUER
    | IS_TENANT_SYS_POOL
    | JOB
    | JSON
    | K_SIZE
    | KEY_BLOCK_SIZE
    | KEYSTORE
    | KEY_VERSION
    | KVCACHE
    | LAG
    | LAST
    | LAST_VALUE
    | LEAD
    | LEADER
    | LEAVES
    | LESS
    | LIMIT
    | LINESTRING
    | LIST
    | LISTAGG
    | LOCAL
    | LOCALITY
    | LOCKED
    | LOCKS
    | LOGONLY_REPLICA_NUM
    | LOGS
    | M_SIZE
    | MAJOR
    | MANAGEMENT
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
    | MAX_USED_PART_ID
    | MAX_USER_CONNECTIONS
    | MEDIUM
    | MEMORY
    | MEMTABLE
    | MESSAGE_TEXT
    | META
    | MICROSECOND
    | MIGRATE
    | MIGRATION
    | MIN_CPU
    | MIN_IOPS
    | MIN_MEMORY
    | MINOR
    | MIN_ROWS
    | MINUTE
    | MONTH
    | MOVE
    | MOVEMENT
    | MULTILINESTRING
    | MULTIPOINT
    | MULTIPOLYGON
    | MUTEX
    | MYSQL_ERRNO
    | NAME
    | NAMES
    | NATIONAL
    | NCHAR
    | NDB
    | NDBCLUSTER
    | NO
    | NODEGROUP
    | NOW
    | NO_WAIT
    | NTILE
    | NTH_VALUE
    | NVARCHAR
    | NVARCHAR2
    | OCCUR
    | OLD_PASSWORD
    | OLD_KEY
    | OLTP
    | OVER
    | ONE
    | ONE_SHOT
    | OPTIONS
    | OUTLINE
    | OWNER
    | P_SIZE
    | PACK_KEYS
    | PAGE
    | PARAMETERS
    | PARSER
    | PARTIAL
    | PARTITION_ID
    | PARTITIONING
    | PARTITIONS
    | PERCENT_RANK
    | PASSWORD_LOCK_TIME
    | PASSWORD_VERIFY_FUNCTION
    | PAUSE
    | PERCENTAGE
    | PHASE
    | PLANREGRESS
    | PLUGIN
    | PLUGIN_DIR
    | PLUGINS
    | POINT
    | POLICY
    | POLYGON
    | POOL
    | PORT
    | POSITION
    | PREPARE
    | PRESERVE
    | PREV
    | PRIMARY_ZONE
    | PROCESS
    | PROCESSLIST
    | PROFILES
    | PROGRESSIVE_MERGE_NUM
    | PROXY
    | QUARTER
    | QUERY
    | QUICK
    | RANK
    | RATIO_TO_REPORT
    | READ_ONLY
    | REBUILD
    | RECYCLE
    | RECYCLEBIN
    | REDACTION
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
    | RESOURCE_POOL_LIST
    | RESPECT
    | RESTART
    | RESTORE
    | RESUME
    | RETURNED_SQLSTATE
    | RETURNS
    | REVERSE
    | REWRITE_MERGE_VERSION
    | ROLLUP
    | ROOT
    | ROOTTABLE
    | ROUTINE
    | ROW_COUNT
    | ROW_FORMAT
    | RTREE
    | RUN
    | SAMPLE
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
    | SESSION_USER
    | SET_MASTER_CLUSTER
    | SET_SLAVE_CLUSTER
    | SET_TP
    | SHRINK
    | SHOW
    | SHUTDOWN
    | SIGNED
    | SIMPLE
    | SLAVE
    | SLOW
    | SOCKET
    | SONAME
    | SOUNDS
    | SOURCE
    | SPACE
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
    | STARTS
    | STATS_AUTO_RECALC
    | STATS_PERSISTENT
    | STATS_SAMPLE_PAGES
    | STATUS
    | STATEMENTS
    | STORAGE_FORMAT_VERSION
    | STORAGE_FORMAT_WORK_VERSION
    | STORING
    | SUBCLASS_ORIGIN
    | SUBDATE
    | SUBJECT
    | SUBPARTITIONS
    | SUBSTR
    | SUPER
    | SUSPEND
    | SWAPS
    | SWITCHES
    | SYSTEM_USER
    | T_SIZE
    | TABLE_CHECKSUM
    | TABLE_MODE
    | TABLEGROUPS
    | TABLEGROUP_ID
    | TABLE_ID
    | TABLE_NAME
    | TABLET
    | TABLET_SIZE
    | TABLET_MAX_SIZE
    | TASK
    | TEMPLATE
    | TEMPTABLE
    | TENANT
    | TEXT
    | THAN
    | TIMESTAMP
    | TIMESTAMPADD
    | TIMESTAMPDIFF
    | TP_NAME
    | TP_NO
    | TRACE
    | TRADITIONAL
    | TRIM
    | TYPE
    | TYPES
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
    | UNUSUAL
    | UPGRADE
    | USAGE
    | USE_BLOOM_FILTER
    | USE_FRM
    | USER_RESOURCES
    | UNBOUNDED
    | VALID
    | VARIABLES
    | VERBOSE
    | MATERIALIZED
    | WAIT
    | WARNINGS
    | WEEK
    | WEIGHT_STRING
    | WRAPPER
    | X509
    | XA
    | XML
    | YEAR
    | ZONE
    | ZONE_LIST
    | ZONE_TYPE
    | LOCATION
    | VISIBLE
    | INVISIBLE
    | RELY
    | NORELY
    | NOVALIDATE
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

