// Code generated from /work/obparser/obmysql/sql/OBParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package mysql // OBParser
import "github.com/antlr4-go/antlr/v4"

// BaseOBParserListener is a complete listener for a parse tree produced by OBParser.
type BaseOBParserListener struct{}

var _ OBParserListener = &BaseOBParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseOBParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseOBParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseOBParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseOBParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterSql_stmt is called when production sql_stmt is entered.
func (s *BaseOBParserListener) EnterSql_stmt(ctx *Sql_stmtContext) {}

// ExitSql_stmt is called when production sql_stmt is exited.
func (s *BaseOBParserListener) ExitSql_stmt(ctx *Sql_stmtContext) {}

// EnterStmt_list is called when production stmt_list is entered.
func (s *BaseOBParserListener) EnterStmt_list(ctx *Stmt_listContext) {}

// ExitStmt_list is called when production stmt_list is exited.
func (s *BaseOBParserListener) ExitStmt_list(ctx *Stmt_listContext) {}

// EnterStmt is called when production stmt is entered.
func (s *BaseOBParserListener) EnterStmt(ctx *StmtContext) {}

// ExitStmt is called when production stmt is exited.
func (s *BaseOBParserListener) ExitStmt(ctx *StmtContext) {}

// EnterExpr_list is called when production expr_list is entered.
func (s *BaseOBParserListener) EnterExpr_list(ctx *Expr_listContext) {}

// ExitExpr_list is called when production expr_list is exited.
func (s *BaseOBParserListener) ExitExpr_list(ctx *Expr_listContext) {}

// EnterExpr_as_list is called when production expr_as_list is entered.
func (s *BaseOBParserListener) EnterExpr_as_list(ctx *Expr_as_listContext) {}

// ExitExpr_as_list is called when production expr_as_list is exited.
func (s *BaseOBParserListener) ExitExpr_as_list(ctx *Expr_as_listContext) {}

// EnterExpr_with_opt_alias is called when production expr_with_opt_alias is entered.
func (s *BaseOBParserListener) EnterExpr_with_opt_alias(ctx *Expr_with_opt_aliasContext) {}

// ExitExpr_with_opt_alias is called when production expr_with_opt_alias is exited.
func (s *BaseOBParserListener) ExitExpr_with_opt_alias(ctx *Expr_with_opt_aliasContext) {}

// EnterColumn_ref is called when production column_ref is entered.
func (s *BaseOBParserListener) EnterColumn_ref(ctx *Column_refContext) {}

// ExitColumn_ref is called when production column_ref is exited.
func (s *BaseOBParserListener) ExitColumn_ref(ctx *Column_refContext) {}

// EnterComplex_string_literal is called when production complex_string_literal is entered.
func (s *BaseOBParserListener) EnterComplex_string_literal(ctx *Complex_string_literalContext) {}

// ExitComplex_string_literal is called when production complex_string_literal is exited.
func (s *BaseOBParserListener) ExitComplex_string_literal(ctx *Complex_string_literalContext) {}

// EnterCharset_introducer is called when production charset_introducer is entered.
func (s *BaseOBParserListener) EnterCharset_introducer(ctx *Charset_introducerContext) {}

// ExitCharset_introducer is called when production charset_introducer is exited.
func (s *BaseOBParserListener) ExitCharset_introducer(ctx *Charset_introducerContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseOBParserListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseOBParserListener) ExitLiteral(ctx *LiteralContext) {}

// EnterNumber_literal is called when production number_literal is entered.
func (s *BaseOBParserListener) EnterNumber_literal(ctx *Number_literalContext) {}

// ExitNumber_literal is called when production number_literal is exited.
func (s *BaseOBParserListener) ExitNumber_literal(ctx *Number_literalContext) {}

// EnterExpr_const is called when production expr_const is entered.
func (s *BaseOBParserListener) EnterExpr_const(ctx *Expr_constContext) {}

// ExitExpr_const is called when production expr_const is exited.
func (s *BaseOBParserListener) ExitExpr_const(ctx *Expr_constContext) {}

// EnterConf_const is called when production conf_const is entered.
func (s *BaseOBParserListener) EnterConf_const(ctx *Conf_constContext) {}

// ExitConf_const is called when production conf_const is exited.
func (s *BaseOBParserListener) ExitConf_const(ctx *Conf_constContext) {}

// EnterGlobal_or_session_alias is called when production global_or_session_alias is entered.
func (s *BaseOBParserListener) EnterGlobal_or_session_alias(ctx *Global_or_session_aliasContext) {}

// ExitGlobal_or_session_alias is called when production global_or_session_alias is exited.
func (s *BaseOBParserListener) ExitGlobal_or_session_alias(ctx *Global_or_session_aliasContext) {}

// EnterBool_pri is called when production bool_pri is entered.
func (s *BaseOBParserListener) EnterBool_pri(ctx *Bool_priContext) {}

// ExitBool_pri is called when production bool_pri is exited.
func (s *BaseOBParserListener) ExitBool_pri(ctx *Bool_priContext) {}

// EnterPredicate is called when production predicate is entered.
func (s *BaseOBParserListener) EnterPredicate(ctx *PredicateContext) {}

// ExitPredicate is called when production predicate is exited.
func (s *BaseOBParserListener) ExitPredicate(ctx *PredicateContext) {}

// EnterBit_expr is called when production bit_expr is entered.
func (s *BaseOBParserListener) EnterBit_expr(ctx *Bit_exprContext) {}

// ExitBit_expr is called when production bit_expr is exited.
func (s *BaseOBParserListener) ExitBit_expr(ctx *Bit_exprContext) {}

// EnterSimple_expr is called when production simple_expr is entered.
func (s *BaseOBParserListener) EnterSimple_expr(ctx *Simple_exprContext) {}

// ExitSimple_expr is called when production simple_expr is exited.
func (s *BaseOBParserListener) ExitSimple_expr(ctx *Simple_exprContext) {}

// EnterExpr is called when production expr is entered.
func (s *BaseOBParserListener) EnterExpr(ctx *ExprContext) {}

// ExitExpr is called when production expr is exited.
func (s *BaseOBParserListener) ExitExpr(ctx *ExprContext) {}

// EnterNot is called when production not is entered.
func (s *BaseOBParserListener) EnterNot(ctx *NotContext) {}

// ExitNot is called when production not is exited.
func (s *BaseOBParserListener) ExitNot(ctx *NotContext) {}

// EnterNot2 is called when production not2 is entered.
func (s *BaseOBParserListener) EnterNot2(ctx *Not2Context) {}

// ExitNot2 is called when production not2 is exited.
func (s *BaseOBParserListener) ExitNot2(ctx *Not2Context) {}

// EnterSub_query_flag is called when production sub_query_flag is entered.
func (s *BaseOBParserListener) EnterSub_query_flag(ctx *Sub_query_flagContext) {}

// ExitSub_query_flag is called when production sub_query_flag is exited.
func (s *BaseOBParserListener) ExitSub_query_flag(ctx *Sub_query_flagContext) {}

// EnterIn_expr is called when production in_expr is entered.
func (s *BaseOBParserListener) EnterIn_expr(ctx *In_exprContext) {}

// ExitIn_expr is called when production in_expr is exited.
func (s *BaseOBParserListener) ExitIn_expr(ctx *In_exprContext) {}

// EnterCase_expr is called when production case_expr is entered.
func (s *BaseOBParserListener) EnterCase_expr(ctx *Case_exprContext) {}

// ExitCase_expr is called when production case_expr is exited.
func (s *BaseOBParserListener) ExitCase_expr(ctx *Case_exprContext) {}

// EnterWindow_function is called when production window_function is entered.
func (s *BaseOBParserListener) EnterWindow_function(ctx *Window_functionContext) {}

// ExitWindow_function is called when production window_function is exited.
func (s *BaseOBParserListener) ExitWindow_function(ctx *Window_functionContext) {}

// EnterFirst_or_last is called when production first_or_last is entered.
func (s *BaseOBParserListener) EnterFirst_or_last(ctx *First_or_lastContext) {}

// ExitFirst_or_last is called when production first_or_last is exited.
func (s *BaseOBParserListener) ExitFirst_or_last(ctx *First_or_lastContext) {}

// EnterRespect_or_ignore is called when production respect_or_ignore is entered.
func (s *BaseOBParserListener) EnterRespect_or_ignore(ctx *Respect_or_ignoreContext) {}

// ExitRespect_or_ignore is called when production respect_or_ignore is exited.
func (s *BaseOBParserListener) ExitRespect_or_ignore(ctx *Respect_or_ignoreContext) {}

// EnterWin_fun_first_last_params is called when production win_fun_first_last_params is entered.
func (s *BaseOBParserListener) EnterWin_fun_first_last_params(ctx *Win_fun_first_last_paramsContext) {}

// ExitWin_fun_first_last_params is called when production win_fun_first_last_params is exited.
func (s *BaseOBParserListener) ExitWin_fun_first_last_params(ctx *Win_fun_first_last_paramsContext) {}

// EnterWin_fun_lead_lag_params is called when production win_fun_lead_lag_params is entered.
func (s *BaseOBParserListener) EnterWin_fun_lead_lag_params(ctx *Win_fun_lead_lag_paramsContext) {}

// ExitWin_fun_lead_lag_params is called when production win_fun_lead_lag_params is exited.
func (s *BaseOBParserListener) ExitWin_fun_lead_lag_params(ctx *Win_fun_lead_lag_paramsContext) {}

// EnterNew_generalized_window_clause is called when production new_generalized_window_clause is entered.
func (s *BaseOBParserListener) EnterNew_generalized_window_clause(ctx *New_generalized_window_clauseContext) {}

// ExitNew_generalized_window_clause is called when production new_generalized_window_clause is exited.
func (s *BaseOBParserListener) ExitNew_generalized_window_clause(ctx *New_generalized_window_clauseContext) {}

// EnterNew_generalized_window_clause_with_blanket is called when production new_generalized_window_clause_with_blanket is entered.
func (s *BaseOBParserListener) EnterNew_generalized_window_clause_with_blanket(ctx *New_generalized_window_clause_with_blanketContext) {}

// ExitNew_generalized_window_clause_with_blanket is called when production new_generalized_window_clause_with_blanket is exited.
func (s *BaseOBParserListener) ExitNew_generalized_window_clause_with_blanket(ctx *New_generalized_window_clause_with_blanketContext) {}

// EnterNamed_windows is called when production named_windows is entered.
func (s *BaseOBParserListener) EnterNamed_windows(ctx *Named_windowsContext) {}

// ExitNamed_windows is called when production named_windows is exited.
func (s *BaseOBParserListener) ExitNamed_windows(ctx *Named_windowsContext) {}

// EnterNamed_window is called when production named_window is entered.
func (s *BaseOBParserListener) EnterNamed_window(ctx *Named_windowContext) {}

// ExitNamed_window is called when production named_window is exited.
func (s *BaseOBParserListener) ExitNamed_window(ctx *Named_windowContext) {}

// EnterGeneralized_window_clause is called when production generalized_window_clause is entered.
func (s *BaseOBParserListener) EnterGeneralized_window_clause(ctx *Generalized_window_clauseContext) {}

// ExitGeneralized_window_clause is called when production generalized_window_clause is exited.
func (s *BaseOBParserListener) ExitGeneralized_window_clause(ctx *Generalized_window_clauseContext) {}

// EnterWin_rows_or_range is called when production win_rows_or_range is entered.
func (s *BaseOBParserListener) EnterWin_rows_or_range(ctx *Win_rows_or_rangeContext) {}

// ExitWin_rows_or_range is called when production win_rows_or_range is exited.
func (s *BaseOBParserListener) ExitWin_rows_or_range(ctx *Win_rows_or_rangeContext) {}

// EnterWin_preceding_or_following is called when production win_preceding_or_following is entered.
func (s *BaseOBParserListener) EnterWin_preceding_or_following(ctx *Win_preceding_or_followingContext) {}

// ExitWin_preceding_or_following is called when production win_preceding_or_following is exited.
func (s *BaseOBParserListener) ExitWin_preceding_or_following(ctx *Win_preceding_or_followingContext) {}

// EnterWin_interval is called when production win_interval is entered.
func (s *BaseOBParserListener) EnterWin_interval(ctx *Win_intervalContext) {}

// ExitWin_interval is called when production win_interval is exited.
func (s *BaseOBParserListener) ExitWin_interval(ctx *Win_intervalContext) {}

// EnterWin_bounding is called when production win_bounding is entered.
func (s *BaseOBParserListener) EnterWin_bounding(ctx *Win_boundingContext) {}

// ExitWin_bounding is called when production win_bounding is exited.
func (s *BaseOBParserListener) ExitWin_bounding(ctx *Win_boundingContext) {}

// EnterWin_window is called when production win_window is entered.
func (s *BaseOBParserListener) EnterWin_window(ctx *Win_windowContext) {}

// ExitWin_window is called when production win_window is exited.
func (s *BaseOBParserListener) ExitWin_window(ctx *Win_windowContext) {}

// EnterCase_arg is called when production case_arg is entered.
func (s *BaseOBParserListener) EnterCase_arg(ctx *Case_argContext) {}

// ExitCase_arg is called when production case_arg is exited.
func (s *BaseOBParserListener) ExitCase_arg(ctx *Case_argContext) {}

// EnterWhen_clause_list is called when production when_clause_list is entered.
func (s *BaseOBParserListener) EnterWhen_clause_list(ctx *When_clause_listContext) {}

// ExitWhen_clause_list is called when production when_clause_list is exited.
func (s *BaseOBParserListener) ExitWhen_clause_list(ctx *When_clause_listContext) {}

// EnterWhen_clause is called when production when_clause is entered.
func (s *BaseOBParserListener) EnterWhen_clause(ctx *When_clauseContext) {}

// ExitWhen_clause is called when production when_clause is exited.
func (s *BaseOBParserListener) ExitWhen_clause(ctx *When_clauseContext) {}

// EnterCase_default is called when production case_default is entered.
func (s *BaseOBParserListener) EnterCase_default(ctx *Case_defaultContext) {}

// ExitCase_default is called when production case_default is exited.
func (s *BaseOBParserListener) ExitCase_default(ctx *Case_defaultContext) {}

// EnterFunc_expr is called when production func_expr is entered.
func (s *BaseOBParserListener) EnterFunc_expr(ctx *Func_exprContext) {}

// ExitFunc_expr is called when production func_expr is exited.
func (s *BaseOBParserListener) ExitFunc_expr(ctx *Func_exprContext) {}

// EnterSys_interval_func is called when production sys_interval_func is entered.
func (s *BaseOBParserListener) EnterSys_interval_func(ctx *Sys_interval_funcContext) {}

// ExitSys_interval_func is called when production sys_interval_func is exited.
func (s *BaseOBParserListener) ExitSys_interval_func(ctx *Sys_interval_funcContext) {}

// EnterUtc_timestamp_func is called when production utc_timestamp_func is entered.
func (s *BaseOBParserListener) EnterUtc_timestamp_func(ctx *Utc_timestamp_funcContext) {}

// ExitUtc_timestamp_func is called when production utc_timestamp_func is exited.
func (s *BaseOBParserListener) ExitUtc_timestamp_func(ctx *Utc_timestamp_funcContext) {}

// EnterSysdate_func is called when production sysdate_func is entered.
func (s *BaseOBParserListener) EnterSysdate_func(ctx *Sysdate_funcContext) {}

// ExitSysdate_func is called when production sysdate_func is exited.
func (s *BaseOBParserListener) ExitSysdate_func(ctx *Sysdate_funcContext) {}

// EnterCur_timestamp_func is called when production cur_timestamp_func is entered.
func (s *BaseOBParserListener) EnterCur_timestamp_func(ctx *Cur_timestamp_funcContext) {}

// ExitCur_timestamp_func is called when production cur_timestamp_func is exited.
func (s *BaseOBParserListener) ExitCur_timestamp_func(ctx *Cur_timestamp_funcContext) {}

// EnterNow_synonyms_func is called when production now_synonyms_func is entered.
func (s *BaseOBParserListener) EnterNow_synonyms_func(ctx *Now_synonyms_funcContext) {}

// ExitNow_synonyms_func is called when production now_synonyms_func is exited.
func (s *BaseOBParserListener) ExitNow_synonyms_func(ctx *Now_synonyms_funcContext) {}

// EnterCur_time_func is called when production cur_time_func is entered.
func (s *BaseOBParserListener) EnterCur_time_func(ctx *Cur_time_funcContext) {}

// ExitCur_time_func is called when production cur_time_func is exited.
func (s *BaseOBParserListener) ExitCur_time_func(ctx *Cur_time_funcContext) {}

// EnterCur_date_func is called when production cur_date_func is entered.
func (s *BaseOBParserListener) EnterCur_date_func(ctx *Cur_date_funcContext) {}

// ExitCur_date_func is called when production cur_date_func is exited.
func (s *BaseOBParserListener) ExitCur_date_func(ctx *Cur_date_funcContext) {}

// EnterSubstr_or_substring is called when production substr_or_substring is entered.
func (s *BaseOBParserListener) EnterSubstr_or_substring(ctx *Substr_or_substringContext) {}

// ExitSubstr_or_substring is called when production substr_or_substring is exited.
func (s *BaseOBParserListener) ExitSubstr_or_substring(ctx *Substr_or_substringContext) {}

// EnterSubstr_params is called when production substr_params is entered.
func (s *BaseOBParserListener) EnterSubstr_params(ctx *Substr_paramsContext) {}

// ExitSubstr_params is called when production substr_params is exited.
func (s *BaseOBParserListener) ExitSubstr_params(ctx *Substr_paramsContext) {}

// EnterDate_params is called when production date_params is entered.
func (s *BaseOBParserListener) EnterDate_params(ctx *Date_paramsContext) {}

// ExitDate_params is called when production date_params is exited.
func (s *BaseOBParserListener) ExitDate_params(ctx *Date_paramsContext) {}

// EnterTimestamp_params is called when production timestamp_params is entered.
func (s *BaseOBParserListener) EnterTimestamp_params(ctx *Timestamp_paramsContext) {}

// ExitTimestamp_params is called when production timestamp_params is exited.
func (s *BaseOBParserListener) ExitTimestamp_params(ctx *Timestamp_paramsContext) {}

// EnterDelete_stmt is called when production delete_stmt is entered.
func (s *BaseOBParserListener) EnterDelete_stmt(ctx *Delete_stmtContext) {}

// ExitDelete_stmt is called when production delete_stmt is exited.
func (s *BaseOBParserListener) ExitDelete_stmt(ctx *Delete_stmtContext) {}

// EnterMulti_delete_table is called when production multi_delete_table is entered.
func (s *BaseOBParserListener) EnterMulti_delete_table(ctx *Multi_delete_tableContext) {}

// ExitMulti_delete_table is called when production multi_delete_table is exited.
func (s *BaseOBParserListener) ExitMulti_delete_table(ctx *Multi_delete_tableContext) {}

// EnterUpdate_stmt is called when production update_stmt is entered.
func (s *BaseOBParserListener) EnterUpdate_stmt(ctx *Update_stmtContext) {}

// ExitUpdate_stmt is called when production update_stmt is exited.
func (s *BaseOBParserListener) ExitUpdate_stmt(ctx *Update_stmtContext) {}

// EnterUpdate_asgn_list is called when production update_asgn_list is entered.
func (s *BaseOBParserListener) EnterUpdate_asgn_list(ctx *Update_asgn_listContext) {}

// ExitUpdate_asgn_list is called when production update_asgn_list is exited.
func (s *BaseOBParserListener) ExitUpdate_asgn_list(ctx *Update_asgn_listContext) {}

// EnterUpdate_asgn_factor is called when production update_asgn_factor is entered.
func (s *BaseOBParserListener) EnterUpdate_asgn_factor(ctx *Update_asgn_factorContext) {}

// ExitUpdate_asgn_factor is called when production update_asgn_factor is exited.
func (s *BaseOBParserListener) ExitUpdate_asgn_factor(ctx *Update_asgn_factorContext) {}

// EnterCreate_resource_stmt is called when production create_resource_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_resource_stmt(ctx *Create_resource_stmtContext) {}

// ExitCreate_resource_stmt is called when production create_resource_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_resource_stmt(ctx *Create_resource_stmtContext) {}

