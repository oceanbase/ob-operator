lexer grammar PLLexer;
@members {
public boolean inRangeOperator = false;
}

MERGE
    : M E R G E
    ;

ABORT
    : A B O R T
    ;

ACCEPT
    : A C C E P T
    ;

ACCESS
    : A C C E S S
    ;

ADD
    : A D D
    ;

ALL
    : A L L
    ;

ALTER
    : A L T E R
    ;

AND
    : A N D
    ;

ANY
    : A N Y
    ;

ARRAY
    : A R R A Y
    ;

ARRAYLEN
    : A R R A Y L E N
    ;

AS
    : A S
    ;

ASC
    : A S C
    ;

ASSERT
    : A S S E R T
    ;

ASSIGN
    : A S S I G N
    ;

AT
    : A T
    ;

AUTHORIZATION
    : A U T H O R I Z A T I O N
    ;

AVG
    : A V G
    ;

BASE_TABLE
    : B A S E '_' T A B L E
    ;

BEGIN_KEY
    : B E G I N
    ;

BETWEEN
    : B E T W E E N
    ;

BINARY_INTEGER
    : B I N A R Y '_' I N T E G E R
    ;

BODY
    : B O D Y
    ;

BOOLEAN
    : B O O L E A N
    ;

BY
    : B Y
    ;

BINARY_DOUBLE
    : B I N A R Y '_' D O U B L E
    ;

BINARY_FLOAT
    : B I N A R Y '_' F L O A T
    ;

SIMPLE_DOUBLE
    : S I M P L E '_' D O U B L E
    ;

SIMPLE_FLOAT
    : S I M P L E '_' F L O A T
    ;

CASE
    : C A S E
    ;

CHARACTER
    : C H A R
    | C H A R A C T E R
    ;

CHAR_BASE
    : C H A R '_' B A S E
    ;

CHECK
    : C H E C K
    ;

CLOSE
    : C L O S E
    ;

CLUSTER
    : C L U S T E R
    ;

CLUSTERS
    : C L U S T E R S
    ;

COLAUTH
    : C O L A U T H
    ;

COLUMNS
    : C O L U M N S
    ;

COMMIT
    : C O M M I T
    ;

COMPRESS
    : C O M P R E S S
    ;

CONNECT
    : C O N N E C T
    ;

CONSTANT
    : C O N S T A N T
    ;

COUNT
    : C O U N T
    ;

CRASH
    : C R A S H
    ;

CREATE
    : C R E A T E
    ;

CURRENT
    : C U R R E N T
    ;

CURRVAL
    : C U R R V A L
    ;

CURSOR
    : C U R S O R
    ;

DATABASE
    : D A T A B A S E
    ;

DATA_BASE
    : D A T A '_' B A S E
    ;

DATE
    : D A T E
    ;

DBA
    : D B A
    ;

DEBUGOFF
    : D E B U G O F F
    ;

DEBUGON
    : D E B U G O N
    ;

NUMBER
    : D E C I M A L
    | N U M B E R
    ;

DECLARE
    : D E C L A R E
    ;

DEFAULT
    : D E F A U L T
    ;

DEFINITION
    : D E F I N I T I O N
    ;

DELAY
    : D E L A Y
    ;

DELETE
    : D E L E T E
    ;

DELTA
    : D E L T A
    ;

DESC
    : D E S C
    ;

DIGITS
    : D I G I T S
    ;

DISPOSE
    : D I S P O S E
    ;

DISTINCT
    : D I S T I N C T
    ;

DO
    : D O
    ;

DROP
    : D R O P
    ;

ELSE
    : E L S E
    ;

ELSIF
    : E L S I F
    ;

END_KEY
    : E N D
    ;

ENTRY
    : E N T R Y
    ;

EXCEPTION
    : E X C E P T I O N
    ;

EXCEPTIONS
    : E X C E P T I O N S
    ;

EXCEPTION_INIT
    : E X C E P T I O N '_' I N I T
    ;

EXISTS
    : E X I S T S
    ;

