; Name: Missing Reentrancy Guard
; Description: External functions that should check reentrancy guard but don't
; Concepts: locked

(function_item
  (function
    name: (identifier) @func_name
    (#match? @func_name "^(deposit|withdraw|flashLoan|borrow)$"))
  body: (block) @func_body
  (#not-match? @func_body "${locked}"))