// EnterOpt_resource_unit_option_list is called when production opt_resource_unit_option_list is entered.
func (s *BaseOBParserListener) EnterOpt_resource_unit_option_list(ctx *Opt_resource_unit_option_listContext) {}

// ExitOpt_resource_unit_option_list is called when production opt_resource_unit_option_list is exited.
func (s *BaseOBParserListener) ExitOpt_resource_unit_option_list(ctx *Opt_resource_unit_option_listContext) {}

// EnterResource_unit_option is called when production resource_unit_option is entered.
func (s *BaseOBParserListener) EnterResource_unit_option(ctx *Resource_unit_optionContext) {}

// ExitResource_unit_option is called when production resource_unit_option is exited.
func (s *BaseOBParserListener) ExitResource_unit_option(ctx *Resource_unit_optionContext) {}

// EnterOpt_create_resource_pool_option_list is called when production opt_create_resource_pool_option_list is entered.
func (s *BaseOBParserListener) EnterOpt_create_resource_pool_option_list(ctx *Opt_create_resource_pool_option_listContext) {}

// ExitOpt_create_resource_pool_option_list is called when production opt_create_resource_pool_option_list is exited.
func (s *BaseOBParserListener) ExitOpt_create_resource_pool_option_list(ctx *Opt_create_resource_pool_option_listContext) {}

// EnterCreate_resource_pool_option is called when production create_resource_pool_option is entered.
func (s *BaseOBParserListener) EnterCreate_resource_pool_option(ctx *Create_resource_pool_optionContext) {}

// ExitCreate_resource_pool_option is called when production create_resource_pool_option is exited.
func (s *BaseOBParserListener) ExitCreate_resource_pool_option(ctx *Create_resource_pool_optionContext) {}

// EnterAlter_resource_pool_option_list is called when production alter_resource_pool_option_list is entered.
func (s *BaseOBParserListener) EnterAlter_resource_pool_option_list(ctx *Alter_resource_pool_option_listContext) {}

// ExitAlter_resource_pool_option_list is called when production alter_resource_pool_option_list is exited.
func (s *BaseOBParserListener) ExitAlter_resource_pool_option_list(ctx *Alter_resource_pool_option_listContext) {}

// EnterUnit_id_list is called when production unit_id_list is entered.
func (s *BaseOBParserListener) EnterUnit_id_list(ctx *Unit_id_listContext) {}

// ExitUnit_id_list is called when production unit_id_list is exited.
func (s *BaseOBParserListener) ExitUnit_id_list(ctx *Unit_id_listContext) {}

// EnterAlter_resource_pool_option is called when production alter_resource_pool_option is entered.
func (s *BaseOBParserListener) EnterAlter_resource_pool_option(ctx *Alter_resource_pool_optionContext) {}

// ExitAlter_resource_pool_option is called when production alter_resource_pool_option is exited.
func (s *BaseOBParserListener) ExitAlter_resource_pool_option(ctx *Alter_resource_pool_optionContext) {}

// EnterAlter_resource_stmt is called when production alter_resource_stmt is entered.
func (s *BaseOBParserListener) EnterAlter_resource_stmt(ctx *Alter_resource_stmtContext) {}

// ExitAlter_resource_stmt is called when production alter_resource_stmt is exited.
func (s *BaseOBParserListener) ExitAlter_resource_stmt(ctx *Alter_resource_stmtContext) {}

// EnterDrop_resource_stmt is called when production drop_resource_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_resource_stmt(ctx *Drop_resource_stmtContext) {}

// ExitDrop_resource_stmt is called when production drop_resource_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_resource_stmt(ctx *Drop_resource_stmtContext) {}

// EnterCreate_tenant_stmt is called when production create_tenant_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_tenant_stmt(ctx *Create_tenant_stmtContext) {}

// ExitCreate_tenant_stmt is called when production create_tenant_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_tenant_stmt(ctx *Create_tenant_stmtContext) {}

// EnterOpt_tenant_option_list is called when production opt_tenant_option_list is entered.
func (s *BaseOBParserListener) EnterOpt_tenant_option_list(ctx *Opt_tenant_option_listContext) {}

// ExitOpt_tenant_option_list is called when production opt_tenant_option_list is exited.
func (s *BaseOBParserListener) ExitOpt_tenant_option_list(ctx *Opt_tenant_option_listContext) {}

// EnterTenant_option is called when production tenant_option is entered.
func (s *BaseOBParserListener) EnterTenant_option(ctx *Tenant_optionContext) {}

// ExitTenant_option is called when production tenant_option is exited.
func (s *BaseOBParserListener) ExitTenant_option(ctx *Tenant_optionContext) {}

// EnterZone_list is called when production zone_list is entered.
func (s *BaseOBParserListener) EnterZone_list(ctx *Zone_listContext) {}

// ExitZone_list is called when production zone_list is exited.
func (s *BaseOBParserListener) ExitZone_list(ctx *Zone_listContext) {}

// EnterResource_pool_list is called when production resource_pool_list is entered.
func (s *BaseOBParserListener) EnterResource_pool_list(ctx *Resource_pool_listContext) {}

// ExitResource_pool_list is called when production resource_pool_list is exited.
func (s *BaseOBParserListener) ExitResource_pool_list(ctx *Resource_pool_listContext) {}

// EnterAlter_tenant_stmt is called when production alter_tenant_stmt is entered.
func (s *BaseOBParserListener) EnterAlter_tenant_stmt(ctx *Alter_tenant_stmtContext) {}

// ExitAlter_tenant_stmt is called when production alter_tenant_stmt is exited.
func (s *BaseOBParserListener) ExitAlter_tenant_stmt(ctx *Alter_tenant_stmtContext) {}

// EnterDrop_tenant_stmt is called when production drop_tenant_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_tenant_stmt(ctx *Drop_tenant_stmtContext) {}

// ExitDrop_tenant_stmt is called when production drop_tenant_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_tenant_stmt(ctx *Drop_tenant_stmtContext) {}

// EnterCreate_database_stmt is called when production create_database_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_database_stmt(ctx *Create_database_stmtContext) {}

// ExitCreate_database_stmt is called when production create_database_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_database_stmt(ctx *Create_database_stmtContext) {}

// EnterDatabase_key is called when production database_key is entered.
func (s *BaseOBParserListener) EnterDatabase_key(ctx *Database_keyContext) {}

// ExitDatabase_key is called when production database_key is exited.
func (s *BaseOBParserListener) ExitDatabase_key(ctx *Database_keyContext) {}

// EnterDatabase_factor is called when production database_factor is entered.
func (s *BaseOBParserListener) EnterDatabase_factor(ctx *Database_factorContext) {}

// ExitDatabase_factor is called when production database_factor is exited.
func (s *BaseOBParserListener) ExitDatabase_factor(ctx *Database_factorContext) {}

// EnterDatabase_option_list is called when production database_option_list is entered.
func (s *BaseOBParserListener) EnterDatabase_option_list(ctx *Database_option_listContext) {}

// ExitDatabase_option_list is called when production database_option_list is exited.
func (s *BaseOBParserListener) ExitDatabase_option_list(ctx *Database_option_listContext) {}

// EnterCharset_key is called when production charset_key is entered.
func (s *BaseOBParserListener) EnterCharset_key(ctx *Charset_keyContext) {}

// ExitCharset_key is called when production charset_key is exited.
func (s *BaseOBParserListener) ExitCharset_key(ctx *Charset_keyContext) {}

// EnterDatabase_option is called when production database_option is entered.
func (s *BaseOBParserListener) EnterDatabase_option(ctx *Database_optionContext) {}

// ExitDatabase_option is called when production database_option is exited.
func (s *BaseOBParserListener) ExitDatabase_option(ctx *Database_optionContext) {}

// EnterRead_only_or_write is called when production read_only_or_write is entered.
func (s *BaseOBParserListener) EnterRead_only_or_write(ctx *Read_only_or_writeContext) {}

// ExitRead_only_or_write is called when production read_only_or_write is exited.
func (s *BaseOBParserListener) ExitRead_only_or_write(ctx *Read_only_or_writeContext) {}

// EnterDrop_database_stmt is called when production drop_database_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_database_stmt(ctx *Drop_database_stmtContext) {}

// ExitDrop_database_stmt is called when production drop_database_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_database_stmt(ctx *Drop_database_stmtContext) {}

// EnterAlter_database_stmt is called when production alter_database_stmt is entered.
func (s *BaseOBParserListener) EnterAlter_database_stmt(ctx *Alter_database_stmtContext) {}

// ExitAlter_database_stmt is called when production alter_database_stmt is exited.
func (s *BaseOBParserListener) ExitAlter_database_stmt(ctx *Alter_database_stmtContext) {}

// EnterLoad_data_stmt is called when production load_data_stmt is entered.
func (s *BaseOBParserListener) EnterLoad_data_stmt(ctx *Load_data_stmtContext) {}

// ExitLoad_data_stmt is called when production load_data_stmt is exited.
func (s *BaseOBParserListener) ExitLoad_data_stmt(ctx *Load_data_stmtContext) {}

// EnterLoad_data_with_opt_hint is called when production load_data_with_opt_hint is entered.
func (s *BaseOBParserListener) EnterLoad_data_with_opt_hint(ctx *Load_data_with_opt_hintContext) {}

// ExitLoad_data_with_opt_hint is called when production load_data_with_opt_hint is exited.
func (s *BaseOBParserListener) ExitLoad_data_with_opt_hint(ctx *Load_data_with_opt_hintContext) {}

// EnterLines_or_rows is called when production lines_or_rows is entered.
func (s *BaseOBParserListener) EnterLines_or_rows(ctx *Lines_or_rowsContext) {}

// ExitLines_or_rows is called when production lines_or_rows is exited.
func (s *BaseOBParserListener) ExitLines_or_rows(ctx *Lines_or_rowsContext) {}

// EnterField_or_vars_list is called when production field_or_vars_list is entered.
func (s *BaseOBParserListener) EnterField_or_vars_list(ctx *Field_or_vars_listContext) {}

// ExitField_or_vars_list is called when production field_or_vars_list is exited.
func (s *BaseOBParserListener) ExitField_or_vars_list(ctx *Field_or_vars_listContext) {}

// EnterField_or_vars is called when production field_or_vars is entered.
func (s *BaseOBParserListener) EnterField_or_vars(ctx *Field_or_varsContext) {}

// ExitField_or_vars is called when production field_or_vars is exited.
func (s *BaseOBParserListener) ExitField_or_vars(ctx *Field_or_varsContext) {}

// EnterLoad_set_list is called when production load_set_list is entered.
func (s *BaseOBParserListener) EnterLoad_set_list(ctx *Load_set_listContext) {}

// ExitLoad_set_list is called when production load_set_list is exited.
func (s *BaseOBParserListener) ExitLoad_set_list(ctx *Load_set_listContext) {}

// EnterLoad_set_element is called when production load_set_element is entered.
func (s *BaseOBParserListener) EnterLoad_set_element(ctx *Load_set_elementContext) {}

// ExitLoad_set_element is called when production load_set_element is exited.
func (s *BaseOBParserListener) ExitLoad_set_element(ctx *Load_set_elementContext) {}

// EnterUse_database_stmt is called when production use_database_stmt is entered.
func (s *BaseOBParserListener) EnterUse_database_stmt(ctx *Use_database_stmtContext) {}

// ExitUse_database_stmt is called when production use_database_stmt is exited.
func (s *BaseOBParserListener) ExitUse_database_stmt(ctx *Use_database_stmtContext) {}

// EnterCreate_synonym_stmt is called when production create_synonym_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_synonym_stmt(ctx *Create_synonym_stmtContext) {}

// ExitCreate_synonym_stmt is called when production create_synonym_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_synonym_stmt(ctx *Create_synonym_stmtContext) {}

// EnterSynonym_name is called when production synonym_name is entered.
func (s *BaseOBParserListener) EnterSynonym_name(ctx *Synonym_nameContext) {}

// ExitSynonym_name is called when production synonym_name is exited.
func (s *BaseOBParserListener) ExitSynonym_name(ctx *Synonym_nameContext) {}

// EnterSynonym_object is called when production synonym_object is entered.
func (s *BaseOBParserListener) EnterSynonym_object(ctx *Synonym_objectContext) {}

// ExitSynonym_object is called when production synonym_object is exited.
func (s *BaseOBParserListener) ExitSynonym_object(ctx *Synonym_objectContext) {}

// EnterDrop_synonym_stmt is called when production drop_synonym_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_synonym_stmt(ctx *Drop_synonym_stmtContext) {}

// ExitDrop_synonym_stmt is called when production drop_synonym_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_synonym_stmt(ctx *Drop_synonym_stmtContext) {}

// EnterTemporary_option is called when production temporary_option is entered.
func (s *BaseOBParserListener) EnterTemporary_option(ctx *Temporary_optionContext) {}

// ExitTemporary_option is called when production temporary_option is exited.
func (s *BaseOBParserListener) ExitTemporary_option(ctx *Temporary_optionContext) {}

// EnterCreate_table_like_stmt is called when production create_table_like_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_table_like_stmt(ctx *Create_table_like_stmtContext) {}

// ExitCreate_table_like_stmt is called when production create_table_like_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_table_like_stmt(ctx *Create_table_like_stmtContext) {}

// EnterCreate_table_stmt is called when production create_table_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_table_stmt(ctx *Create_table_stmtContext) {}

// ExitCreate_table_stmt is called when production create_table_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_table_stmt(ctx *Create_table_stmtContext) {}

// EnterRet_type is called when production ret_type is entered.
func (s *BaseOBParserListener) EnterRet_type(ctx *Ret_typeContext) {}

// ExitRet_type is called when production ret_type is exited.
func (s *BaseOBParserListener) ExitRet_type(ctx *Ret_typeContext) {}

// EnterCreate_function_stmt is called when production create_function_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_function_stmt(ctx *Create_function_stmtContext) {}

// ExitCreate_function_stmt is called when production create_function_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_function_stmt(ctx *Create_function_stmtContext) {}

// EnterDrop_function_stmt is called when production drop_function_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_function_stmt(ctx *Drop_function_stmtContext) {}

// ExitDrop_function_stmt is called when production drop_function_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_function_stmt(ctx *Drop_function_stmtContext) {}

// EnterTable_element_list is called when production table_element_list is entered.
func (s *BaseOBParserListener) EnterTable_element_list(ctx *Table_element_listContext) {}

// ExitTable_element_list is called when production table_element_list is exited.
func (s *BaseOBParserListener) ExitTable_element_list(ctx *Table_element_listContext) {}

// EnterTable_element is called when production table_element is entered.
func (s *BaseOBParserListener) EnterTable_element(ctx *Table_elementContext) {}

// ExitTable_element is called when production table_element is exited.
func (s *BaseOBParserListener) ExitTable_element(ctx *Table_elementContext) {}

// EnterOpt_reference_option_list is called when production opt_reference_option_list is entered.
func (s *BaseOBParserListener) EnterOpt_reference_option_list(ctx *Opt_reference_option_listContext) {}

// ExitOpt_reference_option_list is called when production opt_reference_option_list is exited.
func (s *BaseOBParserListener) ExitOpt_reference_option_list(ctx *Opt_reference_option_listContext) {}

// EnterReference_option is called when production reference_option is entered.
func (s *BaseOBParserListener) EnterReference_option(ctx *Reference_optionContext) {}

// ExitReference_option is called when production reference_option is exited.
func (s *BaseOBParserListener) ExitReference_option(ctx *Reference_optionContext) {}

// EnterReference_action is called when production reference_action is entered.
func (s *BaseOBParserListener) EnterReference_action(ctx *Reference_actionContext) {}

// ExitReference_action is called when production reference_action is exited.
func (s *BaseOBParserListener) ExitReference_action(ctx *Reference_actionContext) {}

// EnterMatch_action is called when production match_action is entered.
func (s *BaseOBParserListener) EnterMatch_action(ctx *Match_actionContext) {}

// ExitMatch_action is called when production match_action is exited.
func (s *BaseOBParserListener) ExitMatch_action(ctx *Match_actionContext) {}

// EnterColumn_definition is called when production column_definition is entered.
func (s *BaseOBParserListener) EnterColumn_definition(ctx *Column_definitionContext) {}

// ExitColumn_definition is called when production column_definition is exited.
func (s *BaseOBParserListener) ExitColumn_definition(ctx *Column_definitionContext) {}

// EnterOpt_generated_column_attribute_list is called when production opt_generated_column_attribute_list is entered.
func (s *BaseOBParserListener) EnterOpt_generated_column_attribute_list(ctx *Opt_generated_column_attribute_listContext) {}

// ExitOpt_generated_column_attribute_list is called when production opt_generated_column_attribute_list is exited.
func (s *BaseOBParserListener) ExitOpt_generated_column_attribute_list(ctx *Opt_generated_column_attribute_listContext) {}

// EnterGenerated_column_attribute is called when production generated_column_attribute is entered.
func (s *BaseOBParserListener) EnterGenerated_column_attribute(ctx *Generated_column_attributeContext) {}

// ExitGenerated_column_attribute is called when production generated_column_attribute is exited.
func (s *BaseOBParserListener) ExitGenerated_column_attribute(ctx *Generated_column_attributeContext) {}

// EnterColumn_definition_ref is called when production column_definition_ref is entered.
func (s *BaseOBParserListener) EnterColumn_definition_ref(ctx *Column_definition_refContext) {}

// ExitColumn_definition_ref is called when production column_definition_ref is exited.
func (s *BaseOBParserListener) ExitColumn_definition_ref(ctx *Column_definition_refContext) {}

// EnterColumn_definition_list is called when production column_definition_list is entered.
func (s *BaseOBParserListener) EnterColumn_definition_list(ctx *Column_definition_listContext) {}

// ExitColumn_definition_list is called when production column_definition_list is exited.
func (s *BaseOBParserListener) ExitColumn_definition_list(ctx *Column_definition_listContext) {}

// EnterCast_data_type is called when production cast_data_type is entered.
func (s *BaseOBParserListener) EnterCast_data_type(ctx *Cast_data_typeContext) {}

// ExitCast_data_type is called when production cast_data_type is exited.
func (s *BaseOBParserListener) ExitCast_data_type(ctx *Cast_data_typeContext) {}

// EnterCast_datetime_type_i is called when production cast_datetime_type_i is entered.
func (s *BaseOBParserListener) EnterCast_datetime_type_i(ctx *Cast_datetime_type_iContext) {}

// ExitCast_datetime_type_i is called when production cast_datetime_type_i is exited.
func (s *BaseOBParserListener) ExitCast_datetime_type_i(ctx *Cast_datetime_type_iContext) {}

// EnterData_type is called when production data_type is entered.
func (s *BaseOBParserListener) EnterData_type(ctx *Data_typeContext) {}

// ExitData_type is called when production data_type is exited.
func (s *BaseOBParserListener) ExitData_type(ctx *Data_typeContext) {}

// EnterString_list is called when production string_list is entered.
func (s *BaseOBParserListener) EnterString_list(ctx *String_listContext) {}

// ExitString_list is called when production string_list is exited.
func (s *BaseOBParserListener) ExitString_list(ctx *String_listContext) {}

// EnterText_string is called when production text_string is entered.
func (s *BaseOBParserListener) EnterText_string(ctx *Text_stringContext) {}

// ExitText_string is called when production text_string is exited.
func (s *BaseOBParserListener) ExitText_string(ctx *Text_stringContext) {}

// EnterInt_type_i is called when production int_type_i is entered.
func (s *BaseOBParserListener) EnterInt_type_i(ctx *Int_type_iContext) {}

// ExitInt_type_i is called when production int_type_i is exited.
func (s *BaseOBParserListener) ExitInt_type_i(ctx *Int_type_iContext) {}

// EnterFloat_type_i is called when production float_type_i is entered.
func (s *BaseOBParserListener) EnterFloat_type_i(ctx *Float_type_iContext) {}

// ExitFloat_type_i is called when production float_type_i is exited.
func (s *BaseOBParserListener) ExitFloat_type_i(ctx *Float_type_iContext) {}

// EnterDatetime_type_i is called when production datetime_type_i is entered.
func (s *BaseOBParserListener) EnterDatetime_type_i(ctx *Datetime_type_iContext) {}

// ExitDatetime_type_i is called when production datetime_type_i is exited.
func (s *BaseOBParserListener) ExitDatetime_type_i(ctx *Datetime_type_iContext) {}