EXIT
    : E X I T
    ;

FALSE
    : F A L S E
    ;

FETCH
    : F E T C H
    ;

FLOAT
    : F L O A T
    ;

FORM
    : F O R M
    ;

FROM
    : F R O M
    ;

FUNCTION
    : F U N C T I O N
    ;

GENERIC
    : G E N E R I C
    ;

GOTO
    : G O T O
    ;

GRANT
    : G R A N T
    ;

GROUP
    : G R O U P
    ;

HAVING
    : H A V I N G
    ;

IDENTIFIED
    : I D E N T I F I E D
    ;

IF
    : I F
    ;

IN
    : I N
    ;

INDEX
    : I N D E X
    ;

INDEXES
    : I N D E X E S
    ;

INDICATOR
    : I N D I C A T O R
    ;

INSERT
    : I N S E R T
    ;

INTEGER
    : I N T E G E R
    | I N T
    ;

INTERSECT
    : I N T E R S E C T
    ;

INTERVAL
    : I N T E R V A L
    ;

INTO
    : I N T O
    ;

IS
    : I S
    ;

LEVEL
    : L E V E L
    ;

LIKE
    : L I K E
    ;

LIMITED
    : L I M I T E D
    ;

LOOP
    : L O O P
    ;

MAX
    : M A X
    ;

MIN
    : M I N
    ;

MINUS
    : M I N U S
    ;

MLSLABEL
    : M L S L A B E L
    ;

MOD
    : M O D
    ;

MODE
    : M O D E
    ;

NATURAL
    : N A T U R A L
    ;

NCHAR
    : N C H A R
    ;

NVARCHAR
    : N V A R C H A R
    ;

NVARCHAR2
    : N V A R C H A R '2'
    ;

NEW
    : N E W
    ;

NEXTVAL
    : N E X T V A L
    ;

NOCOMPRESS
    : N O C O M P R E S S
    ;

NO
    : N O
    ;

NOT
    : N O T
    ;

NULLX
    : N U L L
    ;

NUMBER_BASE
    : N U M B E R '_' B A S E
    ;

OF
    : O F
    ;

ON
    : O N
    ;

OPEN
    : O P E N
    ;

OPTION
    : O P T I O N
    ;

OR
    : O R
    ;

ORDER
    : O R D E R
    ;

OTHERS
    : O T H E R S
    ;

OUT
    : O U T
    ;

PACKAGE_P
    : P A C K A G E
    ;

PARTITION
    : P A R T I T I O N
    ;

PCTFREE
    : P C T F R E E
    ;

POSITIVE
    : P O S I T I V E
    ;

PRAGMA
    : P R A G M A
    ;

PRIOR
    : P R I O R
    ;

PRIVATE
    : P R I V A T E
    ;

PROCEDURE
    : P R O C E D U R E
    ;

PUBLIC
    : P U B L I C
    ;

RAISE
    : R A I S E
    ;

REAL
    : R E A L
    ;

RECORD
    : R E C O R D
    ;

RELEASE
    : R E L E A S E
    ;

REMR
    : R E M R
    ;

RENAME
    : R E N A M E
    ;

RESOURCE
    : R E S O U R C E
    ;

RETURN
    : R E T U R N
    ;

REVERSE
    : R E V E R S E
    ;

REVOKE
    : R E V O K E
    ;

ROLLBACK
    : R O L L B A C K
    ;

ROWID
    : R O W I D
    ;

ROWLABEL
    : R O W L A B E L
    ;

ROWNUM
    : R O W N U M
    ;

ROWTYPE
    : R O W T Y P E
    ;

RUN
    : R U N
    ;

SAVEPOINT
    : S A V E P O I N T
    ;

SCHEMA
    : S C H E M A
    ;

SQL_KEYWORD
    : S E L E C T
    ;

SEPARATE
    : S E P A R A T E
    ;

SET
    : S E T
    ;

SIZE
    : S I Z E
    ;

SMALLINT
    : S M A L L I N T
    ;

