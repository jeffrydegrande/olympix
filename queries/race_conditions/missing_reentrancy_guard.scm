; Name: Missing Reentrancy Protection
; Description: State-changing function lacks reentrancy protection
; Severity: Medium
; References: Common smart contract vulnerability

; Find state-changing functions without reentrancy guards
(function_item
  (function
    name: (identifier) @func_name
    (#match? @func_name "^(deposit|withdraw|flash_loan|activate_market)$"))
  body: (block) @func_body
  (#not-match? @func_body "reentrancy|guard|mutex|lock"))