// EnterDate_year_type_i is called when production date_year_type_i is entered.
func (s *BaseOBParserListener) EnterDate_year_type_i(ctx *Date_year_type_iContext) {}

// ExitDate_year_type_i is called when production date_year_type_i is exited.
func (s *BaseOBParserListener) ExitDate_year_type_i(ctx *Date_year_type_iContext) {}

// EnterText_type_i is called when production text_type_i is entered.
func (s *BaseOBParserListener) EnterText_type_i(ctx *Text_type_iContext) {}

// ExitText_type_i is called when production text_type_i is exited.
func (s *BaseOBParserListener) ExitText_type_i(ctx *Text_type_iContext) {}

// EnterBlob_type_i is called when production blob_type_i is entered.
func (s *BaseOBParserListener) EnterBlob_type_i(ctx *Blob_type_iContext) {}

// ExitBlob_type_i is called when production blob_type_i is exited.
func (s *BaseOBParserListener) ExitBlob_type_i(ctx *Blob_type_iContext) {}

// EnterString_length_i is called when production string_length_i is entered.
func (s *BaseOBParserListener) EnterString_length_i(ctx *String_length_iContext) {}

// ExitString_length_i is called when production string_length_i is exited.
func (s *BaseOBParserListener) ExitString_length_i(ctx *String_length_iContext) {}

// EnterCollation_name is called when production collation_name is entered.
func (s *BaseOBParserListener) EnterCollation_name(ctx *Collation_nameContext) {}

// ExitCollation_name is called when production collation_name is exited.
func (s *BaseOBParserListener) ExitCollation_name(ctx *Collation_nameContext) {}

// EnterTrans_param_name is called when production trans_param_name is entered.
func (s *BaseOBParserListener) EnterTrans_param_name(ctx *Trans_param_nameContext) {}

// ExitTrans_param_name is called when production trans_param_name is exited.
func (s *BaseOBParserListener) ExitTrans_param_name(ctx *Trans_param_nameContext) {}

// EnterTrans_param_value is called when production trans_param_value is entered.
func (s *BaseOBParserListener) EnterTrans_param_value(ctx *Trans_param_valueContext) {}

// ExitTrans_param_value is called when production trans_param_value is exited.
func (s *BaseOBParserListener) ExitTrans_param_value(ctx *Trans_param_valueContext) {}

// EnterCharset_name is called when production charset_name is entered.
func (s *BaseOBParserListener) EnterCharset_name(ctx *Charset_nameContext) {}

// ExitCharset_name is called when production charset_name is exited.
func (s *BaseOBParserListener) ExitCharset_name(ctx *Charset_nameContext) {}

// EnterCharset_name_or_default is called when production charset_name_or_default is entered.
func (s *BaseOBParserListener) EnterCharset_name_or_default(ctx *Charset_name_or_defaultContext) {}

// ExitCharset_name_or_default is called when production charset_name_or_default is exited.
func (s *BaseOBParserListener) ExitCharset_name_or_default(ctx *Charset_name_or_defaultContext) {}

// EnterCollation is called when production collation is entered.
func (s *BaseOBParserListener) EnterCollation(ctx *CollationContext) {}

// ExitCollation is called when production collation is exited.
func (s *BaseOBParserListener) ExitCollation(ctx *CollationContext) {}

// EnterOpt_column_attribute_list is called when production opt_column_attribute_list is entered.
func (s *BaseOBParserListener) EnterOpt_column_attribute_list(ctx *Opt_column_attribute_listContext) {}

// ExitOpt_column_attribute_list is called when production opt_column_attribute_list is exited.
func (s *BaseOBParserListener) ExitOpt_column_attribute_list(ctx *Opt_column_attribute_listContext) {}

// EnterColumn_attribute is called when production column_attribute is entered.
func (s *BaseOBParserListener) EnterColumn_attribute(ctx *Column_attributeContext) {}

// ExitColumn_attribute is called when production column_attribute is exited.
func (s *BaseOBParserListener) ExitColumn_attribute(ctx *Column_attributeContext) {}

// EnterNow_or_signed_literal is called when production now_or_signed_literal is entered.
func (s *BaseOBParserListener) EnterNow_or_signed_literal(ctx *Now_or_signed_literalContext) {}

// ExitNow_or_signed_literal is called when production now_or_signed_literal is exited.
func (s *BaseOBParserListener) ExitNow_or_signed_literal(ctx *Now_or_signed_literalContext) {}

// EnterSigned_literal is called when production signed_literal is entered.
func (s *BaseOBParserListener) EnterSigned_literal(ctx *Signed_literalContext) {}

// ExitSigned_literal is called when production signed_literal is exited.
func (s *BaseOBParserListener) ExitSigned_literal(ctx *Signed_literalContext) {}

// EnterOpt_comma is called when production opt_comma is entered.
func (s *BaseOBParserListener) EnterOpt_comma(ctx *Opt_commaContext) {}

// ExitOpt_comma is called when production opt_comma is exited.
func (s *BaseOBParserListener) ExitOpt_comma(ctx *Opt_commaContext) {}

// EnterTable_option_list_space_seperated is called when production table_option_list_space_seperated is entered.
func (s *BaseOBParserListener) EnterTable_option_list_space_seperated(ctx *Table_option_list_space_seperatedContext) {}

// ExitTable_option_list_space_seperated is called when production table_option_list_space_seperated is exited.
func (s *BaseOBParserListener) ExitTable_option_list_space_seperated(ctx *Table_option_list_space_seperatedContext) {}

// EnterTable_option_list is called when production table_option_list is entered.
func (s *BaseOBParserListener) EnterTable_option_list(ctx *Table_option_listContext) {}

// ExitTable_option_list is called when production table_option_list is exited.
func (s *BaseOBParserListener) ExitTable_option_list(ctx *Table_option_listContext) {}

// EnterPrimary_zone_name is called when production primary_zone_name is entered.
func (s *BaseOBParserListener) EnterPrimary_zone_name(ctx *Primary_zone_nameContext) {}

// ExitPrimary_zone_name is called when production primary_zone_name is exited.
func (s *BaseOBParserListener) ExitPrimary_zone_name(ctx *Primary_zone_nameContext) {}

// EnterTablespace is called when production tablespace is entered.
func (s *BaseOBParserListener) EnterTablespace(ctx *TablespaceContext) {}

// ExitTablespace is called when production tablespace is exited.
func (s *BaseOBParserListener) ExitTablespace(ctx *TablespaceContext) {}

// EnterLocality_name is called when production locality_name is entered.
func (s *BaseOBParserListener) EnterLocality_name(ctx *Locality_nameContext) {}

// ExitLocality_name is called when production locality_name is exited.
func (s *BaseOBParserListener) ExitLocality_name(ctx *Locality_nameContext) {}

// EnterTable_option is called when production table_option is entered.
func (s *BaseOBParserListener) EnterTable_option(ctx *Table_optionContext) {}

// ExitTable_option is called when production table_option is exited.
func (s *BaseOBParserListener) ExitTable_option(ctx *Table_optionContext) {}

// EnterRelation_name_or_string is called when production relation_name_or_string is entered.
func (s *BaseOBParserListener) EnterRelation_name_or_string(ctx *Relation_name_or_stringContext) {}

// ExitRelation_name_or_string is called when production relation_name_or_string is exited.
func (s *BaseOBParserListener) ExitRelation_name_or_string(ctx *Relation_name_or_stringContext) {}

// EnterOpt_equal_mark is called when production opt_equal_mark is entered.
func (s *BaseOBParserListener) EnterOpt_equal_mark(ctx *Opt_equal_markContext) {}

// ExitOpt_equal_mark is called when production opt_equal_mark is exited.
func (s *BaseOBParserListener) ExitOpt_equal_mark(ctx *Opt_equal_markContext) {}

// EnterPartition_option is called when production partition_option is entered.
func (s *BaseOBParserListener) EnterPartition_option(ctx *Partition_optionContext) {}

// ExitPartition_option is called when production partition_option is exited.
func (s *BaseOBParserListener) ExitPartition_option(ctx *Partition_optionContext) {}

// EnterOpt_partition_option is called when production opt_partition_option is entered.
func (s *BaseOBParserListener) EnterOpt_partition_option(ctx *Opt_partition_optionContext) {}

// ExitOpt_partition_option is called when production opt_partition_option is exited.
func (s *BaseOBParserListener) ExitOpt_partition_option(ctx *Opt_partition_optionContext) {}

// EnterHash_partition_option is called when production hash_partition_option is entered.
func (s *BaseOBParserListener) EnterHash_partition_option(ctx *Hash_partition_optionContext) {}

// ExitHash_partition_option is called when production hash_partition_option is exited.
func (s *BaseOBParserListener) ExitHash_partition_option(ctx *Hash_partition_optionContext) {}

// EnterList_partition_option is called when production list_partition_option is entered.
func (s *BaseOBParserListener) EnterList_partition_option(ctx *List_partition_optionContext) {}

// ExitList_partition_option is called when production list_partition_option is exited.
func (s *BaseOBParserListener) ExitList_partition_option(ctx *List_partition_optionContext) {}

// EnterKey_partition_option is called when production key_partition_option is entered.
func (s *BaseOBParserListener) EnterKey_partition_option(ctx *Key_partition_optionContext) {}

// ExitKey_partition_option is called when production key_partition_option is exited.
func (s *BaseOBParserListener) ExitKey_partition_option(ctx *Key_partition_optionContext) {}

// EnterRange_partition_option is called when production range_partition_option is entered.
func (s *BaseOBParserListener) EnterRange_partition_option(ctx *Range_partition_optionContext) {}

// ExitRange_partition_option is called when production range_partition_option is exited.
func (s *BaseOBParserListener) ExitRange_partition_option(ctx *Range_partition_optionContext) {}

// EnterOpt_column_partition_option is called when production opt_column_partition_option is entered.
func (s *BaseOBParserListener) EnterOpt_column_partition_option(ctx *Opt_column_partition_optionContext) {}

// ExitOpt_column_partition_option is called when production opt_column_partition_option is exited.
func (s *BaseOBParserListener) ExitOpt_column_partition_option(ctx *Opt_column_partition_optionContext) {}

// EnterColumn_partition_option is called when production column_partition_option is entered.
func (s *BaseOBParserListener) EnterColumn_partition_option(ctx *Column_partition_optionContext) {}

// ExitColumn_partition_option is called when production column_partition_option is exited.
func (s *BaseOBParserListener) ExitColumn_partition_option(ctx *Column_partition_optionContext) {}

// EnterAux_column_list is called when production aux_column_list is entered.
func (s *BaseOBParserListener) EnterAux_column_list(ctx *Aux_column_listContext) {}

// ExitAux_column_list is called when production aux_column_list is exited.
func (s *BaseOBParserListener) ExitAux_column_list(ctx *Aux_column_listContext) {}

// EnterVertical_column_name is called when production vertical_column_name is entered.
func (s *BaseOBParserListener) EnterVertical_column_name(ctx *Vertical_column_nameContext) {}

// ExitVertical_column_name is called when production vertical_column_name is exited.
func (s *BaseOBParserListener) ExitVertical_column_name(ctx *Vertical_column_nameContext) {}

// EnterColumn_name_list is called when production column_name_list is entered.
func (s *BaseOBParserListener) EnterColumn_name_list(ctx *Column_name_listContext) {}

// ExitColumn_name_list is called when production column_name_list is exited.
func (s *BaseOBParserListener) ExitColumn_name_list(ctx *Column_name_listContext) {}

// EnterSubpartition_option is called when production subpartition_option is entered.
func (s *BaseOBParserListener) EnterSubpartition_option(ctx *Subpartition_optionContext) {}

// ExitSubpartition_option is called when production subpartition_option is exited.
func (s *BaseOBParserListener) ExitSubpartition_option(ctx *Subpartition_optionContext) {}

// EnterOpt_list_partition_list is called when production opt_list_partition_list is entered.
func (s *BaseOBParserListener) EnterOpt_list_partition_list(ctx *Opt_list_partition_listContext) {}

// ExitOpt_list_partition_list is called when production opt_list_partition_list is exited.
func (s *BaseOBParserListener) ExitOpt_list_partition_list(ctx *Opt_list_partition_listContext) {}

// EnterOpt_list_subpartition_list is called when production opt_list_subpartition_list is entered.
func (s *BaseOBParserListener) EnterOpt_list_subpartition_list(ctx *Opt_list_subpartition_listContext) {}

// ExitOpt_list_subpartition_list is called when production opt_list_subpartition_list is exited.
func (s *BaseOBParserListener) ExitOpt_list_subpartition_list(ctx *Opt_list_subpartition_listContext) {}

// EnterOpt_range_partition_list is called when production opt_range_partition_list is entered.
func (s *BaseOBParserListener) EnterOpt_range_partition_list(ctx *Opt_range_partition_listContext) {}

// ExitOpt_range_partition_list is called when production opt_range_partition_list is exited.
func (s *BaseOBParserListener) ExitOpt_range_partition_list(ctx *Opt_range_partition_listContext) {}

// EnterOpt_range_subpartition_list is called when production opt_range_subpartition_list is entered.
func (s *BaseOBParserListener) EnterOpt_range_subpartition_list(ctx *Opt_range_subpartition_listContext) {}

// ExitOpt_range_subpartition_list is called when production opt_range_subpartition_list is exited.
func (s *BaseOBParserListener) ExitOpt_range_subpartition_list(ctx *Opt_range_subpartition_listContext) {}

// EnterList_partition_list is called when production list_partition_list is entered.
func (s *BaseOBParserListener) EnterList_partition_list(ctx *List_partition_listContext) {}

// ExitList_partition_list is called when production list_partition_list is exited.
func (s *BaseOBParserListener) ExitList_partition_list(ctx *List_partition_listContext) {}

// EnterList_subpartition_list is called when production list_subpartition_list is entered.
func (s *BaseOBParserListener) EnterList_subpartition_list(ctx *List_subpartition_listContext) {}

// ExitList_subpartition_list is called when production list_subpartition_list is exited.
func (s *BaseOBParserListener) ExitList_subpartition_list(ctx *List_subpartition_listContext) {}

// EnterList_subpartition_element is called when production list_subpartition_element is entered.
func (s *BaseOBParserListener) EnterList_subpartition_element(ctx *List_subpartition_elementContext) {}

// ExitList_subpartition_element is called when production list_subpartition_element is exited.
func (s *BaseOBParserListener) ExitList_subpartition_element(ctx *List_subpartition_elementContext) {}

// EnterList_partition_element is called when production list_partition_element is entered.
func (s *BaseOBParserListener) EnterList_partition_element(ctx *List_partition_elementContext) {}

// ExitList_partition_element is called when production list_partition_element is exited.
func (s *BaseOBParserListener) ExitList_partition_element(ctx *List_partition_elementContext) {}

// EnterList_partition_expr is called when production list_partition_expr is entered.
func (s *BaseOBParserListener) EnterList_partition_expr(ctx *List_partition_exprContext) {}

// ExitList_partition_expr is called when production list_partition_expr is exited.
func (s *BaseOBParserListener) ExitList_partition_expr(ctx *List_partition_exprContext) {}

// EnterList_expr is called when production list_expr is entered.
func (s *BaseOBParserListener) EnterList_expr(ctx *List_exprContext) {}

// ExitList_expr is called when production list_expr is exited.
func (s *BaseOBParserListener) ExitList_expr(ctx *List_exprContext) {}

// EnterRange_partition_list is called when production range_partition_list is entered.
func (s *BaseOBParserListener) EnterRange_partition_list(ctx *Range_partition_listContext) {}

// ExitRange_partition_list is called when production range_partition_list is exited.
func (s *BaseOBParserListener) ExitRange_partition_list(ctx *Range_partition_listContext) {}

// EnterRange_partition_element is called when production range_partition_element is entered.
func (s *BaseOBParserListener) EnterRange_partition_element(ctx *Range_partition_elementContext) {}

// ExitRange_partition_element is called when production range_partition_element is exited.
func (s *BaseOBParserListener) ExitRange_partition_element(ctx *Range_partition_elementContext) {}

// EnterRange_subpartition_element is called when production range_subpartition_element is entered.
func (s *BaseOBParserListener) EnterRange_subpartition_element(ctx *Range_subpartition_elementContext) {}

// ExitRange_subpartition_element is called when production range_subpartition_element is exited.
func (s *BaseOBParserListener) ExitRange_subpartition_element(ctx *Range_subpartition_elementContext) {}

// EnterRange_subpartition_list is called when production range_subpartition_list is entered.
func (s *BaseOBParserListener) EnterRange_subpartition_list(ctx *Range_subpartition_listContext) {}

// ExitRange_subpartition_list is called when production range_subpartition_list is exited.
func (s *BaseOBParserListener) ExitRange_subpartition_list(ctx *Range_subpartition_listContext) {}

// EnterRange_partition_expr is called when production range_partition_expr is entered.
func (s *BaseOBParserListener) EnterRange_partition_expr(ctx *Range_partition_exprContext) {}

// ExitRange_partition_expr is called when production range_partition_expr is exited.
func (s *BaseOBParserListener) ExitRange_partition_expr(ctx *Range_partition_exprContext) {}

// EnterRange_expr_list is called when production range_expr_list is entered.
func (s *BaseOBParserListener) EnterRange_expr_list(ctx *Range_expr_listContext) {}

// ExitRange_expr_list is called when production range_expr_list is exited.
func (s *BaseOBParserListener) ExitRange_expr_list(ctx *Range_expr_listContext) {}

// EnterRange_expr is called when production range_expr is entered.
func (s *BaseOBParserListener) EnterRange_expr(ctx *Range_exprContext) {}

// ExitRange_expr is called when production range_expr is exited.
func (s *BaseOBParserListener) ExitRange_expr(ctx *Range_exprContext) {}

// EnterInt_or_decimal is called when production int_or_decimal is entered.
func (s *BaseOBParserListener) EnterInt_or_decimal(ctx *Int_or_decimalContext) {}

// ExitInt_or_decimal is called when production int_or_decimal is exited.
func (s *BaseOBParserListener) ExitInt_or_decimal(ctx *Int_or_decimalContext) {}

// EnterTg_hash_partition_option is called when production tg_hash_partition_option is entered.
func (s *BaseOBParserListener) EnterTg_hash_partition_option(ctx *Tg_hash_partition_optionContext) {}

// ExitTg_hash_partition_option is called when production tg_hash_partition_option is exited.
func (s *BaseOBParserListener) ExitTg_hash_partition_option(ctx *Tg_hash_partition_optionContext) {}

// EnterTg_key_partition_option is called when production tg_key_partition_option is entered.
func (s *BaseOBParserListener) EnterTg_key_partition_option(ctx *Tg_key_partition_optionContext) {}

// ExitTg_key_partition_option is called when production tg_key_partition_option is exited.
func (s *BaseOBParserListener) ExitTg_key_partition_option(ctx *Tg_key_partition_optionContext) {}

// EnterTg_range_partition_option is called when production tg_range_partition_option is entered.
func (s *BaseOBParserListener) EnterTg_range_partition_option(ctx *Tg_range_partition_optionContext) {}

// ExitTg_range_partition_option is called when production tg_range_partition_option is exited.
func (s *BaseOBParserListener) ExitTg_range_partition_option(ctx *Tg_range_partition_optionContext) {}

// EnterTg_list_partition_option is called when production tg_list_partition_option is entered.
func (s *BaseOBParserListener) EnterTg_list_partition_option(ctx *Tg_list_partition_optionContext) {}

// ExitTg_list_partition_option is called when production tg_list_partition_option is exited.
func (s *BaseOBParserListener) ExitTg_list_partition_option(ctx *Tg_list_partition_optionContext) {}

// EnterTg_subpartition_option is called when production tg_subpartition_option is entered.
func (s *BaseOBParserListener) EnterTg_subpartition_option(ctx *Tg_subpartition_optionContext) {}

// ExitTg_subpartition_option is called when production tg_subpartition_option is exited.
func (s *BaseOBParserListener) ExitTg_subpartition_option(ctx *Tg_subpartition_optionContext) {}

// EnterRow_format_option is called when production row_format_option is entered.
func (s *BaseOBParserListener) EnterRow_format_option(ctx *Row_format_optionContext) {}

// ExitRow_format_option is called when production row_format_option is exited.
func (s *BaseOBParserListener) ExitRow_format_option(ctx *Row_format_optionContext) {}

