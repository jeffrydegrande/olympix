; Name: Empty Market Special Case
; Description: Function contains special handling for empty markets, a critical vulnerability
; Severity: Critical
; References: zkLend hack (Feb 2025)

; Find functions with empty market special case
(if_expression
  condition: (binary_expression
    left: (identifier) @supply_var
    operator: "=="
    right: (numeric_literal) @zero)
  (#match? @supply_var "^(total_supply|totalSupply|supply|ztoken_supply)$")
  (#eq? @zero "0"))