SPACE
    : S P A C E
    ;

SQL
    : S Q L
    ;

SQLCODE
    : S Q L C O D E
    ;

SQLERRM
    : S Q L E R R M
    ;

START
    : S T A R T
    ;

STATEMENT
    : S T A T E M E N T
    ;

STDDEV
    : S T D D E V
    ;

SUBTYPE
    : S U B T Y P E
    ;

SUM
    : S U M
    ;

TABAUTH
    : T A B A U T H
    ;

TABLE
    : T A B L E
    ;

TABLES
    : T A B L E S
    ;

TASK
    : T A S K
    ;

TERMINATE
    : T E R M I N A T E
    ;

THEN
    : T H E N
    ;

TO
    : T O
    ;

TRIGGER
    : T R I G G E R
    ;

TRUE
    : T R U E
    ;

TYPE
    : T Y P E
    ;

UNION
    : U N I O N
    ;

UNIQUE
    : U N I Q U E
    ;

UPDATE
    : U P D A T E
    ;

USE
    : U S E
    ;

USING_NLS_COMP
    : U S I N G '_' N L S '_' C O M P
    ;

VALUES
    : V A L U E S
    ;

VARCHAR2
    : V A R C H A R
    | V A R C H A R '2'
    ;

VARIANCE
    : V A R I A N C E
    ;

VIEW
    : V I E W
    ;

VIEWS
    : V I E W S
    ;

WHEN
    : W H E N
    ;

WHERE
    : W H E R E
    ;

WHILE
    : W H I L E
    ;

WITH
    : W I T H
    ;

WORK
    : W O R K
    ;

XOR
    : X O R
    ;

BINARY
    : B I N A R Y
    ;

BOOL
    : B O O L
    ;

CLOB
    : C L O B
    ;

BLOB
    : B L O B
    ;

CONSTRUCTOR
    : C O N S T R U C T O R
    ;

CONTINUE
    : C O N T I N U E
    ;

DEBUG
    : D E B U G
    ;

EXTERNAL
    : E X T E R N A L
    ;

FINAL
    : F I N A L
    ;

INSTANTIABLE
    : I N S T A N T I A B L E
    ;

LANGUAGE
    : L A N G U A G E
    ;

MAP
    : M A P
    ;

MEMBER
    : M E M B E R
    ;

NATURALN
    : N A T U R A L N
    ;

NOCOPY
    : N O C O P Y
    ;

NUMERIC
    : N U M E R I C
    ;

OVERRIDING
    : O V E R R I D I N G
    ;

PLS_INTEGER
    : P L S '_' I N T E G E R
    ;

POSITIVEN
    : P O S I T I V E N
    ;

RAW
    : R A W
    ;

REUSE
    : R E U S E
    ;

SELF
    : S E L F
    ;

SIGNTYPE
    : S I G N T Y P E
    ;

SIMPLE_INTEGER
    : S I M P L E '_' I N T E G E R
    ;

STATIC
    : S T A T I C
    ;

TIMESTAMP
    : T I M E S T A M P
    ;

LABEL_LEFT
    : '<<'
    ;

LABEL_RIGHT
    : '>>'
    ;

ASSIGN_OPERATOR
    : ':='
    ;

RANGE_OPERATOR
    : '..' {inRangeOperator=false;} ->mode(DEFAULT_MODE)
    ;

PARAM_ASSIGN_OPERATOR
    : '=>'
    ;

INTNUM
    : [0-9]+
    ;