// EnterCreate_tablegroup_stmt is called when production create_tablegroup_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_tablegroup_stmt(ctx *Create_tablegroup_stmtContext) {}

// ExitCreate_tablegroup_stmt is called when production create_tablegroup_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_tablegroup_stmt(ctx *Create_tablegroup_stmtContext) {}

// EnterDrop_tablegroup_stmt is called when production drop_tablegroup_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_tablegroup_stmt(ctx *Drop_tablegroup_stmtContext) {}

// ExitDrop_tablegroup_stmt is called when production drop_tablegroup_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_tablegroup_stmt(ctx *Drop_tablegroup_stmtContext) {}

// EnterAlter_tablegroup_stmt is called when production alter_tablegroup_stmt is entered.
func (s *BaseOBParserListener) EnterAlter_tablegroup_stmt(ctx *Alter_tablegroup_stmtContext) {}

// ExitAlter_tablegroup_stmt is called when production alter_tablegroup_stmt is exited.
func (s *BaseOBParserListener) ExitAlter_tablegroup_stmt(ctx *Alter_tablegroup_stmtContext) {}

// EnterTablegroup_option_list_space_seperated is called when production tablegroup_option_list_space_seperated is entered.
func (s *BaseOBParserListener) EnterTablegroup_option_list_space_seperated(ctx *Tablegroup_option_list_space_seperatedContext) {}

// ExitTablegroup_option_list_space_seperated is called when production tablegroup_option_list_space_seperated is exited.
func (s *BaseOBParserListener) ExitTablegroup_option_list_space_seperated(ctx *Tablegroup_option_list_space_seperatedContext) {}

// EnterTablegroup_option_list is called when production tablegroup_option_list is entered.
func (s *BaseOBParserListener) EnterTablegroup_option_list(ctx *Tablegroup_option_listContext) {}

// ExitTablegroup_option_list is called when production tablegroup_option_list is exited.
func (s *BaseOBParserListener) ExitTablegroup_option_list(ctx *Tablegroup_option_listContext) {}

// EnterTablegroup_option is called when production tablegroup_option is entered.
func (s *BaseOBParserListener) EnterTablegroup_option(ctx *Tablegroup_optionContext) {}

// ExitTablegroup_option is called when production tablegroup_option is exited.
func (s *BaseOBParserListener) ExitTablegroup_option(ctx *Tablegroup_optionContext) {}

// EnterAlter_tablegroup_actions is called when production alter_tablegroup_actions is entered.
func (s *BaseOBParserListener) EnterAlter_tablegroup_actions(ctx *Alter_tablegroup_actionsContext) {}

// ExitAlter_tablegroup_actions is called when production alter_tablegroup_actions is exited.
func (s *BaseOBParserListener) ExitAlter_tablegroup_actions(ctx *Alter_tablegroup_actionsContext) {}

// EnterAlter_tablegroup_action is called when production alter_tablegroup_action is entered.
func (s *BaseOBParserListener) EnterAlter_tablegroup_action(ctx *Alter_tablegroup_actionContext) {}

// ExitAlter_tablegroup_action is called when production alter_tablegroup_action is exited.
func (s *BaseOBParserListener) ExitAlter_tablegroup_action(ctx *Alter_tablegroup_actionContext) {}

// EnterDefault_tablegroup is called when production default_tablegroup is entered.
func (s *BaseOBParserListener) EnterDefault_tablegroup(ctx *Default_tablegroupContext) {}

// ExitDefault_tablegroup is called when production default_tablegroup is exited.
func (s *BaseOBParserListener) ExitDefault_tablegroup(ctx *Default_tablegroupContext) {}

// EnterCreate_view_stmt is called when production create_view_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_view_stmt(ctx *Create_view_stmtContext) {}

// ExitCreate_view_stmt is called when production create_view_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_view_stmt(ctx *Create_view_stmtContext) {}

// EnterView_select_stmt is called when production view_select_stmt is entered.
func (s *BaseOBParserListener) EnterView_select_stmt(ctx *View_select_stmtContext) {}

// ExitView_select_stmt is called when production view_select_stmt is exited.
func (s *BaseOBParserListener) ExitView_select_stmt(ctx *View_select_stmtContext) {}

// EnterView_name is called when production view_name is entered.
func (s *BaseOBParserListener) EnterView_name(ctx *View_nameContext) {}

// ExitView_name is called when production view_name is exited.
func (s *BaseOBParserListener) ExitView_name(ctx *View_nameContext) {}

// EnterCreate_index_stmt is called when production create_index_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_index_stmt(ctx *Create_index_stmtContext) {}

// ExitCreate_index_stmt is called when production create_index_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_index_stmt(ctx *Create_index_stmtContext) {}

// EnterIndex_name is called when production index_name is entered.
func (s *BaseOBParserListener) EnterIndex_name(ctx *Index_nameContext) {}

// ExitIndex_name is called when production index_name is exited.
func (s *BaseOBParserListener) ExitIndex_name(ctx *Index_nameContext) {}

// EnterOpt_constraint_name is called when production opt_constraint_name is entered.
func (s *BaseOBParserListener) EnterOpt_constraint_name(ctx *Opt_constraint_nameContext) {}

// ExitOpt_constraint_name is called when production opt_constraint_name is exited.
func (s *BaseOBParserListener) ExitOpt_constraint_name(ctx *Opt_constraint_nameContext) {}

// EnterConstraint_name is called when production constraint_name is entered.
func (s *BaseOBParserListener) EnterConstraint_name(ctx *Constraint_nameContext) {}

// ExitConstraint_name is called when production constraint_name is exited.
func (s *BaseOBParserListener) ExitConstraint_name(ctx *Constraint_nameContext) {}

// EnterSort_column_list is called when production sort_column_list is entered.
func (s *BaseOBParserListener) EnterSort_column_list(ctx *Sort_column_listContext) {}

// ExitSort_column_list is called when production sort_column_list is exited.
func (s *BaseOBParserListener) ExitSort_column_list(ctx *Sort_column_listContext) {}

// EnterSort_column_key is called when production sort_column_key is entered.
func (s *BaseOBParserListener) EnterSort_column_key(ctx *Sort_column_keyContext) {}

// ExitSort_column_key is called when production sort_column_key is exited.
func (s *BaseOBParserListener) ExitSort_column_key(ctx *Sort_column_keyContext) {}

// EnterOpt_index_options is called when production opt_index_options is entered.
func (s *BaseOBParserListener) EnterOpt_index_options(ctx *Opt_index_optionsContext) {}

// ExitOpt_index_options is called when production opt_index_options is exited.
func (s *BaseOBParserListener) ExitOpt_index_options(ctx *Opt_index_optionsContext) {}

// EnterIndex_option is called when production index_option is entered.
func (s *BaseOBParserListener) EnterIndex_option(ctx *Index_optionContext) {}

// ExitIndex_option is called when production index_option is exited.
func (s *BaseOBParserListener) ExitIndex_option(ctx *Index_optionContext) {}

// EnterIndex_using_algorithm is called when production index_using_algorithm is entered.
func (s *BaseOBParserListener) EnterIndex_using_algorithm(ctx *Index_using_algorithmContext) {}

// ExitIndex_using_algorithm is called when production index_using_algorithm is exited.
func (s *BaseOBParserListener) ExitIndex_using_algorithm(ctx *Index_using_algorithmContext) {}

// EnterDrop_table_stmt is called when production drop_table_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_table_stmt(ctx *Drop_table_stmtContext) {}

// ExitDrop_table_stmt is called when production drop_table_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_table_stmt(ctx *Drop_table_stmtContext) {}

// EnterTable_or_tables is called when production table_or_tables is entered.
func (s *BaseOBParserListener) EnterTable_or_tables(ctx *Table_or_tablesContext) {}

// ExitTable_or_tables is called when production table_or_tables is exited.
func (s *BaseOBParserListener) ExitTable_or_tables(ctx *Table_or_tablesContext) {}

// EnterDrop_view_stmt is called when production drop_view_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_view_stmt(ctx *Drop_view_stmtContext) {}

// ExitDrop_view_stmt is called when production drop_view_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_view_stmt(ctx *Drop_view_stmtContext) {}

// EnterTable_list is called when production table_list is entered.
func (s *BaseOBParserListener) EnterTable_list(ctx *Table_listContext) {}

// ExitTable_list is called when production table_list is exited.
func (s *BaseOBParserListener) ExitTable_list(ctx *Table_listContext) {}

// EnterDrop_index_stmt is called when production drop_index_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_index_stmt(ctx *Drop_index_stmtContext) {}

// ExitDrop_index_stmt is called when production drop_index_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_index_stmt(ctx *Drop_index_stmtContext) {}

// EnterInsert_stmt is called when production insert_stmt is entered.
func (s *BaseOBParserListener) EnterInsert_stmt(ctx *Insert_stmtContext) {}

// ExitInsert_stmt is called when production insert_stmt is exited.
func (s *BaseOBParserListener) ExitInsert_stmt(ctx *Insert_stmtContext) {}

// EnterSingle_table_insert is called when production single_table_insert is entered.
func (s *BaseOBParserListener) EnterSingle_table_insert(ctx *Single_table_insertContext) {}

// ExitSingle_table_insert is called when production single_table_insert is exited.
func (s *BaseOBParserListener) ExitSingle_table_insert(ctx *Single_table_insertContext) {}

// EnterValues_clause is called when production values_clause is entered.
func (s *BaseOBParserListener) EnterValues_clause(ctx *Values_clauseContext) {}

// ExitValues_clause is called when production values_clause is exited.
func (s *BaseOBParserListener) ExitValues_clause(ctx *Values_clauseContext) {}

// EnterValue_or_values is called when production value_or_values is entered.
func (s *BaseOBParserListener) EnterValue_or_values(ctx *Value_or_valuesContext) {}

// ExitValue_or_values is called when production value_or_values is exited.
func (s *BaseOBParserListener) ExitValue_or_values(ctx *Value_or_valuesContext) {}

// EnterReplace_with_opt_hint is called when production replace_with_opt_hint is entered.
func (s *BaseOBParserListener) EnterReplace_with_opt_hint(ctx *Replace_with_opt_hintContext) {}

// ExitReplace_with_opt_hint is called when production replace_with_opt_hint is exited.
func (s *BaseOBParserListener) ExitReplace_with_opt_hint(ctx *Replace_with_opt_hintContext) {}

// EnterInsert_with_opt_hint is called when production insert_with_opt_hint is entered.
func (s *BaseOBParserListener) EnterInsert_with_opt_hint(ctx *Insert_with_opt_hintContext) {}

// ExitInsert_with_opt_hint is called when production insert_with_opt_hint is exited.
func (s *BaseOBParserListener) ExitInsert_with_opt_hint(ctx *Insert_with_opt_hintContext) {}

// EnterColumn_list is called when production column_list is entered.
func (s *BaseOBParserListener) EnterColumn_list(ctx *Column_listContext) {}

// ExitColumn_list is called when production column_list is exited.
func (s *BaseOBParserListener) ExitColumn_list(ctx *Column_listContext) {}

// EnterInsert_vals_list is called when production insert_vals_list is entered.
func (s *BaseOBParserListener) EnterInsert_vals_list(ctx *Insert_vals_listContext) {}

// ExitInsert_vals_list is called when production insert_vals_list is exited.
func (s *BaseOBParserListener) ExitInsert_vals_list(ctx *Insert_vals_listContext) {}

// EnterInsert_vals is called when production insert_vals is entered.
func (s *BaseOBParserListener) EnterInsert_vals(ctx *Insert_valsContext) {}

// ExitInsert_vals is called when production insert_vals is exited.
func (s *BaseOBParserListener) ExitInsert_vals(ctx *Insert_valsContext) {}

// EnterExpr_or_default is called when production expr_or_default is entered.
func (s *BaseOBParserListener) EnterExpr_or_default(ctx *Expr_or_defaultContext) {}

// ExitExpr_or_default is called when production expr_or_default is exited.
func (s *BaseOBParserListener) ExitExpr_or_default(ctx *Expr_or_defaultContext) {}

// EnterSelect_stmt is called when production select_stmt is entered.
func (s *BaseOBParserListener) EnterSelect_stmt(ctx *Select_stmtContext) {}

// ExitSelect_stmt is called when production select_stmt is exited.
func (s *BaseOBParserListener) ExitSelect_stmt(ctx *Select_stmtContext) {}

// EnterSelect_into is called when production select_into is entered.
func (s *BaseOBParserListener) EnterSelect_into(ctx *Select_intoContext) {}

// ExitSelect_into is called when production select_into is exited.
func (s *BaseOBParserListener) ExitSelect_into(ctx *Select_intoContext) {}

// EnterSelect_with_parens is called when production select_with_parens is entered.
func (s *BaseOBParserListener) EnterSelect_with_parens(ctx *Select_with_parensContext) {}

// ExitSelect_with_parens is called when production select_with_parens is exited.
func (s *BaseOBParserListener) ExitSelect_with_parens(ctx *Select_with_parensContext) {}

// EnterSelect_no_parens is called when production select_no_parens is entered.
func (s *BaseOBParserListener) EnterSelect_no_parens(ctx *Select_no_parensContext) {}

// ExitSelect_no_parens is called when production select_no_parens is exited.
func (s *BaseOBParserListener) ExitSelect_no_parens(ctx *Select_no_parensContext) {}

// EnterNo_table_select is called when production no_table_select is entered.
func (s *BaseOBParserListener) EnterNo_table_select(ctx *No_table_selectContext) {}

// ExitNo_table_select is called when production no_table_select is exited.
func (s *BaseOBParserListener) ExitNo_table_select(ctx *No_table_selectContext) {}

// EnterSelect_clause is called when production select_clause is entered.
func (s *BaseOBParserListener) EnterSelect_clause(ctx *Select_clauseContext) {}

// ExitSelect_clause is called when production select_clause is exited.
func (s *BaseOBParserListener) ExitSelect_clause(ctx *Select_clauseContext) {}

// EnterSelect_clause_set_with_order_and_limit is called when production select_clause_set_with_order_and_limit is entered.
func (s *BaseOBParserListener) EnterSelect_clause_set_with_order_and_limit(ctx *Select_clause_set_with_order_and_limitContext) {}

// ExitSelect_clause_set_with_order_and_limit is called when production select_clause_set_with_order_and_limit is exited.
func (s *BaseOBParserListener) ExitSelect_clause_set_with_order_and_limit(ctx *Select_clause_set_with_order_and_limitContext) {}

// EnterSelect_clause_set is called when production select_clause_set is entered.
func (s *BaseOBParserListener) EnterSelect_clause_set(ctx *Select_clause_setContext) {}

// ExitSelect_clause_set is called when production select_clause_set is exited.
func (s *BaseOBParserListener) ExitSelect_clause_set(ctx *Select_clause_setContext) {}

// EnterSelect_clause_set_right is called when production select_clause_set_right is entered.
func (s *BaseOBParserListener) EnterSelect_clause_set_right(ctx *Select_clause_set_rightContext) {}

// ExitSelect_clause_set_right is called when production select_clause_set_right is exited.
func (s *BaseOBParserListener) ExitSelect_clause_set_right(ctx *Select_clause_set_rightContext) {}

// EnterSelect_clause_set_left is called when production select_clause_set_left is entered.
func (s *BaseOBParserListener) EnterSelect_clause_set_left(ctx *Select_clause_set_leftContext) {}

// ExitSelect_clause_set_left is called when production select_clause_set_left is exited.
func (s *BaseOBParserListener) ExitSelect_clause_set_left(ctx *Select_clause_set_leftContext) {}

// EnterNo_table_select_with_order_and_limit is called when production no_table_select_with_order_and_limit is entered.
func (s *BaseOBParserListener) EnterNo_table_select_with_order_and_limit(ctx *No_table_select_with_order_and_limitContext) {}

// ExitNo_table_select_with_order_and_limit is called when production no_table_select_with_order_and_limit is exited.
func (s *BaseOBParserListener) ExitNo_table_select_with_order_and_limit(ctx *No_table_select_with_order_and_limitContext) {}

// EnterSimple_select_with_order_and_limit is called when production simple_select_with_order_and_limit is entered.
func (s *BaseOBParserListener) EnterSimple_select_with_order_and_limit(ctx *Simple_select_with_order_and_limitContext) {}

// ExitSimple_select_with_order_and_limit is called when production simple_select_with_order_and_limit is exited.
func (s *BaseOBParserListener) ExitSimple_select_with_order_and_limit(ctx *Simple_select_with_order_and_limitContext) {}

// EnterSelect_with_parens_with_order_and_limit is called when production select_with_parens_with_order_and_limit is entered.
func (s *BaseOBParserListener) EnterSelect_with_parens_with_order_and_limit(ctx *Select_with_parens_with_order_and_limitContext) {}

// ExitSelect_with_parens_with_order_and_limit is called when production select_with_parens_with_order_and_limit is exited.
func (s *BaseOBParserListener) ExitSelect_with_parens_with_order_and_limit(ctx *Select_with_parens_with_order_and_limitContext) {}

// EnterSelect_with_opt_hint is called when production select_with_opt_hint is entered.
func (s *BaseOBParserListener) EnterSelect_with_opt_hint(ctx *Select_with_opt_hintContext) {}

// ExitSelect_with_opt_hint is called when production select_with_opt_hint is exited.
func (s *BaseOBParserListener) ExitSelect_with_opt_hint(ctx *Select_with_opt_hintContext) {}

// EnterUpdate_with_opt_hint is called when production update_with_opt_hint is entered.
func (s *BaseOBParserListener) EnterUpdate_with_opt_hint(ctx *Update_with_opt_hintContext) {}

// ExitUpdate_with_opt_hint is called when production update_with_opt_hint is exited.
func (s *BaseOBParserListener) ExitUpdate_with_opt_hint(ctx *Update_with_opt_hintContext) {}

// EnterDelete_with_opt_hint is called when production delete_with_opt_hint is entered.
func (s *BaseOBParserListener) EnterDelete_with_opt_hint(ctx *Delete_with_opt_hintContext) {}

// ExitDelete_with_opt_hint is called when production delete_with_opt_hint is exited.
func (s *BaseOBParserListener) ExitDelete_with_opt_hint(ctx *Delete_with_opt_hintContext) {}

// EnterSimple_select is called when production simple_select is entered.
func (s *BaseOBParserListener) EnterSimple_select(ctx *Simple_selectContext) {}

// ExitSimple_select is called when production simple_select is exited.
func (s *BaseOBParserListener) ExitSimple_select(ctx *Simple_selectContext) {}

// EnterSet_type_union is called when production set_type_union is entered.
func (s *BaseOBParserListener) EnterSet_type_union(ctx *Set_type_unionContext) {}

// ExitSet_type_union is called when production set_type_union is exited.
func (s *BaseOBParserListener) ExitSet_type_union(ctx *Set_type_unionContext) {}

// EnterSet_type_other is called when production set_type_other is entered.
func (s *BaseOBParserListener) EnterSet_type_other(ctx *Set_type_otherContext) {}

// ExitSet_type_other is called when production set_type_other is exited.
func (s *BaseOBParserListener) ExitSet_type_other(ctx *Set_type_otherContext) {}

// EnterSet_type is called when production set_type is entered.
func (s *BaseOBParserListener) EnterSet_type(ctx *Set_typeContext) {}

// ExitSet_type is called when production set_type is exited.
func (s *BaseOBParserListener) ExitSet_type(ctx *Set_typeContext) {}

// EnterSet_expression_option is called when production set_expression_option is entered.
func (s *BaseOBParserListener) EnterSet_expression_option(ctx *Set_expression_optionContext) {}

// ExitSet_expression_option is called when production set_expression_option is exited.
func (s *BaseOBParserListener) ExitSet_expression_option(ctx *Set_expression_optionContext) {}

// EnterOpt_hint_value is called when production opt_hint_value is entered.
func (s *BaseOBParserListener) EnterOpt_hint_value(ctx *Opt_hint_valueContext) {}

// ExitOpt_hint_value is called when production opt_hint_value is exited.
func (s *BaseOBParserListener) ExitOpt_hint_value(ctx *Opt_hint_valueContext) {}

// EnterLimit_clause is called when production limit_clause is entered.
func (s *BaseOBParserListener) EnterLimit_clause(ctx *Limit_clauseContext) {}

// ExitLimit_clause is called when production limit_clause is exited.
func (s *BaseOBParserListener) ExitLimit_clause(ctx *Limit_clauseContext) {}

