// Code generated from /work/obparser/obmysql/sql/OBParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package mysql // OBParser
import "github.com/antlr4-go/antlr/v4"


// A complete Visitor for a parse tree produced by OBParser.
type OBParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by OBParser#sql_stmt.
	VisitSql_stmt(ctx *Sql_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#stmt_list.
	VisitStmt_list(ctx *Stmt_listContext) interface{}

	// Visit a parse tree produced by OBParser#stmt.
	VisitStmt(ctx *StmtContext) interface{}

	// Visit a parse tree produced by OBParser#expr_list.
	VisitExpr_list(ctx *Expr_listContext) interface{}

	// Visit a parse tree produced by OBParser#expr_as_list.
	VisitExpr_as_list(ctx *Expr_as_listContext) interface{}

	// Visit a parse tree produced by OBParser#expr_with_opt_alias.
	VisitExpr_with_opt_alias(ctx *Expr_with_opt_aliasContext) interface{}

	// Visit a parse tree produced by OBParser#column_ref.
	VisitColumn_ref(ctx *Column_refContext) interface{}

	// Visit a parse tree produced by OBParser#complex_string_literal.
	VisitComplex_string_literal(ctx *Complex_string_literalContext) interface{}

	// Visit a parse tree produced by OBParser#charset_introducer.
	VisitCharset_introducer(ctx *Charset_introducerContext) interface{}

	// Visit a parse tree produced by OBParser#literal.
	VisitLiteral(ctx *LiteralContext) interface{}

	// Visit a parse tree produced by OBParser#number_literal.
	VisitNumber_literal(ctx *Number_literalContext) interface{}

	// Visit a parse tree produced by OBParser#expr_const.
	VisitExpr_const(ctx *Expr_constContext) interface{}

	// Visit a parse tree produced by OBParser#conf_const.
	VisitConf_const(ctx *Conf_constContext) interface{}

	// Visit a parse tree produced by OBParser#global_or_session_alias.
	VisitGlobal_or_session_alias(ctx *Global_or_session_aliasContext) interface{}

	// Visit a parse tree produced by OBParser#bool_pri.
	VisitBool_pri(ctx *Bool_priContext) interface{}

	// Visit a parse tree produced by OBParser#predicate.
	VisitPredicate(ctx *PredicateContext) interface{}

	// Visit a parse tree produced by OBParser#bit_expr.
	VisitBit_expr(ctx *Bit_exprContext) interface{}

	// Visit a parse tree produced by OBParser#simple_expr.
	VisitSimple_expr(ctx *Simple_exprContext) interface{}

	// Visit a parse tree produced by OBParser#expr.
	VisitExpr(ctx *ExprContext) interface{}

	// Visit a parse tree produced by OBParser#not.
	VisitNot(ctx *NotContext) interface{}

	// Visit a parse tree produced by OBParser#not2.
	VisitNot2(ctx *Not2Context) interface{}

	// Visit a parse tree produced by OBParser#sub_query_flag.
	VisitSub_query_flag(ctx *Sub_query_flagContext) interface{}

	// Visit a parse tree produced by OBParser#in_expr.
	VisitIn_expr(ctx *In_exprContext) interface{}

	// Visit a parse tree produced by OBParser#case_expr.
	VisitCase_expr(ctx *Case_exprContext) interface{}

	// Visit a parse tree produced by OBParser#window_function.
	VisitWindow_function(ctx *Window_functionContext) interface{}

	// Visit a parse tree produced by OBParser#first_or_last.
	VisitFirst_or_last(ctx *First_or_lastContext) interface{}

	// Visit a parse tree produced by OBParser#respect_or_ignore.
	VisitRespect_or_ignore(ctx *Respect_or_ignoreContext) interface{}

	// Visit a parse tree produced by OBParser#win_fun_first_last_params.
	VisitWin_fun_first_last_params(ctx *Win_fun_first_last_paramsContext) interface{}

	// Visit a parse tree produced by OBParser#win_fun_lead_lag_params.
	VisitWin_fun_lead_lag_params(ctx *Win_fun_lead_lag_paramsContext) interface{}

	// Visit a parse tree produced by OBParser#new_generalized_window_clause.
	VisitNew_generalized_window_clause(ctx *New_generalized_window_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#new_generalized_window_clause_with_blanket.
	VisitNew_generalized_window_clause_with_blanket(ctx *New_generalized_window_clause_with_blanketContext) interface{}

	// Visit a parse tree produced by OBParser#named_windows.
	VisitNamed_windows(ctx *Named_windowsContext) interface{}

	// Visit a parse tree produced by OBParser#named_window.
	VisitNamed_window(ctx *Named_windowContext) interface{}

	// Visit a parse tree produced by OBParser#generalized_window_clause.
	VisitGeneralized_window_clause(ctx *Generalized_window_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#win_rows_or_range.
	VisitWin_rows_or_range(ctx *Win_rows_or_rangeContext) interface{}

	// Visit a parse tree produced by OBParser#win_preceding_or_following.
	VisitWin_preceding_or_following(ctx *Win_preceding_or_followingContext) interface{}

	// Visit a parse tree produced by OBParser#win_interval.
	VisitWin_interval(ctx *Win_intervalContext) interface{}

	// Visit a parse tree produced by OBParser#win_bounding.
	VisitWin_bounding(ctx *Win_boundingContext) interface{}

	// Visit a parse tree produced by OBParser#win_window.
	VisitWin_window(ctx *Win_windowContext) interface{}

	// Visit a parse tree produced by OBParser#case_arg.
	VisitCase_arg(ctx *Case_argContext) interface{}

	// Visit a parse tree produced by OBParser#when_clause_list.
	VisitWhen_clause_list(ctx *When_clause_listContext) interface{}

	// Visit a parse tree produced by OBParser#when_clause.
	VisitWhen_clause(ctx *When_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#case_default.
	VisitCase_default(ctx *Case_defaultContext) interface{}

	// Visit a parse tree produced by OBParser#func_expr.
	VisitFunc_expr(ctx *Func_exprContext) interface{}

	// Visit a parse tree produced by OBParser#sys_interval_func.
	VisitSys_interval_func(ctx *Sys_interval_funcContext) interface{}

	// Visit a parse tree produced by OBParser#utc_timestamp_func.
	VisitUtc_timestamp_func(ctx *Utc_timestamp_funcContext) interface{}

	// Visit a parse tree produced by OBParser#sysdate_func.
	VisitSysdate_func(ctx *Sysdate_funcContext) interface{}

	// Visit a parse tree produced by OBParser#cur_timestamp_func.
	VisitCur_timestamp_func(ctx *Cur_timestamp_funcContext) interface{}

	// Visit a parse tree produced by OBParser#now_synonyms_func.
	VisitNow_synonyms_func(ctx *Now_synonyms_funcContext) interface{}

	// Visit a parse tree produced by OBParser#cur_time_func.
	VisitCur_time_func(ctx *Cur_time_funcContext) interface{}

	// Visit a parse tree produced by OBParser#cur_date_func.
	VisitCur_date_func(ctx *Cur_date_funcContext) interface{}

	// Visit a parse tree produced by OBParser#substr_or_substring.
	VisitSubstr_or_substring(ctx *Substr_or_substringContext) interface{}

	// Visit a parse tree produced by OBParser#substr_params.
	VisitSubstr_params(ctx *Substr_paramsContext) interface{}

	// Visit a parse tree produced by OBParser#date_params.
	VisitDate_params(ctx *Date_paramsContext) interface{}

	// Visit a parse tree produced by OBParser#timestamp_params.
	VisitTimestamp_params(ctx *Timestamp_paramsContext) interface{}

	// Visit a parse tree produced by OBParser#delete_stmt.
	VisitDelete_stmt(ctx *Delete_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#multi_delete_table.
	VisitMulti_delete_table(ctx *Multi_delete_tableContext) interface{}

	// Visit a parse tree produced by OBParser#update_stmt.
	VisitUpdate_stmt(ctx *Update_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#update_asgn_list.
	VisitUpdate_asgn_list(ctx *Update_asgn_listContext) interface{}

	// Visit a parse tree produced by OBParser#update_asgn_factor.
	VisitUpdate_asgn_factor(ctx *Update_asgn_factorContext) interface{}

	// Visit a parse tree produced by OBParser#create_resource_stmt.
	VisitCreate_resource_stmt(ctx *Create_resource_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#opt_resource_unit_option_list.
	VisitOpt_resource_unit_option_list(ctx *Opt_resource_unit_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#resource_unit_option.
	VisitResource_unit_option(ctx *Resource_unit_optionContext) interface{}

	// Visit a parse tree produced by OBParser#opt_create_resource_pool_option_list.
	VisitOpt_create_resource_pool_option_list(ctx *Opt_create_resource_pool_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#create_resource_pool_option.
	VisitCreate_resource_pool_option(ctx *Create_resource_pool_optionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_resource_pool_option_list.
	VisitAlter_resource_pool_option_list(ctx *Alter_resource_pool_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#unit_id_list.
	VisitUnit_id_list(ctx *Unit_id_listContext) interface{}

	// Visit a parse tree produced by OBParser#alter_resource_pool_option.
	VisitAlter_resource_pool_option(ctx *Alter_resource_pool_optionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_resource_stmt.
	VisitAlter_resource_stmt(ctx *Alter_resource_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#drop_resource_stmt.
	VisitDrop_resource_stmt(ctx *Drop_resource_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#create_tenant_stmt.
	VisitCreate_tenant_stmt(ctx *Create_tenant_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#opt_tenant_option_list.
	VisitOpt_tenant_option_list(ctx *Opt_tenant_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#tenant_option.
	VisitTenant_option(ctx *Tenant_optionContext) interface{}

	// Visit a parse tree produced by OBParser#zone_list.
	VisitZone_list(ctx *Zone_listContext) interface{}

	// Visit a parse tree produced by OBParser#resource_pool_list.
	VisitResource_pool_list(ctx *Resource_pool_listContext) interface{}

	// Visit a parse tree produced by OBParser#alter_tenant_stmt.
	VisitAlter_tenant_stmt(ctx *Alter_tenant_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#drop_tenant_stmt.
	VisitDrop_tenant_stmt(ctx *Drop_tenant_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#create_database_stmt.
	VisitCreate_database_stmt(ctx *Create_database_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#database_key.
	VisitDatabase_key(ctx *Database_keyContext) interface{}

	// Visit a parse tree produced by OBParser#database_factor.
	VisitDatabase_factor(ctx *Database_factorContext) interface{}

	// Visit a parse tree produced by OBParser#database_option_list.
	VisitDatabase_option_list(ctx *Database_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#charset_key.
	VisitCharset_key(ctx *Charset_keyContext) interface{}

	// Visit a parse tree produced by OBParser#database_option.
	VisitDatabase_option(ctx *Database_optionContext) interface{}

	// Visit a parse tree produced by OBParser#read_only_or_write.
	VisitRead_only_or_write(ctx *Read_only_or_writeContext) interface{}

	// Visit a parse tree produced by OBParser#drop_database_stmt.
	VisitDrop_database_stmt(ctx *Drop_database_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#alter_database_stmt.
	VisitAlter_database_stmt(ctx *Alter_database_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#load_data_stmt.
	VisitLoad_data_stmt(ctx *Load_data_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#load_data_with_opt_hint.
	VisitLoad_data_with_opt_hint(ctx *Load_data_with_opt_hintContext) interface{}

	// Visit a parse tree produced by OBParser#lines_or_rows.
	VisitLines_or_rows(ctx *Lines_or_rowsContext) interface{}

	// Visit a parse tree produced by OBParser#field_or_vars_list.
	VisitField_or_vars_list(ctx *Field_or_vars_listContext) interface{}

	// Visit a parse tree produced by OBParser#field_or_vars.
	VisitField_or_vars(ctx *Field_or_varsContext) interface{}

	// Visit a parse tree produced by OBParser#load_set_list.
	VisitLoad_set_list(ctx *Load_set_listContext) interface{}

	// Visit a parse tree produced by OBParser#load_set_element.
	VisitLoad_set_element(ctx *Load_set_elementContext) interface{}

	// Visit a parse tree produced by OBParser#use_database_stmt.
	VisitUse_database_stmt(ctx *Use_database_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#create_synonym_stmt.
	VisitCreate_synonym_stmt(ctx *Create_synonym_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#synonym_name.
	VisitSynonym_name(ctx *Synonym_nameContext) interface{}

	// Visit a parse tree produced by OBParser#synonym_object.
	VisitSynonym_object(ctx *Synonym_objectContext) interface{}

	// Visit a parse tree produced by OBParser#drop_synonym_stmt.
	VisitDrop_synonym_stmt(ctx *Drop_synonym_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#temporary_option.
	VisitTemporary_option(ctx *Temporary_optionContext) interface{}

	// Visit a parse tree produced by OBParser#create_table_like_stmt.
	VisitCreate_table_like_stmt(ctx *Create_table_like_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#create_table_stmt.
	VisitCreate_table_stmt(ctx *Create_table_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#ret_type.
	VisitRet_type(ctx *Ret_typeContext) interface{}

	// Visit a parse tree produced by OBParser#create_function_stmt.
	VisitCreate_function_stmt(ctx *Create_function_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#drop_function_stmt.
	VisitDrop_function_stmt(ctx *Drop_function_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#table_element_list.
	VisitTable_element_list(ctx *Table_element_listContext) interface{}

	// Visit a parse tree produced by OBParser#table_element.
	VisitTable_element(ctx *Table_elementContext) interface{}

	// Visit a parse tree produced by OBParser#opt_reference_option_list.
	VisitOpt_reference_option_list(ctx *Opt_reference_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#reference_option.
	VisitReference_option(ctx *Reference_optionContext) interface{}

	// Visit a parse tree produced by OBParser#reference_action.
	VisitReference_action(ctx *Reference_actionContext) interface{}

	// Visit a parse tree produced by OBParser#match_action.
	VisitMatch_action(ctx *Match_actionContext) interface{}

	// Visit a parse tree produced by OBParser#column_definition.
	VisitColumn_definition(ctx *Column_definitionContext) interface{}

	// Visit a parse tree produced by OBParser#opt_generated_column_attribute_list.
	VisitOpt_generated_column_attribute_list(ctx *Opt_generated_column_attribute_listContext) interface{}

	// Visit a parse tree produced by OBParser#generated_column_attribute.
	VisitGenerated_column_attribute(ctx *Generated_column_attributeContext) interface{}

	// Visit a parse tree produced by OBParser#column_definition_ref.
	VisitColumn_definition_ref(ctx *Column_definition_refContext) interface{}

	// Visit a parse tree produced by OBParser#column_definition_list.
	VisitColumn_definition_list(ctx *Column_definition_listContext) interface{}

	// Visit a parse tree produced by OBParser#cast_data_type.
	VisitCast_data_type(ctx *Cast_data_typeContext) interface{}

	// Visit a parse tree produced by OBParser#cast_datetime_type_i.
	VisitCast_datetime_type_i(ctx *Cast_datetime_type_iContext) interface{}

	// Visit a parse tree produced by OBParser#data_type.
	VisitData_type(ctx *Data_typeContext) interface{}

	// Visit a parse tree produced by OBParser#string_list.
	VisitString_list(ctx *String_listContext) interface{}

	// Visit a parse tree produced by OBParser#text_string.
	VisitText_string(ctx *Text_stringContext) interface{}

	// Visit a parse tree produced by OBParser#int_type_i.
	VisitInt_type_i(ctx *Int_type_iContext) interface{}

	// Visit a parse tree produced by OBParser#float_type_i.
	VisitFloat_type_i(ctx *Float_type_iContext) interface{}

	// Visit a parse tree produced by OBParser#datetime_type_i.
	VisitDatetime_type_i(ctx *Datetime_type_iContext) interface{}

	// Visit a parse tree produced by OBParser#date_year_type_i.
	VisitDate_year_type_i(ctx *Date_year_type_iContext) interface{}

	// Visit a parse tree produced by OBParser#text_type_i.
	VisitText_type_i(ctx *Text_type_iContext) interface{}

	// Visit a parse tree produced by OBParser#blob_type_i.
	VisitBlob_type_i(ctx *Blob_type_iContext) interface{}

	// Visit a parse tree produced by OBParser#string_length_i.
	VisitString_length_i(ctx *String_length_iContext) interface{}

	// Visit a parse tree produced by OBParser#collation_name.
	VisitCollation_name(ctx *Collation_nameContext) interface{}

	// Visit a parse tree produced by OBParser#trans_param_name.
	VisitTrans_param_name(ctx *Trans_param_nameContext) interface{}

	// Visit a parse tree produced by OBParser#trans_param_value.
	VisitTrans_param_value(ctx *Trans_param_valueContext) interface{}

	// Visit a parse tree produced by OBParser#charset_name.
	VisitCharset_name(ctx *Charset_nameContext) interface{}

	// Visit a parse tree produced by OBParser#charset_name_or_default.
	VisitCharset_name_or_default(ctx *Charset_name_or_defaultContext) interface{}

	// Visit a parse tree produced by OBParser#collation.
	VisitCollation(ctx *CollationContext) interface{}

	// Visit a parse tree produced by OBParser#opt_column_attribute_list.
	VisitOpt_column_attribute_list(ctx *Opt_column_attribute_listContext) interface{}

	// Visit a parse tree produced by OBParser#column_attribute.
	VisitColumn_attribute(ctx *Column_attributeContext) interface{}

	// Visit a parse tree produced by OBParser#now_or_signed_literal.
	VisitNow_or_signed_literal(ctx *Now_or_signed_literalContext) interface{}

	// Visit a parse tree produced by OBParser#signed_literal.
	VisitSigned_literal(ctx *Signed_literalContext) interface{}

	// Visit a parse tree produced by OBParser#opt_comma.
	VisitOpt_comma(ctx *Opt_commaContext) interface{}

	// Visit a parse tree produced by OBParser#table_option_list_space_seperated.
	VisitTable_option_list_space_seperated(ctx *Table_option_list_space_seperatedContext) interface{}

	// Visit a parse tree produced by OBParser#table_option_list.
	VisitTable_option_list(ctx *Table_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#primary_zone_name.
	VisitPrimary_zone_name(ctx *Primary_zone_nameContext) interface{}

	// Visit a parse tree produced by OBParser#tablespace.
	VisitTablespace(ctx *TablespaceContext) interface{}

	// Visit a parse tree produced by OBParser#locality_name.
	VisitLocality_name(ctx *Locality_nameContext) interface{}

	// Visit a parse tree produced by OBParser#table_option.
	VisitTable_option(ctx *Table_optionContext) interface{}

	// Visit a parse tree produced by OBParser#relation_name_or_string.
	VisitRelation_name_or_string(ctx *Relation_name_or_stringContext) interface{}

	// Visit a parse tree produced by OBParser#opt_equal_mark.
	VisitOpt_equal_mark(ctx *Opt_equal_markContext) interface{}

	// Visit a parse tree produced by OBParser#partition_option.
	VisitPartition_option(ctx *Partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#opt_partition_option.
	VisitOpt_partition_option(ctx *Opt_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#hash_partition_option.
	VisitHash_partition_option(ctx *Hash_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#list_partition_option.
	VisitList_partition_option(ctx *List_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#key_partition_option.
	VisitKey_partition_option(ctx *Key_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#range_partition_option.
	VisitRange_partition_option(ctx *Range_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#opt_column_partition_option.
	VisitOpt_column_partition_option(ctx *Opt_column_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#column_partition_option.
	VisitColumn_partition_option(ctx *Column_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#aux_column_list.
	VisitAux_column_list(ctx *Aux_column_listContext) interface{}

	// Visit a parse tree produced by OBParser#vertical_column_name.
	VisitVertical_column_name(ctx *Vertical_column_nameContext) interface{}

	// Visit a parse tree produced by OBParser#column_name_list.
	VisitColumn_name_list(ctx *Column_name_listContext) interface{}

	// Visit a parse tree produced by OBParser#subpartition_option.
	VisitSubpartition_option(ctx *Subpartition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#opt_list_partition_list.
	VisitOpt_list_partition_list(ctx *Opt_list_partition_listContext) interface{}

	// Visit a parse tree produced by OBParser#opt_list_subpartition_list.
	VisitOpt_list_subpartition_list(ctx *Opt_list_subpartition_listContext) interface{}

	// Visit a parse tree produced by OBParser#opt_range_partition_list.
	VisitOpt_range_partition_list(ctx *Opt_range_partition_listContext) interface{}

	// Visit a parse tree produced by OBParser#opt_range_subpartition_list.
	VisitOpt_range_subpartition_list(ctx *Opt_range_subpartition_listContext) interface{}

	// Visit a parse tree produced by OBParser#list_partition_list.
	VisitList_partition_list(ctx *List_partition_listContext) interface{}

	// Visit a parse tree produced by OBParser#list_subpartition_list.
	VisitList_subpartition_list(ctx *List_subpartition_listContext) interface{}

	// Visit a parse tree produced by OBParser#list_subpartition_element.
	VisitList_subpartition_element(ctx *List_subpartition_elementContext) interface{}

	// Visit a parse tree produced by OBParser#list_partition_element.
	VisitList_partition_element(ctx *List_partition_elementContext) interface{}

	// Visit a parse tree produced by OBParser#list_partition_expr.
	VisitList_partition_expr(ctx *List_partition_exprContext) interface{}

	// Visit a parse tree produced by OBParser#list_expr.
	VisitList_expr(ctx *List_exprContext) interface{}

	// Visit a parse tree produced by OBParser#range_partition_list.
	VisitRange_partition_list(ctx *Range_partition_listContext) interface{}

	// Visit a parse tree produced by OBParser#range_partition_element.
	VisitRange_partition_element(ctx *Range_partition_elementContext) interface{}

	// Visit a parse tree produced by OBParser#range_subpartition_element.
	VisitRange_subpartition_element(ctx *Range_subpartition_elementContext) interface{}

	// Visit a parse tree produced by OBParser#range_subpartition_list.
	VisitRange_subpartition_list(ctx *Range_subpartition_listContext) interface{}

	// Visit a parse tree produced by OBParser#range_partition_expr.
	VisitRange_partition_expr(ctx *Range_partition_exprContext) interface{}

	// Visit a parse tree produced by OBParser#range_expr_list.
	VisitRange_expr_list(ctx *Range_expr_listContext) interface{}

	// Visit a parse tree produced by OBParser#range_expr.
	VisitRange_expr(ctx *Range_exprContext) interface{}

	// Visit a parse tree produced by OBParser#int_or_decimal.
	VisitInt_or_decimal(ctx *Int_or_decimalContext) interface{}

	// Visit a parse tree produced by OBParser#tg_hash_partition_option.
	VisitTg_hash_partition_option(ctx *Tg_hash_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#tg_key_partition_option.
	VisitTg_key_partition_option(ctx *Tg_key_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#tg_range_partition_option.
	VisitTg_range_partition_option(ctx *Tg_range_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#tg_list_partition_option.
	VisitTg_list_partition_option(ctx *Tg_list_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#tg_subpartition_option.
	VisitTg_subpartition_option(ctx *Tg_subpartition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#row_format_option.
	VisitRow_format_option(ctx *Row_format_optionContext) interface{}

	// Visit a parse tree produced by OBParser#create_tablegroup_stmt.
	VisitCreate_tablegroup_stmt(ctx *Create_tablegroup_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#drop_tablegroup_stmt.
	VisitDrop_tablegroup_stmt(ctx *Drop_tablegroup_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#alter_tablegroup_stmt.
	VisitAlter_tablegroup_stmt(ctx *Alter_tablegroup_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#tablegroup_option_list_space_seperated.
	VisitTablegroup_option_list_space_seperated(ctx *Tablegroup_option_list_space_seperatedContext) interface{}

	// Visit a parse tree produced by OBParser#tablegroup_option_list.
	VisitTablegroup_option_list(ctx *Tablegroup_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#tablegroup_option.
	VisitTablegroup_option(ctx *Tablegroup_optionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_tablegroup_actions.
	VisitAlter_tablegroup_actions(ctx *Alter_tablegroup_actionsContext) interface{}

	// Visit a parse tree produced by OBParser#alter_tablegroup_action.
	VisitAlter_tablegroup_action(ctx *Alter_tablegroup_actionContext) interface{}

	// Visit a parse tree produced by OBParser#default_tablegroup.
	VisitDefault_tablegroup(ctx *Default_tablegroupContext) interface{}

	// Visit a parse tree produced by OBParser#create_view_stmt.
	VisitCreate_view_stmt(ctx *Create_view_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#view_select_stmt.
	VisitView_select_stmt(ctx *View_select_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#view_name.
	VisitView_name(ctx *View_nameContext) interface{}

	// Visit a parse tree produced by OBParser#create_index_stmt.
	VisitCreate_index_stmt(ctx *Create_index_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#index_name.
	VisitIndex_name(ctx *Index_nameContext) interface{}

	// Visit a parse tree produced by OBParser#opt_constraint_name.
	VisitOpt_constraint_name(ctx *Opt_constraint_nameContext) interface{}

	// Visit a parse tree produced by OBParser#constraint_name.
	VisitConstraint_name(ctx *Constraint_nameContext) interface{}

	// Visit a parse tree produced by OBParser#sort_column_list.
	VisitSort_column_list(ctx *Sort_column_listContext) interface{}

	// Visit a parse tree produced by OBParser#sort_column_key.
	VisitSort_column_key(ctx *Sort_column_keyContext) interface{}

	// Visit a parse tree produced by OBParser#opt_index_options.
	VisitOpt_index_options(ctx *Opt_index_optionsContext) interface{}

	// Visit a parse tree produced by OBParser#index_option.
	VisitIndex_option(ctx *Index_optionContext) interface{}

	// Visit a parse tree produced by OBParser#index_using_algorithm.
	VisitIndex_using_algorithm(ctx *Index_using_algorithmContext) interface{}

	// Visit a parse tree produced by OBParser#drop_table_stmt.
	VisitDrop_table_stmt(ctx *Drop_table_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#table_or_tables.
	VisitTable_or_tables(ctx *Table_or_tablesContext) interface{}

	// Visit a parse tree produced by OBParser#drop_view_stmt.
	VisitDrop_view_stmt(ctx *Drop_view_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#table_list.
	VisitTable_list(ctx *Table_listContext) interface{}

	// Visit a parse tree produced by OBParser#drop_index_stmt.
	VisitDrop_index_stmt(ctx *Drop_index_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#insert_stmt.
	VisitInsert_stmt(ctx *Insert_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#single_table_insert.
	VisitSingle_table_insert(ctx *Single_table_insertContext) interface{}

	// Visit a parse tree produced by OBParser#values_clause.
	VisitValues_clause(ctx *Values_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#value_or_values.
	VisitValue_or_values(ctx *Value_or_valuesContext) interface{}

	// Visit a parse tree produced by OBParser#replace_with_opt_hint.
	VisitReplace_with_opt_hint(ctx *Replace_with_opt_hintContext) interface{}

	// Visit a parse tree produced by OBParser#insert_with_opt_hint.
	VisitInsert_with_opt_hint(ctx *Insert_with_opt_hintContext) interface{}

	// Visit a parse tree produced by OBParser#column_list.
	VisitColumn_list(ctx *Column_listContext) interface{}

	// Visit a parse tree produced by OBParser#insert_vals_list.
	VisitInsert_vals_list(ctx *Insert_vals_listContext) interface{}

	// Visit a parse tree produced by OBParser#insert_vals.
	VisitInsert_vals(ctx *Insert_valsContext) interface{}

	// Visit a parse tree produced by OBParser#expr_or_default.
	VisitExpr_or_default(ctx *Expr_or_defaultContext) interface{}

	// Visit a parse tree produced by OBParser#select_stmt.
	VisitSelect_stmt(ctx *Select_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#select_into.
	VisitSelect_into(ctx *Select_intoContext) interface{}

	// Visit a parse tree produced by OBParser#select_with_parens.
	VisitSelect_with_parens(ctx *Select_with_parensContext) interface{}

	// Visit a parse tree produced by OBParser#select_no_parens.
	VisitSelect_no_parens(ctx *Select_no_parensContext) interface{}

	// Visit a parse tree produced by OBParser#no_table_select.
	VisitNo_table_select(ctx *No_table_selectContext) interface{}

	// Visit a parse tree produced by OBParser#select_clause.
	VisitSelect_clause(ctx *Select_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#select_clause_set_with_order_and_limit.
	VisitSelect_clause_set_with_order_and_limit(ctx *Select_clause_set_with_order_and_limitContext) interface{}

	// Visit a parse tree produced by OBParser#select_clause_set.
	VisitSelect_clause_set(ctx *Select_clause_setContext) interface{}

	// Visit a parse tree produced by OBParser#select_clause_set_right.
	VisitSelect_clause_set_right(ctx *Select_clause_set_rightContext) interface{}

	// Visit a parse tree produced by OBParser#select_clause_set_left.
	VisitSelect_clause_set_left(ctx *Select_clause_set_leftContext) interface{}

	// Visit a parse tree produced by OBParser#no_table_select_with_order_and_limit.
	VisitNo_table_select_with_order_and_limit(ctx *No_table_select_with_order_and_limitContext) interface{}

	// Visit a parse tree produced by OBParser#simple_select_with_order_and_limit.
	VisitSimple_select_with_order_and_limit(ctx *Simple_select_with_order_and_limitContext) interface{}

	// Visit a parse tree produced by OBParser#select_with_parens_with_order_and_limit.
	VisitSelect_with_parens_with_order_and_limit(ctx *Select_with_parens_with_order_and_limitContext) interface{}

	// Visit a parse tree produced by OBParser#select_with_opt_hint.
	VisitSelect_with_opt_hint(ctx *Select_with_opt_hintContext) interface{}

	// Visit a parse tree produced by OBParser#update_with_opt_hint.
	VisitUpdate_with_opt_hint(ctx *Update_with_opt_hintContext) interface{}

	// Visit a parse tree produced by OBParser#delete_with_opt_hint.
	VisitDelete_with_opt_hint(ctx *Delete_with_opt_hintContext) interface{}

	// Visit a parse tree produced by OBParser#simple_select.
	VisitSimple_select(ctx *Simple_selectContext) interface{}

	// Visit a parse tree produced by OBParser#set_type_union.
	VisitSet_type_union(ctx *Set_type_unionContext) interface{}

	// Visit a parse tree produced by OBParser#set_type_other.
	VisitSet_type_other(ctx *Set_type_otherContext) interface{}

	// Visit a parse tree produced by OBParser#set_type.
	VisitSet_type(ctx *Set_typeContext) interface{}

	// Visit a parse tree produced by OBParser#set_expression_option.
	VisitSet_expression_option(ctx *Set_expression_optionContext) interface{}

	// Visit a parse tree produced by OBParser#opt_hint_value.
	VisitOpt_hint_value(ctx *Opt_hint_valueContext) interface{}

	// Visit a parse tree produced by OBParser#limit_clause.
	VisitLimit_clause(ctx *Limit_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#into_clause.
	VisitInto_clause(ctx *Into_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#into_opt.
	VisitInto_opt(ctx *Into_optContext) interface{}

	// Visit a parse tree produced by OBParser#into_var_list.
	VisitInto_var_list(ctx *Into_var_listContext) interface{}

	// Visit a parse tree produced by OBParser#into_var.
	VisitInto_var(ctx *Into_varContext) interface{}

	// Visit a parse tree produced by OBParser#field_opt.
	VisitField_opt(ctx *Field_optContext) interface{}

	// Visit a parse tree produced by OBParser#field_term_list.
	VisitField_term_list(ctx *Field_term_listContext) interface{}

	// Visit a parse tree produced by OBParser#field_term.
	VisitField_term(ctx *Field_termContext) interface{}

	// Visit a parse tree produced by OBParser#line_opt.
	VisitLine_opt(ctx *Line_optContext) interface{}

	// Visit a parse tree produced by OBParser#line_term_list.
	VisitLine_term_list(ctx *Line_term_listContext) interface{}

	// Visit a parse tree produced by OBParser#line_term.
	VisitLine_term(ctx *Line_termContext) interface{}

	// Visit a parse tree produced by OBParser#hint_list_with_end.
	VisitHint_list_with_end(ctx *Hint_list_with_endContext) interface{}

	// Visit a parse tree produced by OBParser#opt_hint_list.
	VisitOpt_hint_list(ctx *Opt_hint_listContext) interface{}

	// Visit a parse tree produced by OBParser#hint_options.
	VisitHint_options(ctx *Hint_optionsContext) interface{}

	// Visit a parse tree produced by OBParser#name_list.
	VisitName_list(ctx *Name_listContext) interface{}

	// Visit a parse tree produced by OBParser#hint_option.
	VisitHint_option(ctx *Hint_optionContext) interface{}

	// Visit a parse tree produced by OBParser#consistency_level.
	VisitConsistency_level(ctx *Consistency_levelContext) interface{}

	// Visit a parse tree produced by OBParser#use_plan_cache_type.
	VisitUse_plan_cache_type(ctx *Use_plan_cache_typeContext) interface{}

	// Visit a parse tree produced by OBParser#use_jit_type.
	VisitUse_jit_type(ctx *Use_jit_typeContext) interface{}

	// Visit a parse tree produced by OBParser#distribute_method.
	VisitDistribute_method(ctx *Distribute_methodContext) interface{}

	// Visit a parse tree produced by OBParser#limit_expr.
	VisitLimit_expr(ctx *Limit_exprContext) interface{}

	// Visit a parse tree produced by OBParser#opt_for_update_wait.
	VisitOpt_for_update_wait(ctx *Opt_for_update_waitContext) interface{}

	// Visit a parse tree produced by OBParser#parameterized_trim.
	VisitParameterized_trim(ctx *Parameterized_trimContext) interface{}

	// Visit a parse tree produced by OBParser#groupby_clause.
	VisitGroupby_clause(ctx *Groupby_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#sort_list_for_group_by.
	VisitSort_list_for_group_by(ctx *Sort_list_for_group_byContext) interface{}

	// Visit a parse tree produced by OBParser#sort_key_for_group_by.
	VisitSort_key_for_group_by(ctx *Sort_key_for_group_byContext) interface{}

	// Visit a parse tree produced by OBParser#order_by.
	VisitOrder_by(ctx *Order_byContext) interface{}

	// Visit a parse tree produced by OBParser#sort_list.
	VisitSort_list(ctx *Sort_listContext) interface{}

	// Visit a parse tree produced by OBParser#sort_key.
	VisitSort_key(ctx *Sort_keyContext) interface{}

	// Visit a parse tree produced by OBParser#query_expression_option_list.
	VisitQuery_expression_option_list(ctx *Query_expression_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#query_expression_option.
	VisitQuery_expression_option(ctx *Query_expression_optionContext) interface{}

	// Visit a parse tree produced by OBParser#projection.
	VisitProjection(ctx *ProjectionContext) interface{}

	// Visit a parse tree produced by OBParser#select_expr_list.
	VisitSelect_expr_list(ctx *Select_expr_listContext) interface{}

	// Visit a parse tree produced by OBParser#from_list.
	VisitFrom_list(ctx *From_listContext) interface{}

	// Visit a parse tree produced by OBParser#table_references.
	VisitTable_references(ctx *Table_referencesContext) interface{}

	// Visit a parse tree produced by OBParser#table_reference.
	VisitTable_reference(ctx *Table_referenceContext) interface{}

	// Visit a parse tree produced by OBParser#table_factor.
	VisitTable_factor(ctx *Table_factorContext) interface{}

	// Visit a parse tree produced by OBParser#tbl_name.
	VisitTbl_name(ctx *Tbl_nameContext) interface{}

	// Visit a parse tree produced by OBParser#dml_table_name.
	VisitDml_table_name(ctx *Dml_table_nameContext) interface{}

	// Visit a parse tree produced by OBParser#seed.
	VisitSeed(ctx *SeedContext) interface{}

	// Visit a parse tree produced by OBParser#opt_seed.
	VisitOpt_seed(ctx *Opt_seedContext) interface{}

	// Visit a parse tree produced by OBParser#sample_percent.
	VisitSample_percent(ctx *Sample_percentContext) interface{}

	// Visit a parse tree produced by OBParser#sample_clause.
	VisitSample_clause(ctx *Sample_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#table_subquery.
	VisitTable_subquery(ctx *Table_subqueryContext) interface{}

	// Visit a parse tree produced by OBParser#use_partition.
	VisitUse_partition(ctx *Use_partitionContext) interface{}

	// Visit a parse tree produced by OBParser#index_hint_type.
	VisitIndex_hint_type(ctx *Index_hint_typeContext) interface{}

	// Visit a parse tree produced by OBParser#key_or_index.
	VisitKey_or_index(ctx *Key_or_indexContext) interface{}

	// Visit a parse tree produced by OBParser#index_hint_scope.
	VisitIndex_hint_scope(ctx *Index_hint_scopeContext) interface{}

	// Visit a parse tree produced by OBParser#index_element.
	VisitIndex_element(ctx *Index_elementContext) interface{}

	// Visit a parse tree produced by OBParser#index_list.
	VisitIndex_list(ctx *Index_listContext) interface{}

	// Visit a parse tree produced by OBParser#index_hint_definition.
	VisitIndex_hint_definition(ctx *Index_hint_definitionContext) interface{}

	// Visit a parse tree produced by OBParser#index_hint_list.
	VisitIndex_hint_list(ctx *Index_hint_listContext) interface{}

	// Visit a parse tree produced by OBParser#relation_factor.
	VisitRelation_factor(ctx *Relation_factorContext) interface{}

	// Visit a parse tree produced by OBParser#relation_with_star_list.
	VisitRelation_with_star_list(ctx *Relation_with_star_listContext) interface{}

	// Visit a parse tree produced by OBParser#relation_factor_with_star.
	VisitRelation_factor_with_star(ctx *Relation_factor_with_starContext) interface{}

	// Visit a parse tree produced by OBParser#normal_relation_factor.
	VisitNormal_relation_factor(ctx *Normal_relation_factorContext) interface{}

	// Visit a parse tree produced by OBParser#dot_relation_factor.
	VisitDot_relation_factor(ctx *Dot_relation_factorContext) interface{}

	// Visit a parse tree produced by OBParser#relation_factor_in_hint.
	VisitRelation_factor_in_hint(ctx *Relation_factor_in_hintContext) interface{}

	// Visit a parse tree produced by OBParser#qb_name_option.
	VisitQb_name_option(ctx *Qb_name_optionContext) interface{}

	// Visit a parse tree produced by OBParser#relation_factor_in_hint_list.
	VisitRelation_factor_in_hint_list(ctx *Relation_factor_in_hint_listContext) interface{}

	// Visit a parse tree produced by OBParser#relation_sep_option.
	VisitRelation_sep_option(ctx *Relation_sep_optionContext) interface{}

	// Visit a parse tree produced by OBParser#relation_factor_in_pq_hint.
	VisitRelation_factor_in_pq_hint(ctx *Relation_factor_in_pq_hintContext) interface{}

	// Visit a parse tree produced by OBParser#relation_factor_in_leading_hint.
	VisitRelation_factor_in_leading_hint(ctx *Relation_factor_in_leading_hintContext) interface{}

	// Visit a parse tree produced by OBParser#relation_factor_in_leading_hint_list.
	VisitRelation_factor_in_leading_hint_list(ctx *Relation_factor_in_leading_hint_listContext) interface{}

	// Visit a parse tree produced by OBParser#relation_factor_in_leading_hint_list_entry.
	VisitRelation_factor_in_leading_hint_list_entry(ctx *Relation_factor_in_leading_hint_list_entryContext) interface{}

	// Visit a parse tree produced by OBParser#relation_factor_in_use_join_hint_list.
	VisitRelation_factor_in_use_join_hint_list(ctx *Relation_factor_in_use_join_hint_listContext) interface{}

	// Visit a parse tree produced by OBParser#tracing_num_list.
	VisitTracing_num_list(ctx *Tracing_num_listContext) interface{}

	// Visit a parse tree produced by OBParser#join_condition.
	VisitJoin_condition(ctx *Join_conditionContext) interface{}

	// Visit a parse tree produced by OBParser#joined_table.
	VisitJoined_table(ctx *Joined_tableContext) interface{}

	// Visit a parse tree produced by OBParser#natural_join_type.
	VisitNatural_join_type(ctx *Natural_join_typeContext) interface{}

	// Visit a parse tree produced by OBParser#inner_join_type.
	VisitInner_join_type(ctx *Inner_join_typeContext) interface{}

	// Visit a parse tree produced by OBParser#outer_join_type.
	VisitOuter_join_type(ctx *Outer_join_typeContext) interface{}

	// Visit a parse tree produced by OBParser#analyze_stmt.
	VisitAnalyze_stmt(ctx *Analyze_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#create_outline_stmt.
	VisitCreate_outline_stmt(ctx *Create_outline_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#alter_outline_stmt.
	VisitAlter_outline_stmt(ctx *Alter_outline_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#drop_outline_stmt.
	VisitDrop_outline_stmt(ctx *Drop_outline_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#explain_stmt.
	VisitExplain_stmt(ctx *Explain_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#explain_or_desc.
	VisitExplain_or_desc(ctx *Explain_or_descContext) interface{}

	// Visit a parse tree produced by OBParser#explainable_stmt.
	VisitExplainable_stmt(ctx *Explainable_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#format_name.
	VisitFormat_name(ctx *Format_nameContext) interface{}

	// Visit a parse tree produced by OBParser#show_stmt.
	VisitShow_stmt(ctx *Show_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#databases_or_schemas.
	VisitDatabases_or_schemas(ctx *Databases_or_schemasContext) interface{}

	// Visit a parse tree produced by OBParser#opt_for_grant_user.
	VisitOpt_for_grant_user(ctx *Opt_for_grant_userContext) interface{}

	// Visit a parse tree produced by OBParser#columns_or_fields.
	VisitColumns_or_fields(ctx *Columns_or_fieldsContext) interface{}

	// Visit a parse tree produced by OBParser#database_or_schema.
	VisitDatabase_or_schema(ctx *Database_or_schemaContext) interface{}

	// Visit a parse tree produced by OBParser#index_or_indexes_or_keys.
	VisitIndex_or_indexes_or_keys(ctx *Index_or_indexes_or_keysContext) interface{}

	// Visit a parse tree produced by OBParser#from_or_in.
	VisitFrom_or_in(ctx *From_or_inContext) interface{}

	// Visit a parse tree produced by OBParser#help_stmt.
	VisitHelp_stmt(ctx *Help_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#create_tablespace_stmt.
	VisitCreate_tablespace_stmt(ctx *Create_tablespace_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#permanent_tablespace.
	VisitPermanent_tablespace(ctx *Permanent_tablespaceContext) interface{}

	// Visit a parse tree produced by OBParser#permanent_tablespace_option.
	VisitPermanent_tablespace_option(ctx *Permanent_tablespace_optionContext) interface{}

	// Visit a parse tree produced by OBParser#drop_tablespace_stmt.
	VisitDrop_tablespace_stmt(ctx *Drop_tablespace_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#alter_tablespace_actions.
	VisitAlter_tablespace_actions(ctx *Alter_tablespace_actionsContext) interface{}

	// Visit a parse tree produced by OBParser#alter_tablespace_action.
	VisitAlter_tablespace_action(ctx *Alter_tablespace_actionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_tablespace_stmt.
	VisitAlter_tablespace_stmt(ctx *Alter_tablespace_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#rotate_master_key_stmt.
	VisitRotate_master_key_stmt(ctx *Rotate_master_key_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#permanent_tablespace_options.
	VisitPermanent_tablespace_options(ctx *Permanent_tablespace_optionsContext) interface{}

	// Visit a parse tree produced by OBParser#create_user_stmt.
	VisitCreate_user_stmt(ctx *Create_user_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#user_specification_list.
	VisitUser_specification_list(ctx *User_specification_listContext) interface{}

	// Visit a parse tree produced by OBParser#user_specification.
	VisitUser_specification(ctx *User_specificationContext) interface{}

	// Visit a parse tree produced by OBParser#require_specification.
	VisitRequire_specification(ctx *Require_specificationContext) interface{}

	// Visit a parse tree produced by OBParser#tls_option_list.
	VisitTls_option_list(ctx *Tls_option_listContext) interface{}

	// Visit a parse tree produced by OBParser#tls_option.
	VisitTls_option(ctx *Tls_optionContext) interface{}

	// Visit a parse tree produced by OBParser#user.
	VisitUser(ctx *UserContext) interface{}

	// Visit a parse tree produced by OBParser#opt_host_name.
	VisitOpt_host_name(ctx *Opt_host_nameContext) interface{}

	// Visit a parse tree produced by OBParser#user_with_host_name.
	VisitUser_with_host_name(ctx *User_with_host_nameContext) interface{}

	// Visit a parse tree produced by OBParser#password.
	VisitPassword(ctx *PasswordContext) interface{}

	// Visit a parse tree produced by OBParser#drop_user_stmt.
	VisitDrop_user_stmt(ctx *Drop_user_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#user_list.
	VisitUser_list(ctx *User_listContext) interface{}

	// Visit a parse tree produced by OBParser#set_password_stmt.
	VisitSet_password_stmt(ctx *Set_password_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#opt_for_user.
	VisitOpt_for_user(ctx *Opt_for_userContext) interface{}

	// Visit a parse tree produced by OBParser#rename_user_stmt.
	VisitRename_user_stmt(ctx *Rename_user_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#rename_info.
	VisitRename_info(ctx *Rename_infoContext) interface{}

	// Visit a parse tree produced by OBParser#rename_list.
	VisitRename_list(ctx *Rename_listContext) interface{}

	// Visit a parse tree produced by OBParser#lock_user_stmt.
	VisitLock_user_stmt(ctx *Lock_user_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#lock_spec_mysql57.
	VisitLock_spec_mysql57(ctx *Lock_spec_mysql57Context) interface{}

	// Visit a parse tree produced by OBParser#lock_tables_stmt.
	VisitLock_tables_stmt(ctx *Lock_tables_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#unlock_tables_stmt.
	VisitUnlock_tables_stmt(ctx *Unlock_tables_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#lock_table_list.
	VisitLock_table_list(ctx *Lock_table_listContext) interface{}

	// Visit a parse tree produced by OBParser#lock_table.
	VisitLock_table(ctx *Lock_tableContext) interface{}

	// Visit a parse tree produced by OBParser#lock_type.
	VisitLock_type(ctx *Lock_typeContext) interface{}

	// Visit a parse tree produced by OBParser#begin_stmt.
	VisitBegin_stmt(ctx *Begin_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#commit_stmt.
	VisitCommit_stmt(ctx *Commit_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#rollback_stmt.
	VisitRollback_stmt(ctx *Rollback_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#kill_stmt.
	VisitKill_stmt(ctx *Kill_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#grant_stmt.
	VisitGrant_stmt(ctx *Grant_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#grant_privileges.
	VisitGrant_privileges(ctx *Grant_privilegesContext) interface{}

	// Visit a parse tree produced by OBParser#priv_type_list.
	VisitPriv_type_list(ctx *Priv_type_listContext) interface{}

	// Visit a parse tree produced by OBParser#priv_type.
	VisitPriv_type(ctx *Priv_typeContext) interface{}

	// Visit a parse tree produced by OBParser#priv_level.
	VisitPriv_level(ctx *Priv_levelContext) interface{}

	// Visit a parse tree produced by OBParser#grant_options.
	VisitGrant_options(ctx *Grant_optionsContext) interface{}

	// Visit a parse tree produced by OBParser#revoke_stmt.
	VisitRevoke_stmt(ctx *Revoke_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#prepare_stmt.
	VisitPrepare_stmt(ctx *Prepare_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#stmt_name.
	VisitStmt_name(ctx *Stmt_nameContext) interface{}

	// Visit a parse tree produced by OBParser#preparable_stmt.
	VisitPreparable_stmt(ctx *Preparable_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#variable_set_stmt.
	VisitVariable_set_stmt(ctx *Variable_set_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#sys_var_and_val_list.
	VisitSys_var_and_val_list(ctx *Sys_var_and_val_listContext) interface{}

	// Visit a parse tree produced by OBParser#var_and_val_list.
	VisitVar_and_val_list(ctx *Var_and_val_listContext) interface{}

	// Visit a parse tree produced by OBParser#set_expr_or_default.
	VisitSet_expr_or_default(ctx *Set_expr_or_defaultContext) interface{}

	// Visit a parse tree produced by OBParser#var_and_val.
	VisitVar_and_val(ctx *Var_and_valContext) interface{}

	// Visit a parse tree produced by OBParser#sys_var_and_val.
	VisitSys_var_and_val(ctx *Sys_var_and_valContext) interface{}

	// Visit a parse tree produced by OBParser#scope_or_scope_alias.
	VisitScope_or_scope_alias(ctx *Scope_or_scope_aliasContext) interface{}

	// Visit a parse tree produced by OBParser#to_or_eq.
	VisitTo_or_eq(ctx *To_or_eqContext) interface{}

	// Visit a parse tree produced by OBParser#execute_stmt.
	VisitExecute_stmt(ctx *Execute_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#argument_list.
	VisitArgument_list(ctx *Argument_listContext) interface{}

	// Visit a parse tree produced by OBParser#argument.
	VisitArgument(ctx *ArgumentContext) interface{}

	// Visit a parse tree produced by OBParser#deallocate_prepare_stmt.
	VisitDeallocate_prepare_stmt(ctx *Deallocate_prepare_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#deallocate_or_drop.
	VisitDeallocate_or_drop(ctx *Deallocate_or_dropContext) interface{}

	// Visit a parse tree produced by OBParser#truncate_table_stmt.
	VisitTruncate_table_stmt(ctx *Truncate_table_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#rename_table_stmt.
	VisitRename_table_stmt(ctx *Rename_table_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#rename_table_actions.
	VisitRename_table_actions(ctx *Rename_table_actionsContext) interface{}

	// Visit a parse tree produced by OBParser#rename_table_action.
	VisitRename_table_action(ctx *Rename_table_actionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_table_stmt.
	VisitAlter_table_stmt(ctx *Alter_table_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#alter_table_actions.
	VisitAlter_table_actions(ctx *Alter_table_actionsContext) interface{}

	// Visit a parse tree produced by OBParser#alter_table_action.
	VisitAlter_table_action(ctx *Alter_table_actionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_constraint_option.
	VisitAlter_constraint_option(ctx *Alter_constraint_optionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_partition_option.
	VisitAlter_partition_option(ctx *Alter_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#opt_partition_range_or_list.
	VisitOpt_partition_range_or_list(ctx *Opt_partition_range_or_listContext) interface{}

	// Visit a parse tree produced by OBParser#alter_tg_partition_option.
	VisitAlter_tg_partition_option(ctx *Alter_tg_partition_optionContext) interface{}

	// Visit a parse tree produced by OBParser#drop_partition_name_list.
	VisitDrop_partition_name_list(ctx *Drop_partition_name_listContext) interface{}

	// Visit a parse tree produced by OBParser#modify_partition_info.
	VisitModify_partition_info(ctx *Modify_partition_infoContext) interface{}

	// Visit a parse tree produced by OBParser#modify_tg_partition_info.
	VisitModify_tg_partition_info(ctx *Modify_tg_partition_infoContext) interface{}

	// Visit a parse tree produced by OBParser#alter_index_option.
	VisitAlter_index_option(ctx *Alter_index_optionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_foreign_key_action.
	VisitAlter_foreign_key_action(ctx *Alter_foreign_key_actionContext) interface{}

	// Visit a parse tree produced by OBParser#visibility_option.
	VisitVisibility_option(ctx *Visibility_optionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_column_option.
	VisitAlter_column_option(ctx *Alter_column_optionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_tablegroup_option.
	VisitAlter_tablegroup_option(ctx *Alter_tablegroup_optionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_column_behavior.
	VisitAlter_column_behavior(ctx *Alter_column_behaviorContext) interface{}

	// Visit a parse tree produced by OBParser#flashback_stmt.
	VisitFlashback_stmt(ctx *Flashback_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#purge_stmt.
	VisitPurge_stmt(ctx *Purge_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#optimize_stmt.
	VisitOptimize_stmt(ctx *Optimize_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#dump_memory_stmt.
	VisitDump_memory_stmt(ctx *Dump_memory_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#alter_system_stmt.
	VisitAlter_system_stmt(ctx *Alter_system_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#change_tenant_name_or_tenant_id.
	VisitChange_tenant_name_or_tenant_id(ctx *Change_tenant_name_or_tenant_idContext) interface{}

	// Visit a parse tree produced by OBParser#cache_type.
	VisitCache_type(ctx *Cache_typeContext) interface{}

	// Visit a parse tree produced by OBParser#balance_task_type.
	VisitBalance_task_type(ctx *Balance_task_typeContext) interface{}

	// Visit a parse tree produced by OBParser#tenant_list_tuple.
	VisitTenant_list_tuple(ctx *Tenant_list_tupleContext) interface{}

	// Visit a parse tree produced by OBParser#tenant_name_list.
	VisitTenant_name_list(ctx *Tenant_name_listContext) interface{}

	// Visit a parse tree produced by OBParser#flush_scope.
	VisitFlush_scope(ctx *Flush_scopeContext) interface{}

	// Visit a parse tree produced by OBParser#server_info_list.
	VisitServer_info_list(ctx *Server_info_listContext) interface{}

	// Visit a parse tree produced by OBParser#server_info.
	VisitServer_info(ctx *Server_infoContext) interface{}

	// Visit a parse tree produced by OBParser#server_action.
	VisitServer_action(ctx *Server_actionContext) interface{}

	// Visit a parse tree produced by OBParser#server_list.
	VisitServer_list(ctx *Server_listContext) interface{}

	// Visit a parse tree produced by OBParser#zone_action.
	VisitZone_action(ctx *Zone_actionContext) interface{}

	// Visit a parse tree produced by OBParser#ip_port.
	VisitIp_port(ctx *Ip_portContext) interface{}

	// Visit a parse tree produced by OBParser#zone_desc.
	VisitZone_desc(ctx *Zone_descContext) interface{}

	// Visit a parse tree produced by OBParser#server_or_zone.
	VisitServer_or_zone(ctx *Server_or_zoneContext) interface{}

	// Visit a parse tree produced by OBParser#add_or_alter_zone_option.
	VisitAdd_or_alter_zone_option(ctx *Add_or_alter_zone_optionContext) interface{}

	// Visit a parse tree produced by OBParser#add_or_alter_zone_options.
	VisitAdd_or_alter_zone_options(ctx *Add_or_alter_zone_optionsContext) interface{}

	// Visit a parse tree produced by OBParser#alter_or_change_or_modify.
	VisitAlter_or_change_or_modify(ctx *Alter_or_change_or_modifyContext) interface{}

	// Visit a parse tree produced by OBParser#partition_id_desc.
	VisitPartition_id_desc(ctx *Partition_id_descContext) interface{}

	// Visit a parse tree produced by OBParser#partition_id_or_server_or_zone.
	VisitPartition_id_or_server_or_zone(ctx *Partition_id_or_server_or_zoneContext) interface{}

	// Visit a parse tree produced by OBParser#migrate_action.
	VisitMigrate_action(ctx *Migrate_actionContext) interface{}

	// Visit a parse tree produced by OBParser#change_actions.
	VisitChange_actions(ctx *Change_actionsContext) interface{}

	// Visit a parse tree produced by OBParser#change_action.
	VisitChange_action(ctx *Change_actionContext) interface{}

	// Visit a parse tree produced by OBParser#replica_type.
	VisitReplica_type(ctx *Replica_typeContext) interface{}

	// Visit a parse tree produced by OBParser#suspend_or_resume.
	VisitSuspend_or_resume(ctx *Suspend_or_resumeContext) interface{}

	// Visit a parse tree produced by OBParser#baseline_id_expr.
	VisitBaseline_id_expr(ctx *Baseline_id_exprContext) interface{}

	// Visit a parse tree produced by OBParser#sql_id_expr.
	VisitSql_id_expr(ctx *Sql_id_exprContext) interface{}

	// Visit a parse tree produced by OBParser#baseline_asgn_factor.
	VisitBaseline_asgn_factor(ctx *Baseline_asgn_factorContext) interface{}

	// Visit a parse tree produced by OBParser#tenant_name.
	VisitTenant_name(ctx *Tenant_nameContext) interface{}

	// Visit a parse tree produced by OBParser#cache_name.
	VisitCache_name(ctx *Cache_nameContext) interface{}

	// Visit a parse tree produced by OBParser#file_id.
	VisitFile_id(ctx *File_idContext) interface{}

	// Visit a parse tree produced by OBParser#cancel_task_type.
	VisitCancel_task_type(ctx *Cancel_task_typeContext) interface{}

	// Visit a parse tree produced by OBParser#alter_system_set_parameter_actions.
	VisitAlter_system_set_parameter_actions(ctx *Alter_system_set_parameter_actionsContext) interface{}

	// Visit a parse tree produced by OBParser#alter_system_set_parameter_action.
	VisitAlter_system_set_parameter_action(ctx *Alter_system_set_parameter_actionContext) interface{}

	// Visit a parse tree produced by OBParser#alter_system_settp_actions.
	VisitAlter_system_settp_actions(ctx *Alter_system_settp_actionsContext) interface{}

	// Visit a parse tree produced by OBParser#settp_option.
	VisitSettp_option(ctx *Settp_optionContext) interface{}

	// Visit a parse tree produced by OBParser#cluster_role.
	VisitCluster_role(ctx *Cluster_roleContext) interface{}

	// Visit a parse tree produced by OBParser#partition_role.
	VisitPartition_role(ctx *Partition_roleContext) interface{}

	// Visit a parse tree produced by OBParser#upgrade_action.
	VisitUpgrade_action(ctx *Upgrade_actionContext) interface{}

	// Visit a parse tree produced by OBParser#set_names_stmt.
	VisitSet_names_stmt(ctx *Set_names_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#set_charset_stmt.
	VisitSet_charset_stmt(ctx *Set_charset_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#set_transaction_stmt.
	VisitSet_transaction_stmt(ctx *Set_transaction_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#transaction_characteristics.
	VisitTransaction_characteristics(ctx *Transaction_characteristicsContext) interface{}

	// Visit a parse tree produced by OBParser#transaction_access_mode.
	VisitTransaction_access_mode(ctx *Transaction_access_modeContext) interface{}

	// Visit a parse tree produced by OBParser#isolation_level.
	VisitIsolation_level(ctx *Isolation_levelContext) interface{}

	// Visit a parse tree produced by OBParser#create_savepoint_stmt.
	VisitCreate_savepoint_stmt(ctx *Create_savepoint_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#rollback_savepoint_stmt.
	VisitRollback_savepoint_stmt(ctx *Rollback_savepoint_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#release_savepoint_stmt.
	VisitRelease_savepoint_stmt(ctx *Release_savepoint_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#alter_cluster_stmt.
	VisitAlter_cluster_stmt(ctx *Alter_cluster_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#cluster_action.
	VisitCluster_action(ctx *Cluster_actionContext) interface{}

	// Visit a parse tree produced by OBParser#switchover_cluster_stmt.
	VisitSwitchover_cluster_stmt(ctx *Switchover_cluster_stmtContext) interface{}

	// Visit a parse tree produced by OBParser#commit_switchover_clause.
	VisitCommit_switchover_clause(ctx *Commit_switchover_clauseContext) interface{}

	// Visit a parse tree produced by OBParser#cluster_name.
	VisitCluster_name(ctx *Cluster_nameContext) interface{}

	// Visit a parse tree produced by OBParser#var_name.
	VisitVar_name(ctx *Var_nameContext) interface{}

	// Visit a parse tree produced by OBParser#column_name.
	VisitColumn_name(ctx *Column_nameContext) interface{}

	// Visit a parse tree produced by OBParser#relation_name.
	VisitRelation_name(ctx *Relation_nameContext) interface{}

	// Visit a parse tree produced by OBParser#function_name.
	VisitFunction_name(ctx *Function_nameContext) interface{}

	// Visit a parse tree produced by OBParser#column_label.
	VisitColumn_label(ctx *Column_labelContext) interface{}

	// Visit a parse tree produced by OBParser#date_unit.
	VisitDate_unit(ctx *Date_unitContext) interface{}

	// Visit a parse tree produced by OBParser#unreserved_keyword.
	VisitUnreserved_keyword(ctx *Unreserved_keywordContext) interface{}

	// Visit a parse tree produced by OBParser#unreserved_keyword_normal.
	VisitUnreserved_keyword_normal(ctx *Unreserved_keyword_normalContext) interface{}

	// Visit a parse tree produced by OBParser#unreserved_keyword_special.
	VisitUnreserved_keyword_special(ctx *Unreserved_keyword_specialContext) interface{}

	// Visit a parse tree produced by OBParser#empty.
	VisitEmpty(ctx *EmptyContext) interface{}

	// Visit a parse tree produced by OBParser#forward_expr.
	VisitForward_expr(ctx *Forward_exprContext) interface{}

	// Visit a parse tree produced by OBParser#forward_sql_stmt.
	VisitForward_sql_stmt(ctx *Forward_sql_stmtContext) interface{}

}