DECIMAL_VAL
    : ([0-9]+ E [-+]?[0-9]+ F | [0-9]+'.'[0-9]* E [-+]?[0-9]+ F | '.'[0-9]+ E [-+]?[0-9]+ F ) {!inRangeOperator}?
    | ([0-9]+ E [-+]?[0-9]+ D | [0-9]+'.'[0-9]* E [-+]?[0-9]+ D | '.'[0-9]+ E [-+]?[0-9]+ D ) {!inRangeOperator}?
    | ([0-9]+ E [-+]?[0-9]+ | [0-9]+'.'[0-9]* E [-+]?[0-9]+ | '.'[0-9]+ E [-+]?[0-9]+) {!inRangeOperator}?
    | ([0-9]+'.'[0-9]* F | [0-9]+ F | '.'[0-9]+ F ) {!inRangeOperator}?
    | ([0-9]+'.'[0-9]* D | [0-9]+ D | '.'[0-9]+ D ) {!inRangeOperator}?
    | ([0-9]+'.'[0-9]* | '.'[0-9]+) {!inRangeOperator}?
    ;

JAVA
    : J A V A
    ;

MONTH
    : M O N T H
    ;

AFTER
    : A F T E R
    ;

SETTINGS
    : S E T T I N G S
    ;

YEAR
    : Y E A R
    ;

EACH
    : E A C H
    ;

PARALLEL_ENABLE
    : P A R A L L E L '_' E N A B L E
    ;

DECLARATION
    : D E C L A R A T I O N
    ;

VARCHAR
    : V A R C H A R
    ;

SERIALLY_REUSABLE
    : S E R I A L L Y '_' R E U S A B L E
    ;

CALL
    : C A L L
    ;

VARIABLE
    : V A R I A B L E
    ;

INSTEAD
    : I N S T E A D
    ;

RELIES_ON
    : R E L I E S '_' O N
    ;

LONG
    : L O N G
    ;

COLLECT
    : C O L L E C T
    ;

UNDER
    : U N D E R
    ;

REF
    : R E F
    ;

RightBracket
    : R I G H T B R A C K E T
    ;

IMMEDIATE
    : I M M E D I A T E
    ;

EDITIONABLE
    : E D I T I O N A B L E
    ;

REPLACE
    : R E P L A C E
    ;

VARYING
    : V A R Y I N G
    ;

DISABLE
    : D I S A B L E
    ;

NONEDITIONABLE
    : N O N E D I T I O N A B L E
    ;

FOR
    : F O R {inRangeOperator=true;}
    ;

NAME
    : N A M E
    ;

USING
    : U S I N G
    ;

YES
    : Y E S
    ;

TIME
    : T I M E
    ;

VALIDATE
    : V A L I D A T E
    ;

TRUST
    : T R U S T
    ;

AUTHID
    : A U T H I D
    ;

BULK
    : B U L K
    ;

DEFINER
    : D E F I N E R
    ;

LeftBracket
    : L E F T B R A C K E T
    ;

BYTE
    : B Y T E
    ;

LOCAL
    : L O C A L
    ;

RNPS
    : R N P S
    ;

HASH
    : H A S H
    ;

WNPS
    : W N P S
    ;

FORCE
    : F O R C E
    ;

COLLATION
    : C O L L A T I O N
    ;

COMPOUND
    : C O M P O U N D
    ;

CHAR
    : C H A R
    ;

SPECIFICATION
    : S P E C I F I C A T I O N
    ;

ACCESSIBLE
    : A C C E S S I B L E
    ;

SAVE
    : S A V E
    ;

COMPILE
    : C O M P I L E
    ;

COLLATE
    : C O L L A T E
    ;

SELECT
    : S E L E C T
    ;

EXECUTE
    : E X E C U T E
    ;

SQLDATA
    : S Q L D A T A
    ;

PIPELINED
    : P I P E L I N E D
    ;

DAY
    : D A Y
    ;

CURRENT_USER
    : C U R R E N T '_' U S E R
    ;

ZONE
    : Z O N E
    ;

DECIMAL
    : D E C I M A L
    ;

VALUE
    : V A L U E
    ;

WNDS
    : W N D S
    ;

AUTONOMOUS_TRANSACTION
    : A U T O N O M O U S '_' T R A N S A C T I O N
    ;

UDF
    : U D F
    ;

INLINE
    : I N L I N E
    ;

MINUTE
    : M I N U T E
    ;

RESULT_CACHE
    : R E S U L T '_' C A C H E
    ;