// EnterInto_clause is called when production into_clause is entered.
func (s *BaseOBParserListener) EnterInto_clause(ctx *Into_clauseContext) {}

// ExitInto_clause is called when production into_clause is exited.
func (s *BaseOBParserListener) ExitInto_clause(ctx *Into_clauseContext) {}

// EnterInto_opt is called when production into_opt is entered.
func (s *BaseOBParserListener) EnterInto_opt(ctx *Into_optContext) {}

// ExitInto_opt is called when production into_opt is exited.
func (s *BaseOBParserListener) ExitInto_opt(ctx *Into_optContext) {}

// EnterInto_var_list is called when production into_var_list is entered.
func (s *BaseOBParserListener) EnterInto_var_list(ctx *Into_var_listContext) {}

// ExitInto_var_list is called when production into_var_list is exited.
func (s *BaseOBParserListener) ExitInto_var_list(ctx *Into_var_listContext) {}

// EnterInto_var is called when production into_var is entered.
func (s *BaseOBParserListener) EnterInto_var(ctx *Into_varContext) {}

// ExitInto_var is called when production into_var is exited.
func (s *BaseOBParserListener) ExitInto_var(ctx *Into_varContext) {}

// EnterField_opt is called when production field_opt is entered.
func (s *BaseOBParserListener) EnterField_opt(ctx *Field_optContext) {}

// ExitField_opt is called when production field_opt is exited.
func (s *BaseOBParserListener) ExitField_opt(ctx *Field_optContext) {}

// EnterField_term_list is called when production field_term_list is entered.
func (s *BaseOBParserListener) EnterField_term_list(ctx *Field_term_listContext) {}

// ExitField_term_list is called when production field_term_list is exited.
func (s *BaseOBParserListener) ExitField_term_list(ctx *Field_term_listContext) {}

// EnterField_term is called when production field_term is entered.
func (s *BaseOBParserListener) EnterField_term(ctx *Field_termContext) {}

// ExitField_term is called when production field_term is exited.
func (s *BaseOBParserListener) ExitField_term(ctx *Field_termContext) {}

// EnterLine_opt is called when production line_opt is entered.
func (s *BaseOBParserListener) EnterLine_opt(ctx *Line_optContext) {}

// ExitLine_opt is called when production line_opt is exited.
func (s *BaseOBParserListener) ExitLine_opt(ctx *Line_optContext) {}

// EnterLine_term_list is called when production line_term_list is entered.
func (s *BaseOBParserListener) EnterLine_term_list(ctx *Line_term_listContext) {}

// ExitLine_term_list is called when production line_term_list is exited.
func (s *BaseOBParserListener) ExitLine_term_list(ctx *Line_term_listContext) {}

// EnterLine_term is called when production line_term is entered.
func (s *BaseOBParserListener) EnterLine_term(ctx *Line_termContext) {}

// ExitLine_term is called when production line_term is exited.
func (s *BaseOBParserListener) ExitLine_term(ctx *Line_termContext) {}

// EnterHint_list_with_end is called when production hint_list_with_end is entered.
func (s *BaseOBParserListener) EnterHint_list_with_end(ctx *Hint_list_with_endContext) {}

// ExitHint_list_with_end is called when production hint_list_with_end is exited.
func (s *BaseOBParserListener) ExitHint_list_with_end(ctx *Hint_list_with_endContext) {}

// EnterOpt_hint_list is called when production opt_hint_list is entered.
func (s *BaseOBParserListener) EnterOpt_hint_list(ctx *Opt_hint_listContext) {}

// ExitOpt_hint_list is called when production opt_hint_list is exited.
func (s *BaseOBParserListener) ExitOpt_hint_list(ctx *Opt_hint_listContext) {}

// EnterHint_options is called when production hint_options is entered.
func (s *BaseOBParserListener) EnterHint_options(ctx *Hint_optionsContext) {}

// ExitHint_options is called when production hint_options is exited.
func (s *BaseOBParserListener) ExitHint_options(ctx *Hint_optionsContext) {}

// EnterName_list is called when production name_list is entered.
func (s *BaseOBParserListener) EnterName_list(ctx *Name_listContext) {}

// ExitName_list is called when production name_list is exited.
func (s *BaseOBParserListener) ExitName_list(ctx *Name_listContext) {}

// EnterHint_option is called when production hint_option is entered.
func (s *BaseOBParserListener) EnterHint_option(ctx *Hint_optionContext) {}

// ExitHint_option is called when production hint_option is exited.
func (s *BaseOBParserListener) ExitHint_option(ctx *Hint_optionContext) {}

// EnterConsistency_level is called when production consistency_level is entered.
func (s *BaseOBParserListener) EnterConsistency_level(ctx *Consistency_levelContext) {}

// ExitConsistency_level is called when production consistency_level is exited.
func (s *BaseOBParserListener) ExitConsistency_level(ctx *Consistency_levelContext) {}

// EnterUse_plan_cache_type is called when production use_plan_cache_type is entered.
func (s *BaseOBParserListener) EnterUse_plan_cache_type(ctx *Use_plan_cache_typeContext) {}

// ExitUse_plan_cache_type is called when production use_plan_cache_type is exited.
func (s *BaseOBParserListener) ExitUse_plan_cache_type(ctx *Use_plan_cache_typeContext) {}

// EnterUse_jit_type is called when production use_jit_type is entered.
func (s *BaseOBParserListener) EnterUse_jit_type(ctx *Use_jit_typeContext) {}

// ExitUse_jit_type is called when production use_jit_type is exited.
func (s *BaseOBParserListener) ExitUse_jit_type(ctx *Use_jit_typeContext) {}

// EnterDistribute_method is called when production distribute_method is entered.
func (s *BaseOBParserListener) EnterDistribute_method(ctx *Distribute_methodContext) {}

// ExitDistribute_method is called when production distribute_method is exited.
func (s *BaseOBParserListener) ExitDistribute_method(ctx *Distribute_methodContext) {}

// EnterLimit_expr is called when production limit_expr is entered.
func (s *BaseOBParserListener) EnterLimit_expr(ctx *Limit_exprContext) {}

// ExitLimit_expr is called when production limit_expr is exited.
func (s *BaseOBParserListener) ExitLimit_expr(ctx *Limit_exprContext) {}

// EnterOpt_for_update_wait is called when production opt_for_update_wait is entered.
func (s *BaseOBParserListener) EnterOpt_for_update_wait(ctx *Opt_for_update_waitContext) {}

// ExitOpt_for_update_wait is called when production opt_for_update_wait is exited.
func (s *BaseOBParserListener) ExitOpt_for_update_wait(ctx *Opt_for_update_waitContext) {}

// EnterParameterized_trim is called when production parameterized_trim is entered.
func (s *BaseOBParserListener) EnterParameterized_trim(ctx *Parameterized_trimContext) {}

// ExitParameterized_trim is called when production parameterized_trim is exited.
func (s *BaseOBParserListener) ExitParameterized_trim(ctx *Parameterized_trimContext) {}

// EnterGroupby_clause is called when production groupby_clause is entered.
func (s *BaseOBParserListener) EnterGroupby_clause(ctx *Groupby_clauseContext) {}

// ExitGroupby_clause is called when production groupby_clause is exited.
func (s *BaseOBParserListener) ExitGroupby_clause(ctx *Groupby_clauseContext) {}

// EnterSort_list_for_group_by is called when production sort_list_for_group_by is entered.
func (s *BaseOBParserListener) EnterSort_list_for_group_by(ctx *Sort_list_for_group_byContext) {}

// ExitSort_list_for_group_by is called when production sort_list_for_group_by is exited.
func (s *BaseOBParserListener) ExitSort_list_for_group_by(ctx *Sort_list_for_group_byContext) {}

// EnterSort_key_for_group_by is called when production sort_key_for_group_by is entered.
func (s *BaseOBParserListener) EnterSort_key_for_group_by(ctx *Sort_key_for_group_byContext) {}

// ExitSort_key_for_group_by is called when production sort_key_for_group_by is exited.
func (s *BaseOBParserListener) ExitSort_key_for_group_by(ctx *Sort_key_for_group_byContext) {}

// EnterOrder_by is called when production order_by is entered.
func (s *BaseOBParserListener) EnterOrder_by(ctx *Order_byContext) {}

// ExitOrder_by is called when production order_by is exited.
func (s *BaseOBParserListener) ExitOrder_by(ctx *Order_byContext) {}

// EnterSort_list is called when production sort_list is entered.
func (s *BaseOBParserListener) EnterSort_list(ctx *Sort_listContext) {}

// ExitSort_list is called when production sort_list is exited.
func (s *BaseOBParserListener) ExitSort_list(ctx *Sort_listContext) {}

// EnterSort_key is called when production sort_key is entered.
func (s *BaseOBParserListener) EnterSort_key(ctx *Sort_keyContext) {}

// ExitSort_key is called when production sort_key is exited.
func (s *BaseOBParserListener) ExitSort_key(ctx *Sort_keyContext) {}

// EnterQuery_expression_option_list is called when production query_expression_option_list is entered.
func (s *BaseOBParserListener) EnterQuery_expression_option_list(ctx *Query_expression_option_listContext) {}

// ExitQuery_expression_option_list is called when production query_expression_option_list is exited.
func (s *BaseOBParserListener) ExitQuery_expression_option_list(ctx *Query_expression_option_listContext) {}

// EnterQuery_expression_option is called when production query_expression_option is entered.
func (s *BaseOBParserListener) EnterQuery_expression_option(ctx *Query_expression_optionContext) {}

// ExitQuery_expression_option is called when production query_expression_option is exited.
func (s *BaseOBParserListener) ExitQuery_expression_option(ctx *Query_expression_optionContext) {}

// EnterProjection is called when production projection is entered.
func (s *BaseOBParserListener) EnterProjection(ctx *ProjectionContext) {}

// ExitProjection is called when production projection is exited.
func (s *BaseOBParserListener) ExitProjection(ctx *ProjectionContext) {}

// EnterSelect_expr_list is called when production select_expr_list is entered.
func (s *BaseOBParserListener) EnterSelect_expr_list(ctx *Select_expr_listContext) {}

// ExitSelect_expr_list is called when production select_expr_list is exited.
func (s *BaseOBParserListener) ExitSelect_expr_list(ctx *Select_expr_listContext) {}

// EnterFrom_list is called when production from_list is entered.
func (s *BaseOBParserListener) EnterFrom_list(ctx *From_listContext) {}

// ExitFrom_list is called when production from_list is exited.
func (s *BaseOBParserListener) ExitFrom_list(ctx *From_listContext) {}

// EnterTable_references is called when production table_references is entered.
func (s *BaseOBParserListener) EnterTable_references(ctx *Table_referencesContext) {}

// ExitTable_references is called when production table_references is exited.
func (s *BaseOBParserListener) ExitTable_references(ctx *Table_referencesContext) {}

// EnterTable_reference is called when production table_reference is entered.
func (s *BaseOBParserListener) EnterTable_reference(ctx *Table_referenceContext) {}

// ExitTable_reference is called when production table_reference is exited.
func (s *BaseOBParserListener) ExitTable_reference(ctx *Table_referenceContext) {}

// EnterTable_factor is called when production table_factor is entered.
func (s *BaseOBParserListener) EnterTable_factor(ctx *Table_factorContext) {}

// ExitTable_factor is called when production table_factor is exited.
func (s *BaseOBParserListener) ExitTable_factor(ctx *Table_factorContext) {}

// EnterTbl_name is called when production tbl_name is entered.
func (s *BaseOBParserListener) EnterTbl_name(ctx *Tbl_nameContext) {}

// ExitTbl_name is called when production tbl_name is exited.
func (s *BaseOBParserListener) ExitTbl_name(ctx *Tbl_nameContext) {}

// EnterDml_table_name is called when production dml_table_name is entered.
func (s *BaseOBParserListener) EnterDml_table_name(ctx *Dml_table_nameContext) {}

// ExitDml_table_name is called when production dml_table_name is exited.
func (s *BaseOBParserListener) ExitDml_table_name(ctx *Dml_table_nameContext) {}

// EnterSeed is called when production seed is entered.
func (s *BaseOBParserListener) EnterSeed(ctx *SeedContext) {}

// ExitSeed is called when production seed is exited.
func (s *BaseOBParserListener) ExitSeed(ctx *SeedContext) {}

// EnterOpt_seed is called when production opt_seed is entered.
func (s *BaseOBParserListener) EnterOpt_seed(ctx *Opt_seedContext) {}

// ExitOpt_seed is called when production opt_seed is exited.
func (s *BaseOBParserListener) ExitOpt_seed(ctx *Opt_seedContext) {}

// EnterSample_percent is called when production sample_percent is entered.
func (s *BaseOBParserListener) EnterSample_percent(ctx *Sample_percentContext) {}

// ExitSample_percent is called when production sample_percent is exited.
func (s *BaseOBParserListener) ExitSample_percent(ctx *Sample_percentContext) {}

// EnterSample_clause is called when production sample_clause is entered.
func (s *BaseOBParserListener) EnterSample_clause(ctx *Sample_clauseContext) {}

// ExitSample_clause is called when production sample_clause is exited.
func (s *BaseOBParserListener) ExitSample_clause(ctx *Sample_clauseContext) {}

// EnterTable_subquery is called when production table_subquery is entered.
func (s *BaseOBParserListener) EnterTable_subquery(ctx *Table_subqueryContext) {}

// ExitTable_subquery is called when production table_subquery is exited.
func (s *BaseOBParserListener) ExitTable_subquery(ctx *Table_subqueryContext) {}

// EnterUse_partition is called when production use_partition is entered.
func (s *BaseOBParserListener) EnterUse_partition(ctx *Use_partitionContext) {}

// ExitUse_partition is called when production use_partition is exited.
func (s *BaseOBParserListener) ExitUse_partition(ctx *Use_partitionContext) {}

// EnterIndex_hint_type is called when production index_hint_type is entered.
func (s *BaseOBParserListener) EnterIndex_hint_type(ctx *Index_hint_typeContext) {}

// ExitIndex_hint_type is called when production index_hint_type is exited.
func (s *BaseOBParserListener) ExitIndex_hint_type(ctx *Index_hint_typeContext) {}

// EnterKey_or_index is called when production key_or_index is entered.
func (s *BaseOBParserListener) EnterKey_or_index(ctx *Key_or_indexContext) {}

// ExitKey_or_index is called when production key_or_index is exited.
func (s *BaseOBParserListener) ExitKey_or_index(ctx *Key_or_indexContext) {}

// EnterIndex_hint_scope is called when production index_hint_scope is entered.
func (s *BaseOBParserListener) EnterIndex_hint_scope(ctx *Index_hint_scopeContext) {}

// ExitIndex_hint_scope is called when production index_hint_scope is exited.
func (s *BaseOBParserListener) ExitIndex_hint_scope(ctx *Index_hint_scopeContext) {}

// EnterIndex_element is called when production index_element is entered.
func (s *BaseOBParserListener) EnterIndex_element(ctx *Index_elementContext) {}

// ExitIndex_element is called when production index_element is exited.
func (s *BaseOBParserListener) ExitIndex_element(ctx *Index_elementContext) {}

// EnterIndex_list is called when production index_list is entered.
func (s *BaseOBParserListener) EnterIndex_list(ctx *Index_listContext) {}

// ExitIndex_list is called when production index_list is exited.
func (s *BaseOBParserListener) ExitIndex_list(ctx *Index_listContext) {}

// EnterIndex_hint_definition is called when production index_hint_definition is entered.
func (s *BaseOBParserListener) EnterIndex_hint_definition(ctx *Index_hint_definitionContext) {}

// ExitIndex_hint_definition is called when production index_hint_definition is exited.
func (s *BaseOBParserListener) ExitIndex_hint_definition(ctx *Index_hint_definitionContext) {}

// EnterIndex_hint_list is called when production index_hint_list is entered.
func (s *BaseOBParserListener) EnterIndex_hint_list(ctx *Index_hint_listContext) {}

// ExitIndex_hint_list is called when production index_hint_list is exited.
func (s *BaseOBParserListener) ExitIndex_hint_list(ctx *Index_hint_listContext) {}

// EnterRelation_factor is called when production relation_factor is entered.
func (s *BaseOBParserListener) EnterRelation_factor(ctx *Relation_factorContext) {}

// ExitRelation_factor is called when production relation_factor is exited.
func (s *BaseOBParserListener) ExitRelation_factor(ctx *Relation_factorContext) {}

// EnterRelation_with_star_list is called when production relation_with_star_list is entered.
func (s *BaseOBParserListener) EnterRelation_with_star_list(ctx *Relation_with_star_listContext) {}

// ExitRelation_with_star_list is called when production relation_with_star_list is exited.
func (s *BaseOBParserListener) ExitRelation_with_star_list(ctx *Relation_with_star_listContext) {}

// EnterRelation_factor_with_star is called when production relation_factor_with_star is entered.
func (s *BaseOBParserListener) EnterRelation_factor_with_star(ctx *Relation_factor_with_starContext) {}

// ExitRelation_factor_with_star is called when production relation_factor_with_star is exited.
func (s *BaseOBParserListener) ExitRelation_factor_with_star(ctx *Relation_factor_with_starContext) {}

// EnterNormal_relation_factor is called when production normal_relation_factor is entered.
func (s *BaseOBParserListener) EnterNormal_relation_factor(ctx *Normal_relation_factorContext) {}

// ExitNormal_relation_factor is called when production normal_relation_factor is exited.
func (s *BaseOBParserListener) ExitNormal_relation_factor(ctx *Normal_relation_factorContext) {}

// EnterDot_relation_factor is called when production dot_relation_factor is entered.
func (s *BaseOBParserListener) EnterDot_relation_factor(ctx *Dot_relation_factorContext) {}

// ExitDot_relation_factor is called when production dot_relation_factor is exited.
func (s *BaseOBParserListener) ExitDot_relation_factor(ctx *Dot_relation_factorContext) {}

// EnterRelation_factor_in_hint is called when production relation_factor_in_hint is entered.
func (s *BaseOBParserListener) EnterRelation_factor_in_hint(ctx *Relation_factor_in_hintContext) {}

// ExitRelation_factor_in_hint is called when production relation_factor_in_hint is exited.
func (s *BaseOBParserListener) ExitRelation_factor_in_hint(ctx *Relation_factor_in_hintContext) {}

// EnterQb_name_option is called when production qb_name_option is entered.
func (s *BaseOBParserListener) EnterQb_name_option(ctx *Qb_name_optionContext) {}

// ExitQb_name_option is called when production qb_name_option is exited.
func (s *BaseOBParserListener) ExitQb_name_option(ctx *Qb_name_optionContext) {}

// EnterRelation_factor_in_hint_list is called when production relation_factor_in_hint_list is entered.
func (s *BaseOBParserListener) EnterRelation_factor_in_hint_list(ctx *Relation_factor_in_hint_listContext) {}

// ExitRelation_factor_in_hint_list is called when production relation_factor_in_hint_list is exited.
func (s *BaseOBParserListener) ExitRelation_factor_in_hint_list(ctx *Relation_factor_in_hint_listContext) {}

// EnterRelation_sep_option is called when production relation_sep_option is entered.
func (s *BaseOBParserListener) EnterRelation_sep_option(ctx *Relation_sep_optionContext) {}

// ExitRelation_sep_option is called when production relation_sep_option is exited.
func (s *BaseOBParserListener) ExitRelation_sep_option(ctx *Relation_sep_optionContext) {}

// EnterRelation_factor_in_pq_hint is called when production relation_factor_in_pq_hint is entered.
func (s *BaseOBParserListener) EnterRelation_factor_in_pq_hint(ctx *Relation_factor_in_pq_hintContext) {}

// ExitRelation_factor_in_pq_hint is called when production relation_factor_in_pq_hint is exited.
func (s *BaseOBParserListener) ExitRelation_factor_in_pq_hint(ctx *Relation_factor_in_pq_hintContext) {}

// EnterRelation_factor_in_leading_hint is called when production relation_factor_in_leading_hint is entered.
func (s *BaseOBParserListener) EnterRelation_factor_in_leading_hint(ctx *Relation_factor_in_leading_hintContext) {}

// ExitRelation_factor_in_leading_hint is called when production relation_factor_in_leading_hint is exited.
func (s *BaseOBParserListener) ExitRelation_factor_in_leading_hint(ctx *Relation_factor_in_leading_hintContext) {}

