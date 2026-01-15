// Code generated from /work/obparser/obmysql/sql/OBParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package mysql // OBParser
import "github.com/antlr4-go/antlr/v4"

// OBParserListener is a complete listener for a parse tree produced by OBParser.
type OBParserListener interface {
	antlr.ParseTreeListener

	// EnterSql_stmt is called when entering the sql_stmt production.
	EnterSql_stmt(c *Sql_stmtContext)

	// EnterStmt_list is called when entering the stmt_list production.
	EnterStmt_list(c *Stmt_listContext)

	// EnterStmt is called when entering the stmt production.
	EnterStmt(c *StmtContext)

	// EnterExpr_list is called when entering the expr_list production.
	EnterExpr_list(c *Expr_listContext)

	// EnterExpr_as_list is called when entering the expr_as_list production.
	EnterExpr_as_list(c *Expr_as_listContext)

	// EnterExpr_with_opt_alias is called when entering the expr_with_opt_alias production.
	EnterExpr_with_opt_alias(c *Expr_with_opt_aliasContext)

	// EnterColumn_ref is called when entering the column_ref production.
	EnterColumn_ref(c *Column_refContext)

	// EnterComplex_string_literal is called when entering the complex_string_literal production.
	EnterComplex_string_literal(c *Complex_string_literalContext)

	// EnterCharset_introducer is called when entering the charset_introducer production.
	EnterCharset_introducer(c *Charset_introducerContext)

	// EnterLiteral is called when entering the literal production.
	EnterLiteral(c *LiteralContext)

	// EnterNumber_literal is called when entering the number_literal production.
	EnterNumber_literal(c *Number_literalContext)

	// EnterExpr_const is called when entering the expr_const production.
	EnterExpr_const(c *Expr_constContext)

	// EnterConf_const is called when entering the conf_const production.
	EnterConf_const(c *Conf_constContext)

	// EnterGlobal_or_session_alias is called when entering the global_or_session_alias production.
	EnterGlobal_or_session_alias(c *Global_or_session_aliasContext)

	// EnterBool_pri is called when entering the bool_pri production.
	EnterBool_pri(c *Bool_priContext)

	// EnterPredicate is called when entering the predicate production.
	EnterPredicate(c *PredicateContext)

	// EnterBit_expr is called when entering the bit_expr production.
	EnterBit_expr(c *Bit_exprContext)

	// EnterSimple_expr is called when entering the simple_expr production.
	EnterSimple_expr(c *Simple_exprContext)

	// EnterExpr is called when entering the expr production.
	EnterExpr(c *ExprContext)

	// EnterNot is called when entering the not production.
	EnterNot(c *NotContext)

	// EnterNot2 is called when entering the not2 production.
	EnterNot2(c *Not2Context)

	// EnterSub_query_flag is called when entering the sub_query_flag production.
	EnterSub_query_flag(c *Sub_query_flagContext)

	// EnterIn_expr is called when entering the in_expr production.
	EnterIn_expr(c *In_exprContext)

	// EnterCase_expr is called when entering the case_expr production.
	EnterCase_expr(c *Case_exprContext)

	// EnterWindow_function is called when entering the window_function production.
	EnterWindow_function(c *Window_functionContext)

	// EnterFirst_or_last is called when entering the first_or_last production.
	EnterFirst_or_last(c *First_or_lastContext)

	// EnterRespect_or_ignore is called when entering the respect_or_ignore production.
	EnterRespect_or_ignore(c *Respect_or_ignoreContext)

	// EnterWin_fun_first_last_params is called when entering the win_fun_first_last_params production.
	EnterWin_fun_first_last_params(c *Win_fun_first_last_paramsContext)

	// EnterWin_fun_lead_lag_params is called when entering the win_fun_lead_lag_params production.
	EnterWin_fun_lead_lag_params(c *Win_fun_lead_lag_paramsContext)

	// EnterNew_generalized_window_clause is called when entering the new_generalized_window_clause production.
	EnterNew_generalized_window_clause(c *New_generalized_window_clauseContext)

	// EnterNew_generalized_window_clause_with_blanket is called when entering the new_generalized_window_clause_with_blanket production.
	EnterNew_generalized_window_clause_with_blanket(c *New_generalized_window_clause_with_blanketContext)

	// EnterNamed_windows is called when entering the named_windows production.
	EnterNamed_windows(c *Named_windowsContext)

	// EnterNamed_window is called when entering the named_window production.
	EnterNamed_window(c *Named_windowContext)

	// EnterGeneralized_window_clause is called when entering the generalized_window_clause production.
	EnterGeneralized_window_clause(c *Generalized_window_clauseContext)

	// EnterWin_rows_or_range is called when entering the win_rows_or_range production.
	EnterWin_rows_or_range(c *Win_rows_or_rangeContext)

	// EnterWin_preceding_or_following is called when entering the win_preceding_or_following production.
	EnterWin_preceding_or_following(c *Win_preceding_or_followingContext)

	// EnterWin_interval is called when entering the win_interval production.
	EnterWin_interval(c *Win_intervalContext)

	// EnterWin_bounding is called when entering the win_bounding production.
	EnterWin_bounding(c *Win_boundingContext)

	// EnterWin_window is called when entering the win_window production.
	EnterWin_window(c *Win_windowContext)

	// EnterCase_arg is called when entering the case_arg production.
	EnterCase_arg(c *Case_argContext)

	// EnterWhen_clause_list is called when entering the when_clause_list production.
	EnterWhen_clause_list(c *When_clause_listContext)

	// EnterWhen_clause is called when entering the when_clause production.
	EnterWhen_clause(c *When_clauseContext)

	// EnterCase_default is called when entering the case_default production.
	EnterCase_default(c *Case_defaultContext)

	// EnterFunc_expr is called when entering the func_expr production.
	EnterFunc_expr(c *Func_exprContext)

	// EnterSys_interval_func is called when entering the sys_interval_func production.
	EnterSys_interval_func(c *Sys_interval_funcContext)

	// EnterUtc_timestamp_func is called when entering the utc_timestamp_func production.
	EnterUtc_timestamp_func(c *Utc_timestamp_funcContext)

	// EnterSysdate_func is called when entering the sysdate_func production.
	EnterSysdate_func(c *Sysdate_funcContext)

	// EnterCur_timestamp_func is called when entering the cur_timestamp_func production.
	EnterCur_timestamp_func(c *Cur_timestamp_funcContext)

	// EnterNow_synonyms_func is called when entering the now_synonyms_func production.
	EnterNow_synonyms_func(c *Now_synonyms_funcContext)

	// EnterCur_time_func is called when entering the cur_time_func production.
	EnterCur_time_func(c *Cur_time_funcContext)

	// EnterCur_date_func is called when entering the cur_date_func production.
	EnterCur_date_func(c *Cur_date_funcContext)

	// EnterSubstr_or_substring is called when entering the substr_or_substring production.
	EnterSubstr_or_substring(c *Substr_or_substringContext)

	// EnterSubstr_params is called when entering the substr_params production.
	EnterSubstr_params(c *Substr_paramsContext)

	// EnterDate_params is called when entering the date_params production.
	EnterDate_params(c *Date_paramsContext)

	// EnterTimestamp_params is called when entering the timestamp_params production.
	EnterTimestamp_params(c *Timestamp_paramsContext)

	// EnterDelete_stmt is called when entering the delete_stmt production.
	EnterDelete_stmt(c *Delete_stmtContext)

	// EnterMulti_delete_table is called when entering the multi_delete_table production.
	EnterMulti_delete_table(c *Multi_delete_tableContext)

	// EnterUpdate_stmt is called when entering the update_stmt production.
	EnterUpdate_stmt(c *Update_stmtContext)

	// EnterUpdate_asgn_list is called when entering the update_asgn_list production.
	EnterUpdate_asgn_list(c *Update_asgn_listContext)

	// EnterUpdate_asgn_factor is called when entering the update_asgn_factor production.
	EnterUpdate_asgn_factor(c *Update_asgn_factorContext)

	// EnterCreate_resource_stmt is called when entering the create_resource_stmt production.
	EnterCreate_resource_stmt(c *Create_resource_stmtContext)

	// EnterOpt_resource_unit_option_list is called when entering the opt_resource_unit_option_list production.
	EnterOpt_resource_unit_option_list(c *Opt_resource_unit_option_listContext)

	// EnterResource_unit_option is called when entering the resource_unit_option production.
	EnterResource_unit_option(c *Resource_unit_optionContext)

	// EnterOpt_create_resource_pool_option_list is called when entering the opt_create_resource_pool_option_list production.
	EnterOpt_create_resource_pool_option_list(c *Opt_create_resource_pool_option_listContext)

	// EnterCreate_resource_pool_option is called when entering the create_resource_pool_option production.
	EnterCreate_resource_pool_option(c *Create_resource_pool_optionContext)

	// EnterAlter_resource_pool_option_list is called when entering the alter_resource_pool_option_list production.
	EnterAlter_resource_pool_option_list(c *Alter_resource_pool_option_listContext)

	// EnterUnit_id_list is called when entering the unit_id_list production.
	EnterUnit_id_list(c *Unit_id_listContext)

	// EnterAlter_resource_pool_option is called when entering the alter_resource_pool_option production.
	EnterAlter_resource_pool_option(c *Alter_resource_pool_optionContext)

	// EnterAlter_resource_stmt is called when entering the alter_resource_stmt production.
	EnterAlter_resource_stmt(c *Alter_resource_stmtContext)

	// EnterDrop_resource_stmt is called when entering the drop_resource_stmt production.
	EnterDrop_resource_stmt(c *Drop_resource_stmtContext)

	// EnterCreate_tenant_stmt is called when entering the create_tenant_stmt production.
	EnterCreate_tenant_stmt(c *Create_tenant_stmtContext)

	// EnterOpt_tenant_option_list is called when entering the opt_tenant_option_list production.
	EnterOpt_tenant_option_list(c *Opt_tenant_option_listContext)

	// EnterTenant_option is called when entering the tenant_option production.
	EnterTenant_option(c *Tenant_optionContext)

	// EnterZone_list is called when entering the zone_list production.
	EnterZone_list(c *Zone_listContext)

	// EnterResource_pool_list is called when entering the resource_pool_list production.
	EnterResource_pool_list(c *Resource_pool_listContext)

	// EnterAlter_tenant_stmt is called when entering the alter_tenant_stmt production.
	EnterAlter_tenant_stmt(c *Alter_tenant_stmtContext)

	// EnterDrop_tenant_stmt is called when entering the drop_tenant_stmt production.
	EnterDrop_tenant_stmt(c *Drop_tenant_stmtContext)

	// EnterCreate_database_stmt is called when entering the create_database_stmt production.
	EnterCreate_database_stmt(c *Create_database_stmtContext)

	// EnterDatabase_key is called when entering the database_key production.
	EnterDatabase_key(c *Database_keyContext)

	// EnterDatabase_factor is called when entering the database_factor production.
	EnterDatabase_factor(c *Database_factorContext)

	// EnterDatabase_option_list is called when entering the database_option_list production.
	EnterDatabase_option_list(c *Database_option_listContext)

	// EnterCharset_key is called when entering the charset_key production.
	EnterCharset_key(c *Charset_keyContext)

	// EnterDatabase_option is called when entering the database_option production.
	EnterDatabase_option(c *Database_optionContext)

	// EnterRead_only_or_write is called when entering the read_only_or_write production.
	EnterRead_only_or_write(c *Read_only_or_writeContext)

	// EnterDrop_database_stmt is called when entering the drop_database_stmt production.
	EnterDrop_database_stmt(c *Drop_database_stmtContext)

	// EnterAlter_database_stmt is called when entering the alter_database_stmt production.
	EnterAlter_database_stmt(c *Alter_database_stmtContext)

	// EnterLoad_data_stmt is called when entering the load_data_stmt production.
	EnterLoad_data_stmt(c *Load_data_stmtContext)

	// EnterLoad_data_with_opt_hint is called when entering the load_data_with_opt_hint production.
	EnterLoad_data_with_opt_hint(c *Load_data_with_opt_hintContext)

	// EnterLines_or_rows is called when entering the lines_or_rows production.
	EnterLines_or_rows(c *Lines_or_rowsContext)

	// EnterField_or_vars_list is called when entering the field_or_vars_list production.
	EnterField_or_vars_list(c *Field_or_vars_listContext)

	// EnterField_or_vars is called when entering the field_or_vars production.
	EnterField_or_vars(c *Field_or_varsContext)

	// EnterLoad_set_list is called when entering the load_set_list production.
	EnterLoad_set_list(c *Load_set_listContext)

	// EnterLoad_set_element is called when entering the load_set_element production.
	EnterLoad_set_element(c *Load_set_elementContext)

	// EnterUse_database_stmt is called when entering the use_database_stmt production.
	EnterUse_database_stmt(c *Use_database_stmtContext)

	// EnterCreate_synonym_stmt is called when entering the create_synonym_stmt production.
	EnterCreate_synonym_stmt(c *Create_synonym_stmtContext)

	// EnterSynonym_name is called when entering the synonym_name production.
	EnterSynonym_name(c *Synonym_nameContext)

	// EnterSynonym_object is called when entering the synonym_object production.
	EnterSynonym_object(c *Synonym_objectContext)

	// EnterDrop_synonym_stmt is called when entering the drop_synonym_stmt production.
	EnterDrop_synonym_stmt(c *Drop_synonym_stmtContext)

	// EnterTemporary_option is called when entering the temporary_option production.
	EnterTemporary_option(c *Temporary_optionContext)

	// EnterCreate_table_like_stmt is called when entering the create_table_like_stmt production.
	EnterCreate_table_like_stmt(c *Create_table_like_stmtContext)

	// EnterCreate_table_stmt is called when entering the create_table_stmt production.
	EnterCreate_table_stmt(c *Create_table_stmtContext)

	// EnterRet_type is called when entering the ret_type production.
	EnterRet_type(c *Ret_typeContext)

	// EnterCreate_function_stmt is called when entering the create_function_stmt production.
	EnterCreate_function_stmt(c *Create_function_stmtContext)

	// EnterDrop_function_stmt is called when entering the drop_function_stmt production.
	EnterDrop_function_stmt(c *Drop_function_stmtContext)

	// EnterTable_element_list is called when entering the table_element_list production.
	EnterTable_element_list(c *Table_element_listContext)

	// EnterTable_element is called when entering the table_element production.
	EnterTable_element(c *Table_elementContext)

	// EnterOpt_reference_option_list is called when entering the opt_reference_option_list production.
	EnterOpt_reference_option_list(c *Opt_reference_option_listContext)

	// EnterReference_option is called when entering the reference_option production.
	EnterReference_option(c *Reference_optionContext)

	// EnterReference_action is called when entering the reference_action production.
	EnterReference_action(c *Reference_actionContext)

	// EnterMatch_action is called when entering the match_action production.
	EnterMatch_action(c *Match_actionContext)

	// EnterColumn_definition is called when entering the column_definition production.
	EnterColumn_definition(c *Column_definitionContext)

	// EnterOpt_generated_column_attribute_list is called when entering the opt_generated_column_attribute_list production.
	EnterOpt_generated_column_attribute_list(c *Opt_generated_column_attribute_listContext)

	// EnterGenerated_column_attribute is called when entering the generated_column_attribute production.
	EnterGenerated_column_attribute(c *Generated_column_attributeContext)

	// EnterColumn_definition_ref is called when entering the column_definition_ref production.
	EnterColumn_definition_ref(c *Column_definition_refContext)

	// EnterColumn_definition_list is called when entering the column_definition_list production.
	EnterColumn_definition_list(c *Column_definition_listContext)

	// EnterCast_data_type is called when entering the cast_data_type production.
	EnterCast_data_type(c *Cast_data_typeContext)

	// EnterCast_datetime_type_i is called when entering the cast_datetime_type_i production.
	EnterCast_datetime_type_i(c *Cast_datetime_type_iContext)

	// EnterData_type is called when entering the data_type production.
	EnterData_type(c *Data_typeContext)

	// EnterString_list is called when entering the string_list production.
	EnterString_list(c *String_listContext)

	// EnterText_string is called when entering the text_string production.
	EnterText_string(c *Text_stringContext)

	// EnterInt_type_i is called when entering the int_type_i production.
	EnterInt_type_i(c *Int_type_iContext)

	// EnterFloat_type_i is called when entering the float_type_i production.
	EnterFloat_type_i(c *Float_type_iContext)

	// EnterDatetime_type_i is called when entering the datetime_type_i production.
	EnterDatetime_type_i(c *Datetime_type_iContext)

	// EnterDate_year_type_i is called when entering the date_year_type_i production.
	EnterDate_year_type_i(c *Date_year_type_iContext)

	// EnterText_type_i is called when entering the text_type_i production.
	EnterText_type_i(c *Text_type_iContext)

	// EnterBlob_type_i is called when entering the blob_type_i production.
	EnterBlob_type_i(c *Blob_type_iContext)

	// EnterString_length_i is called when entering the string_length_i production.
	EnterString_length_i(c *String_length_iContext)

	// EnterCollation_name is called when entering the collation_name production.
	EnterCollation_name(c *Collation_nameContext)

	// EnterTrans_param_name is called when entering the trans_param_name production.
	EnterTrans_param_name(c *Trans_param_nameContext)

	// EnterTrans_param_value is called when entering the trans_param_value production.
	EnterTrans_param_value(c *Trans_param_valueContext)

	// EnterCharset_name is called when entering the charset_name production.
	EnterCharset_name(c *Charset_nameContext)

	// EnterCharset_name_or_default is called when entering the charset_name_or_default production.
	EnterCharset_name_or_default(c *Charset_name_or_defaultContext)

	// EnterCollation is called when entering the collation production.
	EnterCollation(c *CollationContext)

	// EnterOpt_column_attribute_list is called when entering the opt_column_attribute_list production.
	EnterOpt_column_attribute_list(c *Opt_column_attribute_listContext)

	// EnterColumn_attribute is called when entering the column_attribute production.
	EnterColumn_attribute(c *Column_attributeContext)

	// EnterNow_or_signed_literal is called when entering the now_or_signed_literal production.
	EnterNow_or_signed_literal(c *Now_or_signed_literalContext)

	// EnterSigned_literal is called when entering the signed_literal production.
	EnterSigned_literal(c *Signed_literalContext)

	// EnterOpt_comma is called when entering the opt_comma production.
	EnterOpt_comma(c *Opt_commaContext)

	// EnterTable_option_list_space_seperated is called when entering the table_option_list_space_seperated production.
	EnterTable_option_list_space_seperated(c *Table_option_list_space_seperatedContext)

	// EnterTable_option_list is called when entering the table_option_list production.
	EnterTable_option_list(c *Table_option_listContext)

	// EnterPrimary_zone_name is called when entering the primary_zone_name production.
	EnterPrimary_zone_name(c *Primary_zone_nameContext)

	// EnterTablespace is called when entering the tablespace production.
	EnterTablespace(c *TablespaceContext)

	// EnterLocality_name is called when entering the locality_name production.
	EnterLocality_name(c *Locality_nameContext)

	// EnterTable_option is called when entering the table_option production.
	EnterTable_option(c *Table_optionContext)

	// EnterRelation_name_or_string is called when entering the relation_name_or_string production.
	EnterRelation_name_or_string(c *Relation_name_or_stringContext)

	// EnterOpt_equal_mark is called when entering the opt_equal_mark production.
	EnterOpt_equal_mark(c *Opt_equal_markContext)

	// EnterPartition_option is called when entering the partition_option production.
	EnterPartition_option(c *Partition_optionContext)

	// EnterOpt_partition_option is called when entering the opt_partition_option production.
	EnterOpt_partition_option(c *Opt_partition_optionContext)

	// EnterHash_partition_option is called when entering the hash_partition_option production.
	EnterHash_partition_option(c *Hash_partition_optionContext)

	// EnterList_partition_option is called when entering the list_partition_option production.
	EnterList_partition_option(c *List_partition_optionContext)

	// EnterKey_partition_option is called when entering the key_partition_option production.
	EnterKey_partition_option(c *Key_partition_optionContext)

	// EnterRange_partition_option is called when entering the range_partition_option production.
	EnterRange_partition_option(c *Range_partition_optionContext)

	// EnterOpt_column_partition_option is called when entering the opt_column_partition_option production.
	EnterOpt_column_partition_option(c *Opt_column_partition_optionContext)

	// EnterColumn_partition_option is called when entering the column_partition_option production.
	EnterColumn_partition_option(c *Column_partition_optionContext)

	// EnterAux_column_list is called when entering the aux_column_list production.
	EnterAux_column_list(c *Aux_column_listContext)

	// EnterVertical_column_name is called when entering the vertical_column_name production.
	EnterVertical_column_name(c *Vertical_column_nameContext)

	// EnterColumn_name_list is called when entering the column_name_list production.
	EnterColumn_name_list(c *Column_name_listContext)

	// EnterSubpartition_option is called when entering the subpartition_option production.
	EnterSubpartition_option(c *Subpartition_optionContext)

	// EnterOpt_list_partition_list is called when entering the opt_list_partition_list production.
	EnterOpt_list_partition_list(c *Opt_list_partition_listContext)

	// EnterOpt_list_subpartition_list is called when entering the opt_list_subpartition_list production.
	EnterOpt_list_subpartition_list(c *Opt_list_subpartition_listContext)

	// EnterOpt_range_partition_list is called when entering the opt_range_partition_list production.
	EnterOpt_range_partition_list(c *Opt_range_partition_listContext)

	// EnterOpt_range_subpartition_list is called when entering the opt_range_subpartition_list production.
	EnterOpt_range_subpartition_list(c *Opt_range_subpartition_listContext)

	// EnterList_partition_list is called when entering the list_partition_list production.
	EnterList_partition_list(c *List_partition_listContext)

	// EnterList_subpartition_list is called when entering the list_subpartition_list production.
	EnterList_subpartition_list(c *List_subpartition_listContext)

	// EnterList_subpartition_element is called when entering the list_subpartition_element production.
	EnterList_subpartition_element(c *List_subpartition_elementContext)

	// EnterList_partition_element is called when entering the list_partition_element production.
	EnterList_partition_element(c *List_partition_elementContext)

	// EnterList_partition_expr is called when entering the list_partition_expr production.
	EnterList_partition_expr(c *List_partition_exprContext)

	// EnterList_expr is called when entering the list_expr production.
	EnterList_expr(c *List_exprContext)

	// EnterRange_partition_list is called when entering the range_partition_list production.
	EnterRange_partition_list(c *Range_partition_listContext)

	// EnterRange_partition_element is called when entering the range_partition_element production.
	EnterRange_partition_element(c *Range_partition_elementContext)

	// EnterRange_subpartition_element is called when entering the range_subpartition_element production.
	EnterRange_subpartition_element(c *Range_subpartition_elementContext)

	// EnterRange_subpartition_list is called when entering the range_subpartition_list production.
	EnterRange_subpartition_list(c *Range_subpartition_listContext)

	// EnterRange_partition_expr is called when entering the range_partition_expr production.
	EnterRange_partition_expr(c *Range_partition_exprContext)

	// EnterRange_expr_list is called when entering the range_expr_list production.
	EnterRange_expr_list(c *Range_expr_listContext)

	// EnterRange_expr is called when entering the range_expr production.
	EnterRange_expr(c *Range_exprContext)

	// EnterInt_or_decimal is called when entering the int_or_decimal production.
	EnterInt_or_decimal(c *Int_or_decimalContext)

	// EnterTg_hash_partition_option is called when entering the tg_hash_partition_option production.
	EnterTg_hash_partition_option(c *Tg_hash_partition_optionContext)

	// EnterTg_key_partition_option is called when entering the tg_key_partition_option production.
	EnterTg_key_partition_option(c *Tg_key_partition_optionContext)

	// EnterTg_range_partition_option is called when entering the tg_range_partition_option production.
	EnterTg_range_partition_option(c *Tg_range_partition_optionContext)

	// EnterTg_list_partition_option is called when entering the tg_list_partition_option production.
	EnterTg_list_partition_option(c *Tg_list_partition_optionContext)

	// EnterTg_subpartition_option is called when entering the tg_subpartition_option production.
	EnterTg_subpartition_option(c *Tg_subpartition_optionContext)

	// EnterRow_format_option is called when entering the row_format_option production.
	EnterRow_format_option(c *Row_format_optionContext)

	// EnterCreate_tablegroup_stmt is called when entering the create_tablegroup_stmt production.
	EnterCreate_tablegroup_stmt(c *Create_tablegroup_stmtContext)

	// EnterDrop_tablegroup_stmt is called when entering the drop_tablegroup_stmt production.
	EnterDrop_tablegroup_stmt(c *Drop_tablegroup_stmtContext)

	// EnterAlter_tablegroup_stmt is called when entering the alter_tablegroup_stmt production.
	EnterAlter_tablegroup_stmt(c *Alter_tablegroup_stmtContext)

	// EnterTablegroup_option_list_space_seperated is called when entering the tablegroup_option_list_space_seperated production.
	EnterTablegroup_option_list_space_seperated(c *Tablegroup_option_list_space_seperatedContext)

	// EnterTablegroup_option_list is called when entering the tablegroup_option_list production.
	EnterTablegroup_option_list(c *Tablegroup_option_listContext)

	// EnterTablegroup_option is called when entering the tablegroup_option production.
	EnterTablegroup_option(c *Tablegroup_optionContext)

	// EnterAlter_tablegroup_actions is called when entering the alter_tablegroup_actions production.
	EnterAlter_tablegroup_actions(c *Alter_tablegroup_actionsContext)

	// EnterAlter_tablegroup_action is called when entering the alter_tablegroup_action production.
	EnterAlter_tablegroup_action(c *Alter_tablegroup_actionContext)

	// EnterDefault_tablegroup is called when entering the default_tablegroup production.
	EnterDefault_tablegroup(c *Default_tablegroupContext)

	// EnterCreate_view_stmt is called when entering the create_view_stmt production.
	EnterCreate_view_stmt(c *Create_view_stmtContext)

	// EnterView_select_stmt is called when entering the view_select_stmt production.
	EnterView_select_stmt(c *View_select_stmtContext)

	// EnterView_name is called when entering the view_name production.
	EnterView_name(c *View_nameContext)

	// EnterCreate_index_stmt is called when entering the create_index_stmt production.
	EnterCreate_index_stmt(c *Create_index_stmtContext)

	// EnterIndex_name is called when entering the index_name production.
	EnterIndex_name(c *Index_nameContext)

	// EnterOpt_constraint_name is called when entering the opt_constraint_name production.
	EnterOpt_constraint_name(c *Opt_constraint_nameContext)

	// EnterConstraint_name is called when entering the constraint_name production.
	EnterConstraint_name(c *Constraint_nameContext)

	// EnterSort_column_list is called when entering the sort_column_list production.
	EnterSort_column_list(c *Sort_column_listContext)

	// EnterSort_column_key is called when entering the sort_column_key production.
	EnterSort_column_key(c *Sort_column_keyContext)

	// EnterOpt_index_options is called when entering the opt_index_options production.
	EnterOpt_index_options(c *Opt_index_optionsContext)

	// EnterIndex_option is called when entering the index_option production.
	EnterIndex_option(c *Index_optionContext)

	// EnterIndex_using_algorithm is called when entering the index_using_algorithm production.
	EnterIndex_using_algorithm(c *Index_using_algorithmContext)

	// EnterDrop_table_stmt is called when entering the drop_table_stmt production.
	EnterDrop_table_stmt(c *Drop_table_stmtContext)

	// EnterTable_or_tables is called when entering the table_or_tables production.
	EnterTable_or_tables(c *Table_or_tablesContext)

	// EnterDrop_view_stmt is called when entering the drop_view_stmt production.
	EnterDrop_view_stmt(c *Drop_view_stmtContext)

	// EnterTable_list is called when entering the table_list production.
	EnterTable_list(c *Table_listContext)

	// EnterDrop_index_stmt is called when entering the drop_index_stmt production.
	EnterDrop_index_stmt(c *Drop_index_stmtContext)

	// EnterInsert_stmt is called when entering the insert_stmt production.
	EnterInsert_stmt(c *Insert_stmtContext)

	// EnterSingle_table_insert is called when entering the single_table_insert production.
	EnterSingle_table_insert(c *Single_table_insertContext)

	// EnterValues_clause is called when entering the values_clause production.
	EnterValues_clause(c *Values_clauseContext)

	// EnterValue_or_values is called when entering the value_or_values production.
	EnterValue_or_values(c *Value_or_valuesContext)

	// EnterReplace_with_opt_hint is called when entering the replace_with_opt_hint production.
	EnterReplace_with_opt_hint(c *Replace_with_opt_hintContext)

	// EnterInsert_with_opt_hint is called when entering the insert_with_opt_hint production.
	EnterInsert_with_opt_hint(c *Insert_with_opt_hintContext)

	// EnterColumn_list is called when entering the column_list production.
	EnterColumn_list(c *Column_listContext)

	// EnterInsert_vals_list is called when entering the insert_vals_list production.
	EnterInsert_vals_list(c *Insert_vals_listContext)

	// EnterInsert_vals is called when entering the insert_vals production.
	EnterInsert_vals(c *Insert_valsContext)

	// EnterExpr_or_default is called when entering the expr_or_default production.
	EnterExpr_or_default(c *Expr_or_defaultContext)

	// EnterSelect_stmt is called when entering the select_stmt production.
	EnterSelect_stmt(c *Select_stmtContext)

	// EnterSelect_into is called when entering the select_into production.
	EnterSelect_into(c *Select_intoContext)

	// EnterSelect_with_parens is called when entering the select_with_parens production.
	EnterSelect_with_parens(c *Select_with_parensContext)

	// EnterSelect_no_parens is called when entering the select_no_parens production.
	EnterSelect_no_parens(c *Select_no_parensContext)

	// EnterNo_table_select is called when entering the no_table_select production.
	EnterNo_table_select(c *No_table_selectContext)

	// EnterSelect_clause is called when entering the select_clause production.
	EnterSelect_clause(c *Select_clauseContext)

	// EnterSelect_clause_set_with_order_and_limit is called when entering the select_clause_set_with_order_and_limit production.
	EnterSelect_clause_set_with_order_and_limit(c *Select_clause_set_with_order_and_limitContext)

	// EnterSelect_clause_set is called when entering the select_clause_set production.
	EnterSelect_clause_set(c *Select_clause_setContext)

	// EnterSelect_clause_set_right is called when entering the select_clause_set_right production.
	EnterSelect_clause_set_right(c *Select_clause_set_rightContext)

	// EnterSelect_clause_set_left is called when entering the select_clause_set_left production.
	EnterSelect_clause_set_left(c *Select_clause_set_leftContext)

	// EnterNo_table_select_with_order_and_limit is called when entering the no_table_select_with_order_and_limit production.
	EnterNo_table_select_with_order_and_limit(c *No_table_select_with_order_and_limitContext)

	// EnterSimple_select_with_order_and_limit is called when entering the simple_select_with_order_and_limit production.
	EnterSimple_select_with_order_and_limit(c *Simple_select_with_order_and_limitContext)

	// EnterSelect_with_parens_with_order_and_limit is called when entering the select_with_parens_with_order_and_limit production.
	EnterSelect_with_parens_with_order_and_limit(c *Select_with_parens_with_order_and_limitContext)

	// EnterSelect_with_opt_hint is called when entering the select_with_opt_hint production.
	EnterSelect_with_opt_hint(c *Select_with_opt_hintContext)

	// EnterUpdate_with_opt_hint is called when entering the update_with_opt_hint production.
	EnterUpdate_with_opt_hint(c *Update_with_opt_hintContext)

	// EnterDelete_with_opt_hint is called when entering the delete_with_opt_hint production.
	EnterDelete_with_opt_hint(c *Delete_with_opt_hintContext)

	// EnterSimple_select is called when entering the simple_select production.
	EnterSimple_select(c *Simple_selectContext)

	// EnterSet_type_union is called when entering the set_type_union production.
	EnterSet_type_union(c *Set_type_unionContext)

	// EnterSet_type_other is called when entering the set_type_other production.
	EnterSet_type_other(c *Set_type_otherContext)

	// EnterSet_type is called when entering the set_type production.
	EnterSet_type(c *Set_typeContext)

	// EnterSet_expression_option is called when entering the set_expression_option production.
	EnterSet_expression_option(c *Set_expression_optionContext)

	// EnterOpt_hint_value is called when entering the opt_hint_value production.
	EnterOpt_hint_value(c *Opt_hint_valueContext)

	// EnterLimit_clause is called when entering the limit_clause production.
	EnterLimit_clause(c *Limit_clauseContext)

	// EnterInto_clause is called when entering the into_clause production.
	EnterInto_clause(c *Into_clauseContext)

	// EnterInto_opt is called when entering the into_opt production.
	EnterInto_opt(c *Into_optContext)

	// EnterInto_var_list is called when entering the into_var_list production.
	EnterInto_var_list(c *Into_var_listContext)

	// EnterInto_var is called when entering the into_var production.
	EnterInto_var(c *Into_varContext)

	// EnterField_opt is called when entering the field_opt production.
	EnterField_opt(c *Field_optContext)

	// EnterField_term_list is called when entering the field_term_list production.
	EnterField_term_list(c *Field_term_listContext)

	// EnterField_term is called when entering the field_term production.
	EnterField_term(c *Field_termContext)

	// EnterLine_opt is called when entering the line_opt production.
	EnterLine_opt(c *Line_optContext)

	// EnterLine_term_list is called when entering the line_term_list production.
	EnterLine_term_list(c *Line_term_listContext)

	// EnterLine_term is called when entering the line_term production.
	EnterLine_term(c *Line_termContext)

	// EnterHint_list_with_end is called when entering the hint_list_with_end production.
	EnterHint_list_with_end(c *Hint_list_with_endContext)

	// EnterOpt_hint_list is called when entering the opt_hint_list production.
	EnterOpt_hint_list(c *Opt_hint_listContext)

	// EnterHint_options is called when entering the hint_options production.
	EnterHint_options(c *Hint_optionsContext)

	// EnterName_list is called when entering the name_list production.
	EnterName_list(c *Name_listContext)

	// EnterHint_option is called when entering the hint_option production.
	EnterHint_option(c *Hint_optionContext)

	// EnterConsistency_level is called when entering the consistency_level production.
	EnterConsistency_level(c *Consistency_levelContext)

	// EnterUse_plan_cache_type is called when entering the use_plan_cache_type production.
	EnterUse_plan_cache_type(c *Use_plan_cache_typeContext)

	// EnterUse_jit_type is called when entering the use_jit_type production.
	EnterUse_jit_type(c *Use_jit_typeContext)

	// EnterDistribute_method is called when entering the distribute_method production.
	EnterDistribute_method(c *Distribute_methodContext)

	// EnterLimit_expr is called when entering the limit_expr production.
	EnterLimit_expr(c *Limit_exprContext)

	// EnterOpt_for_update_wait is called when entering the opt_for_update_wait production.
	EnterOpt_for_update_wait(c *Opt_for_update_waitContext)

	// EnterParameterized_trim is called when entering the parameterized_trim production.
	EnterParameterized_trim(c *Parameterized_trimContext)

	// EnterGroupby_clause is called when entering the groupby_clause production.
	EnterGroupby_clause(c *Groupby_clauseContext)

	// EnterSort_list_for_group_by is called when entering the sort_list_for_group_by production.
	EnterSort_list_for_group_by(c *Sort_list_for_group_byContext)

	// EnterSort_key_for_group_by is called when entering the sort_key_for_group_by production.
	EnterSort_key_for_group_by(c *Sort_key_for_group_byContext)

	// EnterOrder_by is called when entering the order_by production.
	EnterOrder_by(c *Order_byContext)

	// EnterSort_list is called when entering the sort_list production.
	EnterSort_list(c *Sort_listContext)

	// EnterSort_key is called when entering the sort_key production.
	EnterSort_key(c *Sort_keyContext)

	// EnterQuery_expression_option_list is called when entering the query_expression_option_list production.
	EnterQuery_expression_option_list(c *Query_expression_option_listContext)

	// EnterQuery_expression_option is called when entering the query_expression_option production.
	EnterQuery_expression_option(c *Query_expression_optionContext)

	// EnterProjection is called when entering the projection production.
	EnterProjection(c *ProjectionContext)

	// EnterSelect_expr_list is called when entering the select_expr_list production.
	EnterSelect_expr_list(c *Select_expr_listContext)

	// EnterFrom_list is called when entering the from_list production.
	EnterFrom_list(c *From_listContext)

	// EnterTable_references is called when entering the table_references production.
	EnterTable_references(c *Table_referencesContext)

	// EnterTable_reference is called when entering the table_reference production.
	EnterTable_reference(c *Table_referenceContext)

	// EnterTable_factor is called when entering the table_factor production.
	EnterTable_factor(c *Table_factorContext)

	// EnterTbl_name is called when entering the tbl_name production.
	EnterTbl_name(c *Tbl_nameContext)

	// EnterDml_table_name is called when entering the dml_table_name production.
	EnterDml_table_name(c *Dml_table_nameContext)

	// EnterSeed is called when entering the seed production.
	EnterSeed(c *SeedContext)

	// EnterOpt_seed is called when entering the opt_seed production.
	EnterOpt_seed(c *Opt_seedContext)

	// EnterSample_percent is called when entering the sample_percent production.
	EnterSample_percent(c *Sample_percentContext)

	// EnterSample_clause is called when entering the sample_clause production.
	EnterSample_clause(c *Sample_clauseContext)

	// EnterTable_subquery is called when entering the table_subquery production.
	EnterTable_subquery(c *Table_subqueryContext)

	// EnterUse_partition is called when entering the use_partition production.
	EnterUse_partition(c *Use_partitionContext)

	// EnterIndex_hint_type is called when entering the index_hint_type production.
	EnterIndex_hint_type(c *Index_hint_typeContext)

	// EnterKey_or_index is called when entering the key_or_index production.
	EnterKey_or_index(c *Key_or_indexContext)

	// EnterIndex_hint_scope is called when entering the index_hint_scope production.
	EnterIndex_hint_scope(c *Index_hint_scopeContext)

	// EnterIndex_element is called when entering the index_element production.
	EnterIndex_element(c *Index_elementContext)

	// EnterIndex_list is called when entering the index_list production.
	EnterIndex_list(c *Index_listContext)

	// EnterIndex_hint_definition is called when entering the index_hint_definition production.
	EnterIndex_hint_definition(c *Index_hint_definitionContext)

	// EnterIndex_hint_list is called when entering the index_hint_list production.
	EnterIndex_hint_list(c *Index_hint_listContext)

	// EnterRelation_factor is called when entering the relation_factor production.
	EnterRelation_factor(c *Relation_factorContext)

	// EnterRelation_with_star_list is called when entering the relation_with_star_list production.
	EnterRelation_with_star_list(c *Relation_with_star_listContext)

	// EnterRelation_factor_with_star is called when entering the relation_factor_with_star production.
	EnterRelation_factor_with_star(c *Relation_factor_with_starContext)

	// EnterNormal_relation_factor is called when entering the normal_relation_factor production.
	EnterNormal_relation_factor(c *Normal_relation_factorContext)

	// EnterDot_relation_factor is called when entering the dot_relation_factor production.
	EnterDot_relation_factor(c *Dot_relation_factorContext)

	// EnterRelation_factor_in_hint is called when entering the relation_factor_in_hint production.
	EnterRelation_factor_in_hint(c *Relation_factor_in_hintContext)

	// EnterQb_name_option is called when entering the qb_name_option production.
	EnterQb_name_option(c *Qb_name_optionContext)

	// EnterRelation_factor_in_hint_list is called when entering the relation_factor_in_hint_list production.
	EnterRelation_factor_in_hint_list(c *Relation_factor_in_hint_listContext)

	// EnterRelation_sep_option is called when entering the relation_sep_option production.
	EnterRelation_sep_option(c *Relation_sep_optionContext)

	// EnterRelation_factor_in_pq_hint is called when entering the relation_factor_in_pq_hint production.
	EnterRelation_factor_in_pq_hint(c *Relation_factor_in_pq_hintContext)

	// EnterRelation_factor_in_leading_hint is called when entering the relation_factor_in_leading_hint production.
	EnterRelation_factor_in_leading_hint(c *Relation_factor_in_leading_hintContext)

	// EnterRelation_factor_in_leading_hint_list is called when entering the relation_factor_in_leading_hint_list production.
	EnterRelation_factor_in_leading_hint_list(c *Relation_factor_in_leading_hint_listContext)

	// EnterRelation_factor_in_leading_hint_list_entry is called when entering the relation_factor_in_leading_hint_list_entry production.
	EnterRelation_factor_in_leading_hint_list_entry(c *Relation_factor_in_leading_hint_list_entryContext)

	// EnterRelation_factor_in_use_join_hint_list is called when entering the relation_factor_in_use_join_hint_list production.
	EnterRelation_factor_in_use_join_hint_list(c *Relation_factor_in_use_join_hint_listContext)

	// EnterTracing_num_list is called when entering the tracing_num_list production.
	EnterTracing_num_list(c *Tracing_num_listContext)

	// EnterJoin_condition is called when entering the join_condition production.
	EnterJoin_condition(c *Join_conditionContext)

	// EnterJoined_table is called when entering the joined_table production.
	EnterJoined_table(c *Joined_tableContext)

	// EnterNatural_join_type is called when entering the natural_join_type production.
	EnterNatural_join_type(c *Natural_join_typeContext)

	// EnterInner_join_type is called when entering the inner_join_type production.
	EnterInner_join_type(c *Inner_join_typeContext)

	// EnterOuter_join_type is called when entering the outer_join_type production.
	EnterOuter_join_type(c *Outer_join_typeContext)

	// EnterAnalyze_stmt is called when entering the analyze_stmt production.
	EnterAnalyze_stmt(c *Analyze_stmtContext)

	// EnterCreate_outline_stmt is called when entering the create_outline_stmt production.
	EnterCreate_outline_stmt(c *Create_outline_stmtContext)

	// EnterAlter_outline_stmt is called when entering the alter_outline_stmt production.
	EnterAlter_outline_stmt(c *Alter_outline_stmtContext)

	// EnterDrop_outline_stmt is called when entering the drop_outline_stmt production.
	EnterDrop_outline_stmt(c *Drop_outline_stmtContext)

	// EnterExplain_stmt is called when entering the explain_stmt production.
	EnterExplain_stmt(c *Explain_stmtContext)

	// EnterExplain_or_desc is called when entering the explain_or_desc production.
	EnterExplain_or_desc(c *Explain_or_descContext)

	// EnterExplainable_stmt is called when entering the explainable_stmt production.
	EnterExplainable_stmt(c *Explainable_stmtContext)

	// EnterFormat_name is called when entering the format_name production.
	EnterFormat_name(c *Format_nameContext)

	// EnterShow_stmt is called when entering the show_stmt production.
	EnterShow_stmt(c *Show_stmtContext)

	// EnterDatabases_or_schemas is called when entering the databases_or_schemas production.
	EnterDatabases_or_schemas(c *Databases_or_schemasContext)

	// EnterOpt_for_grant_user is called when entering the opt_for_grant_user production.
	EnterOpt_for_grant_user(c *Opt_for_grant_userContext)

	// EnterColumns_or_fields is called when entering the columns_or_fields production.
	EnterColumns_or_fields(c *Columns_or_fieldsContext)

	// EnterDatabase_or_schema is called when entering the database_or_schema production.
	EnterDatabase_or_schema(c *Database_or_schemaContext)

	// EnterIndex_or_indexes_or_keys is called when entering the index_or_indexes_or_keys production.
	EnterIndex_or_indexes_or_keys(c *Index_or_indexes_or_keysContext)

	// EnterFrom_or_in is called when entering the from_or_in production.
	EnterFrom_or_in(c *From_or_inContext)

	// EnterHelp_stmt is called when entering the help_stmt production.
	EnterHelp_stmt(c *Help_stmtContext)

	// EnterCreate_tablespace_stmt is called when entering the create_tablespace_stmt production.
	EnterCreate_tablespace_stmt(c *Create_tablespace_stmtContext)

	// EnterPermanent_tablespace is called when entering the permanent_tablespace production.
	EnterPermanent_tablespace(c *Permanent_tablespaceContext)

	// EnterPermanent_tablespace_option is called when entering the permanent_tablespace_option production.
	EnterPermanent_tablespace_option(c *Permanent_tablespace_optionContext)

	// EnterDrop_tablespace_stmt is called when entering the drop_tablespace_stmt production.
	EnterDrop_tablespace_stmt(c *Drop_tablespace_stmtContext)

	// EnterAlter_tablespace_actions is called when entering the alter_tablespace_actions production.
	EnterAlter_tablespace_actions(c *Alter_tablespace_actionsContext)

	// EnterAlter_tablespace_action is called when entering the alter_tablespace_action production.
	EnterAlter_tablespace_action(c *Alter_tablespace_actionContext)

	// EnterAlter_tablespace_stmt is called when entering the alter_tablespace_stmt production.
	EnterAlter_tablespace_stmt(c *Alter_tablespace_stmtContext)

	// EnterRotate_master_key_stmt is called when entering the rotate_master_key_stmt production.
	EnterRotate_master_key_stmt(c *Rotate_master_key_stmtContext)

	// EnterPermanent_tablespace_options is called when entering the permanent_tablespace_options production.
	EnterPermanent_tablespace_options(c *Permanent_tablespace_optionsContext)

	// EnterCreate_user_stmt is called when entering the create_user_stmt production.
	EnterCreate_user_stmt(c *Create_user_stmtContext)

	// EnterUser_specification_list is called when entering the user_specification_list production.
	EnterUser_specification_list(c *User_specification_listContext)

	// EnterUser_specification is called when entering the user_specification production.
	EnterUser_specification(c *User_specificationContext)

	// EnterRequire_specification is called when entering the require_specification production.
	EnterRequire_specification(c *Require_specificationContext)

	// EnterTls_option_list is called when entering the tls_option_list production.
	EnterTls_option_list(c *Tls_option_listContext)

	// EnterTls_option is called when entering the tls_option production.
	EnterTls_option(c *Tls_optionContext)

	// EnterUser is called when entering the user production.
	EnterUser(c *UserContext)

	// EnterOpt_host_name is called when entering the opt_host_name production.
	EnterOpt_host_name(c *Opt_host_nameContext)

	// EnterUser_with_host_name is called when entering the user_with_host_name production.
	EnterUser_with_host_name(c *User_with_host_nameContext)

	// EnterPassword is called when entering the password production.
	EnterPassword(c *PasswordContext)

	// EnterDrop_user_stmt is called when entering the drop_user_stmt production.
	EnterDrop_user_stmt(c *Drop_user_stmtContext)

	// EnterUser_list is called when entering the user_list production.
	EnterUser_list(c *User_listContext)

	// EnterSet_password_stmt is called when entering the set_password_stmt production.
	EnterSet_password_stmt(c *Set_password_stmtContext)

	// EnterOpt_for_user is called when entering the opt_for_user production.
	EnterOpt_for_user(c *Opt_for_userContext)

	// EnterRename_user_stmt is called when entering the rename_user_stmt production.
	EnterRename_user_stmt(c *Rename_user_stmtContext)

	// EnterRename_info is called when entering the rename_info production.
	EnterRename_info(c *Rename_infoContext)

	// EnterRename_list is called when entering the rename_list production.
	EnterRename_list(c *Rename_listContext)

	// EnterLock_user_stmt is called when entering the lock_user_stmt production.
	EnterLock_user_stmt(c *Lock_user_stmtContext)

	// EnterLock_spec_mysql57 is called when entering the lock_spec_mysql57 production.
	EnterLock_spec_mysql57(c *Lock_spec_mysql57Context)

	// EnterLock_tables_stmt is called when entering the lock_tables_stmt production.
	EnterLock_tables_stmt(c *Lock_tables_stmtContext)

	// EnterUnlock_tables_stmt is called when entering the unlock_tables_stmt production.
	EnterUnlock_tables_stmt(c *Unlock_tables_stmtContext)

	// EnterLock_table_list is called when entering the lock_table_list production.
	EnterLock_table_list(c *Lock_table_listContext)

	// EnterLock_table is called when entering the lock_table production.
	EnterLock_table(c *Lock_tableContext)

	// EnterLock_type is called when entering the lock_type production.
	EnterLock_type(c *Lock_typeContext)

	// EnterBegin_stmt is called when entering the begin_stmt production.
	EnterBegin_stmt(c *Begin_stmtContext)

	// EnterCommit_stmt is called when entering the commit_stmt production.
	EnterCommit_stmt(c *Commit_stmtContext)

	// EnterRollback_stmt is called when entering the rollback_stmt production.
	EnterRollback_stmt(c *Rollback_stmtContext)

	// EnterKill_stmt is called when entering the kill_stmt production.
	EnterKill_stmt(c *Kill_stmtContext)

	// EnterGrant_stmt is called when entering the grant_stmt production.
	EnterGrant_stmt(c *Grant_stmtContext)

	// EnterGrant_privileges is called when entering the grant_privileges production.
	EnterGrant_privileges(c *Grant_privilegesContext)

	// EnterPriv_type_list is called when entering the priv_type_list production.
	EnterPriv_type_list(c *Priv_type_listContext)

	// EnterPriv_type is called when entering the priv_type production.
	EnterPriv_type(c *Priv_typeContext)

	// EnterPriv_level is called when entering the priv_level production.
	EnterPriv_level(c *Priv_levelContext)

	// EnterGrant_options is called when entering the grant_options production.
	EnterGrant_options(c *Grant_optionsContext)

	// EnterRevoke_stmt is called when entering the revoke_stmt production.
	EnterRevoke_stmt(c *Revoke_stmtContext)

	// EnterPrepare_stmt is called when entering the prepare_stmt production.
	EnterPrepare_stmt(c *Prepare_stmtContext)

	// EnterStmt_name is called when entering the stmt_name production.
	EnterStmt_name(c *Stmt_nameContext)

	// EnterPreparable_stmt is called when entering the preparable_stmt production.
	EnterPreparable_stmt(c *Preparable_stmtContext)

	// EnterVariable_set_stmt is called when entering the variable_set_stmt production.
	EnterVariable_set_stmt(c *Variable_set_stmtContext)

	// EnterSys_var_and_val_list is called when entering the sys_var_and_val_list production.
	EnterSys_var_and_val_list(c *Sys_var_and_val_listContext)

	// EnterVar_and_val_list is called when entering the var_and_val_list production.
	EnterVar_and_val_list(c *Var_and_val_listContext)

	// EnterSet_expr_or_default is called when entering the set_expr_or_default production.
	EnterSet_expr_or_default(c *Set_expr_or_defaultContext)

	// EnterVar_and_val is called when entering the var_and_val production.
	EnterVar_and_val(c *Var_and_valContext)

	// EnterSys_var_and_val is called when entering the sys_var_and_val production.
	EnterSys_var_and_val(c *Sys_var_and_valContext)

	// EnterScope_or_scope_alias is called when entering the scope_or_scope_alias production.
	EnterScope_or_scope_alias(c *Scope_or_scope_aliasContext)

	// EnterTo_or_eq is called when entering the to_or_eq production.
	EnterTo_or_eq(c *To_or_eqContext)

	// EnterExecute_stmt is called when entering the execute_stmt production.
	EnterExecute_stmt(c *Execute_stmtContext)

	// EnterArgument_list is called when entering the argument_list production.
	EnterArgument_list(c *Argument_listContext)

	// EnterArgument is called when entering the argument production.
	EnterArgument(c *ArgumentContext)

	// EnterDeallocate_prepare_stmt is called when entering the deallocate_prepare_stmt production.
	EnterDeallocate_prepare_stmt(c *Deallocate_prepare_stmtContext)

	// EnterDeallocate_or_drop is called when entering the deallocate_or_drop production.
	EnterDeallocate_or_drop(c *Deallocate_or_dropContext)

	// EnterTruncate_table_stmt is called when entering the truncate_table_stmt production.
	EnterTruncate_table_stmt(c *Truncate_table_stmtContext)

	// EnterRename_table_stmt is called when entering the rename_table_stmt production.
	EnterRename_table_stmt(c *Rename_table_stmtContext)

	// EnterRename_table_actions is called when entering the rename_table_actions production.
	EnterRename_table_actions(c *Rename_table_actionsContext)

	// EnterRename_table_action is called when entering the rename_table_action production.
	EnterRename_table_action(c *Rename_table_actionContext)

	// EnterAlter_table_stmt is called when entering the alter_table_stmt production.
	EnterAlter_table_stmt(c *Alter_table_stmtContext)

	// EnterAlter_table_actions is called when entering the alter_table_actions production.
	EnterAlter_table_actions(c *Alter_table_actionsContext)

	// EnterAlter_table_action is called when entering the alter_table_action production.
	EnterAlter_table_action(c *Alter_table_actionContext)

	// EnterAlter_constraint_option is called when entering the alter_constraint_option production.
	EnterAlter_constraint_option(c *Alter_constraint_optionContext)

	// EnterAlter_partition_option is called when entering the alter_partition_option production.
	EnterAlter_partition_option(c *Alter_partition_optionContext)

	// EnterOpt_partition_range_or_list is called when entering the opt_partition_range_or_list production.
	EnterOpt_partition_range_or_list(c *Opt_partition_range_or_listContext)

	// EnterAlter_tg_partition_option is called when entering the alter_tg_partition_option production.
	EnterAlter_tg_partition_option(c *Alter_tg_partition_optionContext)

	// EnterDrop_partition_name_list is called when entering the drop_partition_name_list production.
	EnterDrop_partition_name_list(c *Drop_partition_name_listContext)

	// EnterModify_partition_info is called when entering the modify_partition_info production.
	EnterModify_partition_info(c *Modify_partition_infoContext)

	// EnterModify_tg_partition_info is called when entering the modify_tg_partition_info production.
	EnterModify_tg_partition_info(c *Modify_tg_partition_infoContext)

	// EnterAlter_index_option is called when entering the alter_index_option production.
	EnterAlter_index_option(c *Alter_index_optionContext)

	// EnterAlter_foreign_key_action is called when entering the alter_foreign_key_action production.
	EnterAlter_foreign_key_action(c *Alter_foreign_key_actionContext)

	// EnterVisibility_option is called when entering the visibility_option production.
	EnterVisibility_option(c *Visibility_optionContext)

	// EnterAlter_column_option is called when entering the alter_column_option production.
	EnterAlter_column_option(c *Alter_column_optionContext)

	// EnterAlter_tablegroup_option is called when entering the alter_tablegroup_option production.
	EnterAlter_tablegroup_option(c *Alter_tablegroup_optionContext)

	// EnterAlter_column_behavior is called when entering the alter_column_behavior production.
	EnterAlter_column_behavior(c *Alter_column_behaviorContext)

	// EnterFlashback_stmt is called when entering the flashback_stmt production.
	EnterFlashback_stmt(c *Flashback_stmtContext)

	// EnterPurge_stmt is called when entering the purge_stmt production.
	EnterPurge_stmt(c *Purge_stmtContext)

	// EnterOptimize_stmt is called when entering the optimize_stmt production.
	EnterOptimize_stmt(c *Optimize_stmtContext)

	// EnterDump_memory_stmt is called when entering the dump_memory_stmt production.
	EnterDump_memory_stmt(c *Dump_memory_stmtContext)

	// EnterAlter_system_stmt is called when entering the alter_system_stmt production.
	EnterAlter_system_stmt(c *Alter_system_stmtContext)

	// EnterChange_tenant_name_or_tenant_id is called when entering the change_tenant_name_or_tenant_id production.
	EnterChange_tenant_name_or_tenant_id(c *Change_tenant_name_or_tenant_idContext)

	// EnterCache_type is called when entering the cache_type production.
	EnterCache_type(c *Cache_typeContext)

	// EnterBalance_task_type is called when entering the balance_task_type production.
	EnterBalance_task_type(c *Balance_task_typeContext)

	// EnterTenant_list_tuple is called when entering the tenant_list_tuple production.
	EnterTenant_list_tuple(c *Tenant_list_tupleContext)

	// EnterTenant_name_list is called when entering the tenant_name_list production.
	EnterTenant_name_list(c *Tenant_name_listContext)

	// EnterFlush_scope is called when entering the flush_scope production.
	EnterFlush_scope(c *Flush_scopeContext)

	// EnterServer_info_list is called when entering the server_info_list production.
	EnterServer_info_list(c *Server_info_listContext)

	// EnterServer_info is called when entering the server_info production.
	EnterServer_info(c *Server_infoContext)

	// EnterServer_action is called when entering the server_action production.
	EnterServer_action(c *Server_actionContext)

	// EnterServer_list is called when entering the server_list production.
	EnterServer_list(c *Server_listContext)

	// EnterZone_action is called when entering the zone_action production.
	EnterZone_action(c *Zone_actionContext)

	// EnterIp_port is called when entering the ip_port production.
	EnterIp_port(c *Ip_portContext)

	// EnterZone_desc is called when entering the zone_desc production.
	EnterZone_desc(c *Zone_descContext)

	// EnterServer_or_zone is called when entering the server_or_zone production.
	EnterServer_or_zone(c *Server_or_zoneContext)

	// EnterAdd_or_alter_zone_option is called when entering the add_or_alter_zone_option production.
	EnterAdd_or_alter_zone_option(c *Add_or_alter_zone_optionContext)

	// EnterAdd_or_alter_zone_options is called when entering the add_or_alter_zone_options production.
	EnterAdd_or_alter_zone_options(c *Add_or_alter_zone_optionsContext)

	// EnterAlter_or_change_or_modify is called when entering the alter_or_change_or_modify production.
	EnterAlter_or_change_or_modify(c *Alter_or_change_or_modifyContext)

	// EnterPartition_id_desc is called when entering the partition_id_desc production.
	EnterPartition_id_desc(c *Partition_id_descContext)

	// EnterPartition_id_or_server_or_zone is called when entering the partition_id_or_server_or_zone production.
	EnterPartition_id_or_server_or_zone(c *Partition_id_or_server_or_zoneContext)

	// EnterMigrate_action is called when entering the migrate_action production.
	EnterMigrate_action(c *Migrate_actionContext)

	// EnterChange_actions is called when entering the change_actions production.
	EnterChange_actions(c *Change_actionsContext)

	// EnterChange_action is called when entering the change_action production.
	EnterChange_action(c *Change_actionContext)

	// EnterReplica_type is called when entering the replica_type production.
	EnterReplica_type(c *Replica_typeContext)

	// EnterSuspend_or_resume is called when entering the suspend_or_resume production.
	EnterSuspend_or_resume(c *Suspend_or_resumeContext)

	// EnterBaseline_id_expr is called when entering the baseline_id_expr production.
	EnterBaseline_id_expr(c *Baseline_id_exprContext)

	// EnterSql_id_expr is called when entering the sql_id_expr production.
	EnterSql_id_expr(c *Sql_id_exprContext)

	// EnterBaseline_asgn_factor is called when entering the baseline_asgn_factor production.
	EnterBaseline_asgn_factor(c *Baseline_asgn_factorContext)

	// EnterTenant_name is called when entering the tenant_name production.
	EnterTenant_name(c *Tenant_nameContext)

	// EnterCache_name is called when entering the cache_name production.
	EnterCache_name(c *Cache_nameContext)

	// EnterFile_id is called when entering the file_id production.
	EnterFile_id(c *File_idContext)

	// EnterCancel_task_type is called when entering the cancel_task_type production.
	EnterCancel_task_type(c *Cancel_task_typeContext)

	// EnterAlter_system_set_parameter_actions is called when entering the alter_system_set_parameter_actions production.
	EnterAlter_system_set_parameter_actions(c *Alter_system_set_parameter_actionsContext)

	// EnterAlter_system_set_parameter_action is called when entering the alter_system_set_parameter_action production.
	EnterAlter_system_set_parameter_action(c *Alter_system_set_parameter_actionContext)

	// EnterAlter_system_settp_actions is called when entering the alter_system_settp_actions production.
	EnterAlter_system_settp_actions(c *Alter_system_settp_actionsContext)

	// EnterSettp_option is called when entering the settp_option production.
	EnterSettp_option(c *Settp_optionContext)

	// EnterCluster_role is called when entering the cluster_role production.
	EnterCluster_role(c *Cluster_roleContext)

	// EnterPartition_role is called when entering the partition_role production.
	EnterPartition_role(c *Partition_roleContext)

	// EnterUpgrade_action is called when entering the upgrade_action production.
	EnterUpgrade_action(c *Upgrade_actionContext)

	// EnterSet_names_stmt is called when entering the set_names_stmt production.
	EnterSet_names_stmt(c *Set_names_stmtContext)

	// EnterSet_charset_stmt is called when entering the set_charset_stmt production.
	EnterSet_charset_stmt(c *Set_charset_stmtContext)

	// EnterSet_transaction_stmt is called when entering the set_transaction_stmt production.
	EnterSet_transaction_stmt(c *Set_transaction_stmtContext)

	// EnterTransaction_characteristics is called when entering the transaction_characteristics production.
	EnterTransaction_characteristics(c *Transaction_characteristicsContext)

	// EnterTransaction_access_mode is called when entering the transaction_access_mode production.
	EnterTransaction_access_mode(c *Transaction_access_modeContext)

	// EnterIsolation_level is called when entering the isolation_level production.
	EnterIsolation_level(c *Isolation_levelContext)

	// EnterCreate_savepoint_stmt is called when entering the create_savepoint_stmt production.
	EnterCreate_savepoint_stmt(c *Create_savepoint_stmtContext)

	// EnterRollback_savepoint_stmt is called when entering the rollback_savepoint_stmt production.
	EnterRollback_savepoint_stmt(c *Rollback_savepoint_stmtContext)

	// EnterRelease_savepoint_stmt is called when entering the release_savepoint_stmt production.
	EnterRelease_savepoint_stmt(c *Release_savepoint_stmtContext)

	// EnterAlter_cluster_stmt is called when entering the alter_cluster_stmt production.
	EnterAlter_cluster_stmt(c *Alter_cluster_stmtContext)

	// EnterCluster_action is called when entering the cluster_action production.
	EnterCluster_action(c *Cluster_actionContext)

	// EnterSwitchover_cluster_stmt is called when entering the switchover_cluster_stmt production.
	EnterSwitchover_cluster_stmt(c *Switchover_cluster_stmtContext)

	// EnterCommit_switchover_clause is called when entering the commit_switchover_clause production.
	EnterCommit_switchover_clause(c *Commit_switchover_clauseContext)

	// EnterCluster_name is called when entering the cluster_name production.
	EnterCluster_name(c *Cluster_nameContext)

	// EnterVar_name is called when entering the var_name production.
	EnterVar_name(c *Var_nameContext)

	// EnterColumn_name is called when entering the column_name production.
	EnterColumn_name(c *Column_nameContext)

	// EnterRelation_name is called when entering the relation_name production.
	EnterRelation_name(c *Relation_nameContext)

	// EnterFunction_name is called when entering the function_name production.
	EnterFunction_name(c *Function_nameContext)

	// EnterColumn_label is called when entering the column_label production.
	EnterColumn_label(c *Column_labelContext)

	// EnterDate_unit is called when entering the date_unit production.
	EnterDate_unit(c *Date_unitContext)

	// EnterUnreserved_keyword is called when entering the unreserved_keyword production.
	EnterUnreserved_keyword(c *Unreserved_keywordContext)

	// EnterUnreserved_keyword_normal is called when entering the unreserved_keyword_normal production.
	EnterUnreserved_keyword_normal(c *Unreserved_keyword_normalContext)

	// EnterUnreserved_keyword_special is called when entering the unreserved_keyword_special production.
	EnterUnreserved_keyword_special(c *Unreserved_keyword_specialContext)

	// EnterEmpty is called when entering the empty production.
	EnterEmpty(c *EmptyContext)

	// EnterForward_expr is called when entering the forward_expr production.
	EnterForward_expr(c *Forward_exprContext)

	// EnterForward_sql_stmt is called when entering the forward_sql_stmt production.
	EnterForward_sql_stmt(c *Forward_sql_stmtContext)

	// ExitSql_stmt is called when exiting the sql_stmt production.
	ExitSql_stmt(c *Sql_stmtContext)

	// ExitStmt_list is called when exiting the stmt_list production.
	ExitStmt_list(c *Stmt_listContext)

	// ExitStmt is called when exiting the stmt production.
	ExitStmt(c *StmtContext)

	// ExitExpr_list is called when exiting the expr_list production.
	ExitExpr_list(c *Expr_listContext)

	// ExitExpr_as_list is called when exiting the expr_as_list production.
	ExitExpr_as_list(c *Expr_as_listContext)

	// ExitExpr_with_opt_alias is called when exiting the expr_with_opt_alias production.
	ExitExpr_with_opt_alias(c *Expr_with_opt_aliasContext)

	// ExitColumn_ref is called when exiting the column_ref production.
	ExitColumn_ref(c *Column_refContext)

	// ExitComplex_string_literal is called when exiting the complex_string_literal production.
	ExitComplex_string_literal(c *Complex_string_literalContext)

	// ExitCharset_introducer is called when exiting the charset_introducer production.
	ExitCharset_introducer(c *Charset_introducerContext)

	// ExitLiteral is called when exiting the literal production.
	ExitLiteral(c *LiteralContext)

	// ExitNumber_literal is called when exiting the number_literal production.
	ExitNumber_literal(c *Number_literalContext)

	// ExitExpr_const is called when exiting the expr_const production.
	ExitExpr_const(c *Expr_constContext)

	// ExitConf_const is called when exiting the conf_const production.
	ExitConf_const(c *Conf_constContext)

	// ExitGlobal_or_session_alias is called when exiting the global_or_session_alias production.
	ExitGlobal_or_session_alias(c *Global_or_session_aliasContext)

	// ExitBool_pri is called when exiting the bool_pri production.
	ExitBool_pri(c *Bool_priContext)

	// ExitPredicate is called when exiting the predicate production.
	ExitPredicate(c *PredicateContext)

	// ExitBit_expr is called when exiting the bit_expr production.
	ExitBit_expr(c *Bit_exprContext)

	// ExitSimple_expr is called when exiting the simple_expr production.
	ExitSimple_expr(c *Simple_exprContext)

	// ExitExpr is called when exiting the expr production.
	ExitExpr(c *ExprContext)

	// ExitNot is called when exiting the not production.
	ExitNot(c *NotContext)

	// ExitNot2 is called when exiting the not2 production.
	ExitNot2(c *Not2Context)

	// ExitSub_query_flag is called when exiting the sub_query_flag production.
	ExitSub_query_flag(c *Sub_query_flagContext)

	// ExitIn_expr is called when exiting the in_expr production.
	ExitIn_expr(c *In_exprContext)

	// ExitCase_expr is called when exiting the case_expr production.
	ExitCase_expr(c *Case_exprContext)

	// ExitWindow_function is called when exiting the window_function production.
	ExitWindow_function(c *Window_functionContext)

	// ExitFirst_or_last is called when exiting the first_or_last production.
	ExitFirst_or_last(c *First_or_lastContext)

	// ExitRespect_or_ignore is called when exiting the respect_or_ignore production.
	ExitRespect_or_ignore(c *Respect_or_ignoreContext)

	// ExitWin_fun_first_last_params is called when exiting the win_fun_first_last_params production.
	ExitWin_fun_first_last_params(c *Win_fun_first_last_paramsContext)

	// ExitWin_fun_lead_lag_params is called when exiting the win_fun_lead_lag_params production.
	ExitWin_fun_lead_lag_params(c *Win_fun_lead_lag_paramsContext)

	// ExitNew_generalized_window_clause is called when exiting the new_generalized_window_clause production.
	ExitNew_generalized_window_clause(c *New_generalized_window_clauseContext)

	// ExitNew_generalized_window_clause_with_blanket is called when exiting the new_generalized_window_clause_with_blanket production.
	ExitNew_generalized_window_clause_with_blanket(c *New_generalized_window_clause_with_blanketContext)

	// ExitNamed_windows is called when exiting the named_windows production.
	ExitNamed_windows(c *Named_windowsContext)

	// ExitNamed_window is called when exiting the named_window production.
	ExitNamed_window(c *Named_windowContext)

	// ExitGeneralized_window_clause is called when exiting the generalized_window_clause production.
	ExitGeneralized_window_clause(c *Generalized_window_clauseContext)

	// ExitWin_rows_or_range is called when exiting the win_rows_or_range production.
	ExitWin_rows_or_range(c *Win_rows_or_rangeContext)

	// ExitWin_preceding_or_following is called when exiting the win_preceding_or_following production.
	ExitWin_preceding_or_following(c *Win_preceding_or_followingContext)

	// ExitWin_interval is called when exiting the win_interval production.
	ExitWin_interval(c *Win_intervalContext)

	// ExitWin_bounding is called when exiting the win_bounding production.
	ExitWin_bounding(c *Win_boundingContext)

	// ExitWin_window is called when exiting the win_window production.
	ExitWin_window(c *Win_windowContext)

	// ExitCase_arg is called when exiting the case_arg production.
	ExitCase_arg(c *Case_argContext)

	// ExitWhen_clause_list is called when exiting the when_clause_list production.
	ExitWhen_clause_list(c *When_clause_listContext)

	// ExitWhen_clause is called when exiting the when_clause production.
	ExitWhen_clause(c *When_clauseContext)

	// ExitCase_default is called when exiting the case_default production.
	ExitCase_default(c *Case_defaultContext)

	// ExitFunc_expr is called when exiting the func_expr production.
	ExitFunc_expr(c *Func_exprContext)

	// ExitSys_interval_func is called when exiting the sys_interval_func production.
	ExitSys_interval_func(c *Sys_interval_funcContext)

	// ExitUtc_timestamp_func is called when exiting the utc_timestamp_func production.
	ExitUtc_timestamp_func(c *Utc_timestamp_funcContext)

	// ExitSysdate_func is called when exiting the sysdate_func production.
	ExitSysdate_func(c *Sysdate_funcContext)

	// ExitCur_timestamp_func is called when exiting the cur_timestamp_func production.
	ExitCur_timestamp_func(c *Cur_timestamp_funcContext)

	// ExitNow_synonyms_func is called when exiting the now_synonyms_func production.
	ExitNow_synonyms_func(c *Now_synonyms_funcContext)

	// ExitCur_time_func is called when exiting the cur_time_func production.
	ExitCur_time_func(c *Cur_time_funcContext)

	// ExitCur_date_func is called when exiting the cur_date_func production.
	ExitCur_date_func(c *Cur_date_funcContext)

	// ExitSubstr_or_substring is called when exiting the substr_or_substring production.
	ExitSubstr_or_substring(c *Substr_or_substringContext)

	// ExitSubstr_params is called when exiting the substr_params production.
	ExitSubstr_params(c *Substr_paramsContext)

	// ExitDate_params is called when exiting the date_params production.
	ExitDate_params(c *Date_paramsContext)

	// ExitTimestamp_params is called when exiting the timestamp_params production.
	ExitTimestamp_params(c *Timestamp_paramsContext)

	// ExitDelete_stmt is called when exiting the delete_stmt production.
	ExitDelete_stmt(c *Delete_stmtContext)

	// ExitMulti_delete_table is called when exiting the multi_delete_table production.
	ExitMulti_delete_table(c *Multi_delete_tableContext)

	// ExitUpdate_stmt is called when exiting the update_stmt production.
	ExitUpdate_stmt(c *Update_stmtContext)

	// ExitUpdate_asgn_list is called when exiting the update_asgn_list production.
	ExitUpdate_asgn_list(c *Update_asgn_listContext)

	// ExitUpdate_asgn_factor is called when exiting the update_asgn_factor production.
	ExitUpdate_asgn_factor(c *Update_asgn_factorContext)

	// ExitCreate_resource_stmt is called when exiting the create_resource_stmt production.
	ExitCreate_resource_stmt(c *Create_resource_stmtContext)

	// ExitOpt_resource_unit_option_list is called when exiting the opt_resource_unit_option_list production.
	ExitOpt_resource_unit_option_list(c *Opt_resource_unit_option_listContext)

	// ExitResource_unit_option is called when exiting the resource_unit_option production.
	ExitResource_unit_option(c *Resource_unit_optionContext)

	// ExitOpt_create_resource_pool_option_list is called when exiting the opt_create_resource_pool_option_list production.
	ExitOpt_create_resource_pool_option_list(c *Opt_create_resource_pool_option_listContext)

	// ExitCreate_resource_pool_option is called when exiting the create_resource_pool_option production.
	ExitCreate_resource_pool_option(c *Create_resource_pool_optionContext)

	// ExitAlter_resource_pool_option_list is called when exiting the alter_resource_pool_option_list production.
	ExitAlter_resource_pool_option_list(c *Alter_resource_pool_option_listContext)

	// ExitUnit_id_list is called when exiting the unit_id_list production.
	ExitUnit_id_list(c *Unit_id_listContext)

	// ExitAlter_resource_pool_option is called when exiting the alter_resource_pool_option production.
	ExitAlter_resource_pool_option(c *Alter_resource_pool_optionContext)

	// ExitAlter_resource_stmt is called when exiting the alter_resource_stmt production.
	ExitAlter_resource_stmt(c *Alter_resource_stmtContext)

	// ExitDrop_resource_stmt is called when exiting the drop_resource_stmt production.
	ExitDrop_resource_stmt(c *Drop_resource_stmtContext)

	// ExitCreate_tenant_stmt is called when exiting the create_tenant_stmt production.
	ExitCreate_tenant_stmt(c *Create_tenant_stmtContext)

	// ExitOpt_tenant_option_list is called when exiting the opt_tenant_option_list production.
	ExitOpt_tenant_option_list(c *Opt_tenant_option_listContext)

	// ExitTenant_option is called when exiting the tenant_option production.
	ExitTenant_option(c *Tenant_optionContext)

	// ExitZone_list is called when exiting the zone_list production.
	ExitZone_list(c *Zone_listContext)

	// ExitResource_pool_list is called when exiting the resource_pool_list production.
	ExitResource_pool_list(c *Resource_pool_listContext)

	// ExitAlter_tenant_stmt is called when exiting the alter_tenant_stmt production.
	ExitAlter_tenant_stmt(c *Alter_tenant_stmtContext)

	// ExitDrop_tenant_stmt is called when exiting the drop_tenant_stmt production.
	ExitDrop_tenant_stmt(c *Drop_tenant_stmtContext)

	// ExitCreate_database_stmt is called when exiting the create_database_stmt production.
	ExitCreate_database_stmt(c *Create_database_stmtContext)

	// ExitDatabase_key is called when exiting the database_key production.
	ExitDatabase_key(c *Database_keyContext)

	// ExitDatabase_factor is called when exiting the database_factor production.
	ExitDatabase_factor(c *Database_factorContext)

	// ExitDatabase_option_list is called when exiting the database_option_list production.
	ExitDatabase_option_list(c *Database_option_listContext)

	// ExitCharset_key is called when exiting the charset_key production.
	ExitCharset_key(c *Charset_keyContext)

	// ExitDatabase_option is called when exiting the database_option production.
	ExitDatabase_option(c *Database_optionContext)

	// ExitRead_only_or_write is called when exiting the read_only_or_write production.
	ExitRead_only_or_write(c *Read_only_or_writeContext)

	// ExitDrop_database_stmt is called when exiting the drop_database_stmt production.
	ExitDrop_database_stmt(c *Drop_database_stmtContext)

	// ExitAlter_database_stmt is called when exiting the alter_database_stmt production.
	ExitAlter_database_stmt(c *Alter_database_stmtContext)

	// ExitLoad_data_stmt is called when exiting the load_data_stmt production.
	ExitLoad_data_stmt(c *Load_data_stmtContext)

	// ExitLoad_data_with_opt_hint is called when exiting the load_data_with_opt_hint production.
	ExitLoad_data_with_opt_hint(c *Load_data_with_opt_hintContext)

	// ExitLines_or_rows is called when exiting the lines_or_rows production.
	ExitLines_or_rows(c *Lines_or_rowsContext)

	// ExitField_or_vars_list is called when exiting the field_or_vars_list production.
	ExitField_or_vars_list(c *Field_or_vars_listContext)

	// ExitField_or_vars is called when exiting the field_or_vars production.
	ExitField_or_vars(c *Field_or_varsContext)

	// ExitLoad_set_list is called when exiting the load_set_list production.
	ExitLoad_set_list(c *Load_set_listContext)

	// ExitLoad_set_element is called when exiting the load_set_element production.
	ExitLoad_set_element(c *Load_set_elementContext)

	// ExitUse_database_stmt is called when exiting the use_database_stmt production.
	ExitUse_database_stmt(c *Use_database_stmtContext)

	// ExitCreate_synonym_stmt is called when exiting the create_synonym_stmt production.
	ExitCreate_synonym_stmt(c *Create_synonym_stmtContext)

	// ExitSynonym_name is called when exiting the synonym_name production.
	ExitSynonym_name(c *Synonym_nameContext)

	// ExitSynonym_object is called when exiting the synonym_object production.
	ExitSynonym_object(c *Synonym_objectContext)

	// ExitDrop_synonym_stmt is called when exiting the drop_synonym_stmt production.
	ExitDrop_synonym_stmt(c *Drop_synonym_stmtContext)

	// ExitTemporary_option is called when exiting the temporary_option production.
	ExitTemporary_option(c *Temporary_optionContext)

	// ExitCreate_table_like_stmt is called when exiting the create_table_like_stmt production.
	ExitCreate_table_like_stmt(c *Create_table_like_stmtContext)

	// ExitCreate_table_stmt is called when exiting the create_table_stmt production.
	ExitCreate_table_stmt(c *Create_table_stmtContext)

	// ExitRet_type is called when exiting the ret_type production.
	ExitRet_type(c *Ret_typeContext)

	// ExitCreate_function_stmt is called when exiting the create_function_stmt production.
	ExitCreate_function_stmt(c *Create_function_stmtContext)

	// ExitDrop_function_stmt is called when exiting the drop_function_stmt production.
	ExitDrop_function_stmt(c *Drop_function_stmtContext)

	// ExitTable_element_list is called when exiting the table_element_list production.
	ExitTable_element_list(c *Table_element_listContext)

	// ExitTable_element is called when exiting the table_element production.
	ExitTable_element(c *Table_elementContext)

	// ExitOpt_reference_option_list is called when exiting the opt_reference_option_list production.
	ExitOpt_reference_option_list(c *Opt_reference_option_listContext)

	// ExitReference_option is called when exiting the reference_option production.
	ExitReference_option(c *Reference_optionContext)

	// ExitReference_action is called when exiting the reference_action production.
	ExitReference_action(c *Reference_actionContext)

	// ExitMatch_action is called when exiting the match_action production.
	ExitMatch_action(c *Match_actionContext)

	// ExitColumn_definition is called when exiting the column_definition production.
	ExitColumn_definition(c *Column_definitionContext)

	// ExitOpt_generated_column_attribute_list is called when exiting the opt_generated_column_attribute_list production.
	ExitOpt_generated_column_attribute_list(c *Opt_generated_column_attribute_listContext)

	// ExitGenerated_column_attribute is called when exiting the generated_column_attribute production.
	ExitGenerated_column_attribute(c *Generated_column_attributeContext)

	// ExitColumn_definition_ref is called when exiting the column_definition_ref production.
	ExitColumn_definition_ref(c *Column_definition_refContext)

	// ExitColumn_definition_list is called when exiting the column_definition_list production.
	ExitColumn_definition_list(c *Column_definition_listContext)

	// ExitCast_data_type is called when exiting the cast_data_type production.
	ExitCast_data_type(c *Cast_data_typeContext)

	// ExitCast_datetime_type_i is called when exiting the cast_datetime_type_i production.
	ExitCast_datetime_type_i(c *Cast_datetime_type_iContext)

	// ExitData_type is called when exiting the data_type production.
	ExitData_type(c *Data_typeContext)

	// ExitString_list is called when exiting the string_list production.
	ExitString_list(c *String_listContext)

	// ExitText_string is called when exiting the text_string production.
	ExitText_string(c *Text_stringContext)

	// ExitInt_type_i is called when exiting the int_type_i production.
	ExitInt_type_i(c *Int_type_iContext)

	// ExitFloat_type_i is called when exiting the float_type_i production.
	ExitFloat_type_i(c *Float_type_iContext)

	// ExitDatetime_type_i is called when exiting the datetime_type_i production.
	ExitDatetime_type_i(c *Datetime_type_iContext)

	// ExitDate_year_type_i is called when exiting the date_year_type_i production.
	ExitDate_year_type_i(c *Date_year_type_iContext)

	// ExitText_type_i is called when exiting the text_type_i production.
	ExitText_type_i(c *Text_type_iContext)

	// ExitBlob_type_i is called when exiting the blob_type_i production.
	ExitBlob_type_i(c *Blob_type_iContext)

	// ExitString_length_i is called when exiting the string_length_i production.
	ExitString_length_i(c *String_length_iContext)

	// ExitCollation_name is called when exiting the collation_name production.
	ExitCollation_name(c *Collation_nameContext)

	// ExitTrans_param_name is called when exiting the trans_param_name production.
	ExitTrans_param_name(c *Trans_param_nameContext)

	// ExitTrans_param_value is called when exiting the trans_param_value production.
	ExitTrans_param_value(c *Trans_param_valueContext)

	// ExitCharset_name is called when exiting the charset_name production.
	ExitCharset_name(c *Charset_nameContext)

	// ExitCharset_name_or_default is called when exiting the charset_name_or_default production.
	ExitCharset_name_or_default(c *Charset_name_or_defaultContext)

	// ExitCollation is called when exiting the collation production.
	ExitCollation(c *CollationContext)

	// ExitOpt_column_attribute_list is called when exiting the opt_column_attribute_list production.
	ExitOpt_column_attribute_list(c *Opt_column_attribute_listContext)

	// ExitColumn_attribute is called when exiting the column_attribute production.
	ExitColumn_attribute(c *Column_attributeContext)

	// ExitNow_or_signed_literal is called when exiting the now_or_signed_literal production.
	ExitNow_or_signed_literal(c *Now_or_signed_literalContext)

	// ExitSigned_literal is called when exiting the signed_literal production.
	ExitSigned_literal(c *Signed_literalContext)

	// ExitOpt_comma is called when exiting the opt_comma production.
	ExitOpt_comma(c *Opt_commaContext)

	// ExitTable_option_list_space_seperated is called when exiting the table_option_list_space_seperated production.
	ExitTable_option_list_space_seperated(c *Table_option_list_space_seperatedContext)

	// ExitTable_option_list is called when exiting the table_option_list production.
	ExitTable_option_list(c *Table_option_listContext)

	// ExitPrimary_zone_name is called when exiting the primary_zone_name production.
	ExitPrimary_zone_name(c *Primary_zone_nameContext)

	// ExitTablespace is called when exiting the tablespace production.
	ExitTablespace(c *TablespaceContext)

	// ExitLocality_name is called when exiting the locality_name production.
	ExitLocality_name(c *Locality_nameContext)

	// ExitTable_option is called when exiting the table_option production.
	ExitTable_option(c *Table_optionContext)

	// ExitRelation_name_or_string is called when exiting the relation_name_or_string production.
	ExitRelation_name_or_string(c *Relation_name_or_stringContext)

	// ExitOpt_equal_mark is called when exiting the opt_equal_mark production.
	ExitOpt_equal_mark(c *Opt_equal_markContext)

	// ExitPartition_option is called when exiting the partition_option production.
	ExitPartition_option(c *Partition_optionContext)

	// ExitOpt_partition_option is called when exiting the opt_partition_option production.
	ExitOpt_partition_option(c *Opt_partition_optionContext)

	// ExitHash_partition_option is called when exiting the hash_partition_option production.
	ExitHash_partition_option(c *Hash_partition_optionContext)

	// ExitList_partition_option is called when exiting the list_partition_option production.
	ExitList_partition_option(c *List_partition_optionContext)

	// ExitKey_partition_option is called when exiting the key_partition_option production.
	ExitKey_partition_option(c *Key_partition_optionContext)

	// ExitRange_partition_option is called when exiting the range_partition_option production.
	ExitRange_partition_option(c *Range_partition_optionContext)

	// ExitOpt_column_partition_option is called when exiting the opt_column_partition_option production.
	ExitOpt_column_partition_option(c *Opt_column_partition_optionContext)

	// ExitColumn_partition_option is called when exiting the column_partition_option production.
	ExitColumn_partition_option(c *Column_partition_optionContext)

	// ExitAux_column_list is called when exiting the aux_column_list production.
	ExitAux_column_list(c *Aux_column_listContext)

	// ExitVertical_column_name is called when exiting the vertical_column_name production.
	ExitVertical_column_name(c *Vertical_column_nameContext)

	// ExitColumn_name_list is called when exiting the column_name_list production.
	ExitColumn_name_list(c *Column_name_listContext)

	// ExitSubpartition_option is called when exiting the subpartition_option production.
	ExitSubpartition_option(c *Subpartition_optionContext)

	// ExitOpt_list_partition_list is called when exiting the opt_list_partition_list production.
	ExitOpt_list_partition_list(c *Opt_list_partition_listContext)

	// ExitOpt_list_subpartition_list is called when exiting the opt_list_subpartition_list production.
	ExitOpt_list_subpartition_list(c *Opt_list_subpartition_listContext)

	// ExitOpt_range_partition_list is called when exiting the opt_range_partition_list production.
	ExitOpt_range_partition_list(c *Opt_range_partition_listContext)

	// ExitOpt_range_subpartition_list is called when exiting the opt_range_subpartition_list production.
	ExitOpt_range_subpartition_list(c *Opt_range_subpartition_listContext)

	// ExitList_partition_list is called when exiting the list_partition_list production.
	ExitList_partition_list(c *List_partition_listContext)

	// ExitList_subpartition_list is called when exiting the list_subpartition_list production.
	ExitList_subpartition_list(c *List_subpartition_listContext)

	// ExitList_subpartition_element is called when exiting the list_subpartition_element production.
	ExitList_subpartition_element(c *List_subpartition_elementContext)

	// ExitList_partition_element is called when exiting the list_partition_element production.
	ExitList_partition_element(c *List_partition_elementContext)

	// ExitList_partition_expr is called when exiting the list_partition_expr production.
	ExitList_partition_expr(c *List_partition_exprContext)

	// ExitList_expr is called when exiting the list_expr production.
	ExitList_expr(c *List_exprContext)

	// ExitRange_partition_list is called when exiting the range_partition_list production.
	ExitRange_partition_list(c *Range_partition_listContext)

	// ExitRange_partition_element is called when exiting the range_partition_element production.
	ExitRange_partition_element(c *Range_partition_elementContext)

	// ExitRange_subpartition_element is called when exiting the range_subpartition_element production.
	ExitRange_subpartition_element(c *Range_subpartition_elementContext)

	// ExitRange_subpartition_list is called when exiting the range_subpartition_list production.
	ExitRange_subpartition_list(c *Range_subpartition_listContext)

	// ExitRange_partition_expr is called when exiting the range_partition_expr production.
	ExitRange_partition_expr(c *Range_partition_exprContext)

	// ExitRange_expr_list is called when exiting the range_expr_list production.
	ExitRange_expr_list(c *Range_expr_listContext)

	// ExitRange_expr is called when exiting the range_expr production.
	ExitRange_expr(c *Range_exprContext)

	// ExitInt_or_decimal is called when exiting the int_or_decimal production.
	ExitInt_or_decimal(c *Int_or_decimalContext)

	// ExitTg_hash_partition_option is called when exiting the tg_hash_partition_option production.
	ExitTg_hash_partition_option(c *Tg_hash_partition_optionContext)

	// ExitTg_key_partition_option is called when exiting the tg_key_partition_option production.
	ExitTg_key_partition_option(c *Tg_key_partition_optionContext)

	// ExitTg_range_partition_option is called when exiting the tg_range_partition_option production.
	ExitTg_range_partition_option(c *Tg_range_partition_optionContext)

	// ExitTg_list_partition_option is called when exiting the tg_list_partition_option production.
	ExitTg_list_partition_option(c *Tg_list_partition_optionContext)

	// ExitTg_subpartition_option is called when exiting the tg_subpartition_option production.
	ExitTg_subpartition_option(c *Tg_subpartition_optionContext)

	// ExitRow_format_option is called when exiting the row_format_option production.
	ExitRow_format_option(c *Row_format_optionContext)

	// ExitCreate_tablegroup_stmt is called when exiting the create_tablegroup_stmt production.
	ExitCreate_tablegroup_stmt(c *Create_tablegroup_stmtContext)

	// ExitDrop_tablegroup_stmt is called when exiting the drop_tablegroup_stmt production.
	ExitDrop_tablegroup_stmt(c *Drop_tablegroup_stmtContext)

	// ExitAlter_tablegroup_stmt is called when exiting the alter_tablegroup_stmt production.
	ExitAlter_tablegroup_stmt(c *Alter_tablegroup_stmtContext)

	// ExitTablegroup_option_list_space_seperated is called when exiting the tablegroup_option_list_space_seperated production.
	ExitTablegroup_option_list_space_seperated(c *Tablegroup_option_list_space_seperatedContext)

	// ExitTablegroup_option_list is called when exiting the tablegroup_option_list production.
	ExitTablegroup_option_list(c *Tablegroup_option_listContext)

	// ExitTablegroup_option is called when exiting the tablegroup_option production.
	ExitTablegroup_option(c *Tablegroup_optionContext)

	// ExitAlter_tablegroup_actions is called when exiting the alter_tablegroup_actions production.
	ExitAlter_tablegroup_actions(c *Alter_tablegroup_actionsContext)

	// ExitAlter_tablegroup_action is called when exiting the alter_tablegroup_action production.
	ExitAlter_tablegroup_action(c *Alter_tablegroup_actionContext)

	// ExitDefault_tablegroup is called when exiting the default_tablegroup production.
	ExitDefault_tablegroup(c *Default_tablegroupContext)

	// ExitCreate_view_stmt is called when exiting the create_view_stmt production.
	ExitCreate_view_stmt(c *Create_view_stmtContext)

	// ExitView_select_stmt is called when exiting the view_select_stmt production.
	ExitView_select_stmt(c *View_select_stmtContext)

	// ExitView_name is called when exiting the view_name production.
	ExitView_name(c *View_nameContext)

	// ExitCreate_index_stmt is called when exiting the create_index_stmt production.
	ExitCreate_index_stmt(c *Create_index_stmtContext)

	// ExitIndex_name is called when exiting the index_name production.
	ExitIndex_name(c *Index_nameContext)

	// ExitOpt_constraint_name is called when exiting the opt_constraint_name production.
	ExitOpt_constraint_name(c *Opt_constraint_nameContext)

	// ExitConstraint_name is called when exiting the constraint_name production.
	ExitConstraint_name(c *Constraint_nameContext)

	// ExitSort_column_list is called when exiting the sort_column_list production.
	ExitSort_column_list(c *Sort_column_listContext)

	// ExitSort_column_key is called when exiting the sort_column_key production.
	ExitSort_column_key(c *Sort_column_keyContext)

	// ExitOpt_index_options is called when exiting the opt_index_options production.
	ExitOpt_index_options(c *Opt_index_optionsContext)

	// ExitIndex_option is called when exiting the index_option production.
	ExitIndex_option(c *Index_optionContext)

	// ExitIndex_using_algorithm is called when exiting the index_using_algorithm production.
	ExitIndex_using_algorithm(c *Index_using_algorithmContext)

	// ExitDrop_table_stmt is called when exiting the drop_table_stmt production.
	ExitDrop_table_stmt(c *Drop_table_stmtContext)

	// ExitTable_or_tables is called when exiting the table_or_tables production.
	ExitTable_or_tables(c *Table_or_tablesContext)

	// ExitDrop_view_stmt is called when exiting the drop_view_stmt production.
	ExitDrop_view_stmt(c *Drop_view_stmtContext)

	// ExitTable_list is called when exiting the table_list production.
	ExitTable_list(c *Table_listContext)

	// ExitDrop_index_stmt is called when exiting the drop_index_stmt production.
	ExitDrop_index_stmt(c *Drop_index_stmtContext)

	// ExitInsert_stmt is called when exiting the insert_stmt production.
	ExitInsert_stmt(c *Insert_stmtContext)

	// ExitSingle_table_insert is called when exiting the single_table_insert production.
	ExitSingle_table_insert(c *Single_table_insertContext)

	// ExitValues_clause is called when exiting the values_clause production.
	ExitValues_clause(c *Values_clauseContext)

	// ExitValue_or_values is called when exiting the value_or_values production.
	ExitValue_or_values(c *Value_or_valuesContext)

	// ExitReplace_with_opt_hint is called when exiting the replace_with_opt_hint production.
	ExitReplace_with_opt_hint(c *Replace_with_opt_hintContext)

	// ExitInsert_with_opt_hint is called when exiting the insert_with_opt_hint production.
	ExitInsert_with_opt_hint(c *Insert_with_opt_hintContext)

	// ExitColumn_list is called when exiting the column_list production.
	ExitColumn_list(c *Column_listContext)

	// ExitInsert_vals_list is called when exiting the insert_vals_list production.
	ExitInsert_vals_list(c *Insert_vals_listContext)

	// ExitInsert_vals is called when exiting the insert_vals production.
	ExitInsert_vals(c *Insert_valsContext)

	// ExitExpr_or_default is called when exiting the expr_or_default production.
	ExitExpr_or_default(c *Expr_or_defaultContext)

	// ExitSelect_stmt is called when exiting the select_stmt production.
	ExitSelect_stmt(c *Select_stmtContext)

	// ExitSelect_into is called when exiting the select_into production.
	ExitSelect_into(c *Select_intoContext)

	// ExitSelect_with_parens is called when exiting the select_with_parens production.
	ExitSelect_with_parens(c *Select_with_parensContext)

	// ExitSelect_no_parens is called when exiting the select_no_parens production.
	ExitSelect_no_parens(c *Select_no_parensContext)

	// ExitNo_table_select is called when exiting the no_table_select production.
	ExitNo_table_select(c *No_table_selectContext)

	// ExitSelect_clause is called when exiting the select_clause production.
	ExitSelect_clause(c *Select_clauseContext)

	// ExitSelect_clause_set_with_order_and_limit is called when exiting the select_clause_set_with_order_and_limit production.
	ExitSelect_clause_set_with_order_and_limit(c *Select_clause_set_with_order_and_limitContext)

	// ExitSelect_clause_set is called when exiting the select_clause_set production.
	ExitSelect_clause_set(c *Select_clause_setContext)

	// ExitSelect_clause_set_right is called when exiting the select_clause_set_right production.
	ExitSelect_clause_set_right(c *Select_clause_set_rightContext)

	// ExitSelect_clause_set_left is called when exiting the select_clause_set_left production.
	ExitSelect_clause_set_left(c *Select_clause_set_leftContext)

	// ExitNo_table_select_with_order_and_limit is called when exiting the no_table_select_with_order_and_limit production.
	ExitNo_table_select_with_order_and_limit(c *No_table_select_with_order_and_limitContext)

	// ExitSimple_select_with_order_and_limit is called when exiting the simple_select_with_order_and_limit production.
	ExitSimple_select_with_order_and_limit(c *Simple_select_with_order_and_limitContext)

	// ExitSelect_with_parens_with_order_and_limit is called when exiting the select_with_parens_with_order_and_limit production.
	ExitSelect_with_parens_with_order_and_limit(c *Select_with_parens_with_order_and_limitContext)

	// ExitSelect_with_opt_hint is called when exiting the select_with_opt_hint production.
	ExitSelect_with_opt_hint(c *Select_with_opt_hintContext)

	// ExitUpdate_with_opt_hint is called when exiting the update_with_opt_hint production.
	ExitUpdate_with_opt_hint(c *Update_with_opt_hintContext)

	// ExitDelete_with_opt_hint is called when exiting the delete_with_opt_hint production.
	ExitDelete_with_opt_hint(c *Delete_with_opt_hintContext)

	// ExitSimple_select is called when exiting the simple_select production.
	ExitSimple_select(c *Simple_selectContext)

	// ExitSet_type_union is called when exiting the set_type_union production.
	ExitSet_type_union(c *Set_type_unionContext)

	// ExitSet_type_other is called when exiting the set_type_other production.
	ExitSet_type_other(c *Set_type_otherContext)

	// ExitSet_type is called when exiting the set_type production.
	ExitSet_type(c *Set_typeContext)

	// ExitSet_expression_option is called when exiting the set_expression_option production.
	ExitSet_expression_option(c *Set_expression_optionContext)

	// ExitOpt_hint_value is called when exiting the opt_hint_value production.
	ExitOpt_hint_value(c *Opt_hint_valueContext)

	// ExitLimit_clause is called when exiting the limit_clause production.
	ExitLimit_clause(c *Limit_clauseContext)

	// ExitInto_clause is called when exiting the into_clause production.
	ExitInto_clause(c *Into_clauseContext)

	// ExitInto_opt is called when exiting the into_opt production.
	ExitInto_opt(c *Into_optContext)

	// ExitInto_var_list is called when exiting the into_var_list production.
	ExitInto_var_list(c *Into_var_listContext)

	// ExitInto_var is called when exiting the into_var production.
	ExitInto_var(c *Into_varContext)

	// ExitField_opt is called when exiting the field_opt production.
	ExitField_opt(c *Field_optContext)

	// ExitField_term_list is called when exiting the field_term_list production.
	ExitField_term_list(c *Field_term_listContext)

	// ExitField_term is called when exiting the field_term production.
	ExitField_term(c *Field_termContext)

	// ExitLine_opt is called when exiting the line_opt production.
	ExitLine_opt(c *Line_optContext)

	// ExitLine_term_list is called when exiting the line_term_list production.
	ExitLine_term_list(c *Line_term_listContext)

	// ExitLine_term is called when exiting the line_term production.
	ExitLine_term(c *Line_termContext)

	// ExitHint_list_with_end is called when exiting the hint_list_with_end production.
	ExitHint_list_with_end(c *Hint_list_with_endContext)

	// ExitOpt_hint_list is called when exiting the opt_hint_list production.
	ExitOpt_hint_list(c *Opt_hint_listContext)

	// ExitHint_options is called when exiting the hint_options production.
	ExitHint_options(c *Hint_optionsContext)

	// ExitName_list is called when exiting the name_list production.
	ExitName_list(c *Name_listContext)

	// ExitHint_option is called when exiting the hint_option production.
	ExitHint_option(c *Hint_optionContext)

	// ExitConsistency_level is called when exiting the consistency_level production.
	ExitConsistency_level(c *Consistency_levelContext)

	// ExitUse_plan_cache_type is called when exiting the use_plan_cache_type production.
	ExitUse_plan_cache_type(c *Use_plan_cache_typeContext)

	// ExitUse_jit_type is called when exiting the use_jit_type production.
	ExitUse_jit_type(c *Use_jit_typeContext)

	// ExitDistribute_method is called when exiting the distribute_method production.
	ExitDistribute_method(c *Distribute_methodContext)

	// ExitLimit_expr is called when exiting the limit_expr production.
	ExitLimit_expr(c *Limit_exprContext)

	// ExitOpt_for_update_wait is called when exiting the opt_for_update_wait production.
	ExitOpt_for_update_wait(c *Opt_for_update_waitContext)

	// ExitParameterized_trim is called when exiting the parameterized_trim production.
	ExitParameterized_trim(c *Parameterized_trimContext)

	// ExitGroupby_clause is called when exiting the groupby_clause production.
	ExitGroupby_clause(c *Groupby_clauseContext)

	// ExitSort_list_for_group_by is called when exiting the sort_list_for_group_by production.
	ExitSort_list_for_group_by(c *Sort_list_for_group_byContext)

	// ExitSort_key_for_group_by is called when exiting the sort_key_for_group_by production.
	ExitSort_key_for_group_by(c *Sort_key_for_group_byContext)

	// ExitOrder_by is called when exiting the order_by production.
	ExitOrder_by(c *Order_byContext)

	// ExitSort_list is called when exiting the sort_list production.
	ExitSort_list(c *Sort_listContext)

	// ExitSort_key is called when exiting the sort_key production.
	ExitSort_key(c *Sort_keyContext)

	// ExitQuery_expression_option_list is called when exiting the query_expression_option_list production.
	ExitQuery_expression_option_list(c *Query_expression_option_listContext)

	// ExitQuery_expression_option is called when exiting the query_expression_option production.
	ExitQuery_expression_option(c *Query_expression_optionContext)

	// ExitProjection is called when exiting the projection production.
	ExitProjection(c *ProjectionContext)

	// ExitSelect_expr_list is called when exiting the select_expr_list production.
	ExitSelect_expr_list(c *Select_expr_listContext)

	// ExitFrom_list is called when exiting the from_list production.
	ExitFrom_list(c *From_listContext)

	// ExitTable_references is called when exiting the table_references production.
	ExitTable_references(c *Table_referencesContext)

	// ExitTable_reference is called when exiting the table_reference production.
	ExitTable_reference(c *Table_referenceContext)

	// ExitTable_factor is called when exiting the table_factor production.
	ExitTable_factor(c *Table_factorContext)

	// ExitTbl_name is called when exiting the tbl_name production.
	ExitTbl_name(c *Tbl_nameContext)

	// ExitDml_table_name is called when exiting the dml_table_name production.
	ExitDml_table_name(c *Dml_table_nameContext)

	// ExitSeed is called when exiting the seed production.
	ExitSeed(c *SeedContext)

	// ExitOpt_seed is called when exiting the opt_seed production.
	ExitOpt_seed(c *Opt_seedContext)

	// ExitSample_percent is called when exiting the sample_percent production.
	ExitSample_percent(c *Sample_percentContext)

	// ExitSample_clause is called when exiting the sample_clause production.
	ExitSample_clause(c *Sample_clauseContext)

	// ExitTable_subquery is called when exiting the table_subquery production.
	ExitTable_subquery(c *Table_subqueryContext)

	// ExitUse_partition is called when exiting the use_partition production.
	ExitUse_partition(c *Use_partitionContext)

	// ExitIndex_hint_type is called when exiting the index_hint_type production.
	ExitIndex_hint_type(c *Index_hint_typeContext)

	// ExitKey_or_index is called when exiting the key_or_index production.
	ExitKey_or_index(c *Key_or_indexContext)

	// ExitIndex_hint_scope is called when exiting the index_hint_scope production.
	ExitIndex_hint_scope(c *Index_hint_scopeContext)

	// ExitIndex_element is called when exiting the index_element production.
	ExitIndex_element(c *Index_elementContext)

	// ExitIndex_list is called when exiting the index_list production.
	ExitIndex_list(c *Index_listContext)

	// ExitIndex_hint_definition is called when exiting the index_hint_definition production.
	ExitIndex_hint_definition(c *Index_hint_definitionContext)

	// ExitIndex_hint_list is called when exiting the index_hint_list production.
	ExitIndex_hint_list(c *Index_hint_listContext)

	// ExitRelation_factor is called when exiting the relation_factor production.
	ExitRelation_factor(c *Relation_factorContext)

	// ExitRelation_with_star_list is called when exiting the relation_with_star_list production.
	ExitRelation_with_star_list(c *Relation_with_star_listContext)

	// ExitRelation_factor_with_star is called when exiting the relation_factor_with_star production.
	ExitRelation_factor_with_star(c *Relation_factor_with_starContext)

	// ExitNormal_relation_factor is called when exiting the normal_relation_factor production.
	ExitNormal_relation_factor(c *Normal_relation_factorContext)

	// ExitDot_relation_factor is called when exiting the dot_relation_factor production.
	ExitDot_relation_factor(c *Dot_relation_factorContext)

	// ExitRelation_factor_in_hint is called when exiting the relation_factor_in_hint production.
	ExitRelation_factor_in_hint(c *Relation_factor_in_hintContext)

	// ExitQb_name_option is called when exiting the qb_name_option production.
	ExitQb_name_option(c *Qb_name_optionContext)

	// ExitRelation_factor_in_hint_list is called when exiting the relation_factor_in_hint_list production.
	ExitRelation_factor_in_hint_list(c *Relation_factor_in_hint_listContext)

	// ExitRelation_sep_option is called when exiting the relation_sep_option production.
	ExitRelation_sep_option(c *Relation_sep_optionContext)

	// ExitRelation_factor_in_pq_hint is called when exiting the relation_factor_in_pq_hint production.
	ExitRelation_factor_in_pq_hint(c *Relation_factor_in_pq_hintContext)

	// ExitRelation_factor_in_leading_hint is called when exiting the relation_factor_in_leading_hint production.
	ExitRelation_factor_in_leading_hint(c *Relation_factor_in_leading_hintContext)

	// ExitRelation_factor_in_leading_hint_list is called when exiting the relation_factor_in_leading_hint_list production.
	ExitRelation_factor_in_leading_hint_list(c *Relation_factor_in_leading_hint_listContext)

	// ExitRelation_factor_in_leading_hint_list_entry is called when exiting the relation_factor_in_leading_hint_list_entry production.
	ExitRelation_factor_in_leading_hint_list_entry(c *Relation_factor_in_leading_hint_list_entryContext)

	// ExitRelation_factor_in_use_join_hint_list is called when exiting the relation_factor_in_use_join_hint_list production.
	ExitRelation_factor_in_use_join_hint_list(c *Relation_factor_in_use_join_hint_listContext)

	// ExitTracing_num_list is called when exiting the tracing_num_list production.
	ExitTracing_num_list(c *Tracing_num_listContext)

	// ExitJoin_condition is called when exiting the join_condition production.
	ExitJoin_condition(c *Join_conditionContext)

	// ExitJoined_table is called when exiting the joined_table production.
	ExitJoined_table(c *Joined_tableContext)

	// ExitNatural_join_type is called when exiting the natural_join_type production.
	ExitNatural_join_type(c *Natural_join_typeContext)

	// ExitInner_join_type is called when exiting the inner_join_type production.
	ExitInner_join_type(c *Inner_join_typeContext)

	// ExitOuter_join_type is called when exiting the outer_join_type production.
	ExitOuter_join_type(c *Outer_join_typeContext)

	// ExitAnalyze_stmt is called when exiting the analyze_stmt production.
	ExitAnalyze_stmt(c *Analyze_stmtContext)

	// ExitCreate_outline_stmt is called when exiting the create_outline_stmt production.
	ExitCreate_outline_stmt(c *Create_outline_stmtContext)

	// ExitAlter_outline_stmt is called when exiting the alter_outline_stmt production.
	ExitAlter_outline_stmt(c *Alter_outline_stmtContext)

	// ExitDrop_outline_stmt is called when exiting the drop_outline_stmt production.
	ExitDrop_outline_stmt(c *Drop_outline_stmtContext)

	// ExitExplain_stmt is called when exiting the explain_stmt production.
	ExitExplain_stmt(c *Explain_stmtContext)

	// ExitExplain_or_desc is called when exiting the explain_or_desc production.
	ExitExplain_or_desc(c *Explain_or_descContext)

	// ExitExplainable_stmt is called when exiting the explainable_stmt production.
	ExitExplainable_stmt(c *Explainable_stmtContext)

	// ExitFormat_name is called when exiting the format_name production.
	ExitFormat_name(c *Format_nameContext)

	// ExitShow_stmt is called when exiting the show_stmt production.
	ExitShow_stmt(c *Show_stmtContext)

	// ExitDatabases_or_schemas is called when exiting the databases_or_schemas production.
	ExitDatabases_or_schemas(c *Databases_or_schemasContext)

	// ExitOpt_for_grant_user is called when exiting the opt_for_grant_user production.
	ExitOpt_for_grant_user(c *Opt_for_grant_userContext)

	// ExitColumns_or_fields is called when exiting the columns_or_fields production.
	ExitColumns_or_fields(c *Columns_or_fieldsContext)

	// ExitDatabase_or_schema is called when exiting the database_or_schema production.
	ExitDatabase_or_schema(c *Database_or_schemaContext)

	// ExitIndex_or_indexes_or_keys is called when exiting the index_or_indexes_or_keys production.
	ExitIndex_or_indexes_or_keys(c *Index_or_indexes_or_keysContext)

	// ExitFrom_or_in is called when exiting the from_or_in production.
	ExitFrom_or_in(c *From_or_inContext)

	// ExitHelp_stmt is called when exiting the help_stmt production.
	ExitHelp_stmt(c *Help_stmtContext)

	// ExitCreate_tablespace_stmt is called when exiting the create_tablespace_stmt production.
	ExitCreate_tablespace_stmt(c *Create_tablespace_stmtContext)

	// ExitPermanent_tablespace is called when exiting the permanent_tablespace production.
	ExitPermanent_tablespace(c *Permanent_tablespaceContext)

	// ExitPermanent_tablespace_option is called when exiting the permanent_tablespace_option production.
	ExitPermanent_tablespace_option(c *Permanent_tablespace_optionContext)

	// ExitDrop_tablespace_stmt is called when exiting the drop_tablespace_stmt production.
	ExitDrop_tablespace_stmt(c *Drop_tablespace_stmtContext)

	// ExitAlter_tablespace_actions is called when exiting the alter_tablespace_actions production.
	ExitAlter_tablespace_actions(c *Alter_tablespace_actionsContext)

	// ExitAlter_tablespace_action is called when exiting the alter_tablespace_action production.
	ExitAlter_tablespace_action(c *Alter_tablespace_actionContext)

	// ExitAlter_tablespace_stmt is called when exiting the alter_tablespace_stmt production.
	ExitAlter_tablespace_stmt(c *Alter_tablespace_stmtContext)

	// ExitRotate_master_key_stmt is called when exiting the rotate_master_key_stmt production.
	ExitRotate_master_key_stmt(c *Rotate_master_key_stmtContext)

	// ExitPermanent_tablespace_options is called when exiting the permanent_tablespace_options production.
	ExitPermanent_tablespace_options(c *Permanent_tablespace_optionsContext)

	// ExitCreate_user_stmt is called when exiting the create_user_stmt production.
	ExitCreate_user_stmt(c *Create_user_stmtContext)

	// ExitUser_specification_list is called when exiting the user_specification_list production.
	ExitUser_specification_list(c *User_specification_listContext)

	// ExitUser_specification is called when exiting the user_specification production.
	ExitUser_specification(c *User_specificationContext)

	// ExitRequire_specification is called when exiting the require_specification production.
	ExitRequire_specification(c *Require_specificationContext)

	// ExitTls_option_list is called when exiting the tls_option_list production.
	ExitTls_option_list(c *Tls_option_listContext)

	// ExitTls_option is called when exiting the tls_option production.
	ExitTls_option(c *Tls_optionContext)

	// ExitUser is called when exiting the user production.
	ExitUser(c *UserContext)

	// ExitOpt_host_name is called when exiting the opt_host_name production.
	ExitOpt_host_name(c *Opt_host_nameContext)

	// ExitUser_with_host_name is called when exiting the user_with_host_name production.
	ExitUser_with_host_name(c *User_with_host_nameContext)

	// ExitPassword is called when exiting the password production.
	ExitPassword(c *PasswordContext)

	// ExitDrop_user_stmt is called when exiting the drop_user_stmt production.
	ExitDrop_user_stmt(c *Drop_user_stmtContext)

	// ExitUser_list is called when exiting the user_list production.
	ExitUser_list(c *User_listContext)

	// ExitSet_password_stmt is called when exiting the set_password_stmt production.
	ExitSet_password_stmt(c *Set_password_stmtContext)

	// ExitOpt_for_user is called when exiting the opt_for_user production.
	ExitOpt_for_user(c *Opt_for_userContext)

	// ExitRename_user_stmt is called when exiting the rename_user_stmt production.
	ExitRename_user_stmt(c *Rename_user_stmtContext)

	// ExitRename_info is called when exiting the rename_info production.
	ExitRename_info(c *Rename_infoContext)

	// ExitRename_list is called when exiting the rename_list production.
	ExitRename_list(c *Rename_listContext)

	// ExitLock_user_stmt is called when exiting the lock_user_stmt production.
	ExitLock_user_stmt(c *Lock_user_stmtContext)

	// ExitLock_spec_mysql57 is called when exiting the lock_spec_mysql57 production.
	ExitLock_spec_mysql57(c *Lock_spec_mysql57Context)

	// ExitLock_tables_stmt is called when exiting the lock_tables_stmt production.
	ExitLock_tables_stmt(c *Lock_tables_stmtContext)

	// ExitUnlock_tables_stmt is called when exiting the unlock_tables_stmt production.
	ExitUnlock_tables_stmt(c *Unlock_tables_stmtContext)

	// ExitLock_table_list is called when exiting the lock_table_list production.
	ExitLock_table_list(c *Lock_table_listContext)

	// ExitLock_table is called when exiting the lock_table production.
	ExitLock_table(c *Lock_tableContext)

	// ExitLock_type is called when exiting the lock_type production.
	ExitLock_type(c *Lock_typeContext)

	// ExitBegin_stmt is called when exiting the begin_stmt production.
	ExitBegin_stmt(c *Begin_stmtContext)

	// ExitCommit_stmt is called when exiting the commit_stmt production.
	ExitCommit_stmt(c *Commit_stmtContext)

	// ExitRollback_stmt is called when exiting the rollback_stmt production.
	ExitRollback_stmt(c *Rollback_stmtContext)

	// ExitKill_stmt is called when exiting the kill_stmt production.
	ExitKill_stmt(c *Kill_stmtContext)

	// ExitGrant_stmt is called when exiting the grant_stmt production.
	ExitGrant_stmt(c *Grant_stmtContext)

	// ExitGrant_privileges is called when exiting the grant_privileges production.
	ExitGrant_privileges(c *Grant_privilegesContext)

	// ExitPriv_type_list is called when exiting the priv_type_list production.
	ExitPriv_type_list(c *Priv_type_listContext)

	// ExitPriv_type is called when exiting the priv_type production.
	ExitPriv_type(c *Priv_typeContext)

	// ExitPriv_level is called when exiting the priv_level production.
	ExitPriv_level(c *Priv_levelContext)

	// ExitGrant_options is called when exiting the grant_options production.
	ExitGrant_options(c *Grant_optionsContext)

	// ExitRevoke_stmt is called when exiting the revoke_stmt production.
	ExitRevoke_stmt(c *Revoke_stmtContext)

	// ExitPrepare_stmt is called when exiting the prepare_stmt production.
	ExitPrepare_stmt(c *Prepare_stmtContext)

	// ExitStmt_name is called when exiting the stmt_name production.
	ExitStmt_name(c *Stmt_nameContext)

	// ExitPreparable_stmt is called when exiting the preparable_stmt production.
	ExitPreparable_stmt(c *Preparable_stmtContext)

	// ExitVariable_set_stmt is called when exiting the variable_set_stmt production.
	ExitVariable_set_stmt(c *Variable_set_stmtContext)

	// ExitSys_var_and_val_list is called when exiting the sys_var_and_val_list production.
	ExitSys_var_and_val_list(c *Sys_var_and_val_listContext)

	// ExitVar_and_val_list is called when exiting the var_and_val_list production.
	ExitVar_and_val_list(c *Var_and_val_listContext)

	// ExitSet_expr_or_default is called when exiting the set_expr_or_default production.
	ExitSet_expr_or_default(c *Set_expr_or_defaultContext)

	// ExitVar_and_val is called when exiting the var_and_val production.
	ExitVar_and_val(c *Var_and_valContext)

	// ExitSys_var_and_val is called when exiting the sys_var_and_val production.
	ExitSys_var_and_val(c *Sys_var_and_valContext)

	// ExitScope_or_scope_alias is called when exiting the scope_or_scope_alias production.
	ExitScope_or_scope_alias(c *Scope_or_scope_aliasContext)

	// ExitTo_or_eq is called when exiting the to_or_eq production.
	ExitTo_or_eq(c *To_or_eqContext)

	// ExitExecute_stmt is called when exiting the execute_stmt production.
	ExitExecute_stmt(c *Execute_stmtContext)

	// ExitArgument_list is called when exiting the argument_list production.
	ExitArgument_list(c *Argument_listContext)

	// ExitArgument is called when exiting the argument production.
	ExitArgument(c *ArgumentContext)

	// ExitDeallocate_prepare_stmt is called when exiting the deallocate_prepare_stmt production.
	ExitDeallocate_prepare_stmt(c *Deallocate_prepare_stmtContext)

	// ExitDeallocate_or_drop is called when exiting the deallocate_or_drop production.
	ExitDeallocate_or_drop(c *Deallocate_or_dropContext)

	// ExitTruncate_table_stmt is called when exiting the truncate_table_stmt production.
	ExitTruncate_table_stmt(c *Truncate_table_stmtContext)

	// ExitRename_table_stmt is called when exiting the rename_table_stmt production.
	ExitRename_table_stmt(c *Rename_table_stmtContext)

	// ExitRename_table_actions is called when exiting the rename_table_actions production.
	ExitRename_table_actions(c *Rename_table_actionsContext)

	// ExitRename_table_action is called when exiting the rename_table_action production.
	ExitRename_table_action(c *Rename_table_actionContext)

	// ExitAlter_table_stmt is called when exiting the alter_table_stmt production.
	ExitAlter_table_stmt(c *Alter_table_stmtContext)

	// ExitAlter_table_actions is called when exiting the alter_table_actions production.
	ExitAlter_table_actions(c *Alter_table_actionsContext)

	// ExitAlter_table_action is called when exiting the alter_table_action production.
	ExitAlter_table_action(c *Alter_table_actionContext)

	// ExitAlter_constraint_option is called when exiting the alter_constraint_option production.
	ExitAlter_constraint_option(c *Alter_constraint_optionContext)

	// ExitAlter_partition_option is called when exiting the alter_partition_option production.
	ExitAlter_partition_option(c *Alter_partition_optionContext)

	// ExitOpt_partition_range_or_list is called when exiting the opt_partition_range_or_list production.
	ExitOpt_partition_range_or_list(c *Opt_partition_range_or_listContext)

	// ExitAlter_tg_partition_option is called when exiting the alter_tg_partition_option production.
	ExitAlter_tg_partition_option(c *Alter_tg_partition_optionContext)

	// ExitDrop_partition_name_list is called when exiting the drop_partition_name_list production.
	ExitDrop_partition_name_list(c *Drop_partition_name_listContext)

	// ExitModify_partition_info is called when exiting the modify_partition_info production.
	ExitModify_partition_info(c *Modify_partition_infoContext)

	// ExitModify_tg_partition_info is called when exiting the modify_tg_partition_info production.
	ExitModify_tg_partition_info(c *Modify_tg_partition_infoContext)

	// ExitAlter_index_option is called when exiting the alter_index_option production.
	ExitAlter_index_option(c *Alter_index_optionContext)

	// ExitAlter_foreign_key_action is called when exiting the alter_foreign_key_action production.
	ExitAlter_foreign_key_action(c *Alter_foreign_key_actionContext)

	// ExitVisibility_option is called when exiting the visibility_option production.
	ExitVisibility_option(c *Visibility_optionContext)

	// ExitAlter_column_option is called when exiting the alter_column_option production.
	ExitAlter_column_option(c *Alter_column_optionContext)

	// ExitAlter_tablegroup_option is called when exiting the alter_tablegroup_option production.
	ExitAlter_tablegroup_option(c *Alter_tablegroup_optionContext)

	// ExitAlter_column_behavior is called when exiting the alter_column_behavior production.
	ExitAlter_column_behavior(c *Alter_column_behaviorContext)

	// ExitFlashback_stmt is called when exiting the flashback_stmt production.
	ExitFlashback_stmt(c *Flashback_stmtContext)

	// ExitPurge_stmt is called when exiting the purge_stmt production.
	ExitPurge_stmt(c *Purge_stmtContext)

	// ExitOptimize_stmt is called when exiting the optimize_stmt production.
	ExitOptimize_stmt(c *Optimize_stmtContext)

	// ExitDump_memory_stmt is called when exiting the dump_memory_stmt production.
	ExitDump_memory_stmt(c *Dump_memory_stmtContext)

	// ExitAlter_system_stmt is called when exiting the alter_system_stmt production.
	ExitAlter_system_stmt(c *Alter_system_stmtContext)

	// ExitChange_tenant_name_or_tenant_id is called when exiting the change_tenant_name_or_tenant_id production.
	ExitChange_tenant_name_or_tenant_id(c *Change_tenant_name_or_tenant_idContext)

	// ExitCache_type is called when exiting the cache_type production.
	ExitCache_type(c *Cache_typeContext)

	// ExitBalance_task_type is called when exiting the balance_task_type production.
	ExitBalance_task_type(c *Balance_task_typeContext)

	// ExitTenant_list_tuple is called when exiting the tenant_list_tuple production.
	ExitTenant_list_tuple(c *Tenant_list_tupleContext)

	// ExitTenant_name_list is called when exiting the tenant_name_list production.
	ExitTenant_name_list(c *Tenant_name_listContext)

	// ExitFlush_scope is called when exiting the flush_scope production.
	ExitFlush_scope(c *Flush_scopeContext)

	// ExitServer_info_list is called when exiting the server_info_list production.
	ExitServer_info_list(c *Server_info_listContext)

	// ExitServer_info is called when exiting the server_info production.
	ExitServer_info(c *Server_infoContext)

	// ExitServer_action is called when exiting the server_action production.
	ExitServer_action(c *Server_actionContext)

	// ExitServer_list is called when exiting the server_list production.
	ExitServer_list(c *Server_listContext)

	// ExitZone_action is called when exiting the zone_action production.
	ExitZone_action(c *Zone_actionContext)

	// ExitIp_port is called when exiting the ip_port production.
	ExitIp_port(c *Ip_portContext)

	// ExitZone_desc is called when exiting the zone_desc production.
	ExitZone_desc(c *Zone_descContext)

	// ExitServer_or_zone is called when exiting the server_or_zone production.
	ExitServer_or_zone(c *Server_or_zoneContext)

	// ExitAdd_or_alter_zone_option is called when exiting the add_or_alter_zone_option production.
	ExitAdd_or_alter_zone_option(c *Add_or_alter_zone_optionContext)

	// ExitAdd_or_alter_zone_options is called when exiting the add_or_alter_zone_options production.
	ExitAdd_or_alter_zone_options(c *Add_or_alter_zone_optionsContext)

	// ExitAlter_or_change_or_modify is called when exiting the alter_or_change_or_modify production.
	ExitAlter_or_change_or_modify(c *Alter_or_change_or_modifyContext)

	// ExitPartition_id_desc is called when exiting the partition_id_desc production.
	ExitPartition_id_desc(c *Partition_id_descContext)

	// ExitPartition_id_or_server_or_zone is called when exiting the partition_id_or_server_or_zone production.
	ExitPartition_id_or_server_or_zone(c *Partition_id_or_server_or_zoneContext)

	// ExitMigrate_action is called when exiting the migrate_action production.
	ExitMigrate_action(c *Migrate_actionContext)

	// ExitChange_actions is called when exiting the change_actions production.
	ExitChange_actions(c *Change_actionsContext)

	// ExitChange_action is called when exiting the change_action production.
	ExitChange_action(c *Change_actionContext)

	// ExitReplica_type is called when exiting the replica_type production.
	ExitReplica_type(c *Replica_typeContext)

	// ExitSuspend_or_resume is called when exiting the suspend_or_resume production.
	ExitSuspend_or_resume(c *Suspend_or_resumeContext)

	// ExitBaseline_id_expr is called when exiting the baseline_id_expr production.
	ExitBaseline_id_expr(c *Baseline_id_exprContext)

	// ExitSql_id_expr is called when exiting the sql_id_expr production.
	ExitSql_id_expr(c *Sql_id_exprContext)

	// ExitBaseline_asgn_factor is called when exiting the baseline_asgn_factor production.
	ExitBaseline_asgn_factor(c *Baseline_asgn_factorContext)

	// ExitTenant_name is called when exiting the tenant_name production.
	ExitTenant_name(c *Tenant_nameContext)

	// ExitCache_name is called when exiting the cache_name production.
	ExitCache_name(c *Cache_nameContext)

	// ExitFile_id is called when exiting the file_id production.
	ExitFile_id(c *File_idContext)

	// ExitCancel_task_type is called when exiting the cancel_task_type production.
	ExitCancel_task_type(c *Cancel_task_typeContext)

	// ExitAlter_system_set_parameter_actions is called when exiting the alter_system_set_parameter_actions production.
	ExitAlter_system_set_parameter_actions(c *Alter_system_set_parameter_actionsContext)

	// ExitAlter_system_set_parameter_action is called when exiting the alter_system_set_parameter_action production.
	ExitAlter_system_set_parameter_action(c *Alter_system_set_parameter_actionContext)

	// ExitAlter_system_settp_actions is called when exiting the alter_system_settp_actions production.
	ExitAlter_system_settp_actions(c *Alter_system_settp_actionsContext)

	// ExitSettp_option is called when exiting the settp_option production.
	ExitSettp_option(c *Settp_optionContext)

	// ExitCluster_role is called when exiting the cluster_role production.
	ExitCluster_role(c *Cluster_roleContext)

	// ExitPartition_role is called when exiting the partition_role production.
	ExitPartition_role(c *Partition_roleContext)

	// ExitUpgrade_action is called when exiting the upgrade_action production.
	ExitUpgrade_action(c *Upgrade_actionContext)

	// ExitSet_names_stmt is called when exiting the set_names_stmt production.
	ExitSet_names_stmt(c *Set_names_stmtContext)

	// ExitSet_charset_stmt is called when exiting the set_charset_stmt production.
	ExitSet_charset_stmt(c *Set_charset_stmtContext)

	// ExitSet_transaction_stmt is called when exiting the set_transaction_stmt production.
	ExitSet_transaction_stmt(c *Set_transaction_stmtContext)

	// ExitTransaction_characteristics is called when exiting the transaction_characteristics production.
	ExitTransaction_characteristics(c *Transaction_characteristicsContext)

	// ExitTransaction_access_mode is called when exiting the transaction_access_mode production.
	ExitTransaction_access_mode(c *Transaction_access_modeContext)

	// ExitIsolation_level is called when exiting the isolation_level production.
	ExitIsolation_level(c *Isolation_levelContext)

	// ExitCreate_savepoint_stmt is called when exiting the create_savepoint_stmt production.
	ExitCreate_savepoint_stmt(c *Create_savepoint_stmtContext)

	// ExitRollback_savepoint_stmt is called when exiting the rollback_savepoint_stmt production.
	ExitRollback_savepoint_stmt(c *Rollback_savepoint_stmtContext)

	// ExitRelease_savepoint_stmt is called when exiting the release_savepoint_stmt production.
	ExitRelease_savepoint_stmt(c *Release_savepoint_stmtContext)

	// ExitAlter_cluster_stmt is called when exiting the alter_cluster_stmt production.
	ExitAlter_cluster_stmt(c *Alter_cluster_stmtContext)

	// ExitCluster_action is called when exiting the cluster_action production.
	ExitCluster_action(c *Cluster_actionContext)

	// ExitSwitchover_cluster_stmt is called when exiting the switchover_cluster_stmt production.
	ExitSwitchover_cluster_stmt(c *Switchover_cluster_stmtContext)

	// ExitCommit_switchover_clause is called when exiting the commit_switchover_clause production.
	ExitCommit_switchover_clause(c *Commit_switchover_clauseContext)

	// ExitCluster_name is called when exiting the cluster_name production.
	ExitCluster_name(c *Cluster_nameContext)

	// ExitVar_name is called when exiting the var_name production.
	ExitVar_name(c *Var_nameContext)

	// ExitColumn_name is called when exiting the column_name production.
	ExitColumn_name(c *Column_nameContext)

	// ExitRelation_name is called when exiting the relation_name production.
	ExitRelation_name(c *Relation_nameContext)

	// ExitFunction_name is called when exiting the function_name production.
	ExitFunction_name(c *Function_nameContext)

	// ExitColumn_label is called when exiting the column_label production.
	ExitColumn_label(c *Column_labelContext)

	// ExitDate_unit is called when exiting the date_unit production.
	ExitDate_unit(c *Date_unitContext)

	// ExitUnreserved_keyword is called when exiting the unreserved_keyword production.
	ExitUnreserved_keyword(c *Unreserved_keywordContext)

	// ExitUnreserved_keyword_normal is called when exiting the unreserved_keyword_normal production.
	ExitUnreserved_keyword_normal(c *Unreserved_keyword_normalContext)

	// ExitUnreserved_keyword_special is called when exiting the unreserved_keyword_special production.
	ExitUnreserved_keyword_special(c *Unreserved_keyword_specialContext)

	// ExitEmpty is called when exiting the empty production.
	ExitEmpty(c *EmptyContext)

	// ExitForward_expr is called when exiting the forward_expr production.
	ExitForward_expr(c *Forward_exprContext)

	// ExitForward_sql_stmt is called when exiting the forward_sql_stmt production.
	ExitForward_sql_stmt(c *Forward_sql_stmtContext)
}