ENABLE
    : E N A B L E
    ;

OID
    : O I D
    ;

OBJECT
    : O B J E C T
    ;

RESTRICT_REFERENCES
    : R E S T R I C T '_' R E F E R E N C E S
    ;

ROW
    : R O W
    ;

RANGE
    : R A N G E {inRangeOperator=true;}
    ;

ORADATA
    : O R A D A T A
    ;

HOUR
    : H O U R
    ;

FORALL
    : F O R A L L {inRangeOperator=true;}
    ;

LIMIT
    : L I M I T
    ;

INT
    : I N T
    ;

VARRAY
    : V A R R A Y
    ;

CUSTOMDATUM
    : C U S T O M D A T U M
    ;

DETERMINISTIC
    : D E T E R M I N I S T I C
    ;

RNDS
    : R N D S
    ;

BEFORE
    : B E F O R E
    ;

CHARSET
    : C H A R S E T
    ;

SECOND
    : S E C O N D
    ;

RESULT
    : R E S U L T
    ;

DATE_VALUE
    : D A T E ([ \t\n\r\f]+|('--'(~[\n\r])*))?'\''(~['])*'\''
    | T I M E ([ \t\n\r\f]+|('--'(~[\n\r])*))?'\''(~['])*'\''
    | T I M E S T A M P ([ \t\n\r\f]+|('--'(~[\n\r])*))?'\''(~['])*'\''
    | D A T E ([ \t\n\r\f]+|('--'(~[\n\r])*))?'"'(~["])*'"'
    | T I M E ([ \t\n\r\f]+|('--'(~[\n\r])*))?'"'(~["])*'"'
    | T I M E S T A M P ([ \t\n\r\f]+|('--'(~[\n\r])*))?'"'(~["])*'"'
    ;

DELIMITER
    : ';'
    ;

Equal
    : '='
    ;

IDENT
    : ':'(([A-Za-z]|~[\u0000-\u007F\uD800-\uDBFF])([A-Za-z0-9$_#]|~[\u0000-\u007F\uD800-\uDBFF])*)
    | (([A-Za-z]|~[\u0000-\u007F\uD800-\uDBFF])([A-Za-z0-9$_#]|~[\u0000-\u007F\uD800-\uDBFF])*)
    | '`' ~[`]* '`'
    | '"' (~["]|('""'))* '"'

    ;

Or
    : [|]
    ;

Minus
    : [-]
    ;

Star
    : [*]
    ;

Div
    : [/]
    ;

Not
    : [!]
    ;

Caret
    : [^]
    ;

Colon
    : [:]
    ;

Mod
    : [%]
    ;

Dot
    : [.]
    ;

RightParen
    : [)]
    ;

LeftParen
    : [(]
    ;

Comma
    : [,]
    ;

Plus
    : [+]
    ;

And
    : [&]
    ;

Tilde
    : [~]
    ;

STRING
    : '\'' (~[']|('\'\'')|('\\\''))* '\''
    ;

In_c_comment
    : '/*' .*? '*/'      -> channel(1)
    ;

ANTLR_SKIP
    : '--'[ \t]* .*? '\n'   -> channel(1)
    ;

Blank
    : [ \t\r\n] -> channel(1)    ;

SQL_TOKEN_OR_UNKNOWN
    : (.) -> channel(1)    ;


fragment A : [aA];
fragment B : [bB];
fragment C : [cC];
fragment D : [dD];
fragment E : [eE];
fragment F : [fF];
fragment G : [gG];
fragment H : [hH];
fragment I : [iI];
fragment J : [jJ];
fragment K : [kK];
fragment L : [lL];
fragment M : [mM];
fragment N : [nN];
fragment O : [oO];
fragment P : [pP];
fragment Q : [qQ];
fragment R : [rR];
fragment S : [sS];
fragment T : [tT];
fragment U : [uU];
fragment V : [vV];
fragment W : [wW];
fragment X : [xX];
fragment Y : [yY];
fragment Z : [zZ];
