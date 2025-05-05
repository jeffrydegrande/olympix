; Name: Deposit Without Activation Check
; Description: Deposit function can be called without checking if market is properly activated
; Severity: High
; References: zkLend hack (Feb 2025)

; Find deposit function without activation check
(function_item
  (function
    name: (identifier) @func_name
    (#match? @func_name "^(deposit|provide_liquidity|add_liquidity|mint)$"))
  body: (block) @func_body
  (#not-match? @func_body "is_active|active|isActive|marketActive"))