// EnterRelation_factor_in_leading_hint_list is called when production relation_factor_in_leading_hint_list is entered.
func (s *BaseOBParserListener) EnterRelation_factor_in_leading_hint_list(ctx *Relation_factor_in_leading_hint_listContext) {}

// ExitRelation_factor_in_leading_hint_list is called when production relation_factor_in_leading_hint_list is exited.
func (s *BaseOBParserListener) ExitRelation_factor_in_leading_hint_list(ctx *Relation_factor_in_leading_hint_listContext) {}

// EnterRelation_factor_in_leading_hint_list_entry is called when production relation_factor_in_leading_hint_list_entry is entered.
func (s *BaseOBParserListener) EnterRelation_factor_in_leading_hint_list_entry(ctx *Relation_factor_in_leading_hint_list_entryContext) {}

// ExitRelation_factor_in_leading_hint_list_entry is called when production relation_factor_in_leading_hint_list_entry is exited.
func (s *BaseOBParserListener) ExitRelation_factor_in_leading_hint_list_entry(ctx *Relation_factor_in_leading_hint_list_entryContext) {}

// EnterRelation_factor_in_use_join_hint_list is called when production relation_factor_in_use_join_hint_list is entered.
func (s *BaseOBParserListener) EnterRelation_factor_in_use_join_hint_list(ctx *Relation_factor_in_use_join_hint_listContext) {}

// ExitRelation_factor_in_use_join_hint_list is called when production relation_factor_in_use_join_hint_list is exited.
func (s *BaseOBParserListener) ExitRelation_factor_in_use_join_hint_list(ctx *Relation_factor_in_use_join_hint_listContext) {}

// EnterTracing_num_list is called when production tracing_num_list is entered.
func (s *BaseOBParserListener) EnterTracing_num_list(ctx *Tracing_num_listContext) {}

// ExitTracing_num_list is called when production tracing_num_list is exited.
func (s *BaseOBParserListener) ExitTracing_num_list(ctx *Tracing_num_listContext) {}

// EnterJoin_condition is called when production join_condition is entered.
func (s *BaseOBParserListener) EnterJoin_condition(ctx *Join_conditionContext) {}

// ExitJoin_condition is called when production join_condition is exited.
func (s *BaseOBParserListener) ExitJoin_condition(ctx *Join_conditionContext) {}

// EnterJoined_table is called when production joined_table is entered.
func (s *BaseOBParserListener) EnterJoined_table(ctx *Joined_tableContext) {}

// ExitJoined_table is called when production joined_table is exited.
func (s *BaseOBParserListener) ExitJoined_table(ctx *Joined_tableContext) {}

// EnterNatural_join_type is called when production natural_join_type is entered.
func (s *BaseOBParserListener) EnterNatural_join_type(ctx *Natural_join_typeContext) {}

// ExitNatural_join_type is called when production natural_join_type is exited.
func (s *BaseOBParserListener) ExitNatural_join_type(ctx *Natural_join_typeContext) {}

// EnterInner_join_type is called when production inner_join_type is entered.
func (s *BaseOBParserListener) EnterInner_join_type(ctx *Inner_join_typeContext) {}

// ExitInner_join_type is called when production inner_join_type is exited.
func (s *BaseOBParserListener) ExitInner_join_type(ctx *Inner_join_typeContext) {}

// EnterOuter_join_type is called when production outer_join_type is entered.
func (s *BaseOBParserListener) EnterOuter_join_type(ctx *Outer_join_typeContext) {}

// ExitOuter_join_type is called when production outer_join_type is exited.
func (s *BaseOBParserListener) ExitOuter_join_type(ctx *Outer_join_typeContext) {}

// EnterAnalyze_stmt is called when production analyze_stmt is entered.
func (s *BaseOBParserListener) EnterAnalyze_stmt(ctx *Analyze_stmtContext) {}

// ExitAnalyze_stmt is called when production analyze_stmt is exited.
func (s *BaseOBParserListener) ExitAnalyze_stmt(ctx *Analyze_stmtContext) {}

// EnterCreate_outline_stmt is called when production create_outline_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_outline_stmt(ctx *Create_outline_stmtContext) {}

// ExitCreate_outline_stmt is called when production create_outline_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_outline_stmt(ctx *Create_outline_stmtContext) {}

// EnterAlter_outline_stmt is called when production alter_outline_stmt is entered.
func (s *BaseOBParserListener) EnterAlter_outline_stmt(ctx *Alter_outline_stmtContext) {}

// ExitAlter_outline_stmt is called when production alter_outline_stmt is exited.
func (s *BaseOBParserListener) ExitAlter_outline_stmt(ctx *Alter_outline_stmtContext) {}

// EnterDrop_outline_stmt is called when production drop_outline_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_outline_stmt(ctx *Drop_outline_stmtContext) {}

// ExitDrop_outline_stmt is called when production drop_outline_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_outline_stmt(ctx *Drop_outline_stmtContext) {}

// EnterExplain_stmt is called when production explain_stmt is entered.
func (s *BaseOBParserListener) EnterExplain_stmt(ctx *Explain_stmtContext) {}

// ExitExplain_stmt is called when production explain_stmt is exited.
func (s *BaseOBParserListener) ExitExplain_stmt(ctx *Explain_stmtContext) {}

// EnterExplain_or_desc is called when production explain_or_desc is entered.
func (s *BaseOBParserListener) EnterExplain_or_desc(ctx *Explain_or_descContext) {}

// ExitExplain_or_desc is called when production explain_or_desc is exited.
func (s *BaseOBParserListener) ExitExplain_or_desc(ctx *Explain_or_descContext) {}

// EnterExplainable_stmt is called when production explainable_stmt is entered.
func (s *BaseOBParserListener) EnterExplainable_stmt(ctx *Explainable_stmtContext) {}

// ExitExplainable_stmt is called when production explainable_stmt is exited.
func (s *BaseOBParserListener) ExitExplainable_stmt(ctx *Explainable_stmtContext) {}

// EnterFormat_name is called when production format_name is entered.
func (s *BaseOBParserListener) EnterFormat_name(ctx *Format_nameContext) {}

// ExitFormat_name is called when production format_name is exited.
func (s *BaseOBParserListener) ExitFormat_name(ctx *Format_nameContext) {}

// EnterShow_stmt is called when production show_stmt is entered.
func (s *BaseOBParserListener) EnterShow_stmt(ctx *Show_stmtContext) {}

// ExitShow_stmt is called when production show_stmt is exited.
func (s *BaseOBParserListener) ExitShow_stmt(ctx *Show_stmtContext) {}

// EnterDatabases_or_schemas is called when production databases_or_schemas is entered.
func (s *BaseOBParserListener) EnterDatabases_or_schemas(ctx *Databases_or_schemasContext) {}

// ExitDatabases_or_schemas is called when production databases_or_schemas is exited.
func (s *BaseOBParserListener) ExitDatabases_or_schemas(ctx *Databases_or_schemasContext) {}

// EnterOpt_for_grant_user is called when production opt_for_grant_user is entered.
func (s *BaseOBParserListener) EnterOpt_for_grant_user(ctx *Opt_for_grant_userContext) {}

// ExitOpt_for_grant_user is called when production opt_for_grant_user is exited.
func (s *BaseOBParserListener) ExitOpt_for_grant_user(ctx *Opt_for_grant_userContext) {}

// EnterColumns_or_fields is called when production columns_or_fields is entered.
func (s *BaseOBParserListener) EnterColumns_or_fields(ctx *Columns_or_fieldsContext) {}

// ExitColumns_or_fields is called when production columns_or_fields is exited.
func (s *BaseOBParserListener) ExitColumns_or_fields(ctx *Columns_or_fieldsContext) {}

// EnterDatabase_or_schema is called when production database_or_schema is entered.
func (s *BaseOBParserListener) EnterDatabase_or_schema(ctx *Database_or_schemaContext) {}

// ExitDatabase_or_schema is called when production database_or_schema is exited.
func (s *BaseOBParserListener) ExitDatabase_or_schema(ctx *Database_or_schemaContext) {}

// EnterIndex_or_indexes_or_keys is called when production index_or_indexes_or_keys is entered.
func (s *BaseOBParserListener) EnterIndex_or_indexes_or_keys(ctx *Index_or_indexes_or_keysContext) {}

// ExitIndex_or_indexes_or_keys is called when production index_or_indexes_or_keys is exited.
func (s *BaseOBParserListener) ExitIndex_or_indexes_or_keys(ctx *Index_or_indexes_or_keysContext) {}

// EnterFrom_or_in is called when production from_or_in is entered.
func (s *BaseOBParserListener) EnterFrom_or_in(ctx *From_or_inContext) {}

// ExitFrom_or_in is called when production from_or_in is exited.
func (s *BaseOBParserListener) ExitFrom_or_in(ctx *From_or_inContext) {}

// EnterHelp_stmt is called when production help_stmt is entered.
func (s *BaseOBParserListener) EnterHelp_stmt(ctx *Help_stmtContext) {}

// ExitHelp_stmt is called when production help_stmt is exited.
func (s *BaseOBParserListener) ExitHelp_stmt(ctx *Help_stmtContext) {}

// EnterCreate_tablespace_stmt is called when production create_tablespace_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_tablespace_stmt(ctx *Create_tablespace_stmtContext) {}

// ExitCreate_tablespace_stmt is called when production create_tablespace_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_tablespace_stmt(ctx *Create_tablespace_stmtContext) {}

// EnterPermanent_tablespace is called when production permanent_tablespace is entered.
func (s *BaseOBParserListener) EnterPermanent_tablespace(ctx *Permanent_tablespaceContext) {}

// ExitPermanent_tablespace is called when production permanent_tablespace is exited.
func (s *BaseOBParserListener) ExitPermanent_tablespace(ctx *Permanent_tablespaceContext) {}

// EnterPermanent_tablespace_option is called when production permanent_tablespace_option is entered.
func (s *BaseOBParserListener) EnterPermanent_tablespace_option(ctx *Permanent_tablespace_optionContext) {}

// ExitPermanent_tablespace_option is called when production permanent_tablespace_option is exited.
func (s *BaseOBParserListener) ExitPermanent_tablespace_option(ctx *Permanent_tablespace_optionContext) {}

// EnterDrop_tablespace_stmt is called when production drop_tablespace_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_tablespace_stmt(ctx *Drop_tablespace_stmtContext) {}

// ExitDrop_tablespace_stmt is called when production drop_tablespace_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_tablespace_stmt(ctx *Drop_tablespace_stmtContext) {}

// EnterAlter_tablespace_actions is called when production alter_tablespace_actions is entered.
func (s *BaseOBParserListener) EnterAlter_tablespace_actions(ctx *Alter_tablespace_actionsContext) {}

// ExitAlter_tablespace_actions is called when production alter_tablespace_actions is exited.
func (s *BaseOBParserListener) ExitAlter_tablespace_actions(ctx *Alter_tablespace_actionsContext) {}

// EnterAlter_tablespace_action is called when production alter_tablespace_action is entered.
func (s *BaseOBParserListener) EnterAlter_tablespace_action(ctx *Alter_tablespace_actionContext) {}

// ExitAlter_tablespace_action is called when production alter_tablespace_action is exited.
func (s *BaseOBParserListener) ExitAlter_tablespace_action(ctx *Alter_tablespace_actionContext) {}

// EnterAlter_tablespace_stmt is called when production alter_tablespace_stmt is entered.
func (s *BaseOBParserListener) EnterAlter_tablespace_stmt(ctx *Alter_tablespace_stmtContext) {}

// ExitAlter_tablespace_stmt is called when production alter_tablespace_stmt is exited.
func (s *BaseOBParserListener) ExitAlter_tablespace_stmt(ctx *Alter_tablespace_stmtContext) {}

// EnterRotate_master_key_stmt is called when production rotate_master_key_stmt is entered.
func (s *BaseOBParserListener) EnterRotate_master_key_stmt(ctx *Rotate_master_key_stmtContext) {}

// ExitRotate_master_key_stmt is called when production rotate_master_key_stmt is exited.
func (s *BaseOBParserListener) ExitRotate_master_key_stmt(ctx *Rotate_master_key_stmtContext) {}

// EnterPermanent_tablespace_options is called when production permanent_tablespace_options is entered.
func (s *BaseOBParserListener) EnterPermanent_tablespace_options(ctx *Permanent_tablespace_optionsContext) {}

// ExitPermanent_tablespace_options is called when production permanent_tablespace_options is exited.
func (s *BaseOBParserListener) ExitPermanent_tablespace_options(ctx *Permanent_tablespace_optionsContext) {}

// EnterCreate_user_stmt is called when production create_user_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_user_stmt(ctx *Create_user_stmtContext) {}

// ExitCreate_user_stmt is called when production create_user_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_user_stmt(ctx *Create_user_stmtContext) {}

// EnterUser_specification_list is called when production user_specification_list is entered.
func (s *BaseOBParserListener) EnterUser_specification_list(ctx *User_specification_listContext) {}

// ExitUser_specification_list is called when production user_specification_list is exited.
func (s *BaseOBParserListener) ExitUser_specification_list(ctx *User_specification_listContext) {}

// EnterUser_specification is called when production user_specification is entered.
func (s *BaseOBParserListener) EnterUser_specification(ctx *User_specificationContext) {}

// ExitUser_specification is called when production user_specification is exited.
func (s *BaseOBParserListener) ExitUser_specification(ctx *User_specificationContext) {}

// EnterRequire_specification is called when production require_specification is entered.
func (s *BaseOBParserListener) EnterRequire_specification(ctx *Require_specificationContext) {}

// ExitRequire_specification is called when production require_specification is exited.
func (s *BaseOBParserListener) ExitRequire_specification(ctx *Require_specificationContext) {}

// EnterTls_option_list is called when production tls_option_list is entered.
func (s *BaseOBParserListener) EnterTls_option_list(ctx *Tls_option_listContext) {}

// ExitTls_option_list is called when production tls_option_list is exited.
func (s *BaseOBParserListener) ExitTls_option_list(ctx *Tls_option_listContext) {}

// EnterTls_option is called when production tls_option is entered.
func (s *BaseOBParserListener) EnterTls_option(ctx *Tls_optionContext) {}

// ExitTls_option is called when production tls_option is exited.
func (s *BaseOBParserListener) ExitTls_option(ctx *Tls_optionContext) {}

// EnterUser is called when production user is entered.
func (s *BaseOBParserListener) EnterUser(ctx *UserContext) {}

// ExitUser is called when production user is exited.
func (s *BaseOBParserListener) ExitUser(ctx *UserContext) {}

// EnterOpt_host_name is called when production opt_host_name is entered.
func (s *BaseOBParserListener) EnterOpt_host_name(ctx *Opt_host_nameContext) {}

// ExitOpt_host_name is called when production opt_host_name is exited.
func (s *BaseOBParserListener) ExitOpt_host_name(ctx *Opt_host_nameContext) {}

// EnterUser_with_host_name is called when production user_with_host_name is entered.
func (s *BaseOBParserListener) EnterUser_with_host_name(ctx *User_with_host_nameContext) {}

// ExitUser_with_host_name is called when production user_with_host_name is exited.
func (s *BaseOBParserListener) ExitUser_with_host_name(ctx *User_with_host_nameContext) {}

// EnterPassword is called when production password is entered.
func (s *BaseOBParserListener) EnterPassword(ctx *PasswordContext) {}

// ExitPassword is called when production password is exited.
func (s *BaseOBParserListener) ExitPassword(ctx *PasswordContext) {}

// EnterDrop_user_stmt is called when production drop_user_stmt is entered.
func (s *BaseOBParserListener) EnterDrop_user_stmt(ctx *Drop_user_stmtContext) {}

// ExitDrop_user_stmt is called when production drop_user_stmt is exited.
func (s *BaseOBParserListener) ExitDrop_user_stmt(ctx *Drop_user_stmtContext) {}

// EnterUser_list is called when production user_list is entered.
func (s *BaseOBParserListener) EnterUser_list(ctx *User_listContext) {}

// ExitUser_list is called when production user_list is exited.
func (s *BaseOBParserListener) ExitUser_list(ctx *User_listContext) {}

// EnterSet_password_stmt is called when production set_password_stmt is entered.
func (s *BaseOBParserListener) EnterSet_password_stmt(ctx *Set_password_stmtContext) {}

// ExitSet_password_stmt is called when production set_password_stmt is exited.
func (s *BaseOBParserListener) ExitSet_password_stmt(ctx *Set_password_stmtContext) {}

// EnterOpt_for_user is called when production opt_for_user is entered.
func (s *BaseOBParserListener) EnterOpt_for_user(ctx *Opt_for_userContext) {}

// ExitOpt_for_user is called when production opt_for_user is exited.
func (s *BaseOBParserListener) ExitOpt_for_user(ctx *Opt_for_userContext) {}

// EnterRename_user_stmt is called when production rename_user_stmt is entered.
func (s *BaseOBParserListener) EnterRename_user_stmt(ctx *Rename_user_stmtContext) {}

// ExitRename_user_stmt is called when production rename_user_stmt is exited.
func (s *BaseOBParserListener) ExitRename_user_stmt(ctx *Rename_user_stmtContext) {}

// EnterRename_info is called when production rename_info is entered.
func (s *BaseOBParserListener) EnterRename_info(ctx *Rename_infoContext) {}

// ExitRename_info is called when production rename_info is exited.
func (s *BaseOBParserListener) ExitRename_info(ctx *Rename_infoContext) {}

// EnterRename_list is called when production rename_list is entered.
func (s *BaseOBParserListener) EnterRename_list(ctx *Rename_listContext) {}

// ExitRename_list is called when production rename_list is exited.
func (s *BaseOBParserListener) ExitRename_list(ctx *Rename_listContext) {}

// EnterLock_user_stmt is called when production lock_user_stmt is entered.
func (s *BaseOBParserListener) EnterLock_user_stmt(ctx *Lock_user_stmtContext) {}

// ExitLock_user_stmt is called when production lock_user_stmt is exited.
func (s *BaseOBParserListener) ExitLock_user_stmt(ctx *Lock_user_stmtContext) {}

// EnterLock_spec_mysql57 is called when production lock_spec_mysql57 is entered.
func (s *BaseOBParserListener) EnterLock_spec_mysql57(ctx *Lock_spec_mysql57Context) {}

// ExitLock_spec_mysql57 is called when production lock_spec_mysql57 is exited.
func (s *BaseOBParserListener) ExitLock_spec_mysql57(ctx *Lock_spec_mysql57Context) {}

// EnterLock_tables_stmt is called when production lock_tables_stmt is entered.
func (s *BaseOBParserListener) EnterLock_tables_stmt(ctx *Lock_tables_stmtContext) {}

// ExitLock_tables_stmt is called when production lock_tables_stmt is exited.
func (s *BaseOBParserListener) ExitLock_tables_stmt(ctx *Lock_tables_stmtContext) {}

// EnterUnlock_tables_stmt is called when production unlock_tables_stmt is entered.
func (s *BaseOBParserListener) EnterUnlock_tables_stmt(ctx *Unlock_tables_stmtContext) {}

// ExitUnlock_tables_stmt is called when production unlock_tables_stmt is exited.
func (s *BaseOBParserListener) ExitUnlock_tables_stmt(ctx *Unlock_tables_stmtContext) {}

// EnterLock_table_list is called when production lock_table_list is entered.
func (s *BaseOBParserListener) EnterLock_table_list(ctx *Lock_table_listContext) {}

// ExitLock_table_list is called when production lock_table_list is exited.
func (s *BaseOBParserListener) ExitLock_table_list(ctx *Lock_table_listContext) {}

// EnterLock_table is called when production lock_table is entered.
func (s *BaseOBParserListener) EnterLock_table(ctx *Lock_tableContext) {}

// ExitLock_table is called when production lock_table is exited.
func (s *BaseOBParserListener) ExitLock_table(ctx *Lock_tableContext) {}

// EnterLock_type is called when production lock_type is entered.
func (s *BaseOBParserListener) EnterLock_type(ctx *Lock_typeContext) {}

// ExitLock_type is called when production lock_type is exited.
func (s *BaseOBParserListener) ExitLock_type(ctx *Lock_typeContext) {}

// EnterBegin_stmt is called when production begin_stmt is entered.
func (s *BaseOBParserListener) EnterBegin_stmt(ctx *Begin_stmtContext) {}

