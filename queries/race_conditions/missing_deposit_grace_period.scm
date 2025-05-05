; Name: Missing Deposit Grace Period
; Description: Deposit function lacks protection against front-running attacks
; Severity: Medium
; References: Front-running protection pattern

; Find deposit functions without any time-based protection
(function_item
  (function
    name: (identifier) @func_name
    (#match? @func_name "^(deposit|provide_liquidity|add_liquidity|mint)$"))
  body: (block) @func_body
  (#not-match? @func_body "timestamp|grace|period|timelock|delay"))
