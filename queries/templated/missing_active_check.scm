; Name: Missing Market Activation Check
; Description: Functions that should check market activation status but don't
; Concepts: active

(function_item
  (function
    name: (identifier) @func_name
    (#match? @func_name "^(deposit|withdraw|flashLoan|borrow)$"))
  body: (block) @func_body
  (#not-match? @func_body "${active}"))