// ExitBegin_stmt is called when production begin_stmt is exited.
func (s *BaseOBParserListener) ExitBegin_stmt(ctx *Begin_stmtContext) {}

// EnterCommit_stmt is called when production commit_stmt is entered.
func (s *BaseOBParserListener) EnterCommit_stmt(ctx *Commit_stmtContext) {}

// ExitCommit_stmt is called when production commit_stmt is exited.
func (s *BaseOBParserListener) ExitCommit_stmt(ctx *Commit_stmtContext) {}

// EnterRollback_stmt is called when production rollback_stmt is entered.
func (s *BaseOBParserListener) EnterRollback_stmt(ctx *Rollback_stmtContext) {}

// ExitRollback_stmt is called when production rollback_stmt is exited.
func (s *BaseOBParserListener) ExitRollback_stmt(ctx *Rollback_stmtContext) {}

// EnterKill_stmt is called when production kill_stmt is entered.
func (s *BaseOBParserListener) EnterKill_stmt(ctx *Kill_stmtContext) {}

// ExitKill_stmt is called when production kill_stmt is exited.
func (s *BaseOBParserListener) ExitKill_stmt(ctx *Kill_stmtContext) {}

// EnterGrant_stmt is called when production grant_stmt is entered.
func (s *BaseOBParserListener) EnterGrant_stmt(ctx *Grant_stmtContext) {}

// ExitGrant_stmt is called when production grant_stmt is exited.
func (s *BaseOBParserListener) ExitGrant_stmt(ctx *Grant_stmtContext) {}

// EnterGrant_privileges is called when production grant_privileges is entered.
func (s *BaseOBParserListener) EnterGrant_privileges(ctx *Grant_privilegesContext) {}

// ExitGrant_privileges is called when production grant_privileges is exited.
func (s *BaseOBParserListener) ExitGrant_privileges(ctx *Grant_privilegesContext) {}

// EnterPriv_type_list is called when production priv_type_list is entered.
func (s *BaseOBParserListener) EnterPriv_type_list(ctx *Priv_type_listContext) {}

// ExitPriv_type_list is called when production priv_type_list is exited.
func (s *BaseOBParserListener) ExitPriv_type_list(ctx *Priv_type_listContext) {}

// EnterPriv_type is called when production priv_type is entered.
func (s *BaseOBParserListener) EnterPriv_type(ctx *Priv_typeContext) {}

// ExitPriv_type is called when production priv_type is exited.
func (s *BaseOBParserListener) ExitPriv_type(ctx *Priv_typeContext) {}

// EnterPriv_level is called when production priv_level is entered.
func (s *BaseOBParserListener) EnterPriv_level(ctx *Priv_levelContext) {}

// ExitPriv_level is called when production priv_level is exited.
func (s *BaseOBParserListener) ExitPriv_level(ctx *Priv_levelContext) {}

// EnterGrant_options is called when production grant_options is entered.
func (s *BaseOBParserListener) EnterGrant_options(ctx *Grant_optionsContext) {}

// ExitGrant_options is called when production grant_options is exited.
func (s *BaseOBParserListener) ExitGrant_options(ctx *Grant_optionsContext) {}

// EnterRevoke_stmt is called when production revoke_stmt is entered.
func (s *BaseOBParserListener) EnterRevoke_stmt(ctx *Revoke_stmtContext) {}

// ExitRevoke_stmt is called when production revoke_stmt is exited.
func (s *BaseOBParserListener) ExitRevoke_stmt(ctx *Revoke_stmtContext) {}

// EnterPrepare_stmt is called when production prepare_stmt is entered.
func (s *BaseOBParserListener) EnterPrepare_stmt(ctx *Prepare_stmtContext) {}

// ExitPrepare_stmt is called when production prepare_stmt is exited.
func (s *BaseOBParserListener) ExitPrepare_stmt(ctx *Prepare_stmtContext) {}

// EnterStmt_name is called when production stmt_name is entered.
func (s *BaseOBParserListener) EnterStmt_name(ctx *Stmt_nameContext) {}

// ExitStmt_name is called when production stmt_name is exited.
func (s *BaseOBParserListener) ExitStmt_name(ctx *Stmt_nameContext) {}

// EnterPreparable_stmt is called when production preparable_stmt is entered.
func (s *BaseOBParserListener) EnterPreparable_stmt(ctx *Preparable_stmtContext) {}

// ExitPreparable_stmt is called when production preparable_stmt is exited.
func (s *BaseOBParserListener) ExitPreparable_stmt(ctx *Preparable_stmtContext) {}

// EnterVariable_set_stmt is called when production variable_set_stmt is entered.
func (s *BaseOBParserListener) EnterVariable_set_stmt(ctx *Variable_set_stmtContext) {}

// ExitVariable_set_stmt is called when production variable_set_stmt is exited.
func (s *BaseOBParserListener) ExitVariable_set_stmt(ctx *Variable_set_stmtContext) {}

// EnterSys_var_and_val_list is called when production sys_var_and_val_list is entered.
func (s *BaseOBParserListener) EnterSys_var_and_val_list(ctx *Sys_var_and_val_listContext) {}

// ExitSys_var_and_val_list is called when production sys_var_and_val_list is exited.
func (s *BaseOBParserListener) ExitSys_var_and_val_list(ctx *Sys_var_and_val_listContext) {}

// EnterVar_and_val_list is called when production var_and_val_list is entered.
func (s *BaseOBParserListener) EnterVar_and_val_list(ctx *Var_and_val_listContext) {}

// ExitVar_and_val_list is called when production var_and_val_list is exited.
func (s *BaseOBParserListener) ExitVar_and_val_list(ctx *Var_and_val_listContext) {}

// EnterSet_expr_or_default is called when production set_expr_or_default is entered.
func (s *BaseOBParserListener) EnterSet_expr_or_default(ctx *Set_expr_or_defaultContext) {}

// ExitSet_expr_or_default is called when production set_expr_or_default is exited.
func (s *BaseOBParserListener) ExitSet_expr_or_default(ctx *Set_expr_or_defaultContext) {}

// EnterVar_and_val is called when production var_and_val is entered.
func (s *BaseOBParserListener) EnterVar_and_val(ctx *Var_and_valContext) {}

// ExitVar_and_val is called when production var_and_val is exited.
func (s *BaseOBParserListener) ExitVar_and_val(ctx *Var_and_valContext) {}

// EnterSys_var_and_val is called when production sys_var_and_val is entered.
func (s *BaseOBParserListener) EnterSys_var_and_val(ctx *Sys_var_and_valContext) {}

// ExitSys_var_and_val is called when production sys_var_and_val is exited.
func (s *BaseOBParserListener) ExitSys_var_and_val(ctx *Sys_var_and_valContext) {}

// EnterScope_or_scope_alias is called when production scope_or_scope_alias is entered.
func (s *BaseOBParserListener) EnterScope_or_scope_alias(ctx *Scope_or_scope_aliasContext) {}

// ExitScope_or_scope_alias is called when production scope_or_scope_alias is exited.
func (s *BaseOBParserListener) ExitScope_or_scope_alias(ctx *Scope_or_scope_aliasContext) {}

// EnterTo_or_eq is called when production to_or_eq is entered.
func (s *BaseOBParserListener) EnterTo_or_eq(ctx *To_or_eqContext) {}

// ExitTo_or_eq is called when production to_or_eq is exited.
func (s *BaseOBParserListener) ExitTo_or_eq(ctx *To_or_eqContext) {}

// EnterExecute_stmt is called when production execute_stmt is entered.
func (s *BaseOBParserListener) EnterExecute_stmt(ctx *Execute_stmtContext) {}

// ExitExecute_stmt is called when production execute_stmt is exited.
func (s *BaseOBParserListener) ExitExecute_stmt(ctx *Execute_stmtContext) {}

// EnterArgument_list is called when production argument_list is entered.
func (s *BaseOBParserListener) EnterArgument_list(ctx *Argument_listContext) {}

// ExitArgument_list is called when production argument_list is exited.
func (s *BaseOBParserListener) ExitArgument_list(ctx *Argument_listContext) {}

// EnterArgument is called when production argument is entered.
func (s *BaseOBParserListener) EnterArgument(ctx *ArgumentContext) {}

// ExitArgument is called when production argument is exited.
func (s *BaseOBParserListener) ExitArgument(ctx *ArgumentContext) {}

// EnterDeallocate_prepare_stmt is called when production deallocate_prepare_stmt is entered.
func (s *BaseOBParserListener) EnterDeallocate_prepare_stmt(ctx *Deallocate_prepare_stmtContext) {}

// ExitDeallocate_prepare_stmt is called when production deallocate_prepare_stmt is exited.
func (s *BaseOBParserListener) ExitDeallocate_prepare_stmt(ctx *Deallocate_prepare_stmtContext) {}

// EnterDeallocate_or_drop is called when production deallocate_or_drop is entered.
func (s *BaseOBParserListener) EnterDeallocate_or_drop(ctx *Deallocate_or_dropContext) {}

// ExitDeallocate_or_drop is called when production deallocate_or_drop is exited.
func (s *BaseOBParserListener) ExitDeallocate_or_drop(ctx *Deallocate_or_dropContext) {}

// EnterTruncate_table_stmt is called when production truncate_table_stmt is entered.
func (s *BaseOBParserListener) EnterTruncate_table_stmt(ctx *Truncate_table_stmtContext) {}

// ExitTruncate_table_stmt is called when production truncate_table_stmt is exited.
func (s *BaseOBParserListener) ExitTruncate_table_stmt(ctx *Truncate_table_stmtContext) {}

// EnterRename_table_stmt is called when production rename_table_stmt is entered.
func (s *BaseOBParserListener) EnterRename_table_stmt(ctx *Rename_table_stmtContext) {}

// ExitRename_table_stmt is called when production rename_table_stmt is exited.
func (s *BaseOBParserListener) ExitRename_table_stmt(ctx *Rename_table_stmtContext) {}

// EnterRename_table_actions is called when production rename_table_actions is entered.
func (s *BaseOBParserListener) EnterRename_table_actions(ctx *Rename_table_actionsContext) {}

// ExitRename_table_actions is called when production rename_table_actions is exited.
func (s *BaseOBParserListener) ExitRename_table_actions(ctx *Rename_table_actionsContext) {}

// EnterRename_table_action is called when production rename_table_action is entered.
func (s *BaseOBParserListener) EnterRename_table_action(ctx *Rename_table_actionContext) {}

// ExitRename_table_action is called when production rename_table_action is exited.
func (s *BaseOBParserListener) ExitRename_table_action(ctx *Rename_table_actionContext) {}

// EnterAlter_table_stmt is called when production alter_table_stmt is entered.
func (s *BaseOBParserListener) EnterAlter_table_stmt(ctx *Alter_table_stmtContext) {}

// ExitAlter_table_stmt is called when production alter_table_stmt is exited.
func (s *BaseOBParserListener) ExitAlter_table_stmt(ctx *Alter_table_stmtContext) {}

// EnterAlter_table_actions is called when production alter_table_actions is entered.
func (s *BaseOBParserListener) EnterAlter_table_actions(ctx *Alter_table_actionsContext) {}

// ExitAlter_table_actions is called when production alter_table_actions is exited.
func (s *BaseOBParserListener) ExitAlter_table_actions(ctx *Alter_table_actionsContext) {}

// EnterAlter_table_action is called when production alter_table_action is entered.
func (s *BaseOBParserListener) EnterAlter_table_action(ctx *Alter_table_actionContext) {}

// ExitAlter_table_action is called when production alter_table_action is exited.
func (s *BaseOBParserListener) ExitAlter_table_action(ctx *Alter_table_actionContext) {}

// EnterAlter_constraint_option is called when production alter_constraint_option is entered.
func (s *BaseOBParserListener) EnterAlter_constraint_option(ctx *Alter_constraint_optionContext) {}

// ExitAlter_constraint_option is called when production alter_constraint_option is exited.
func (s *BaseOBParserListener) ExitAlter_constraint_option(ctx *Alter_constraint_optionContext) {}

// EnterAlter_partition_option is called when production alter_partition_option is entered.
func (s *BaseOBParserListener) EnterAlter_partition_option(ctx *Alter_partition_optionContext) {}

// ExitAlter_partition_option is called when production alter_partition_option is exited.
func (s *BaseOBParserListener) ExitAlter_partition_option(ctx *Alter_partition_optionContext) {}

// EnterOpt_partition_range_or_list is called when production opt_partition_range_or_list is entered.
func (s *BaseOBParserListener) EnterOpt_partition_range_or_list(ctx *Opt_partition_range_or_listContext) {}

// ExitOpt_partition_range_or_list is called when production opt_partition_range_or_list is exited.
func (s *BaseOBParserListener) ExitOpt_partition_range_or_list(ctx *Opt_partition_range_or_listContext) {}

// EnterAlter_tg_partition_option is called when production alter_tg_partition_option is entered.
func (s *BaseOBParserListener) EnterAlter_tg_partition_option(ctx *Alter_tg_partition_optionContext) {}

// ExitAlter_tg_partition_option is called when production alter_tg_partition_option is exited.
func (s *BaseOBParserListener) ExitAlter_tg_partition_option(ctx *Alter_tg_partition_optionContext) {}

// EnterDrop_partition_name_list is called when production drop_partition_name_list is entered.
func (s *BaseOBParserListener) EnterDrop_partition_name_list(ctx *Drop_partition_name_listContext) {}

// ExitDrop_partition_name_list is called when production drop_partition_name_list is exited.
func (s *BaseOBParserListener) ExitDrop_partition_name_list(ctx *Drop_partition_name_listContext) {}

// EnterModify_partition_info is called when production modify_partition_info is entered.
func (s *BaseOBParserListener) EnterModify_partition_info(ctx *Modify_partition_infoContext) {}

// ExitModify_partition_info is called when production modify_partition_info is exited.
func (s *BaseOBParserListener) ExitModify_partition_info(ctx *Modify_partition_infoContext) {}

// EnterModify_tg_partition_info is called when production modify_tg_partition_info is entered.
func (s *BaseOBParserListener) EnterModify_tg_partition_info(ctx *Modify_tg_partition_infoContext) {}

// ExitModify_tg_partition_info is called when production modify_tg_partition_info is exited.
func (s *BaseOBParserListener) ExitModify_tg_partition_info(ctx *Modify_tg_partition_infoContext) {}

// EnterAlter_index_option is called when production alter_index_option is entered.
func (s *BaseOBParserListener) EnterAlter_index_option(ctx *Alter_index_optionContext) {}

// ExitAlter_index_option is called when production alter_index_option is exited.
func (s *BaseOBParserListener) ExitAlter_index_option(ctx *Alter_index_optionContext) {}

// EnterAlter_foreign_key_action is called when production alter_foreign_key_action is entered.
func (s *BaseOBParserListener) EnterAlter_foreign_key_action(ctx *Alter_foreign_key_actionContext) {}

// ExitAlter_foreign_key_action is called when production alter_foreign_key_action is exited.
func (s *BaseOBParserListener) ExitAlter_foreign_key_action(ctx *Alter_foreign_key_actionContext) {}

// EnterVisibility_option is called when production visibility_option is entered.
func (s *BaseOBParserListener) EnterVisibility_option(ctx *Visibility_optionContext) {}

// ExitVisibility_option is called when production visibility_option is exited.
func (s *BaseOBParserListener) ExitVisibility_option(ctx *Visibility_optionContext) {}

// EnterAlter_column_option is called when production alter_column_option is entered.
func (s *BaseOBParserListener) EnterAlter_column_option(ctx *Alter_column_optionContext) {}

// ExitAlter_column_option is called when production alter_column_option is exited.
func (s *BaseOBParserListener) ExitAlter_column_option(ctx *Alter_column_optionContext) {}

// EnterAlter_tablegroup_option is called when production alter_tablegroup_option is entered.
func (s *BaseOBParserListener) EnterAlter_tablegroup_option(ctx *Alter_tablegroup_optionContext) {}

// ExitAlter_tablegroup_option is called when production alter_tablegroup_option is exited.
func (s *BaseOBParserListener) ExitAlter_tablegroup_option(ctx *Alter_tablegroup_optionContext) {}

// EnterAlter_column_behavior is called when production alter_column_behavior is entered.
func (s *BaseOBParserListener) EnterAlter_column_behavior(ctx *Alter_column_behaviorContext) {}

// ExitAlter_column_behavior is called when production alter_column_behavior is exited.
func (s *BaseOBParserListener) ExitAlter_column_behavior(ctx *Alter_column_behaviorContext) {}

// EnterFlashback_stmt is called when production flashback_stmt is entered.
func (s *BaseOBParserListener) EnterFlashback_stmt(ctx *Flashback_stmtContext) {}

// ExitFlashback_stmt is called when production flashback_stmt is exited.
func (s *BaseOBParserListener) ExitFlashback_stmt(ctx *Flashback_stmtContext) {}

// EnterPurge_stmt is called when production purge_stmt is entered.
func (s *BaseOBParserListener) EnterPurge_stmt(ctx *Purge_stmtContext) {}

// ExitPurge_stmt is called when production purge_stmt is exited.
func (s *BaseOBParserListener) ExitPurge_stmt(ctx *Purge_stmtContext) {}

// EnterOptimize_stmt is called when production optimize_stmt is entered.
func (s *BaseOBParserListener) EnterOptimize_stmt(ctx *Optimize_stmtContext) {}

// ExitOptimize_stmt is called when production optimize_stmt is exited.
func (s *BaseOBParserListener) ExitOptimize_stmt(ctx *Optimize_stmtContext) {}

// EnterDump_memory_stmt is called when production dump_memory_stmt is entered.
func (s *BaseOBParserListener) EnterDump_memory_stmt(ctx *Dump_memory_stmtContext) {}

// ExitDump_memory_stmt is called when production dump_memory_stmt is exited.
func (s *BaseOBParserListener) ExitDump_memory_stmt(ctx *Dump_memory_stmtContext) {}

// EnterAlter_system_stmt is called when production alter_system_stmt is entered.
func (s *BaseOBParserListener) EnterAlter_system_stmt(ctx *Alter_system_stmtContext) {}

// ExitAlter_system_stmt is called when production alter_system_stmt is exited.
func (s *BaseOBParserListener) ExitAlter_system_stmt(ctx *Alter_system_stmtContext) {}

// EnterChange_tenant_name_or_tenant_id is called when production change_tenant_name_or_tenant_id is entered.
func (s *BaseOBParserListener) EnterChange_tenant_name_or_tenant_id(ctx *Change_tenant_name_or_tenant_idContext) {}

// ExitChange_tenant_name_or_tenant_id is called when production change_tenant_name_or_tenant_id is exited.
func (s *BaseOBParserListener) ExitChange_tenant_name_or_tenant_id(ctx *Change_tenant_name_or_tenant_idContext) {}

// EnterCache_type is called when production cache_type is entered.
func (s *BaseOBParserListener) EnterCache_type(ctx *Cache_typeContext) {}

// ExitCache_type is called when production cache_type is exited.
func (s *BaseOBParserListener) ExitCache_type(ctx *Cache_typeContext) {}

// EnterBalance_task_type is called when production balance_task_type is entered.
func (s *BaseOBParserListener) EnterBalance_task_type(ctx *Balance_task_typeContext) {}

// ExitBalance_task_type is called when production balance_task_type is exited.
func (s *BaseOBParserListener) ExitBalance_task_type(ctx *Balance_task_typeContext) {}

// EnterTenant_list_tuple is called when production tenant_list_tuple is entered.
func (s *BaseOBParserListener) EnterTenant_list_tuple(ctx *Tenant_list_tupleContext) {}

// ExitTenant_list_tuple is called when production tenant_list_tuple is exited.
func (s *BaseOBParserListener) ExitTenant_list_tuple(ctx *Tenant_list_tupleContext) {}

// EnterTenant_name_list is called when production tenant_name_list is entered.
func (s *BaseOBParserListener) EnterTenant_name_list(ctx *Tenant_name_listContext) {}

// ExitTenant_name_list is called when production tenant_name_list is exited.
func (s *BaseOBParserListener) ExitTenant_name_list(ctx *Tenant_name_listContext) {}

// EnterFlush_scope is called when production flush_scope is entered.
func (s *BaseOBParserListener) EnterFlush_scope(ctx *Flush_scopeContext) {}

