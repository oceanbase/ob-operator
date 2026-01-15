// Code generated from /work/obparser/obmysql/sql/OBParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package mysql // OBParser
import "github.com/antlr4-go/antlr/v4"

type BaseOBParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseOBParserVisitor) VisitSql_stmt(ctx *Sql_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitStmt_list(ctx *Stmt_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitStmt(ctx *StmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExpr_list(ctx *Expr_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExpr_as_list(ctx *Expr_as_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExpr_with_opt_alias(ctx *Expr_with_opt_aliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_ref(ctx *Column_refContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitComplex_string_literal(ctx *Complex_string_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCharset_introducer(ctx *Charset_introducerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLiteral(ctx *LiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNumber_literal(ctx *Number_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExpr_const(ctx *Expr_constContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitConf_const(ctx *Conf_constContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitGlobal_or_session_alias(ctx *Global_or_session_aliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitBool_pri(ctx *Bool_priContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPredicate(ctx *PredicateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitBit_expr(ctx *Bit_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSimple_expr(ctx *Simple_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExpr(ctx *ExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNot(ctx *NotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNot2(ctx *Not2Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSub_query_flag(ctx *Sub_query_flagContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIn_expr(ctx *In_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCase_expr(ctx *Case_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWindow_function(ctx *Window_functionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFirst_or_last(ctx *First_or_lastContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRespect_or_ignore(ctx *Respect_or_ignoreContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWin_fun_first_last_params(ctx *Win_fun_first_last_paramsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWin_fun_lead_lag_params(ctx *Win_fun_lead_lag_paramsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNew_generalized_window_clause(ctx *New_generalized_window_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNew_generalized_window_clause_with_blanket(ctx *New_generalized_window_clause_with_blanketContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNamed_windows(ctx *Named_windowsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNamed_window(ctx *Named_windowContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitGeneralized_window_clause(ctx *Generalized_window_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWin_rows_or_range(ctx *Win_rows_or_rangeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWin_preceding_or_following(ctx *Win_preceding_or_followingContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWin_interval(ctx *Win_intervalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWin_bounding(ctx *Win_boundingContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWin_window(ctx *Win_windowContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCase_arg(ctx *Case_argContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWhen_clause_list(ctx *When_clause_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitWhen_clause(ctx *When_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCase_default(ctx *Case_defaultContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFunc_expr(ctx *Func_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSys_interval_func(ctx *Sys_interval_funcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUtc_timestamp_func(ctx *Utc_timestamp_funcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSysdate_func(ctx *Sysdate_funcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCur_timestamp_func(ctx *Cur_timestamp_funcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNow_synonyms_func(ctx *Now_synonyms_funcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCur_time_func(ctx *Cur_time_funcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCur_date_func(ctx *Cur_date_funcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSubstr_or_substring(ctx *Substr_or_substringContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSubstr_params(ctx *Substr_paramsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDate_params(ctx *Date_paramsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTimestamp_params(ctx *Timestamp_paramsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDelete_stmt(ctx *Delete_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitMulti_delete_table(ctx *Multi_delete_tableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUpdate_stmt(ctx *Update_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUpdate_asgn_list(ctx *Update_asgn_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUpdate_asgn_factor(ctx *Update_asgn_factorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_resource_stmt(ctx *Create_resource_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_resource_unit_option_list(ctx *Opt_resource_unit_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitResource_unit_option(ctx *Resource_unit_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_create_resource_pool_option_list(ctx *Opt_create_resource_pool_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_resource_pool_option(ctx *Create_resource_pool_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_resource_pool_option_list(ctx *Alter_resource_pool_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUnit_id_list(ctx *Unit_id_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_resource_pool_option(ctx *Alter_resource_pool_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_resource_stmt(ctx *Alter_resource_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_resource_stmt(ctx *Drop_resource_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_tenant_stmt(ctx *Create_tenant_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_tenant_option_list(ctx *Opt_tenant_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTenant_option(ctx *Tenant_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitZone_list(ctx *Zone_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitResource_pool_list(ctx *Resource_pool_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_tenant_stmt(ctx *Alter_tenant_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_tenant_stmt(ctx *Drop_tenant_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_database_stmt(ctx *Create_database_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDatabase_key(ctx *Database_keyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDatabase_factor(ctx *Database_factorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDatabase_option_list(ctx *Database_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCharset_key(ctx *Charset_keyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDatabase_option(ctx *Database_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRead_only_or_write(ctx *Read_only_or_writeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_database_stmt(ctx *Drop_database_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_database_stmt(ctx *Alter_database_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLoad_data_stmt(ctx *Load_data_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLoad_data_with_opt_hint(ctx *Load_data_with_opt_hintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLines_or_rows(ctx *Lines_or_rowsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitField_or_vars_list(ctx *Field_or_vars_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitField_or_vars(ctx *Field_or_varsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLoad_set_list(ctx *Load_set_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLoad_set_element(ctx *Load_set_elementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUse_database_stmt(ctx *Use_database_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_synonym_stmt(ctx *Create_synonym_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSynonym_name(ctx *Synonym_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSynonym_object(ctx *Synonym_objectContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_synonym_stmt(ctx *Drop_synonym_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTemporary_option(ctx *Temporary_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_table_like_stmt(ctx *Create_table_like_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_table_stmt(ctx *Create_table_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRet_type(ctx *Ret_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_function_stmt(ctx *Create_function_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_function_stmt(ctx *Drop_function_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_element_list(ctx *Table_element_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_element(ctx *Table_elementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_reference_option_list(ctx *Opt_reference_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitReference_option(ctx *Reference_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitReference_action(ctx *Reference_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitMatch_action(ctx *Match_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_definition(ctx *Column_definitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_generated_column_attribute_list(ctx *Opt_generated_column_attribute_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitGenerated_column_attribute(ctx *Generated_column_attributeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_definition_ref(ctx *Column_definition_refContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_definition_list(ctx *Column_definition_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCast_data_type(ctx *Cast_data_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCast_datetime_type_i(ctx *Cast_datetime_type_iContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitData_type(ctx *Data_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitString_list(ctx *String_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitText_string(ctx *Text_stringContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInt_type_i(ctx *Int_type_iContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFloat_type_i(ctx *Float_type_iContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDatetime_type_i(ctx *Datetime_type_iContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDate_year_type_i(ctx *Date_year_type_iContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitText_type_i(ctx *Text_type_iContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitBlob_type_i(ctx *Blob_type_iContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitString_length_i(ctx *String_length_iContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCollation_name(ctx *Collation_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTrans_param_name(ctx *Trans_param_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTrans_param_value(ctx *Trans_param_valueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCharset_name(ctx *Charset_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCharset_name_or_default(ctx *Charset_name_or_defaultContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCollation(ctx *CollationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_column_attribute_list(ctx *Opt_column_attribute_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_attribute(ctx *Column_attributeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNow_or_signed_literal(ctx *Now_or_signed_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSigned_literal(ctx *Signed_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_comma(ctx *Opt_commaContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_option_list_space_seperated(ctx *Table_option_list_space_seperatedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_option_list(ctx *Table_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPrimary_zone_name(ctx *Primary_zone_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTablespace(ctx *TablespaceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLocality_name(ctx *Locality_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_option(ctx *Table_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_name_or_string(ctx *Relation_name_or_stringContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_equal_mark(ctx *Opt_equal_markContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPartition_option(ctx *Partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_partition_option(ctx *Opt_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitHash_partition_option(ctx *Hash_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitList_partition_option(ctx *List_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitKey_partition_option(ctx *Key_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRange_partition_option(ctx *Range_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_column_partition_option(ctx *Opt_column_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_partition_option(ctx *Column_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAux_column_list(ctx *Aux_column_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitVertical_column_name(ctx *Vertical_column_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_name_list(ctx *Column_name_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSubpartition_option(ctx *Subpartition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_list_partition_list(ctx *Opt_list_partition_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_list_subpartition_list(ctx *Opt_list_subpartition_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_range_partition_list(ctx *Opt_range_partition_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_range_subpartition_list(ctx *Opt_range_subpartition_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitList_partition_list(ctx *List_partition_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitList_subpartition_list(ctx *List_subpartition_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitList_subpartition_element(ctx *List_subpartition_elementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitList_partition_element(ctx *List_partition_elementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitList_partition_expr(ctx *List_partition_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitList_expr(ctx *List_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRange_partition_list(ctx *Range_partition_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRange_partition_element(ctx *Range_partition_elementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRange_subpartition_element(ctx *Range_subpartition_elementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRange_subpartition_list(ctx *Range_subpartition_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRange_partition_expr(ctx *Range_partition_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRange_expr_list(ctx *Range_expr_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRange_expr(ctx *Range_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInt_or_decimal(ctx *Int_or_decimalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTg_hash_partition_option(ctx *Tg_hash_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTg_key_partition_option(ctx *Tg_key_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTg_range_partition_option(ctx *Tg_range_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTg_list_partition_option(ctx *Tg_list_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTg_subpartition_option(ctx *Tg_subpartition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRow_format_option(ctx *Row_format_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_tablegroup_stmt(ctx *Create_tablegroup_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_tablegroup_stmt(ctx *Drop_tablegroup_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_tablegroup_stmt(ctx *Alter_tablegroup_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTablegroup_option_list_space_seperated(ctx *Tablegroup_option_list_space_seperatedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTablegroup_option_list(ctx *Tablegroup_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTablegroup_option(ctx *Tablegroup_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_tablegroup_actions(ctx *Alter_tablegroup_actionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_tablegroup_action(ctx *Alter_tablegroup_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDefault_tablegroup(ctx *Default_tablegroupContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_view_stmt(ctx *Create_view_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitView_select_stmt(ctx *View_select_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitView_name(ctx *View_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_index_stmt(ctx *Create_index_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_name(ctx *Index_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_constraint_name(ctx *Opt_constraint_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitConstraint_name(ctx *Constraint_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSort_column_list(ctx *Sort_column_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSort_column_key(ctx *Sort_column_keyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_index_options(ctx *Opt_index_optionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_option(ctx *Index_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_using_algorithm(ctx *Index_using_algorithmContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_table_stmt(ctx *Drop_table_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_or_tables(ctx *Table_or_tablesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_view_stmt(ctx *Drop_view_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_list(ctx *Table_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_index_stmt(ctx *Drop_index_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInsert_stmt(ctx *Insert_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSingle_table_insert(ctx *Single_table_insertContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitValues_clause(ctx *Values_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitValue_or_values(ctx *Value_or_valuesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitReplace_with_opt_hint(ctx *Replace_with_opt_hintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInsert_with_opt_hint(ctx *Insert_with_opt_hintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_list(ctx *Column_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInsert_vals_list(ctx *Insert_vals_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInsert_vals(ctx *Insert_valsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExpr_or_default(ctx *Expr_or_defaultContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_stmt(ctx *Select_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_into(ctx *Select_intoContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_with_parens(ctx *Select_with_parensContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_no_parens(ctx *Select_no_parensContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNo_table_select(ctx *No_table_selectContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_clause(ctx *Select_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_clause_set_with_order_and_limit(ctx *Select_clause_set_with_order_and_limitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_clause_set(ctx *Select_clause_setContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_clause_set_right(ctx *Select_clause_set_rightContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_clause_set_left(ctx *Select_clause_set_leftContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNo_table_select_with_order_and_limit(ctx *No_table_select_with_order_and_limitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSimple_select_with_order_and_limit(ctx *Simple_select_with_order_and_limitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_with_parens_with_order_and_limit(ctx *Select_with_parens_with_order_and_limitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_with_opt_hint(ctx *Select_with_opt_hintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUpdate_with_opt_hint(ctx *Update_with_opt_hintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDelete_with_opt_hint(ctx *Delete_with_opt_hintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSimple_select(ctx *Simple_selectContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSet_type_union(ctx *Set_type_unionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSet_type_other(ctx *Set_type_otherContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSet_type(ctx *Set_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSet_expression_option(ctx *Set_expression_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_hint_value(ctx *Opt_hint_valueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLimit_clause(ctx *Limit_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInto_clause(ctx *Into_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInto_opt(ctx *Into_optContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInto_var_list(ctx *Into_var_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInto_var(ctx *Into_varContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitField_opt(ctx *Field_optContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitField_term_list(ctx *Field_term_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitField_term(ctx *Field_termContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLine_opt(ctx *Line_optContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLine_term_list(ctx *Line_term_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLine_term(ctx *Line_termContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitHint_list_with_end(ctx *Hint_list_with_endContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_hint_list(ctx *Opt_hint_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitHint_options(ctx *Hint_optionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitName_list(ctx *Name_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitHint_option(ctx *Hint_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitConsistency_level(ctx *Consistency_levelContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUse_plan_cache_type(ctx *Use_plan_cache_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUse_jit_type(ctx *Use_jit_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDistribute_method(ctx *Distribute_methodContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLimit_expr(ctx *Limit_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_for_update_wait(ctx *Opt_for_update_waitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitParameterized_trim(ctx *Parameterized_trimContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitGroupby_clause(ctx *Groupby_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSort_list_for_group_by(ctx *Sort_list_for_group_byContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSort_key_for_group_by(ctx *Sort_key_for_group_byContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOrder_by(ctx *Order_byContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSort_list(ctx *Sort_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSort_key(ctx *Sort_keyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitQuery_expression_option_list(ctx *Query_expression_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitQuery_expression_option(ctx *Query_expression_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitProjection(ctx *ProjectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSelect_expr_list(ctx *Select_expr_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFrom_list(ctx *From_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_references(ctx *Table_referencesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_reference(ctx *Table_referenceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_factor(ctx *Table_factorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTbl_name(ctx *Tbl_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDml_table_name(ctx *Dml_table_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSeed(ctx *SeedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_seed(ctx *Opt_seedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSample_percent(ctx *Sample_percentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSample_clause(ctx *Sample_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTable_subquery(ctx *Table_subqueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUse_partition(ctx *Use_partitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_hint_type(ctx *Index_hint_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitKey_or_index(ctx *Key_or_indexContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_hint_scope(ctx *Index_hint_scopeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_element(ctx *Index_elementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_list(ctx *Index_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_hint_definition(ctx *Index_hint_definitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_hint_list(ctx *Index_hint_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_factor(ctx *Relation_factorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_with_star_list(ctx *Relation_with_star_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_factor_with_star(ctx *Relation_factor_with_starContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNormal_relation_factor(ctx *Normal_relation_factorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDot_relation_factor(ctx *Dot_relation_factorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_factor_in_hint(ctx *Relation_factor_in_hintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitQb_name_option(ctx *Qb_name_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_factor_in_hint_list(ctx *Relation_factor_in_hint_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_sep_option(ctx *Relation_sep_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_factor_in_pq_hint(ctx *Relation_factor_in_pq_hintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_factor_in_leading_hint(ctx *Relation_factor_in_leading_hintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_factor_in_leading_hint_list(ctx *Relation_factor_in_leading_hint_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_factor_in_leading_hint_list_entry(ctx *Relation_factor_in_leading_hint_list_entryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_factor_in_use_join_hint_list(ctx *Relation_factor_in_use_join_hint_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTracing_num_list(ctx *Tracing_num_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitJoin_condition(ctx *Join_conditionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitJoined_table(ctx *Joined_tableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitNatural_join_type(ctx *Natural_join_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitInner_join_type(ctx *Inner_join_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOuter_join_type(ctx *Outer_join_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAnalyze_stmt(ctx *Analyze_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_outline_stmt(ctx *Create_outline_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_outline_stmt(ctx *Alter_outline_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_outline_stmt(ctx *Drop_outline_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExplain_stmt(ctx *Explain_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExplain_or_desc(ctx *Explain_or_descContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExplainable_stmt(ctx *Explainable_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFormat_name(ctx *Format_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitShow_stmt(ctx *Show_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDatabases_or_schemas(ctx *Databases_or_schemasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_for_grant_user(ctx *Opt_for_grant_userContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumns_or_fields(ctx *Columns_or_fieldsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDatabase_or_schema(ctx *Database_or_schemaContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIndex_or_indexes_or_keys(ctx *Index_or_indexes_or_keysContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFrom_or_in(ctx *From_or_inContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitHelp_stmt(ctx *Help_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_tablespace_stmt(ctx *Create_tablespace_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPermanent_tablespace(ctx *Permanent_tablespaceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPermanent_tablespace_option(ctx *Permanent_tablespace_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_tablespace_stmt(ctx *Drop_tablespace_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_tablespace_actions(ctx *Alter_tablespace_actionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_tablespace_action(ctx *Alter_tablespace_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_tablespace_stmt(ctx *Alter_tablespace_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRotate_master_key_stmt(ctx *Rotate_master_key_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPermanent_tablespace_options(ctx *Permanent_tablespace_optionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_user_stmt(ctx *Create_user_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUser_specification_list(ctx *User_specification_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUser_specification(ctx *User_specificationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRequire_specification(ctx *Require_specificationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTls_option_list(ctx *Tls_option_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTls_option(ctx *Tls_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUser(ctx *UserContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_host_name(ctx *Opt_host_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUser_with_host_name(ctx *User_with_host_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPassword(ctx *PasswordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_user_stmt(ctx *Drop_user_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUser_list(ctx *User_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSet_password_stmt(ctx *Set_password_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_for_user(ctx *Opt_for_userContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRename_user_stmt(ctx *Rename_user_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRename_info(ctx *Rename_infoContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRename_list(ctx *Rename_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLock_user_stmt(ctx *Lock_user_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLock_spec_mysql57(ctx *Lock_spec_mysql57Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLock_tables_stmt(ctx *Lock_tables_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUnlock_tables_stmt(ctx *Unlock_tables_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLock_table_list(ctx *Lock_table_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLock_table(ctx *Lock_tableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitLock_type(ctx *Lock_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitBegin_stmt(ctx *Begin_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCommit_stmt(ctx *Commit_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRollback_stmt(ctx *Rollback_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitKill_stmt(ctx *Kill_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitGrant_stmt(ctx *Grant_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitGrant_privileges(ctx *Grant_privilegesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPriv_type_list(ctx *Priv_type_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPriv_type(ctx *Priv_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPriv_level(ctx *Priv_levelContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitGrant_options(ctx *Grant_optionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRevoke_stmt(ctx *Revoke_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPrepare_stmt(ctx *Prepare_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitStmt_name(ctx *Stmt_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPreparable_stmt(ctx *Preparable_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitVariable_set_stmt(ctx *Variable_set_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSys_var_and_val_list(ctx *Sys_var_and_val_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitVar_and_val_list(ctx *Var_and_val_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSet_expr_or_default(ctx *Set_expr_or_defaultContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitVar_and_val(ctx *Var_and_valContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSys_var_and_val(ctx *Sys_var_and_valContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitScope_or_scope_alias(ctx *Scope_or_scope_aliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTo_or_eq(ctx *To_or_eqContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitExecute_stmt(ctx *Execute_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitArgument_list(ctx *Argument_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitArgument(ctx *ArgumentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDeallocate_prepare_stmt(ctx *Deallocate_prepare_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDeallocate_or_drop(ctx *Deallocate_or_dropContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTruncate_table_stmt(ctx *Truncate_table_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRename_table_stmt(ctx *Rename_table_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRename_table_actions(ctx *Rename_table_actionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRename_table_action(ctx *Rename_table_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_table_stmt(ctx *Alter_table_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_table_actions(ctx *Alter_table_actionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_table_action(ctx *Alter_table_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_constraint_option(ctx *Alter_constraint_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_partition_option(ctx *Alter_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOpt_partition_range_or_list(ctx *Opt_partition_range_or_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_tg_partition_option(ctx *Alter_tg_partition_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDrop_partition_name_list(ctx *Drop_partition_name_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitModify_partition_info(ctx *Modify_partition_infoContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitModify_tg_partition_info(ctx *Modify_tg_partition_infoContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_index_option(ctx *Alter_index_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_foreign_key_action(ctx *Alter_foreign_key_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitVisibility_option(ctx *Visibility_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_column_option(ctx *Alter_column_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_tablegroup_option(ctx *Alter_tablegroup_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_column_behavior(ctx *Alter_column_behaviorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFlashback_stmt(ctx *Flashback_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPurge_stmt(ctx *Purge_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitOptimize_stmt(ctx *Optimize_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDump_memory_stmt(ctx *Dump_memory_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_system_stmt(ctx *Alter_system_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitChange_tenant_name_or_tenant_id(ctx *Change_tenant_name_or_tenant_idContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCache_type(ctx *Cache_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitBalance_task_type(ctx *Balance_task_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTenant_list_tuple(ctx *Tenant_list_tupleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTenant_name_list(ctx *Tenant_name_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFlush_scope(ctx *Flush_scopeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitServer_info_list(ctx *Server_info_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitServer_info(ctx *Server_infoContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitServer_action(ctx *Server_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitServer_list(ctx *Server_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitZone_action(ctx *Zone_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIp_port(ctx *Ip_portContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitZone_desc(ctx *Zone_descContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitServer_or_zone(ctx *Server_or_zoneContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAdd_or_alter_zone_option(ctx *Add_or_alter_zone_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAdd_or_alter_zone_options(ctx *Add_or_alter_zone_optionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_or_change_or_modify(ctx *Alter_or_change_or_modifyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPartition_id_desc(ctx *Partition_id_descContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPartition_id_or_server_or_zone(ctx *Partition_id_or_server_or_zoneContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitMigrate_action(ctx *Migrate_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitChange_actions(ctx *Change_actionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitChange_action(ctx *Change_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitReplica_type(ctx *Replica_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSuspend_or_resume(ctx *Suspend_or_resumeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitBaseline_id_expr(ctx *Baseline_id_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSql_id_expr(ctx *Sql_id_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitBaseline_asgn_factor(ctx *Baseline_asgn_factorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTenant_name(ctx *Tenant_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCache_name(ctx *Cache_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFile_id(ctx *File_idContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCancel_task_type(ctx *Cancel_task_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_system_set_parameter_actions(ctx *Alter_system_set_parameter_actionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_system_set_parameter_action(ctx *Alter_system_set_parameter_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_system_settp_actions(ctx *Alter_system_settp_actionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSettp_option(ctx *Settp_optionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCluster_role(ctx *Cluster_roleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitPartition_role(ctx *Partition_roleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUpgrade_action(ctx *Upgrade_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSet_names_stmt(ctx *Set_names_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSet_charset_stmt(ctx *Set_charset_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSet_transaction_stmt(ctx *Set_transaction_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTransaction_characteristics(ctx *Transaction_characteristicsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitTransaction_access_mode(ctx *Transaction_access_modeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitIsolation_level(ctx *Isolation_levelContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCreate_savepoint_stmt(ctx *Create_savepoint_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRollback_savepoint_stmt(ctx *Rollback_savepoint_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelease_savepoint_stmt(ctx *Release_savepoint_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitAlter_cluster_stmt(ctx *Alter_cluster_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCluster_action(ctx *Cluster_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitSwitchover_cluster_stmt(ctx *Switchover_cluster_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCommit_switchover_clause(ctx *Commit_switchover_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitCluster_name(ctx *Cluster_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitVar_name(ctx *Var_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_name(ctx *Column_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitRelation_name(ctx *Relation_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitFunction_name(ctx *Function_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitColumn_label(ctx *Column_labelContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitDate_unit(ctx *Date_unitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUnreserved_keyword(ctx *Unreserved_keywordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUnreserved_keyword_normal(ctx *Unreserved_keyword_normalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitUnreserved_keyword_special(ctx *Unreserved_keyword_specialContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitEmpty(ctx *EmptyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitForward_expr(ctx *Forward_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseOBParserVisitor) VisitForward_sql_stmt(ctx *Forward_sql_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}
