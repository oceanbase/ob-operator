##@ Parser Generation

ANTLR_BUILDER_IMAGE = oceanbase/antlr:latest

.PHONY: generate-parser
generate-parser: ## Generate Go parser code from ANTLR4 grammar files using Docker
	@echo "Generating Parser..."
	@mkdir -p internal/sql-analyzer/parser/mysql
	@docker run --rm -u $$(id -u):$$(id -g) -v $$(pwd):/work $(ANTLR_BUILDER_IMAGE) -Dlanguage=Go -o /work/internal/sql-analyzer/parser/mysql -visitor -package mysql /work/obparser/obmysql/sql/OBLexer.g4 /work/obparser/obmysql/sql/OBParser.g4
	@python3 hack/fix_parser.py internal/sql-analyzer/parser/mysql/ob_parser.go
	@echo "Parser generation complete."