// ExitFlush_scope is called when production flush_scope is exited.
func (s *BaseOBParserListener) ExitFlush_scope(ctx *Flush_scopeContext) {}

// EnterServer_info_list is called when production server_info_list is entered.
func (s *BaseOBParserListener) EnterServer_info_list(ctx *Server_info_listContext) {}

// ExitServer_info_list is called when production server_info_list is exited.
func (s *BaseOBParserListener) ExitServer_info_list(ctx *Server_info_listContext) {}

// EnterServer_info is called when production server_info is entered.
func (s *BaseOBParserListener) EnterServer_info(ctx *Server_infoContext) {}

// ExitServer_info is called when production server_info is exited.
func (s *BaseOBParserListener) ExitServer_info(ctx *Server_infoContext) {}

// EnterServer_action is called when production server_action is entered.
func (s *BaseOBParserListener) EnterServer_action(ctx *Server_actionContext) {}

// ExitServer_action is called when production server_action is exited.
func (s *BaseOBParserListener) ExitServer_action(ctx *Server_actionContext) {}

// EnterServer_list is called when production server_list is entered.
func (s *BaseOBParserListener) EnterServer_list(ctx *Server_listContext) {}

// ExitServer_list is called when production server_list is exited.
func (s *BaseOBParserListener) ExitServer_list(ctx *Server_listContext) {}

// EnterZone_action is called when production zone_action is entered.
func (s *BaseOBParserListener) EnterZone_action(ctx *Zone_actionContext) {}

// ExitZone_action is called when production zone_action is exited.
func (s *BaseOBParserListener) ExitZone_action(ctx *Zone_actionContext) {}

// EnterIp_port is called when production ip_port is entered.
func (s *BaseOBParserListener) EnterIp_port(ctx *Ip_portContext) {}

// ExitIp_port is called when production ip_port is exited.
func (s *BaseOBParserListener) ExitIp_port(ctx *Ip_portContext) {}

// EnterZone_desc is called when production zone_desc is entered.
func (s *BaseOBParserListener) EnterZone_desc(ctx *Zone_descContext) {}

// ExitZone_desc is called when production zone_desc is exited.
func (s *BaseOBParserListener) ExitZone_desc(ctx *Zone_descContext) {}

// EnterServer_or_zone is called when production server_or_zone is entered.
func (s *BaseOBParserListener) EnterServer_or_zone(ctx *Server_or_zoneContext) {}

// ExitServer_or_zone is called when production server_or_zone is exited.
func (s *BaseOBParserListener) ExitServer_or_zone(ctx *Server_or_zoneContext) {}

// EnterAdd_or_alter_zone_option is called when production add_or_alter_zone_option is entered.
func (s *BaseOBParserListener) EnterAdd_or_alter_zone_option(ctx *Add_or_alter_zone_optionContext) {}

// ExitAdd_or_alter_zone_option is called when production add_or_alter_zone_option is exited.
func (s *BaseOBParserListener) ExitAdd_or_alter_zone_option(ctx *Add_or_alter_zone_optionContext) {}

// EnterAdd_or_alter_zone_options is called when production add_or_alter_zone_options is entered.
func (s *BaseOBParserListener) EnterAdd_or_alter_zone_options(ctx *Add_or_alter_zone_optionsContext) {}

// ExitAdd_or_alter_zone_options is called when production add_or_alter_zone_options is exited.
func (s *BaseOBParserListener) ExitAdd_or_alter_zone_options(ctx *Add_or_alter_zone_optionsContext) {}

// EnterAlter_or_change_or_modify is called when production alter_or_change_or_modify is entered.
func (s *BaseOBParserListener) EnterAlter_or_change_or_modify(ctx *Alter_or_change_or_modifyContext) {}

// ExitAlter_or_change_or_modify is called when production alter_or_change_or_modify is exited.
func (s *BaseOBParserListener) ExitAlter_or_change_or_modify(ctx *Alter_or_change_or_modifyContext) {}

// EnterPartition_id_desc is called when production partition_id_desc is entered.
func (s *BaseOBParserListener) EnterPartition_id_desc(ctx *Partition_id_descContext) {}

// ExitPartition_id_desc is called when production partition_id_desc is exited.
func (s *BaseOBParserListener) ExitPartition_id_desc(ctx *Partition_id_descContext) {}

// EnterPartition_id_or_server_or_zone is called when production partition_id_or_server_or_zone is entered.
func (s *BaseOBParserListener) EnterPartition_id_or_server_or_zone(ctx *Partition_id_or_server_or_zoneContext) {}

// ExitPartition_id_or_server_or_zone is called when production partition_id_or_server_or_zone is exited.
func (s *BaseOBParserListener) ExitPartition_id_or_server_or_zone(ctx *Partition_id_or_server_or_zoneContext) {}

// EnterMigrate_action is called when production migrate_action is entered.
func (s *BaseOBParserListener) EnterMigrate_action(ctx *Migrate_actionContext) {}

// ExitMigrate_action is called when production migrate_action is exited.
func (s *BaseOBParserListener) ExitMigrate_action(ctx *Migrate_actionContext) {}

// EnterChange_actions is called when production change_actions is entered.
func (s *BaseOBParserListener) EnterChange_actions(ctx *Change_actionsContext) {}

// ExitChange_actions is called when production change_actions is exited.
func (s *BaseOBParserListener) ExitChange_actions(ctx *Change_actionsContext) {}

// EnterChange_action is called when production change_action is entered.
func (s *BaseOBParserListener) EnterChange_action(ctx *Change_actionContext) {}

// ExitChange_action is called when production change_action is exited.
func (s *BaseOBParserListener) ExitChange_action(ctx *Change_actionContext) {}

// EnterReplica_type is called when production replica_type is entered.
func (s *BaseOBParserListener) EnterReplica_type(ctx *Replica_typeContext) {}

// ExitReplica_type is called when production replica_type is exited.
func (s *BaseOBParserListener) ExitReplica_type(ctx *Replica_typeContext) {}

// EnterSuspend_or_resume is called when production suspend_or_resume is entered.
func (s *BaseOBParserListener) EnterSuspend_or_resume(ctx *Suspend_or_resumeContext) {}

// ExitSuspend_or_resume is called when production suspend_or_resume is exited.
func (s *BaseOBParserListener) ExitSuspend_or_resume(ctx *Suspend_or_resumeContext) {}

// EnterBaseline_id_expr is called when production baseline_id_expr is entered.
func (s *BaseOBParserListener) EnterBaseline_id_expr(ctx *Baseline_id_exprContext) {}

// ExitBaseline_id_expr is called when production baseline_id_expr is exited.
func (s *BaseOBParserListener) ExitBaseline_id_expr(ctx *Baseline_id_exprContext) {}

// EnterSql_id_expr is called when production sql_id_expr is entered.
func (s *BaseOBParserListener) EnterSql_id_expr(ctx *Sql_id_exprContext) {}

// ExitSql_id_expr is called when production sql_id_expr is exited.
func (s *BaseOBParserListener) ExitSql_id_expr(ctx *Sql_id_exprContext) {}

// EnterBaseline_asgn_factor is called when production baseline_asgn_factor is entered.
func (s *BaseOBParserListener) EnterBaseline_asgn_factor(ctx *Baseline_asgn_factorContext) {}

// ExitBaseline_asgn_factor is called when production baseline_asgn_factor is exited.
func (s *BaseOBParserListener) ExitBaseline_asgn_factor(ctx *Baseline_asgn_factorContext) {}

// EnterTenant_name is called when production tenant_name is entered.
func (s *BaseOBParserListener) EnterTenant_name(ctx *Tenant_nameContext) {}

// ExitTenant_name is called when production tenant_name is exited.
func (s *BaseOBParserListener) ExitTenant_name(ctx *Tenant_nameContext) {}

// EnterCache_name is called when production cache_name is entered.
func (s *BaseOBParserListener) EnterCache_name(ctx *Cache_nameContext) {}

// ExitCache_name is called when production cache_name is exited.
func (s *BaseOBParserListener) ExitCache_name(ctx *Cache_nameContext) {}

// EnterFile_id is called when production file_id is entered.
func (s *BaseOBParserListener) EnterFile_id(ctx *File_idContext) {}

// ExitFile_id is called when production file_id is exited.
func (s *BaseOBParserListener) ExitFile_id(ctx *File_idContext) {}

// EnterCancel_task_type is called when production cancel_task_type is entered.
func (s *BaseOBParserListener) EnterCancel_task_type(ctx *Cancel_task_typeContext) {}

// ExitCancel_task_type is called when production cancel_task_type is exited.
func (s *BaseOBParserListener) ExitCancel_task_type(ctx *Cancel_task_typeContext) {}

// EnterAlter_system_set_parameter_actions is called when production alter_system_set_parameter_actions is entered.
func (s *BaseOBParserListener) EnterAlter_system_set_parameter_actions(ctx *Alter_system_set_parameter_actionsContext) {}

// ExitAlter_system_set_parameter_actions is called when production alter_system_set_parameter_actions is exited.
func (s *BaseOBParserListener) ExitAlter_system_set_parameter_actions(ctx *Alter_system_set_parameter_actionsContext) {}

// EnterAlter_system_set_parameter_action is called when production alter_system_set_parameter_action is entered.
func (s *BaseOBParserListener) EnterAlter_system_set_parameter_action(ctx *Alter_system_set_parameter_actionContext) {}

// ExitAlter_system_set_parameter_action is called when production alter_system_set_parameter_action is exited.
func (s *BaseOBParserListener) ExitAlter_system_set_parameter_action(ctx *Alter_system_set_parameter_actionContext) {}

// EnterAlter_system_settp_actions is called when production alter_system_settp_actions is entered.
func (s *BaseOBParserListener) EnterAlter_system_settp_actions(ctx *Alter_system_settp_actionsContext) {}

// ExitAlter_system_settp_actions is called when production alter_system_settp_actions is exited.
func (s *BaseOBParserListener) ExitAlter_system_settp_actions(ctx *Alter_system_settp_actionsContext) {}

// EnterSettp_option is called when production settp_option is entered.
func (s *BaseOBParserListener) EnterSettp_option(ctx *Settp_optionContext) {}

// ExitSettp_option is called when production settp_option is exited.
func (s *BaseOBParserListener) ExitSettp_option(ctx *Settp_optionContext) {}

// EnterCluster_role is called when production cluster_role is entered.
func (s *BaseOBParserListener) EnterCluster_role(ctx *Cluster_roleContext) {}

// ExitCluster_role is called when production cluster_role is exited.
func (s *BaseOBParserListener) ExitCluster_role(ctx *Cluster_roleContext) {}

// EnterPartition_role is called when production partition_role is entered.
func (s *BaseOBParserListener) EnterPartition_role(ctx *Partition_roleContext) {}

// ExitPartition_role is called when production partition_role is exited.
func (s *BaseOBParserListener) ExitPartition_role(ctx *Partition_roleContext) {}

// EnterUpgrade_action is called when production upgrade_action is entered.
func (s *BaseOBParserListener) EnterUpgrade_action(ctx *Upgrade_actionContext) {}

// ExitUpgrade_action is called when production upgrade_action is exited.
func (s *BaseOBParserListener) ExitUpgrade_action(ctx *Upgrade_actionContext) {}

// EnterSet_names_stmt is called when production set_names_stmt is entered.
func (s *BaseOBParserListener) EnterSet_names_stmt(ctx *Set_names_stmtContext) {}

// ExitSet_names_stmt is called when production set_names_stmt is exited.
func (s *BaseOBParserListener) ExitSet_names_stmt(ctx *Set_names_stmtContext) {}

// EnterSet_charset_stmt is called when production set_charset_stmt is entered.
func (s *BaseOBParserListener) EnterSet_charset_stmt(ctx *Set_charset_stmtContext) {}

// ExitSet_charset_stmt is called when production set_charset_stmt is exited.
func (s *BaseOBParserListener) ExitSet_charset_stmt(ctx *Set_charset_stmtContext) {}

// EnterSet_transaction_stmt is called when production set_transaction_stmt is entered.
func (s *BaseOBParserListener) EnterSet_transaction_stmt(ctx *Set_transaction_stmtContext) {}

// ExitSet_transaction_stmt is called when production set_transaction_stmt is exited.
func (s *BaseOBParserListener) ExitSet_transaction_stmt(ctx *Set_transaction_stmtContext) {}

// EnterTransaction_characteristics is called when production transaction_characteristics is entered.
func (s *BaseOBParserListener) EnterTransaction_characteristics(ctx *Transaction_characteristicsContext) {}

// ExitTransaction_characteristics is called when production transaction_characteristics is exited.
func (s *BaseOBParserListener) ExitTransaction_characteristics(ctx *Transaction_characteristicsContext) {}

// EnterTransaction_access_mode is called when production transaction_access_mode is entered.
func (s *BaseOBParserListener) EnterTransaction_access_mode(ctx *Transaction_access_modeContext) {}

// ExitTransaction_access_mode is called when production transaction_access_mode is exited.
func (s *BaseOBParserListener) ExitTransaction_access_mode(ctx *Transaction_access_modeContext) {}

// EnterIsolation_level is called when production isolation_level is entered.
func (s *BaseOBParserListener) EnterIsolation_level(ctx *Isolation_levelContext) {}

// ExitIsolation_level is called when production isolation_level is exited.
func (s *BaseOBParserListener) ExitIsolation_level(ctx *Isolation_levelContext) {}

// EnterCreate_savepoint_stmt is called when production create_savepoint_stmt is entered.
func (s *BaseOBParserListener) EnterCreate_savepoint_stmt(ctx *Create_savepoint_stmtContext) {}

// ExitCreate_savepoint_stmt is called when production create_savepoint_stmt is exited.
func (s *BaseOBParserListener) ExitCreate_savepoint_stmt(ctx *Create_savepoint_stmtContext) {}

// EnterRollback_savepoint_stmt is called when production rollback_savepoint_stmt is entered.
func (s *BaseOBParserListener) EnterRollback_savepoint_stmt(ctx *Rollback_savepoint_stmtContext) {}

// ExitRollback_savepoint_stmt is called when production rollback_savepoint_stmt is exited.
func (s *BaseOBParserListener) ExitRollback_savepoint_stmt(ctx *Rollback_savepoint_stmtContext) {}

// EnterRelease_savepoint_stmt is called when production release_savepoint_stmt is entered.
func (s *BaseOBParserListener) EnterRelease_savepoint_stmt(ctx *Release_savepoint_stmtContext) {}

// ExitRelease_savepoint_stmt is called when production release_savepoint_stmt is exited.
func (s *BaseOBParserListener) ExitRelease_savepoint_stmt(ctx *Release_savepoint_stmtContext) {}

// EnterAlter_cluster_stmt is called when production alter_cluster_stmt is entered.
func (s *BaseOBParserListener) EnterAlter_cluster_stmt(ctx *Alter_cluster_stmtContext) {}

// ExitAlter_cluster_stmt is called when production alter_cluster_stmt is exited.
func (s *BaseOBParserListener) ExitAlter_cluster_stmt(ctx *Alter_cluster_stmtContext) {}

// EnterCluster_action is called when production cluster_action is entered.
func (s *BaseOBParserListener) EnterCluster_action(ctx *Cluster_actionContext) {}

// ExitCluster_action is called when production cluster_action is exited.
func (s *BaseOBParserListener) ExitCluster_action(ctx *Cluster_actionContext) {}

// EnterSwitchover_cluster_stmt is called when production switchover_cluster_stmt is entered.
func (s *BaseOBParserListener) EnterSwitchover_cluster_stmt(ctx *Switchover_cluster_stmtContext) {}

// ExitSwitchover_cluster_stmt is called when production switchover_cluster_stmt is exited.
func (s *BaseOBParserListener) ExitSwitchover_cluster_stmt(ctx *Switchover_cluster_stmtContext) {}

// EnterCommit_switchover_clause is called when production commit_switchover_clause is entered.
func (s *BaseOBParserListener) EnterCommit_switchover_clause(ctx *Commit_switchover_clauseContext) {}

// ExitCommit_switchover_clause is called when production commit_switchover_clause is exited.
func (s *BaseOBParserListener) ExitCommit_switchover_clause(ctx *Commit_switchover_clauseContext) {}

// EnterCluster_name is called when production cluster_name is entered.
func (s *BaseOBParserListener) EnterCluster_name(ctx *Cluster_nameContext) {}

// ExitCluster_name is called when production cluster_name is exited.
func (s *BaseOBParserListener) ExitCluster_name(ctx *Cluster_nameContext) {}

// EnterVar_name is called when production var_name is entered.
func (s *BaseOBParserListener) EnterVar_name(ctx *Var_nameContext) {}

// ExitVar_name is called when production var_name is exited.
func (s *BaseOBParserListener) ExitVar_name(ctx *Var_nameContext) {}

// EnterColumn_name is called when production column_name is entered.
func (s *BaseOBParserListener) EnterColumn_name(ctx *Column_nameContext) {}

// ExitColumn_name is called when production column_name is exited.
func (s *BaseOBParserListener) ExitColumn_name(ctx *Column_nameContext) {}

// EnterRelation_name is called when production relation_name is entered.
func (s *BaseOBParserListener) EnterRelation_name(ctx *Relation_nameContext) {}

// ExitRelation_name is called when production relation_name is exited.
func (s *BaseOBParserListener) ExitRelation_name(ctx *Relation_nameContext) {}

// EnterFunction_name is called when production function_name is entered.
func (s *BaseOBParserListener) EnterFunction_name(ctx *Function_nameContext) {}

// ExitFunction_name is called when production function_name is exited.
func (s *BaseOBParserListener) ExitFunction_name(ctx *Function_nameContext) {}

// EnterColumn_label is called when production column_label is entered.
func (s *BaseOBParserListener) EnterColumn_label(ctx *Column_labelContext) {}

// ExitColumn_label is called when production column_label is exited.
func (s *BaseOBParserListener) ExitColumn_label(ctx *Column_labelContext) {}

// EnterDate_unit is called when production date_unit is entered.
func (s *BaseOBParserListener) EnterDate_unit(ctx *Date_unitContext) {}

// ExitDate_unit is called when production date_unit is exited.
func (s *BaseOBParserListener) ExitDate_unit(ctx *Date_unitContext) {}

// EnterUnreserved_keyword is called when production unreserved_keyword is entered.
func (s *BaseOBParserListener) EnterUnreserved_keyword(ctx *Unreserved_keywordContext) {}

// ExitUnreserved_keyword is called when production unreserved_keyword is exited.
func (s *BaseOBParserListener) ExitUnreserved_keyword(ctx *Unreserved_keywordContext) {}

// EnterUnreserved_keyword_normal is called when production unreserved_keyword_normal is entered.
func (s *BaseOBParserListener) EnterUnreserved_keyword_normal(ctx *Unreserved_keyword_normalContext) {}

// ExitUnreserved_keyword_normal is called when production unreserved_keyword_normal is exited.
func (s *BaseOBParserListener) ExitUnreserved_keyword_normal(ctx *Unreserved_keyword_normalContext) {}

// EnterUnreserved_keyword_special is called when production unreserved_keyword_special is entered.
func (s *BaseOBParserListener) EnterUnreserved_keyword_special(ctx *Unreserved_keyword_specialContext) {}

// ExitUnreserved_keyword_special is called when production unreserved_keyword_special is exited.
func (s *BaseOBParserListener) ExitUnreserved_keyword_special(ctx *Unreserved_keyword_specialContext) {}

// EnterEmpty is called when production empty is entered.
func (s *BaseOBParserListener) EnterEmpty(ctx *EmptyContext) {}

// ExitEmpty is called when production empty is exited.
func (s *BaseOBParserListener) ExitEmpty(ctx *EmptyContext) {}

// EnterForward_expr is called when production forward_expr is entered.
func (s *BaseOBParserListener) EnterForward_expr(ctx *Forward_exprContext) {}

// ExitForward_expr is called when production forward_expr is exited.
func (s *BaseOBParserListener) ExitForward_expr(ctx *Forward_exprContext) {}

// EnterForward_sql_stmt is called when production forward_sql_stmt is entered.
func (s *BaseOBParserListener) EnterForward_sql_stmt(ctx *Forward_sql_stmtContext) {}

// ExitForward_sql_stmt is called when production forward_sql_stmt is exited.
func (s *BaseOBParserListener) ExitForward_sql_stmt(ctx *Forward_sql_stmtContext) {}
